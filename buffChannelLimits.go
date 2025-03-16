package main

import (
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup

func writer(stream chan int) {
	defer wg.Done()
	readerTimer := time.NewTimer(time.Second * 2)
	exitTimer := time.NewTimer(time.Second * 5)
	i := 0

	for {
		select {
		case stream <- i:
			fmt.Println("Wrote", i)
			i++
		case <-readerTimer.C:
			fmt.Println("Read", <-stream, "Now accepting one more value.")
		case <-exitTimer.C:
			fmt.Println("Time to go")
			return
		default:
			fmt.Println("Buffers full. Can't write more values.")
			// 	return
		}
		time.Sleep(time.Millisecond * 200)
	}
}

func main() {
	stream := make(chan int, 5)
	wg.Add(1)
	go writer(stream)
	wg.Wait()
}
