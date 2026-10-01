package metrics

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEndpoints(t *testing.T) {
	m := &Counters{Started: time.Now()}
	h := m.Handler(func() int { return 3 }, 10)
	for _, path := range []string{"/healthz", "/readyz", "/metrics", "/stats"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		want := 200
		if path == "/readyz" {
			want = 503
		}
		if w.Code != want {
			t.Fatal(path, w.Code)
		}
		if path == "/metrics" {
			var v map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil || v["queue_depth"] != float64(3) {
				t.Fatal(v, err)
			}
		}
	}
	m.Ready.Store(true)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/readyz", nil))
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
}
