package failure

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	failureDetectorLabels = []string{
		"client_app_id", "server_app_id", "client_addr", "server_addr",
	}

	// failure_detector_active_clients -> the total number of connected clients
	activeClientsGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "failure_detector",
		Name:      "active_clients",
		Help:      "the total number of connected clients",
	}, failureDetectorLabels)

	// failure_detector_heartbeat_interval -> agg. heartbeat intervals from clients
	heartbeatIntervalHist = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "failure_detector",
		Name:      "heartbeat_interval",
		Help:      "agg. heartbeat intervals from clients",
		Buckets:   prometheus.ExponentialBucketsRange(32, 8192, 16),
	}, failureDetectorLabels)

	suspicionHist = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "failure_detector",
		Name:      "suspicion",
		Help:      "per-connection suspicion",
		Buckets:   prometheus.ExponentialBucketsRange(0.001, 16, 16),
	}, failureDetectorLabels)
)
