package failure

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	failureDetectorLabels = []string{
		"client_host_id", "client_pid", "client_process_uuid", "client_region", "server_host_id",
	}

	activeClientsGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "failure_detector",
		Name:      "active_clients",
		Help:      "the total number of connected clients",
	}, failureDetectorLabels)

	heartbeatIntervalGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "failure_detector",
		Name:      "heartbeat_interval",
		Help:      "gauge of heartbeat interval recv from clients",
	}, failureDetectorLabels)

	heartbeatIntervalStDevGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "failure_detector",
		Name:      "heartbeat_interval_stdev",
		Help:      "gauge of heartbeat interval variance recv from clients",
	}, failureDetectorLabels)

	suspicionLevelGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "failure_detector",
		Name:      "suspicion_level",
		Help:      "calculated suspicion level",
	}, failureDetectorLabels)
)
