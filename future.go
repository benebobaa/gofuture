package gofuture

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"
)

var (
	ErrMaxRetriesReached = errors.New("maximum retries reached")
)

// Future represents an asynchronous operation
type Future[T any] struct {
	result T
	err    error
	done   chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

// RetryConfig holds configuration for retry logic
type RetryConfig struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:  3,
		InitialWait: 100 * time.Millisecond,
		MaxWait:     10 * time.Second,
		Multiplier:  2,
	}
}

// Async executes a function asynchronously and returns a Future
func Async[T any](ctx context.Context, f func(context.Context) (T, error)) *Future[T] {
	childCtx, cancel := context.WithCancel(ctx)
	future := &Future[T]{
		done:   make(chan struct{}),
		ctx:    childCtx,
		cancel: cancel,
	}

	go future.execute(f)
	return future
}

// AsyncWithRetry executes a function asynchronously with retry logic
func AsyncWithRetry[T any](ctx context.Context, f func(context.Context) (T, error), config RetryConfig) *Future[T] {
	childCtx, cancel := context.WithCancel(ctx)
	future := &Future[T]{
		done:   make(chan struct{}),
		ctx:    childCtx,
		cancel: cancel,
	}

	go future.executeWithRetry(f, config)
	return future
}

// execute runs the given function and stores the result
func (f *Future[T]) execute(fn func(context.Context) (T, error)) {
	defer close(f.done)
	defer f.cancel()
	f.result, f.err = fn(f.ctx)
}

// executeWithRetry runs the given function with retry logic
func (f *Future[T]) executeWithRetry(fn func(context.Context) (T, error), config RetryConfig) {
	defer close(f.done)
	defer f.cancel()

	var lastErr error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		select {
		case <-f.ctx.Done():
			f.err = f.ctx.Err()
			return
		default:
			if attempt > 0 {
				if !f.wait(attempt, config) {
					return
				}
			}
			f.result, lastErr = fn(f.ctx)
			if lastErr == nil {
				return
			}
		}
	}
	f.err = errors.Join(ErrMaxRetriesReached, lastErr)
}

// wait implements the waiting logic between retries
func (f *Future[T]) wait(attempt int, config RetryConfig) bool {
	waitTime := calculateWaitTime(attempt, config)
	timer := time.NewTimer(waitTime)
	defer timer.Stop()

	select {
	case <-f.ctx.Done():
		f.err = f.ctx.Err()
		return false
	case <-timer.C:
		return true
	}
}

// calculateWaitTime calculates the wait time for the next retry attempt
func calculateWaitTime(attempt int, config RetryConfig) time.Duration {
	wait := float64(config.InitialWait) * math.Pow(config.Multiplier, float64(attempt-1))
	return time.Duration(math.Min(float64(config.MaxWait), wait))
}

// Get retrieves the result of the Future, blocking until it's available or the context is done
func (f *Future[T]) Get(ctx context.Context) (T, error) {
	select {
	case <-f.done:
		return f.result, f.err
	case <-ctx.Done():
		f.cancel() // Cancel the operation if the context is done
		var zero T
		return zero, ctx.Err()
	}
}

// WaitAll waits for all Futures to complete or for the context to be done
func WaitAll[T any](ctx context.Context, futures ...*Future[T]) error {
	errs := make([]error, len(futures))
	var wg sync.WaitGroup

	for i, f := range futures {
		wg.Add(1)
		go func(i int, future *Future[T]) {
			defer wg.Done()
			_, errs[i] = future.Get(ctx)
		}(i, f)
	}

	waitCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		return errors.Join(errs...)
	case <-ctx.Done():
		return ctx.Err()
	}
}
