package metrics

import (
	"encoding/json"
	"net/http"
)

type HubMetrics interface {
	GetActiveWorkers() int64
	GetBytesProcessed() uint64
}

type MetricsServer struct {
	Hub HubMetrics
}

func (s *MetricsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"active_workers":  s.Hub.GetActiveWorkers(),
		"bytes_processed": s.Hub.GetBytesProcessed(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func StartMetricsServer(addr string, hub HubMetrics) error {
	server := &MetricsServer{Hub: hub}
	return http.ListenAndServe(addr, server)
}
