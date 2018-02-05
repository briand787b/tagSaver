package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
)

const (
	port = ":3000"
)

var (
	ctx context.Context
)

func main() {
	l, err := net.Listen("tcp", "localhost"+port)
	if err != nil {
		log.Fatal(err)
	}

	// Context to cancel functions
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// listen for termination signals, e.g. SIGINT
	go func() {
		sigquit := make(chan os.Signal, 1)
		signal.Notify(sigquit, os.Kill, os.Interrupt)

		sig := <-sigquit
		fmt.Printf("caught sig: %+v\n", sig)
		fmt.Printf("Gracefully shuting down service...\n")

		// cancel all running queries and
		// shut down service with status code
		// success
		cancel()
		os.Exit(0)
	}()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConn(c)
	}
}

func handleConn(c net.Conn) {
	defer c.Close()

	buf := newTagBuffer()
	scnr := bufio.NewScanner(c)
	scnr.Split(bufio.ScanWords)
	for scnr.Scan() {
		// w := scnr.Text()
		buf.Add(scnr.Text())
		// fmt.Println(w)
	}

	if err := scnr.Err(); err != nil {
		fmt.Println("error scanning: ", err)
	}

	// flush buffer
	buf.Save()
}
