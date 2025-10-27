package main

import (
	"bytes"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})
}

func TestTelnetClient_ConnectionTimeout(t *testing.T) {
	timeout := 100 * time.Millisecond
	client := NewTelnetClient("127.0.0.1:65000", timeout, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})
	err := client.Connect()
	require.Error(t, err)
}

func TestTelnetClient_ReceiveAfterServerClose(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer func() { _ = l.Close() }()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		in := &bytes.Buffer{}
		out := &bytes.Buffer{}
		timeout := time.Second

		client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
		require.NoError(t, client.Connect())

		err = client.Receive()
		require.True(t, err == nil || errors.Is(err, io.EOF), "expected EOF or nil, got %v", err)
		require.NoError(t, client.Close())
	}()

	go func() {
		defer wg.Done()
		conn, err := l.Accept()
		require.NoError(t, err)
		_ = conn.Close()
	}()

	wg.Wait()
}

func TestTelnetClient_SendAfterClose(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer func() { _ = l.Close() }()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		in := &bytes.Buffer{}
		out := &bytes.Buffer{}
		timeout := time.Second

		client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
		require.NoError(t, client.Connect())
		require.NoError(t, client.Close())

		in.WriteString("test\n")
		err := client.Send()
		require.Error(t, err, "send on closed connection should fail")
	}()

	go func() {
		defer wg.Done()
		conn, err := l.Accept()
		require.NoError(t, err)
		defer func() { _ = conn.Close() }()
		time.Sleep(200 * time.Millisecond)
	}()

	wg.Wait()
}

func TestTelnetClient_MultipleSendReceive(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer func() { _ = l.Close() }()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		in := &bytes.Buffer{}
		out := &bytes.Buffer{}

		timeout := 2 * time.Second
		client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
		require.NoError(t, client.Connect())
		defer func() { _ = client.Close() }()

		for i := 0; i < 3; i++ {
			in.WriteString("ping\n")
			err := client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
		}
		require.Equal(t, "pong\npong\npong\n", out.String())
	}()

	go func() {
		defer wg.Done()
		conn, err := l.Accept()
		require.NoError(t, err)
		defer func() { _ = conn.Close() }()

		buf := make([]byte, 16)
		for i := 0; i < 3; i++ {
			n, err := conn.Read(buf)
			require.NoError(t, err)
			require.Equal(t, "ping\n", string(buf[:n]))
			_, err = conn.Write([]byte("pong\n"))
			require.NoError(t, err)
		}
	}()

	wg.Wait()
}
