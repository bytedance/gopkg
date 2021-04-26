// Copyright 2021 ByteDance Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gopool

import (
	"context"
	"fmt"
	"sync"
)

// defaultPool 不应该被修改或者 Closed，所以保护起来
var defaultPool Pool

var poolMap sync.Map

func init() {
	defaultPool = NewPool("gopool.DefaultPool", 10000, NewConfig())
}

// Go 作为 go 关键字的替代品，拥有 panic recover 的作用
// gopool.Go(func(arg interface{}){
//     ...
// }(nil))
func Go(f func()) {
	CtxGo(context.Background(), f)
}

// 建议使用带 ctx 前缀的，打日志时可以带上 logid，方便调用链追踪
func CtxGo(ctx context.Context, f func()) {
	defaultPool.CtxGo(ctx, f)
}

// 不建议更改大小，容易造成全局其它调用者的问题
func SetCap(cap int32) {
	defaultPool.SetCap(cap)
}

// 设置默认 pool panic 情况下的 handler
func SetPanicHandler(f func(context.Context, interface{})) {
	defaultPool.SetPanicHandler(f)
}

// 把 Pool 注册到全局的 map 里面，可以通过 GetPool 获取
// 如果已经注册过会返回 error
func RegisterPool(p Pool) error {
	_, loaded := poolMap.LoadOrStore(p.Name(), p)
	if loaded {
		return fmt.Errorf("name: %s already registered", p.Name())
	}
	return nil
}

// 通过 name 获取对应的 Pool
func GetPool(name string) Pool {
	p, _ := poolMap.Load(name)
	return p.(Pool)
}
