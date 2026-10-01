// Package client sends v1 Aether frames. Successful sends mean TCP writes,
// not durable Hub acceptance. Retrying a failed send can duplicate a record.
package client

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"gitlab.com/nihitdev/Aether-Log/internal/protocol"
)

type Config struct {
	Address                  string
	Timeout, Retry, MaxRetry time.Duration
	MaxPayload               int
	// OnRetry runs synchronously after a network failure; it must return promptly.
	OnRetry func(error, time.Duration)
}
type Client struct {
	mu     sync.Mutex
	config Config
	conn   net.Conn
}

func New(c Config) (*Client, error) {
	if c.Address == "" {
		return nil, fmt.Errorf("address required")
	}
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Second
	}
	if c.Retry == 0 {
		c.Retry = 200 * time.Millisecond
	}
	if c.MaxRetry == 0 {
		c.MaxRetry = 10 * time.Second
	}
	if c.MaxPayload == 0 {
		c.MaxPayload = 1024 * 1024
	}
	if c.Timeout < 0 || c.Retry < 0 || c.MaxRetry < c.Retry || c.MaxPayload < 0 {
		return nil, fmt.Errorf("invalid client limits")
	}
	return &Client{config: c}, nil
}

// Send blocks with bounded exponential reconnect delays until success or cancellation.
func (c *Client) Send(ctx context.Context, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(payload) > c.config.MaxPayload {
		return protocol.ErrPayloadTooLarge
	}
	delay := c.config.Retry
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var err error
		if c.conn == nil {
			d := net.Dialer{Timeout: c.config.Timeout}
			c.conn, err = d.DialContext(ctx, "tcp", c.config.Address)
		}
		if err == nil {
			deadline := time.Now().Add(c.config.Timeout)
			if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
				deadline = d
			}
			c.conn.SetWriteDeadline(deadline)
			conn := c.conn
			done := make(chan struct{})
			cancel := context.AfterFunc(ctx, func() { conn.SetWriteDeadline(time.Now()); close(done) })
			err = protocol.WriteFrame(conn, protocol.TypeData, payload)
			if !cancel() {
				<-done
			}
			if err == nil {
				return nil
			}
			c.conn.Close()
			c.conn = nil
		}
		if c.config.OnRetry != nil {
			c.config.OnRetry(err, delay)
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("send: %w (last network error: %v)", ctx.Err(), err)
		case <-timer.C:
		}
		if delay < c.config.MaxRetry {
			if delay > c.config.MaxRetry/2 {
				delay = c.config.MaxRetry
			} else {
				delay *= 2
			}
		}
	}
}
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}
