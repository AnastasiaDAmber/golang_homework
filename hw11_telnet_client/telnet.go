package main

import (
	"errors"
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type telnetClient struct {
	address string
	timeout time.Duration
	conn    net.Conn
	in      io.ReadCloser
	out     io.Writer
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

func (t *telnetClient) Connect() error {
	conn, err := net.DialTimeout("tcp", t.address, t.timeout)
	if err != nil {
		return err
	}
	t.conn = conn
	return nil
}

func (t *telnetClient) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}

func (t *telnetClient) Send() error {
	buf := make([]byte, 1024)
	n, err := t.in.Read(buf)
	if n > 0 {
		_, writeErr := t.conn.Write(buf[:n])
		if writeErr != nil {
			return writeErr
		}
	}
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

func (t *telnetClient) Receive() error {
	buf := make([]byte, 1024)
	n, err := t.conn.Read(buf)
	if n > 0 {
		if _, writeErr := t.out.Write(buf[:n]); writeErr != nil {
			return writeErr
		}
	}
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}
