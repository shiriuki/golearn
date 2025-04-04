// This application was developed while learning how to program in Go.
// This is a basic echo server that allows a maximum number of concurrent
// clients. These clients can connect using netcat to the port the server
// is listing. Then the clients will receive back whatever text they sent.
// If they send STOP they will be disconnected, if they send EXIT, they will
// terminate the server and will terminate other clients as well. If a client
// want to connect and the number maximum number of clients has been reach,
// their connection will be refused.
package main

import (
	"bufio"
	"context"
	"echoServer/logger"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"golang.org/x/sync/semaphore"
)

// Defines the maximum number of concurrent clients.
const MAX_CONCURRENT_CLIENTS = 2

var sem = semaphore.NewWeighted(int64(MAX_CONCURRENT_CLIENTS))
var wg sync.WaitGroup
var serverFinished = false

func handleNewConnRequest(conn net.Conn, ctx context.Context, cancelFunc context.CancelFunc) {
	defer wg.Done()
	defer conn.Close()

	canHandle := sem.TryAcquire(1)
	if !canHandle {
		reportCantHandle(conn)
		return
	}
	defer sem.Release(1)
	reportHandlerServing(conn)

	input := make(chan string, 1)
	go getInput(conn, input)

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-input:
			closeClient, closeServer, err := handleMsgFromClient(conn, msg)
			if err != nil && !serverFinished {
				logger.Warn(err.Error())
			} else if closeClient {
				reportClosingHandler(conn)
				return
			} else if closeServer {
				serverFinished = true
				cancelFunc()
				return
			}
			go getInput(conn, input)
		}
	}
}

func getInput(conn net.Conn, input chan string) {
	msg, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil && !serverFinished {
		logger.Warn("couldn't read message from client " + conn.RemoteAddr().Network() + ". Error: " + err.Error())
	}
	input <- msg
}

func handleMsgFromClient(conn net.Conn, msg string) (closeClient bool, closeServer bool, error error) {
	if strings.TrimSpace(string(msg)) == "STOP" {
		return true, false, nil
	}
	if strings.TrimSpace(string(msg)) == "EXIT" {
		return false, true, nil
	}

	conn.Write([]byte("echo: " + msg))
	return false, false, nil
}

func reportClosingHandler(conn net.Conn) {
	logger.Info("Closing handler " + conn.RemoteAddr().String())
}

func reportHandlerServing(conn net.Conn) {
	logger.Info("Handler serving " + conn.RemoteAddr().String())
}

func reportCantHandle(conn net.Conn) {
	conn.Write([]byte("Too many concurrent clients\n"))
}

func getPortNumber() (port string, error error) {
	arguments := os.Args
	if len(arguments) == 1 {
		return "", errors.New("please provide port number")
	}
	return ":" + arguments[1], nil
}

func getListener(port string) (net.Listener, error) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func processConnRequests(listener net.Listener, ctx context.Context, cancelFunc context.CancelFunc) {
	for {
		conn, err := listener.Accept()
		if !serverFinished {
			if err != nil {
				logger.Warn("Can't open client connection: " + err.Error())
			} else {
				wg.Add(1)
				go handleNewConnRequest(conn, ctx, cancelFunc)
			}
		}
	}
}

func main() {
	port, err := getPortNumber()
	if err != nil {
		logger.ErrorAndExit(err.Error())
	}

	listener, err := getListener(port)
	if err != nil {
		logger.ErrorAndExit(err.Error())
	}

	fmt.Println("Server listening port", port)
	defer listener.Close()

	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	go processConnRequests(listener, ctx, cancelFunc)

	select {
	case <-ctx.Done():
		wg.Wait()
		fmt.Println("Bye")
	}
}
