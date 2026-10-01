package client

import (
	"context"
	"net"
	"testing"
	"time"

	"gitlab.com/nihitdev/Aether-Log/internal/protocol"
)

func TestSend(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	result := make(chan protocol.Frame, 1)
	go func() {
		c, err := l.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		f, _ := protocol.ReadFrame(c, 100)
		result <- f
	}()
	c, _ := New(Config{Address: l.Addr().String()})
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := c.Send(ctx, []byte("sdk")); err != nil {
		t.Fatal(err)
	}
	select {
	case f := <-result:
		if string(f.Payload) != "sdk" {
			t.Fatal(f)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}
func TestRetryCancellation(t *testing.T) {
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := l.Addr().String()
	l.Close()
	c, _ := New(Config{Address: addr, Retry: time.Millisecond, MaxRetry: 2 * time.Millisecond})
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if c.Send(ctx, []byte("x")) == nil {
		t.Fatal("unexpected success")
	}
}

func TestReconnectAfterRefusal(t *testing.T) {
	reserved, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := reserved.Addr().String()
	reserved.Close()
	var listener net.Listener
	result := make(chan string, 1)
	c, err := New(Config{Address: addr, Retry: time.Millisecond, MaxRetry: 2 * time.Millisecond, OnRetry: func(networkErr error, delay time.Duration) {
		if listener != nil {
			return
		}
		var err error
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			defer conn.Close()
			f, err := protocol.ReadFrame(conn, 100)
			if err == nil {
				result <- string(f.Payload)
			}
		}()
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	defer func() {
		if listener != nil {
			listener.Close()
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := c.Send(ctx, []byte("reconnected")); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-result:
		if got != "reconnected" {
			t.Fatal(got)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}
