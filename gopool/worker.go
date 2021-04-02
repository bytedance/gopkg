package gopool

import (
	"sync"
	"sync/atomic"
)

var workerPool sync.Pool

var globalCnt uint32

func init() {
	workerPool.New = newWorker
}

type worker struct {
	pool *pool
}

func newWorker() interface{} {
	return &worker{}
}

func (w *worker) run() {
	l := len(w.pool.taskLists)
	go func() {
		for {
			var t *task
			for i := 0; i < l; i++ {
				idx := int(atomic.AddUint32(&globalCnt, 1)) % l
				w.pool.taskLists[idx].Lock()
				if w.pool.taskLists[idx].taskHead != nil {
					t = w.pool.taskLists[idx].taskHead
					w.pool.taskLists[idx].taskHead = w.pool.taskLists[idx].taskHead.next
					atomic.AddInt32(&w.pool.taskCount, -1)
					w.pool.taskLists[idx].Unlock()
					break
				} else {
					if i == l-1 {
						// 最后一次循环，如果没有任务要做了，就释放资源，退出
						w.close()
						w.pool.taskLists[idx].Unlock()
						w.Recycle()
						return
					}
					w.pool.taskLists[idx].Unlock()
				}
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						w.pool.panicHandler(t.ctx, r)
					}
				}()
				t.f()
			}()
			t.Recycle()
		}
	}()
}
func (w *worker) close() {
	w.pool.decWorkerCount()
}

func (w *worker) zero() {
	w.pool = nil
}

func (w *worker) Recycle() {
	w.zero()
	workerPool.Put(w)
}
