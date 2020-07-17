package sync

import "github.com/prometheus/client_golang/prometheus"

const (
	prometheusNamespace = "crsync"
	prometheusSubsystem = "sync"
)

var (
	tagsSyncedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "tags_synced_total",
			Help:      "Number of synchronized tags per repository",
		},
		[]string{
			"source_registry",
			"destination_registry",
			"repository",
		},
	)
)

func init() {
	prometheus.MustRegister(tagsSyncedTotal)
}

func updateTagsSyncedTotal(srcRegistry, dstRegistry, repository string) {
	tagsSyncedTotal.WithLabelValues(
		srcRegistry, dstRegistry, repository,
	).Inc()
}
