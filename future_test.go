package gofuture

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestAsync(t *testing.T) {
	t.Run("Successful execution", func(t *testing.T) {
		ctx := context.Background()
		future := Async(ctx, func(ctx context.Context) (int, error) {
			return 42, nil
		})

		result, err := future.Get(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result != 42 {
			t.Errorf("Expected result 42, got %d", result)
		}
	})

	t.Run("Error propagation", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("test error")
		future := Async(ctx, func(ctx context.Context) (int, error) {
			return 0, expectedErr
		})

		_, err := future.Get(ctx)
		if err != expectedErr {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		future := Async(ctx, func(ctx context.Context) (int, error) {
			<-ctx.Done()
			return 0, ctx.Err()
		})

		cancel()
		_, err := future.Get(context.Background())
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled error, got %v", err)
		}
	})
}

func TestAsyncWithRetry(t *testing.T) {
	t.Run("Successful execution after retries", func(t *testing.T) {
		ctx := context.Background()
		attempts := 0
		config := DefaultRetryConfig()
		config.MaxRetries = 2

		future := AsyncWithRetry(ctx, func(ctx context.Context) (int, error) {
			attempts++
			if attempts <= 2 {
				return 0, errors.New("temporary error")
			}
			return 42, nil
		}, config)

		result, err := future.Get(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result != 42 {
			t.Errorf("Expected result 42, got %d", result)
		}
		if attempts != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("Max retries reached", func(t *testing.T) {
		ctx := context.Background()
		config := DefaultRetryConfig()
		config.MaxRetries = 2

		future := AsyncWithRetry(ctx, func(ctx context.Context) (int, error) {
			return 0, errors.New("persistent error")
		}, config)

		_, err := future.Get(ctx)
		if !errors.Is(err, ErrMaxRetriesReached) {
			t.Errorf("Expected ErrMaxRetriesReached, got %v", err)
		}
	})

	t.Run("Context cancellation during retry", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		config := DefaultRetryConfig()
		config.MaxRetries = 5
		config.InitialWait = 50 * time.Millisecond

		attempts := 0
		future := AsyncWithRetry(ctx, func(ctx context.Context) (int, error) {
			attempts++
			if attempts == 2 {
				cancel()
			}
			return 0, errors.New("temporary error")
		}, config)

		_, err := future.Get(context.Background())
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled error, got %v", err)
		}
		if attempts != 2 {
			t.Errorf("Expected 2 attempts, got %d", attempts)
		}
	})
}

func TestWaitAll(t *testing.T) {
	t.Run("All futures successful", func(t *testing.T) {
		ctx := context.Background()
		futures := []*Future[int]{
			Async(ctx, func(ctx context.Context) (int, error) { return 1, nil }),
			Async(ctx, func(ctx context.Context) (int, error) { return 2, nil }),
			Async(ctx, func(ctx context.Context) (int, error) { return 3, nil }),
		}

		err := WaitAll(ctx, futures...)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Some futures with errors", func(t *testing.T) {
		ctx := context.Background()
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		futures := []*Future[int]{
			Async(ctx, func(ctx context.Context) (int, error) { return 1, nil }),
			Async(ctx, func(ctx context.Context) (int, error) { return 0, err1 }),
			Async(ctx, func(ctx context.Context) (int, error) { return 0, err2 }),
		}

		err := WaitAll(ctx, futures...)
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if !errors.Is(err, err1) || !errors.Is(err, err2) {
			t.Errorf("Expected errors %v and %v, got %v", err1, err2, err)
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		futures := []*Future[int]{
			Async(ctx, func(ctx context.Context) (int, error) {
				time.Sleep(50 * time.Millisecond)
				return 1, nil
			}),
			Async(ctx, func(ctx context.Context) (int, error) {
				time.Sleep(100 * time.Millisecond)
				return 2, nil
			}),
		}

		go func() {
			time.Sleep(25 * time.Millisecond)
			cancel()
		}()

		err := WaitAll(ctx, futures...)
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled error, got %v", err)
		}
	})
}

func TestRaceConditions(t *testing.T) {
	t.Run("Concurrent access to Future", func(t *testing.T) {
		ctx := context.Background()
		future := Async(ctx, func(ctx context.Context) (int, error) {
			time.Sleep(50 * time.Millisecond)
			return 42, nil
		})

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = future.Get(ctx)
			}()
		}

		wg.Wait()
	})
}

func TestDeadlock(t *testing.T) {
	t.Run("No deadlock on early return", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		future := Async(ctx, func(ctx context.Context) (int, error) {
			time.Sleep(1 * time.Second)
			return 42, nil
		})

		_, err := future.Get(ctx)
		if err != context.DeadlineExceeded {
			t.Errorf("Expected context.DeadlineExceeded error, got %v", err)
		}
	})
}

func TestCalculateWaitTime(t *testing.T) {
	config := DefaultRetryConfig()

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{"First attempt", 1, 100 * time.Millisecond},
		{"Second attempt", 2, 200 * time.Millisecond},
		{"Third attempt", 3, 400 * time.Millisecond},
		{"Max wait time", 10, 10 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateWaitTime(tt.attempt, config)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries to be 3, got %d", config.MaxRetries)
	}
	if config.InitialWait != 100*time.Millisecond {
		t.Errorf("Expected InitialWait to be 100ms, got %v", config.InitialWait)
	}
	if config.MaxWait != 10*time.Second {
		t.Errorf("Expected MaxWait to be 10s, got %v", config.MaxWait)
	}
	if config.Multiplier != 2 {
		t.Errorf("Expected Multiplier to be 2, got %f", config.Multiplier)
	}
}
