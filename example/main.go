package main

import (
	"context"
	"errors"
	"fmt"
	async "github.com/benebobaa/gofuture"
	"math/rand"
	"time"
)

// Simulates a database query that might fail
func queryDB(ctx context.Context, id int) (string, error) {
	// Simulate work and potential failure
	delay := time.Duration(rand.Intn(1000)) * time.Millisecond
	select {
	case <-time.After(delay):
		if rand.Float32() < 0.3 { // 30% chance of failure
			return "", fmt.Errorf("DB query failed for id %d", id)
		}
		return fmt.Sprintf("Result for ID %d", id), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// Always fails, used to demonstrate max retries
func alwaysFailTask(ctx context.Context) (string, error) {
	return "", errors.New("this task always fails")
}

// Simulates sending an email
func sendEmailSuccess(ctx context.Context, to string) error {
	time.Sleep(500 * time.Millisecond)
	fmt.Printf("Email sent to %s\n", to)
	return nil
}

func main() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Simple async task
	fmt.Println("1. Simple async task:")
	simpleFuture := async.Async(ctx, func(ctx context.Context) (string, error) {
		return queryDB(ctx, 1)
	})
	result, err := simpleFuture.Get(ctx)
	fmt.Printf("Result: %v, Error: %v\n\n", result, err)

	// 2. Async task with default retry
	fmt.Println("2. Async task with default retry:")
	retryFuture := async.AsyncWithRetry(ctx, func(ctx context.Context) (string, error) {
		return queryDB(ctx, 2)
	}, async.DefaultRetryConfig())
	result, err = retryFuture.Get(ctx)
	fmt.Printf("Result: %v, Error: %v\n\n", result, err)

	// 3. Async task with custom retry config
	fmt.Println("3. Async task with custom retry config:")
	customConfig := async.RetryConfig{
		MaxRetries:  5,
		InitialWait: 50 * time.Millisecond,
		MaxWait:     2 * time.Second,
		Multiplier:  1.5,
	}
	customRetryFuture := async.AsyncWithRetry(ctx, func(ctx context.Context) (string, error) {
		return queryDB(ctx, 3)
	}, customConfig)
	result, err = customRetryFuture.Get(ctx)
	fmt.Printf("Result: %v, Error: %v\n\n", result, err)

	// 4. Async task that always fails (max retries)
	fmt.Println("4. Async task that always fails (max retries):")
	alwaysFailFuture := async.AsyncWithRetry(ctx, alwaysFailTask, async.DefaultRetryConfig())
	result, err = alwaysFailFuture.Get(ctx)
	fmt.Printf("Result: %v, Error: %v\n\n", result, err)

	// 5. Multiple async tasks with WaitAll
	fmt.Println("5. Multiple async tasks with WaitAll:")
	future1 := async.Async(ctx, func(ctx context.Context) (string, error) { return queryDB(ctx, 4) })
	future2 := async.Async(ctx, func(ctx context.Context) (string, error) { return queryDB(ctx, 5) })
	future3 := async.Async(ctx, func(ctx context.Context) (string, error) { return queryDB(ctx, 6) })

	err = async.WaitAll(ctx, future1, future2, future3)
	fmt.Printf("WaitAll Error: %v\n", err)

	result1, err1 := future1.Get(ctx)
	result2, err2 := future2.Get(ctx)
	result3, err3 := future3.Get(ctx)
	fmt.Printf("Result 1: %v, Error: %v\n", result1, err1)
	fmt.Printf("Result 2: %v, Error: %v\n", result2, err2)
	fmt.Printf("Result 3: %v, Error: %v\n\n", result3, err3)

	// 6. Fire-and-forget task (email sending)
	fmt.Println("6. Fire-and-forget task (email sending):")
	_ = async.Async(ctx, func(ctx context.Context) (interface{}, error) {
		return nil, sendEmailSuccess(ctx, "user@example.com")
	})
	fmt.Println("Email sending initiated")

	fmt.Println("Email sending initiated with retry")

	// 7. Send email with retry
	fmt.Println("\n7. Send email with retry:")
	emailConfig := async.RetryConfig{
		MaxRetries:  5,
		InitialWait: 100 * time.Millisecond,
		MaxWait:     1 * time.Second,
		Multiplier:  1.5,
	}
	emailFuture := async.AsyncWithRetry(ctx, func(ctx context.Context) (string, error) {
		err := sendEmail(ctx, "user@example.com")
		if err != nil {
			return "", err
		}
		return "Email sent successfully", nil
	}, emailConfig)

	emailResult, emailErr := emailFuture.Get(ctx)
	fmt.Printf("Email Result: %v, Error: %v\n", emailResult, emailErr)

	// 8. Send Kafka message with default retry
	fmt.Println("\n8. Send Kafka message with default retry:")
	kafkaFuture := async.AsyncWithRetry(ctx, func(ctx context.Context) (string, error) {
		err := sendKafkaMessage(ctx, "example-topic", "Hello, Kafka!")
		if err != nil {
			return "", err
		}
		return "Kafka message sent successfully", nil
	}, async.DefaultRetryConfig())

	kafkaResult, kafkaErr := kafkaFuture.Get(ctx)
	fmt.Printf("Kafka Result: %v, Error: %v\n", kafkaResult, kafkaErr)

	// 9. Multiple operations with different retry configs
	fmt.Println("\n9. Multiple operations with different retry configs:")
	dbFuture := async.AsyncWithRetry(ctx, func(ctx context.Context) (string, error) {
		return queryDB(ctx, 7)
	}, async.DefaultRetryConfig())

	emailFuture2 := async.AsyncWithRetry(ctx, func(ctx context.Context) (string, error) {
		err := sendEmail(ctx, "another@example.com")
		if err != nil {
			return "", err
		}
		return "Second email sent successfully", nil
	}, emailConfig)

	kafkaFuture2 := async.AsyncWithRetry(ctx, func(ctx context.Context) (string, error) {
		err := sendKafkaMessage(ctx, "another-topic", "Hello again, Kafka!")
		if err != nil {
			return "", err
		}
		return "Second Kafka message sent successfully", nil
	}, async.DefaultRetryConfig())

	// Wait for all operations to complete
	err = async.WaitAll(ctx, dbFuture, emailFuture2, kafkaFuture2)
	if err != nil {
		fmt.Printf("WaitAll Error: %v\n", err)
	}

	dbResult, dbErr := dbFuture.Get(ctx)
	emailResult2, emailErr2 := emailFuture2.Get(ctx)
	kafkaResult2, kafkaErr2 := kafkaFuture2.Get(ctx)

	fmt.Printf("DB Result: %v, Error: %v\n", dbResult, dbErr)
	fmt.Printf("Email Result: %v, Error: %v\n", emailResult2, emailErr2)
	fmt.Printf("Kafka Result: %v, Error: %v\n", kafkaResult2, kafkaErr2)

	// Wait a bit to allow the email to be "sent"
	time.Sleep(1 * time.Second)
}

// Simulates sending an email with potential failures
func sendEmail(ctx context.Context, to string) error {
	time.Sleep(500 * time.Millisecond)
	if rand.Float32() < 0.4 { // 40% chance of failure
		return fmt.Errorf("failed to send email to %s", to)
	}
	fmt.Printf("Email sent to %s\n", to)
	return nil
}

// Simulates sending a message to Kafka
func sendKafkaMessage(ctx context.Context, topic, message string) error {
	time.Sleep(300 * time.Millisecond)
	if rand.Float32() < 0.3 { // 30% chance of failure
		return fmt.Errorf("failed to send message to Kafka topic %s", topic)
	}
	fmt.Printf("Message sent to Kafka topic %s: %s\n", topic, message)
	return nil
}
