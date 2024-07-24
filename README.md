# GoFuture

GoFuture is a lightweight, flexible Go library for managing asynchronous operations with support for futures, retries, and concurrent task execution.

## Features

- Asynchronous task execution
- Future-based result handling
- Configurable retry mechanism
- Concurrent execution with `WaitAll`
- Context-aware operations

## Installation

To install GoFuture, use `go get`:

```bash
go get github.com/benebobaa/gofuture
```

## Quick Start

Here's a simple example to get you started:

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/benebobaa/gofuture"
)

func main() {
    ctx := context.Background()
    
    future := gofuture.Async(ctx, func(ctx context.Context) (string, error) {
        time.Sleep(time.Second)
        return "Hello, GoFuture!", nil
    })

    result, err := future.Get(ctx)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    } else {
        fmt.Printf("Result: %s\n", result)
    }
}
```

## Usage

### Simple Async Task

```go
future := gofuture.Async(ctx, func(ctx context.Context) (string, error) {
    return "Async task completed", nil
})
result, err := future.Get(ctx)
```

### Async with Retry

```go
retryConfig := gofuture.DefaultRetryConfig()
future := gofuture.AsyncWithRetry(ctx, func(ctx context.Context) (string, error) {
    // Your potentially failing operation here
    return "Operation successful", nil
}, retryConfig)
result, err := future.Get(ctx)
```

### Wait for Multiple Tasks

```go
future1 := gofuture.Async(ctx, task1)
future2 := gofuture.Async(ctx, task2)
future3 := gofuture.Async(ctx, task3)

err := gofuture.WaitAll(ctx, future1, future2, future3)
```

## Configuration

### Retry Configuration

You can customize the retry behavior:

```go
customConfig := gofuture.RetryConfig{
    MaxRetries:  5,
    InitialWait: 50 * time.Millisecond,
    MaxWait:     2 * time.Second,
    Multiplier:  1.5,
}
```

## Best Practices

1. Always use a context to manage timeouts and cancellations.
2. Handle errors returned by `Get()` method of futures.
3. Use `WaitAll` for efficiently waiting on multiple futures.
4. Customize retry configurations based on your specific use case.

## Contributing

Contributions to GoFuture are welcome! Please feel free to submit a Pull Request.

## License

GoFuture is released under the MIT License. See the LICENSE file for details.
