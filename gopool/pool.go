package gopool

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
)

type Pool interface {
	// 获取到对应 pool 的 name
	Name() string
	// 更新 goroutine pool 的容量
	SetCap(cap int32)
	// 执行 f
	Go(f func())
	// 传入 ctx 和 f，panic 打日志时带上 logid
	CtxGo(ctx context.Context, f func())
	// panic 的时候调用额外的 handler
	SetPanicHandler(f func(context.Context, interface{}))
	// 获取当前正在运行的 goroutine 数量
	WorkerCount() int32
	// Close 会停止接收新的任务，等到旧的任务全部执行完成之后，所有的 worker 会自动退出
	Close()
}

var taskPool sync.Pool

func init() {
	taskPool.New = newTask
}

// ctx 主要是为了打日志的时候用，这样如果有 logid 的话调用链追踪可以查找到
type task struct {
	ctx context.Context
	f   func()
	// 指向下一个 task 的指针
	next *task
}

func (t *task) zero() {
	t.ctx = nil
	t.f = nil
	t.next = nil
}

func (t *task) Recycle() {
	t.zero()
	taskPool.Put(t)
}

func newTask() interface{} {
	return &task{}
}

type taskList struct {
	sync.Mutex
	taskHead *task
	taskTail *task
}

type pool struct {
	// pool 的名字，打 metrics 和打 log 时用到
	name string

	cnt uint32
	// pool 的容量，也就是最大的真正在工作的 goroutine 的数量
	// 为了性能考虑，可能会有小误差
	cap int32
	// 配置信息
	config *Config
	// 任务链表
	taskLists []taskList
	taskCount int32

	// 记录正在运行的 worker 数量
	workerCount int32

	// 用来标记是否关闭
	closed int32

	// worker panic 的时候会调用这个方法
	panicHandler func(context.Context, interface{})
}

// name 必须是不含有空格的，只能含有字母、数字和下划线，否则 metrics 会失败
func NewPool(name string, cap int32, config *Config) Pool {
	p := &pool{
		name:      name,
		cap:       cap,
		config:    config,
		taskLists: make([]taskList, runtime.GOMAXPROCS(0)),
	}
	return p
}

func (p *pool) Name() string {
	return p.name
}

func (p *pool) SetCap(cap int32) {
	atomic.StoreInt32(&p.cap, cap)
}

func (p *pool) Go(f func()) {
	p.CtxGo(context.Background(), f)
}

func (p *pool) CtxGo(ctx context.Context, f func()) {
	t := taskPool.Get().(*task)
	t.ctx = ctx
	t.f = f
	idx := int(atomic.AddUint32(&p.cnt, 1)) % len(p.taskLists)
	p.taskLists[idx].Lock()
	if p.taskLists[idx].taskHead == nil {
		p.taskLists[idx].taskHead = t
		p.taskLists[idx].taskTail = t
	} else {
		p.taskLists[idx].taskTail.next = t
		p.taskLists[idx].taskTail = t
	}
	p.taskLists[idx].Unlock()
	atomic.AddInt32(&p.taskCount, 1)
	// 如果 pool 已经被关闭了，就 panic
	if atomic.LoadInt32(&p.closed) == 1 {
		panic("use closed pool")
	}
	// 满足以下两个条件：
	// 1. task 数量大于阈值
	// 2. 目前的 worker 数量小于上限 p.cap
	// 或者目前没有 worker
	if (atomic.LoadInt32(&p.taskCount) >= p.config.ScaleThreshold && p.WorkerCount() < atomic.LoadInt32(&p.cap)) || p.WorkerCount() == 0 {
		p.incWorkerCount()
		w := workerPool.Get().(*worker)
		w.pool = p
		w.run()
	}
}

// 这里的 Handler 会在 panic 被 recover 之后执行
func (p *pool) SetPanicHandler(f func(context.Context, interface{})) {
	p.panicHandler = f
}

func (p *pool) WorkerCount() int32 {
	return atomic.LoadInt32(&p.workerCount)
}

// Close 会停止接收新的任务，等到旧的任务全部执行完成之后，所有的 worker 会自动退出
func (p *pool) Close() {
	atomic.StoreInt32(&p.closed, 1)
}

func (p *pool) incWorkerCount() {
	atomic.AddInt32(&p.workerCount, 1)
}

func (p *pool) decWorkerCount() {
	atomic.AddInt32(&p.workerCount, -1)
}
