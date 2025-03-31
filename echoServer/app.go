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
	"time"

	"golang.org/x/sync/semaphore"
)

const MAX_CONCURRENT_CLIENTS = 2

var sem = semaphore.NewWeighted(int64(MAX_CONCURRENT_CLIENTS))
var wg sync.WaitGroup
var serverFinished = false

func handleNewConnRequest(conn net.Conn, ctx context.Context, cancelFunc context.CancelFunc) {
	defer wg.Done()
	defer conn.Close()

	if serverFinished {
		return
	}

	canHandle := sem.TryAcquire(1)
	if !canHandle {
		reportCantHandle(conn)
		return
	}
	defer sem.Release(1)
	reportHandlerServing(conn)
	serving := true
	reading := true

	for serving {
		select {
		case <-ctx.Done():
			return
		default:
			if reading {
				reading = false
				go func() {
					closeClient, closeServer, err := handleMsgFromClient(conn)
					if err != nil && !serverFinished {
						logger.Warn(err.Error())
					} else if closeClient {
						reportClosingHandler(conn)
						serving = false
						return
					} else if closeServer {
						serverFinished = true
						cancelFunc()
						return
					}
					reading = true
				}()
			} else {
				time.Sleep(time.Millisecond * 500)
			}
		}
	}
}

func handleMsgFromClient(conn net.Conn) (closeClient bool, closeServer bool, error error) {
	msg, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, false, errors.New("Couldn't read message from client " + err.Error())
	}
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
	for !serverFinished {
		conn, err := listener.Accept()
		if err != nil {
			logger.Warn("Can't open client connection: " + err.Error())
		} else {
			wg.Add(1)
			go handleNewConnRequest(conn, ctx, cancelFunc)
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
