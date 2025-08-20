package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	prometheus.MustRegister(collectors.NewBuildInfoCollector())
}

func StartMetricsServer(listenAddr string) error {
	mux := http.NewServeMux()
	metricsServer := &http.Server{
		Addr:              listenAddr,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       30 * time.Second,
		Handler:           mux,
	}

	mux.HandleFunc("/health", healthHandle)
	mux.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))

	return metricsServer.ListenAndServe()
}

func healthHandle(w http.ResponseWriter, r *http.Request) {}
