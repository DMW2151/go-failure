package failure

import (
	"context"
	"fmt"
	"math"
	"time"

	failproto "github.com/dmw2151/go-failure/proto"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// Node - collection of health detectors for each client sending through the interceptor
type Node struct {
	recentClients map[string]*PhiAccrualDetector // maps senderAddress -> detector
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
		recentClients: make(map[string]*PhiAccrualDetector),
		opts:          nOpts,
		metadata:      nMetadata,
	}
}

// ReceiveHeartbeat - create or update a record in the node's recentClients
func (n *Node) ReceiveHeartbeat(ctx context.Context, clientID string, beatmsg *failproto.Beat) error {

	var (
		arrivalTime time.Time = time.Now()
		delta float64
	)

	// if client process already exists -> update entry in recentClients w. delta since last event
	if detector, ok := n.recentClients[clientID]; ok {

		delta = float64(arrivalTime.Sub(detector.lastHeartbeat) / time.Millisecond)
		detector.AddValue(ctx, arrivalTime)

		log.WithFields(log.Fields{
			"client_app_id": beatmsg.AppID,
			"server_app_id": n.metadata.AppID,
			"client_addr":   clientID,
			"server_addr":   n.metadata.HostAddress,
		}).Debug("heartbeat from client")

		// update histogram metrics
		heartbeatIntervalHist.With(prometheus.Labels{
			"client_app_id": beatmsg.AppID,
			"server_app_id": n.metadata.AppID,
			"client_addr":   clientID,
			"server_addr":   n.metadata.HostAddress,
		}).Observe(delta)

		return nil
	}

	// if client process DNE -> create an entry in recentClients and increment the guage for
	// activeClients
	n.recentClients[clientID] = NewPhiAccrualDetector(arrivalTime, n.opts, &NodeMetadata{
		HostAddress: clientID,
		AppID:       beatmsg.AppID,
	})

	log.WithFields(log.Fields{
		"client_app_id":   beatmsg.AppID,
		"server_app_id":   n.metadata.AppID,
		"client_addr":     clientID,
		"server_addr":     n.metadata.HostAddress,
		"current_clients": len(n.recentClients),
	}).Info("received heartbeat from new client")

	activeClientsGauge.With(prometheus.Labels{
		"client_app_id": beatmsg.AppID,
		"server_app_id": n.metadata.AppID,
		"client_addr":   clientID,
		"server_addr":   n.metadata.HostAddress,
	}).Inc()
	return nil
}

// FailureDetectorInterceptor - Acts as a UnaryServerInterceptor, updates detector node's heartbeat statistics when sees
// an incoming failproto.Beat message
func (n *Node) FailureDetectorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		// if incoming messasge is `*failproto.Beat`; update node's heartbeat statistics
		if msg, ok := req.(*failproto.Beat); ok {
			if p, ok := peer.FromContext(ctx); ok {
				err := n.ReceiveHeartbeat(ctx, p.Addr.String(), msg)
				if err != nil {
					log.WithFields(log.Fields{
						"err":  err,
						"beat": fmt.Sprintf("%+v", msg),
					}).Error("failed to update heartbeat statistics")
				}
			} else {
				// failed to extract peer from context - can't get client info ->> do not add to recentClients
				log.WithFields(log.Fields{
					"beat": fmt.Sprintf("%+v", msg),
				}).Error("failed to extract peer from context")
			}
		}

		// call next interceptor/handler in call chain
		h, err := handler(ctx, req)
		return h, err
	}
}

// WatchConnectedClients - watches all connected clients, if the reap interval
func (n *Node) WatchConnectedClients(ctx context.Context) {

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
	for addr, detector := range n.recentClients {
		phi = detector.Suspicion(calcTimestamp)


		// require the following two conditions
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
			delete(n.recentClients, addr)
		}
	}
}
