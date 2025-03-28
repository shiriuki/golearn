package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup

func simulatedTask(c chan bool, seconds int) {
	time.Sleep(time.Duration(seconds) * time.Second)
	c <- true
}

func task1(ctx context.Context) {
	defer wg.Done()
	c := make(chan bool)
	defer close(c)

	go simulatedTask(c, 2)
	select {
	case <-c:
		fmt.Println("Task 1 finished on time")
		return
	case <-ctx.Done():
		fmt.Println("Task 1 cancelled!", context.Cause(ctx))
		return
	}
}

func task2(ctx context.Context) {
	defer wg.Done()
	c := make(chan bool)
	defer close(c)

	go simulatedTask(c, 7)
	select {
	case <-c:
		fmt.Println("Task 2 finished on time")
		return
	case <-ctx.Done():
		fmt.Println("Task 2 cancelled!", context.Cause(ctx))
		return
	}
}

func main() {
	maxTaskDuration := time.Second * 5
	timeOutErr := errors.New("cancelled by timeout")

	wg.Add(1)
	ctx1 := context.Background()
	ctx1, _ = context.WithTimeoutCause(ctx1, maxTaskDuration, timeOutErr)
	go task1(ctx1)

	wg.Add(1)
	ctx2 := context.Background()
	ctx2, _ = context.WithTimeoutCause(ctx2, maxTaskDuration, timeOutErr)
	go task2(ctx2)

	wg.Wait()
	fmt.Println("App finished")
}
