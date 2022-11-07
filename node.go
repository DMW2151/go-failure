package failure

import (
	"context"
	"math"
	"time"

	failproto "github.com/dmw2151/go-failure/proto"

	log "github.com/sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// Node - collection of health detectors for each client sending through the interceptor
type Node struct {
	RecentClients map[string]*PhiAccrualDetector // maps senderAddress -> detector
	opts          *NodeOptions
	metadata      *NodeMetadata
}

// NodeMetadata - metadata abt. the running grpc application for labeling published metrics
type NodeMetadata struct {
	HostAddress string
	AppID       string
}

// NodeOptions - options for distribution estimation window, purging interval, etc.
type NodeOptions struct {
	EstimationWindowSize int
	ReapInterval         time.Duration
	PurgeGracePeriod     time.Duration
}

// NewFailureDetectorNode - new failure-detecting node
func NewFailureDetectorNode(nOpts *NodeOptions, nMetadata *NodeMetadata) *Node {
	return &Node{
		RecentClients: make(map[string]*PhiAccrualDetector),
		opts:          nOpts,
		metadata:      nMetadata,
	}
}

// ReceiveHeartbeat - create or update a record in the node's RecentClients
func (n *Node) ReceiveHeartbeat(ctx context.Context, clientID string, beatmsg *failproto.Beat) error {

	var (
		arrivalTime time.Time = time.Now()
		phi, delta  float64
	)

	// client process already exists -> update entry in RecentClients w. delta since last event
	if detector, ok := n.RecentClients[clientID]; ok {
		labels := prometheus.Labels{
			"client_app_id": beatmsg.ClientID,
			"server_app_id": n.metadata.AppID,
			"client_addr":   clientID,
			"server_addr":   n.metadata.HostAddress,
		}

		delta = float64(arrivalTime.Sub(detector.lastHeartbeat) / time.Millisecond)

		// note: do not update histogram w. the most recent phi if NaN or Inf, these vals
		// ruin the distribution of the histogram!
		phi = detector.Suspicion(arrivalTime)
		if !(math.IsNaN(phi) || math.IsInf(phi, 1) || math.IsInf(phi, -1)) {
			detector.lastPhi = phi
			suspicionHist.With(labels).Observe(phi)
		}

		// update timedelta, always safe to update w. delta, massive times just fall into +Inf 
		// histogram bucket
		detector.AddValue(ctx, arrivalTime)
		heartbeatIntervalHist.With(labels).Observe(delta)
		return nil
	}

	// if client process DNE -> create an entry in RecentClients and increment the guage for
	// activeClients
	n.RecentClients[clientID] = NewPhiAccrualDetector(arrivalTime, n.opts, &NodeMetadata{
		HostAddress: clientID,
		AppID:       beatmsg.ClientID,
	})

	log.WithFields(log.Fields{
		"client_app_id":   beatmsg.ClientID,
		"server_app_id":   n.metadata.AppID,
		"client_addr":     clientID,
		"server_addr":     n.metadata.HostAddress,
		"current_clients": len(n.RecentClients),
	}).Info("received heartbeat from new client")

	activeClientsGauge.With(prometheus.Labels{
		"client_app_id": beatmsg.ClientID,
		"server_app_id": n.metadata.AppID,
		"client_addr":   clientID,
		"server_addr":   n.metadata.HostAddress,
	}).Inc()
	return nil
}

// WatchConnectedNodes - watches all connected clients, if the reap interval
func (n *Node) WatchConnectedNodes(ctx context.Context) {

	// start ticker
	logTicker := time.NewTicker(n.opts.ReapInterval)
	defer logTicker.Stop()

	for {
		select {
		case t := <-logTicker.C:
			n.PurgeInactiveClients(ctx, t)
		case <-ctx.Done():
			// context cancelled
			log.WithFields(log.Fields{
				"err": ctx.Err(),
			}).Error("conext error")
			return
		}
	}
}

// PurgeNeighbors - calculates phi and removes processes that have been marked suspicious (using
// +inf as suspicion threshold) AND have not been seen within grace period.
func (n *Node) PurgeInactiveClients(ctx context.Context, calcTimestamp time.Time) {

	var phi float64

	// remove clients w. infinite suspicion
	for addr, detector := range n.RecentClients {

		phi = detector.Suspicion(calcTimestamp)
		detector.lastPhi = phi

		// require the following two conditions -
		if (calcTimestamp.Sub(detector.lastHeartbeat) > n.opts.PurgeGracePeriod) && (phi == math.Inf(1)) {

			var labels = prometheus.Labels{
				"client_app_id": detector.metadata.AppID,
				"server_app_id": n.metadata.AppID,
				"client_addr":   addr,
				"server_addr":   n.metadata.HostAddress,
			}

			// if client process suspicion == 1 & age -> decrement the guage for activeClients
			activeClientsGauge.With(labels).Dec()
			if n := heartbeatIntervalHist.DeletePartialMatch(labels); n == 0 {
				log.WithFields(log.Fields{
					"labels": labels,
				}).Warn("failed to remove metrics of (suspected) crashed process")
			}

			log.WithFields(log.Fields{
				"client_app_id": detector.metadata.AppID,
				"server_app_id": n.metadata.AppID,
				"client_addr":   addr,
				"server_addr":   n.metadata.HostAddress,
			}).Info("no heartbeat from client in max suspicion interval, removing")

			// warn:
			delete(n.RecentClients, addr)
		}
	}
}

// FailureDetectorInterceptor - Acts as a UnaryServerInterceptor, updates detector node's heartbeat statistics when sees
// an incoming failproto.Beat message
func (n *Node) FailureDetectorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// todo: add a check for `service-name` here + add a check for incoming *listen* address, not incoming client address!!
		if msg, ok := req.(*failproto.Beat); ok {
			if p, ok := peer.FromContext(ctx); ok {
				go n.ReceiveHeartbeat(ctx, p.Addr.String(), msg)
			}
		}
		h, err := handler(ctx, req)
		return h, err
	}
}
