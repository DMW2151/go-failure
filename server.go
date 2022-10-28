package failure

import (
	"context"
	"math"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	failproto "github.com/dmw2151/go-failure/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	"github.com/prometheus/client_golang/prometheus"
)

// DetectorOptions - config options to control failure detection estimation window, etc...
type DetectorOptions struct {
	WindowSize             int
	ManagementInterval     time.Duration
	PurgeAllSuspectedProcs bool
}

// Server - maintains collection of detectors + metadata. Implements `UnimplementedPhiAccrualServer`
type Server struct {
	failproto.UnimplementedPhiAccrualServer
	registeredProcs map[string]*PhiAccrualDetector
	dOpts           *DetectorOptions
	logger          *log.Logger
	hostID          string
}

// NewFailureDetectorServer - creates new Server
func NewFailureDetectorServer(dOpts *DetectorOptions, logger *log.Logger) (Server, error) {

	// use default logrus text logger if logger DNE
	if logger == nil {
		logger = log.New()
		logger.SetFormatter(&log.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})
	}

	hostname, _ := os.Hostname()
	return Server{
		registeredProcs: make(map[string]*PhiAccrualDetector),
		logger:          logger,
		dOpts:           dOpts,
		hostID:          hostname,
	}, nil
}

// Heartbeat - handler for recv'ing a heartbeat from a client
func (s Server) Heartbeat(ctx context.Context, hb *failproto.Beat) (*emptypb.Empty, error) {

	var (
		detector    *PhiAccrualDetector
		arrivalTime time.Time = time.Now()
	)

	// lookup a client process by UUID && check if exists/DNE
	detector, ok := s.registeredProcs[hb.Uuid]

	// if client process DNE -> create a new entry in the registry of tracked clients
	if !ok {
		s.logger.WithFields(log.Fields{
			"client_process_uuid": hb.Uuid,
		}).Info("registering client process")

		// create && populate detector start values
		detector = NewPhiAccrualDetector(s.dOpts.WindowSize)
		detector.lastHeartbeat = arrivalTime
		detector.Tags = hb.Tags
		s.registeredProcs[hb.Uuid] = detector

		// increment `activeClientsGauge` -> +1 to the total number of connected clients
		activeClientsGauge.With(prometheus.Labels{
			"client_host_id":      detector.Tags["client_host_id"],
			"client_pid":          detector.Tags["client_pid"],
			"client_region":       detector.Tags["client_region"],
			"client_process_uuid": hb.Uuid,
			"server_host_id":      s.hostID,
		}).Inc()
		return &emptypb.Empty{}, nil
	}

	var labels = prometheus.Labels{
		"client_host_id":      detector.Tags["client_host_id"],
		"client_pid":          detector.Tags["client_pid"],
		"client_region":       detector.Tags["client_region"],
		"client_process_uuid": hb.Uuid,
		"server_host_id":      s.hostID,
	}

	// Update suspicions on event
	phi := detector.Suspicion(detector.lastHeartbeat, arrivalTime)
	suspicionLevelGauge.With(labels).Set(phi)

	// if client process (by UUID) already exists -> calculate dela since last event
	// add to interval, update stats, update last arrival time, etc.
	hbDelta := float64(arrivalTime.Sub(detector.lastHeartbeat) / time.Microsecond)
	detector.AddValue(hbDelta)
	detector.lastHeartbeat = arrivalTime

	return &emptypb.Empty{}, nil
}

// CalculateProcessSuspicion - calc phi, updates metrics && returns processes that have been marked suspicious
// for deletion w. +inf suspicion (can/should be configured down)
func (s *Server) CalculateProcessSuspicion(calcTimestamp time.Time, publishMetrics bool) ([]string, error) {

	var (
		deadProcs       []string
		rAvg, rVar, phi float64
	)

	// calc stats, age of last heartbeat & suspicion level, when suspicion is Inf (normally collapses around
	// +10 S.D) then mark as failed & mark for deletion
	for pUuid, pFd := range s.registeredProcs {

		phi = pFd.Suspicion(pFd.lastHeartbeat, calcTimestamp)
		if phi == math.Inf(1) {
			deadProcs = append(deadProcs, pUuid)
		}

		// optionally push metrics to prometheus endpoint e.g. :1234/metrics, see conf options
		if publishMetrics {
			rAvg, rVar = pFd.Parameters()
			var labels = prometheus.Labels{
				"client_host_id":      pFd.Tags["client_host_id"],
				"client_pid":          pFd.Tags["client_pid"],
				"client_region":       pFd.Tags["client_region"],
				"client_process_uuid": pUuid,
				"server_host_id":      s.hostID,
			}
			heartbeatIntervalGauge.With(labels).Set(rAvg)
			heartbeatIntervalStDevGauge.With(labels).Set(math.Pow(rVar, 0.5))
			suspicionLevelGauge.With(labels).Set(phi)
		}
	}
	return deadProcs, nil
}

// PurgeDeadProcs - remove dead procs from server equal or over suspicion threshsold (+inf)
func (s *Server) PurgeDeadProcs(deadProcs []string, publishMetrics bool) error {

	for _, pUuid := range deadProcs {

		d, _ := s.registeredProcs[pUuid]
		delete(s.registeredProcs, pUuid)

		var labels = prometheus.Labels{
			"client_host_id":      d.Tags["client_host_id"],
			"client_pid":          d.Tags["client_pid"],
			"client_region":       d.Tags["client_region"],
			"client_process_uuid": pUuid,
			"server_host_id":      s.hostID,
		}

		// optionally push metrics to prometheus endpoint e.g. :1234/metrics, see conf options
		if publishMetrics {
			for _, met := range []*prometheus.GaugeVec{activeClientsGauge, heartbeatIntervalGauge, heartbeatIntervalStDevGauge, suspicionLevelGauge} {
				if n := met.DeletePartialMatch(labels); n == 0 {
					s.logger.WithFields(log.Fields{
						"labels": labels,
					}).Warn("failed to remove metrics of (suspected) crashed process")
				}
			}
		}

		s.logger.WithFields(log.Fields{
			"client_process_uuid": pUuid,
			"labels":              labels,
		}).Info("removed client process, suspected crash")
	}
	return nil
}

// ManageLifecycle - calc phi && de-registers expired procs on set interval + publishes
func (s *Server) ManageLifecycle(ctx context.Context, publishMetrics bool) {

	var calcTimestamp time.Time

	logTicker := time.NewTicker(s.dOpts.ManagementInterval)
	defer logTicker.Stop()

	for {
		select {
		case <-logTicker.C:
			calcTimestamp = time.Now()

			// calc phi
			deadProcs, err := s.CalculateProcessSuspicion(calcTimestamp, publishMetrics)
			if err != nil {
				s.logger.WithFields(log.Fields{
					"server_host_id": s.hostID,
					"err":            err,
				}).Error("error calculating suspicion")
			}

			// note: making `PurgeSuspectedProcs` an option here -> probably don't want to blow away
			// a node's history from a single bad meessage...
			if s.dOpts.PurgeAllSuspectedProcs {
				err = s.PurgeDeadProcs(deadProcs, publishMetrics)
				if err != nil {
					s.logger.WithFields(log.Fields{
						"server_host_id": s.hostID,
						"err":            err,
					}).Error("error purging suspected procs")
				}
			}

		}
	}
}
