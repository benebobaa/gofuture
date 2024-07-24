package gofuture

import (
	"context"
	"errors"
	"testing"
	"time"
)

func BenchmarkAsync(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		future := Async(ctx, func(ctx context.Context) (int, error) {
			return 42, nil
		})
		_, _ = future.Get(ctx)
	}
}

func BenchmarkAsyncWithRetry(b *testing.B) {
	ctx := context.Background()
	config := DefaultRetryConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		future := AsyncWithRetry(ctx, func(ctx context.Context) (int, error) {
			return 42, nil
		}, config)
		_, _ = future.Get(ctx)
	}
}

func BenchmarkAsyncWithRetryError(b *testing.B) {
	ctx := context.Background()
	config := DefaultRetryConfig()
	config.MaxRetries = 3
	config.InitialWait = time.Microsecond
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		future := AsyncWithRetry(ctx, func(ctx context.Context) (int, error) {
			return 0, errors.New("error")
		}, config)
		_, _ = future.Get(ctx)
	}
}

func BenchmarkWaitAll(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		futures := []*Future[int]{
			Async(ctx, func(ctx context.Context) (int, error) { return 1, nil }),
			Async(ctx, func(ctx context.Context) (int, error) { return 2, nil }),
			Async(ctx, func(ctx context.Context) (int, error) { return 3, nil }),
		}
		_ = WaitAll(ctx, futures...)
	}
}

func BenchmarkCalculateWaitTime(b *testing.B) {
	config := DefaultRetryConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateWaitTime(i%5+1, config)
	}
}

func BenchmarkFutureGet(b *testing.B) {
	ctx := context.Background()
	future := Async(ctx, func(ctx context.Context) (int, error) {
		time.Sleep(time.Millisecond)
		return 42, nil
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = future.Get(ctx)
	}
}

func BenchmarkFutureGetParallel(b *testing.B) {
	ctx := context.Background()
	future := Async(ctx, func(ctx context.Context) (int, error) {
		time.Sleep(time.Millisecond)
		return 42, nil
	})
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = future.Get(ctx)
		}
	})
}

func BenchmarkAsyncComplex(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		future := Async(ctx, func(ctx context.Context) (int, error) {
			time.Sleep(time.Microsecond)
			return fibonacci(20), nil
		})
		_, _ = future.Get(ctx)
	}
}

// Helper function for complex computation
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}
