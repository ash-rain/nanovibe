package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"vibecodepc/server/httputil"
	"vibecodepc/server/services"
)

// RegisterMetricsRoutes registers all metrics API routes.
func RegisterMetricsRoutes(r chi.Router) {
	r.Get("/api/metrics/stream", getMetricsStream)
}

func getMetricsStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		httputil.WriteError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Send an initial reading immediately.
	sendMetrics(w, flusher)

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			sendMetrics(w, flusher)
		}
	}
}

func sendMetrics(w http.ResponseWriter, flusher http.Flusher) {
	metrics, err := services.Read()
	if err != nil {
		return
	}
	data, err := json.Marshal(metrics)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: metrics\ndata: %s\n\n", data)
	flusher.Flush()
}
