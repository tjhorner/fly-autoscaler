package prometheus_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/superfly/fly-autoscaler/prometheus"
)

// newHangingServer returns a server that never responds until the request is abandoned.
func newHangingServer(t *testing.T) *httptest.Server {
	t.Helper()
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	t.Cleanup(func() {
		close(release)
		server.Close()
	})
	return server
}

func TestMetricCollector_CollectMetric_RespectsContext(t *testing.T) {
	server := newHangingServer(t)

	collector, err := prometheus.NewMetricCollector("m", server.URL, "up", "")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err = collector.CollectMetric(ctx, "my-app")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got err %v, want context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("query took %s; context deadline was ignored", elapsed)
	}
}

func TestMachineMetricCollector_CollectMachineMetrics_RespectsContext(t *testing.T) {
	server := newHangingServer(t)

	collector, err := prometheus.NewMachineMetricCollector(server.URL, "up", "")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err = collector.CollectMachineMetrics(ctx, "my-app")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got err %v, want context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("query took %s; context deadline was ignored", elapsed)
	}
}
