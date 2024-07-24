package main

import (
	"context"
	"fmt"
	"time"

	async "github.com/benebobaa/gofuture"
)

func main() {

	fmt.Println("Test")

	ctx := context.Background()

	start := time.Now()

	future1 := async.Async(
		ctx,
		func(ctx context.Context) (string, error) {
			return getData()
		},
	)

	result, err := future1.Get(ctx)
	// result, err := getData()

	fmt.Println("err :: ", err)

	future2 := async.Async(
		ctx,
		func(ctx context.Context) (string, error) {
			return getData2(result)
		},
	)

	result, err = future2.Get(ctx)
	fmt.Println("result :: ", result)

	end := time.Now()
	duration := end.Sub(start)

	fmt.Println("duration :: ", duration)
}

func getData() (string, error) {

	time.Sleep(time.Second)

	return "from data 1", nil
}

func getData2(msg string) (string, error) {
	time.Sleep(time.Second)

	fmt.Println("process :: ", msg)

	time.Sleep(time.Second)

	return "success", nil
}
