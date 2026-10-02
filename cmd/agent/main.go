package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	client "github.com/nihitdev/Aether-Log/sdk/go"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
func run() error {
	addr := flag.String("hub", "127.0.0.1:8080", "Hub address")
	path := flag.String("file", "", "input file; omitted or - means stdin")
	retry := flag.Duration("retry", 200*time.Millisecond, "initial reconnect delay")
	maxRetry := flag.Duration("max-retry", 10*time.Second, "maximum reconnect delay")
	timeout := flag.Duration("connect-timeout", 5*time.Second, "connect and write timeout")
	maxLine := flag.Int("max-line", 1024*1024, "maximum record bytes")
	flag.Parse()
	if *maxLine <= 0 || *maxLine > 16*1024*1024 {
		return fmt.Errorf("max-line must be 1..16777216")
	}
	c, err := client.New(client.Config{Address: *addr, Timeout: *timeout, Retry: *retry, MaxRetry: *maxRetry, MaxPayload: *maxLine, OnRetry: func(err error, delay time.Duration) { log.Printf("send failed: %v; retry in %s", err, delay) }})
	if err != nil {
		return err
	}
	defer c.Close()
	source := os.Stdin
	if *path != "" && *path != "-" {
		source, err = os.Open(*path)
		if err != nil {
			return err
		}
	}
	defer source.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			source.Close()
		case <-done:
		}
	}()
	defer close(done)
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, min(64*1024, *maxLine+2)), *maxLine+2)
	for scanner.Scan() {
		if len(scanner.Bytes()) > *maxLine {
			return fmt.Errorf("record exceeds max-line")
		}
		if err := c.Send(ctx, scanner.Bytes()); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
	}
	if ctx.Err() != nil {
		return nil
	}
	return scanner.Err()
}
