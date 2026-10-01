package metrics

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

type Counters struct {
	Accepted, Rejected, Frames, Bytes, Queued, Batched, Persisted, Dropped, Batches, Failures atomic.Uint64
	Active                                                                                    atomic.Int64
	Ready                                                                                     atomic.Bool
	Started                                                                                   time.Time
}

func (m *Counters) Handler(depth func() int, capacity int) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok\n")) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if !m.Ready.Load() {
			http.Error(w, "not ready", 503)
			return
		}
		w.Write([]byte("ready\n"))
	})
	stats := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"connections_accepted": m.Accepted.Load(), "connections_rejected": m.Rejected.Load(), "connections_active": m.Active.Load(),
			"frames_received": m.Frames.Load(), "bytes_received": m.Bytes.Load(), "records_queued": m.Queued.Load(), "records_batched": m.Batched.Load(),
			"records_persisted": m.Persisted.Load(), "records_dropped": m.Dropped.Load(), "batches_written": m.Batches.Load(), "write_failures": m.Failures.Load(),
			"queue_depth": depth(), "queue_capacity": capacity, "uptime_seconds": time.Since(m.Started).Seconds(), "ready": m.Ready.Load()})
	}
	mux.HandleFunc("/metrics", stats)
	mux.HandleFunc("/stats", stats)
	return mux
}
