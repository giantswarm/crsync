package sync

import "github.com/prometheus/client_golang/prometheus"

const (
	prometheusNamespace = "crsync"
	prometheusSubsystem = "sync"
)

var (
	errorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "errors_total",
			Help:      "Number of errors occurred while synchronising repositories",
		},
	)

	tagsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "tags_total",
			Help:      "Number of tags in repository",
		},
		[]string{
			"registry",
			"repository",
		},
	)
)

func init() {
	prometheus.MustRegister(errorsTotal)
	prometheus.MustRegister(tagsTotal)
}
