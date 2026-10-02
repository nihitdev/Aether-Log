package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nihitdev/Aether-Log/internal/ingest"
	"github.com/nihitdev/Aether-Log/internal/protocol"
	"github.com/nihitdev/Aether-Log/pkg/metrics"
	"github.com/nihitdev/Aether-Log/pkg/storage"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
func run() error {
	addr := flag.String("listen", ":8080", "TCP address")
	httpAddr := flag.String("metrics", ":8081", "HTTP address")
	output := flag.String("log", "aether.log", "output path")
	connections := flag.Int("max-connections", 256, "maximum concurrent connections")
	payload := flag.Uint("max-payload", 1024*1024, "maximum payload bytes")
	capacity := flag.Int("queue-capacity", 1024, "queued records")
	batch := flag.Int("batch-size", 128, "records per batch")
	interval := flag.Duration("flush-interval", time.Second, "batch flush interval")
	idle := flag.Duration("read-timeout", time.Minute, "idle/frame read timeout")
	rotate := flag.Int64("rotate-bytes", 64*1024*1024, "rotation threshold; 0 disables")
	retain := flag.Int("retain", 5, "retained rotations")
	flag.Parse()
	if *connections <= 0 || *payload == 0 || uint64(*payload) > math.MaxUint32 || *idle <= 0 {
		return fmt.Errorf("invalid connection, payload or timeout limits")
	}
	store, err := storage.Open(*output, *rotate, *retain)
	if err != nil {
		return err
	}
	defer store.Close()
	m := &metrics.Counters{Started: time.Now()}
	writer, err := ingest.New(store, *capacity, *batch, *interval, m)
	if err != nil {
		return err
	}
	defer writer.Close()
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	hl, err := net.Listen("tcp", *httpAddr)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	server := &http.Server{Handler: m.Handler(writer.Depth, *capacity), ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second}
	httpDone := make(chan struct{})
	go func() {
		defer close(httpDone)
		if err := server.Serve(hl); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP: %v", err)
			stop()
		}
	}()
	var mu sync.Mutex
	active := map[net.Conn]struct{}{}
	var wg sync.WaitGroup
	m.Ready.Store(true)
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		m.Ready.Store(false)
		listener.Close()
		mu.Lock()
		for c := range active {
			c.Close()
		}
		mu.Unlock()
	}()
	log.Printf("Hub TCP %s HTTP %s output %s", listener.Addr(), hl.Addr(), *output)
	for {
		c, err := listener.Accept()
		if err != nil {
			if ctx.Err() == nil {
				log.Printf("accept: %v", err)
				stop()
			}
			break
		}
		mu.Lock()
		if ctx.Err() != nil || !m.Ready.Load() || len(active) >= *connections {
			mu.Unlock()
			m.Rejected.Add(1)
			c.Close()
			continue
		}
		active[c] = struct{}{}
		mu.Unlock()
		m.Accepted.Add(1)
		m.Active.Add(1)
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { c.Close(); mu.Lock(); delete(active, c); mu.Unlock(); m.Active.Add(-1) }()
			for {
				c.SetReadDeadline(time.Now().Add(*idle))
				f, err := protocol.ReadFrame(c, uint32(*payload))
				if err != nil {
					if ctx.Err() == nil && !errors.Is(err, io.EOF) {
						log.Printf("read %s: %v", c.RemoteAddr(), err)
					}
					return
				}
				if !m.Ready.Load() {
					return
				}
				if f.Type != protocol.TypeData {
					log.Printf("unexpected client control frame from %s", c.RemoteAddr())
					return
				}
				m.Frames.Add(1)
				m.Bytes.Add(uint64(len(f.Payload)))
				if err := writer.Submit(ctx, f.Payload); err != nil {
					return
				}
			}
		}()
	}
	stop()
	<-shutdownDone
	wg.Wait()
	err = writer.Close()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(shutdownCtx)
	server.Close()
	<-httpDone
	if err != nil {
		return err
	}
	return store.Close()
}
