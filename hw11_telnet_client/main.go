package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "usage: go-telnet [--timeout=10s] host port")
		os.Exit(1)
	}

	host, port := flag.Arg(0), flag.Arg(1)
	address := net.JoinHostPort(host, port)

	client := NewTelnetClient(address, *timeout, os.Stdin, os.Stdout)

	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "...Connection error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "...Connected to %s\n", address)

	done := make(chan struct{})

	// чтение из сокета → stdout
	go func() {
		if err := client.Receive(); err != nil && !errors.Is(err, io.EOF) {
			fmt.Fprintf(os.Stderr, "...Receive error: %v\n", err)
		}
		fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
		close(done)
	}()

	// ввод пользователя → сокет
	go func() {
		if err := client.Send(); err != nil {
			fmt.Fprintf(os.Stderr, "...Send error: %v\n", err)
		}
		fmt.Fprintln(os.Stderr, "...EOF")
		close(done)
	}()

	<-done
	_ = client.Close()
}
