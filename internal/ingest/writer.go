// Package ingest implements a bounded queue and a single persistence worker.
package ingest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/nihitdev/Aether-Log/pkg/metrics"
)

type Writer struct {
	queue   chan []byte
	done    chan struct{}
	mu      sync.RWMutex
	closed  bool
	err     error
	metrics *metrics.Counters
}

func New(out io.Writer, capacity, batchSize int, interval time.Duration, m *metrics.Counters) (*Writer, error) {
	if capacity <= 0 || batchSize <= 0 || interval <= 0 {
		return nil, fmt.Errorf("queue, batch size and flush interval must be positive")
	}
	w := &Writer{queue: make(chan []byte, capacity), done: make(chan struct{}), metrics: m}
	go w.run(out, batchSize, interval)
	return w, nil
}

// Submit transfers ownership of data on success. It blocks when the queue is full.
func (w *Writer) Submit(ctx context.Context, data []byte) error {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if w.closed {
		return fmt.Errorf("ingestion closed")
	}
	select {
	case w.queue <- data:
		w.metrics.Queued.Add(1)
		return nil
	case <-ctx.Done():
		w.metrics.Dropped.Add(1)
		return ctx.Err()
	}
}
func (w *Writer) Depth() int { return len(w.queue) }

// Close must follow cancellation/waiting of producers; it drains accepted records.
func (w *Writer) Close() error {
	w.mu.Lock()
	if !w.closed {
		w.closed = true
		close(w.queue)
	}
	w.mu.Unlock()
	<-w.done
	return w.err
}
func (w *Writer) run(out io.Writer, size int, interval time.Duration) {
	defer close(w.done)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	var batch bytes.Buffer
	count := 0
	flush := func() {
		if count == 0 {
			return
		}
		w.metrics.Batched.Add(uint64(count))
		if w.err == nil {
			n, err := out.Write(batch.Bytes())
			if err == nil && n != batch.Len() {
				err = io.ErrShortWrite
			}
			if err != nil {
				w.err = fmt.Errorf("persist batch: %w", err)
				w.metrics.Failures.Add(1)
				w.metrics.Ready.Store(false)
			} else {
				w.metrics.Persisted.Add(uint64(count))
				w.metrics.Batches.Add(1)
			}
		}
		if w.err != nil {
			w.metrics.Dropped.Add(uint64(count))
		}
		batch.Reset()
		count = 0
	}
	for {
		select {
		case data, ok := <-w.queue:
			if !ok {
				flush()
				return
			}
			batch.Write(data)
			batch.WriteByte('\n')
			count++
			if count >= size {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}
