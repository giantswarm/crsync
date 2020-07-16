package sync

import "github.com/prometheus/client_golang/prometheus"

const (
	prometheusNamespace = "crsync"
	prometheusSubsystem = "sync"
)

var (
	tagsSyncedTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
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

func updateTagsSyncedTotal(srcRegistry, dstRegistry, repository string) {
	tagsSyncedTotal.WithLabelValues(
		srcRegistry, dstRegistry, repository,
	).Inc()
}
