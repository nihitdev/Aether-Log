package ingest

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/nihitdev/Aether-Log/pkg/metrics"
)

func TestDrain(t *testing.T) {
	var out bytes.Buffer
	m := &metrics.Counters{}
	w, _ := New(&out, 4, 3, time.Hour, m)
	for _, s := range []string{"a", "b", "c", "d"} {
		if err := w.Submit(context.Background(), []byte(s)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if out.String() != "a\nb\nc\nd\n" || m.Persisted.Load() != 4 || m.Batches.Load() != 2 {
		t.Fatalf("output %q metrics %d", out.String(), m.Persisted.Load())
	}
	if w.Submit(context.Background(), nil) == nil {
		t.Fatal("accepted after close")
	}
}

type gate struct{ entered, release chan struct{} }

func (g *gate) Write(p []byte) (int, error) { g.entered <- struct{}{}; <-g.release; return len(p), nil }
func TestBoundedQueue(t *testing.T) {
	g := &gate{make(chan struct{}, 1), make(chan struct{})}
	m := &metrics.Counters{}
	w, _ := New(g, 1, 1, time.Hour, m)
	w.Submit(context.Background(), []byte("a"))
	<-g.entered
	w.Submit(context.Background(), []byte("b"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !errors.Is(w.Submit(ctx, []byte("c")), context.Canceled) {
		t.Fatal("full queue did not cancel")
	}
	if w.Depth() != 1 || m.Dropped.Load() != 1 {
		t.Fatal("queue metrics")
	}
	close(g.release)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

type broken struct{}

func (broken) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }
func TestWriteFailure(t *testing.T) {
	m := &metrics.Counters{}
	m.Ready.Store(true)
	w, _ := New(broken{}, 2, 2, time.Hour, m)
	w.Submit(context.Background(), []byte("a"))
	w.Submit(context.Background(), []byte("b"))
	if w.Close() == nil || m.Ready.Load() || m.Dropped.Load() != 2 || m.Failures.Load() != 1 {
		t.Fatal("failure not accounted")
	}
}
func TestIntervalFlush(t *testing.T) {
	g := &gate{make(chan struct{}, 1), make(chan struct{})}
	w, _ := New(g, 2, 10, time.Millisecond, &metrics.Counters{})
	w.Submit(context.Background(), []byte("a"))
	select {
	case <-g.entered:
	case <-time.After(time.Second):
		t.Fatal("no interval flush")
	}
	close(g.release)
	w.Close()
}
func BenchmarkBatching(b *testing.B) {
	w, _ := New(io.Discard, 1024, 128, time.Hour, &metrics.Counters{})
	ctx := context.Background()
	p := []byte("record")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Submit(ctx, p)
	}
	w.Close()
}
