package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	Requests     *prometheus.CounterVec
	Errors       *prometheus.CounterVec
	ResponseTime *prometheus.HistogramVec
)

func init() {
	var labelNames = []string{"service", "url"}

	Requests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "Requests",
	}, labelNames)

	Errors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "Errors",
	}, labelNames)

	ResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "ResponseTime",
	}, labelNames)
}
