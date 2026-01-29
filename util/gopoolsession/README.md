# gopoolsession

## Introduction

`gopoolsession` wraps `gopool` to support localsession automatic binding and unbinding. It keeps compatibility with the original `gopool` API, minimizing business code changes.

This package is designed to seamlessly integrate with [CloudWeGo localsession](https://github.com/cloudwego/localsession), allowing session information to be automatically propagated to goroutines in the pool without manual handling.

## Features

- **Automatic Session Management**: Automatically binds and unbinds localsession for tasks
- **API Compatible**: Same interface as `gopool`, no need to modify existing code
- **High Performance**: Built on top of the efficient `gopool` implementation
- **Panic Safe**: Maintains the original panic recovery mechanism

## QuickStart

Just replace your `gopool.Go(func(){...})` with `gopoolsession.Go(func(){...})`. The session will be automatically propagated.

old:
```go
gopool.Go(func() {
    // do your job without session
})```

new:
```go
gopoolsession.Go(func() {
    // do your job with automatic session binding
})```

### Example with localsession

```go
import (
    "github.com/cloudwego/localsession"
    "github.com/bytedance/gopkg/util/gopoolsession"
)

// Set session in the current goroutine
localsession.SetSession(someSession)

gopoolsession.Go(func() {
    // Session is automatically available here
    sess := localsession.CurSession()
    // use the session
})
```

## API

The API is identical to `gopool`:

- `gopoolsession.Go(func())` - Submit an asynchronous task
- `gopoolsession.CtxGo(ctx context.Context, func())` - Submit a task with context

All other `gopool` functions (like `SetCap`, `SetPanicHandler`, etc.) can still be used through the original `gopool` package.