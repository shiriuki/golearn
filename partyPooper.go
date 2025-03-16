package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var wg sync.WaitGroup

func partyPooper(end chan<- bool) {
	defer wg.Done()
	secs := rand.Intn(5) + 1
	fmt.Printf("You are getting %d seconds of fun\n", secs)
	time.Sleep(time.Second * time.Duration(secs))
	end <- true
}

func writer(stream chan<- int, end <-chan bool) {
	defer wg.Done()
	time.Sleep(time.Millisecond * 200)
	for {
		select {
		case stream <- rand.Intn(1000):
		case <-end:
			fmt.Println()
			fmt.Println("Done writing")
			close(stream)
			return
		}
		time.Sleep(time.Millisecond * 50)
	}
}

func reader(stream <-chan int) {
	defer wg.Done()
	for {
		if v, ok := <-stream; !ok {
			fmt.Println("Done reading")
			return
		} else {
			fmt.Printf("%d ", v)
		}
	}
}

func main() {
	stream := make(chan int)
	end := make(chan bool)
	wg.Add(1)
	go partyPooper(end)
	wg.Add(1)
	go writer(stream, end)
	wg.Add(1)
	go reader(stream)
	wg.Wait()
}
