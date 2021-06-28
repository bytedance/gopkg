# hcontext

## 功能

对 context.Context 的包装函数

## usage

```golang
//移除超时信息，保留value信息
ctx = hcontext.WithNoDeadline(ctx)

//移除取消信息，保留超时和value信息
ctx = hcontext.WithNoCancel(ctx)
```
