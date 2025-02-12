package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var MetricApiCallCount = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "circulation_api_call_count",
		Help: "Circulation backend api call count",
	},
)

var MetricApiUsedTime = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "circulation_api_used_time",
		Help: "Circulation backend api used time",
	},
)

var MetricApiErrorCount = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "circulation_api_error_count",
		Help: "Circulation backend api error count",
	},
)

var CurrentlyIndexed = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "circulation_currently_indexed_block",
		Help: "Which block currently indexed.",
	},
)
