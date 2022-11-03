package failure

import (
	"context"
	"fmt"
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
		phi, delta float64
	)

	// if client process already exists -> update entry in recentClients w. delta since last event
	if detector, ok := n.recentClients[clientID]; ok {
		go func(){
			labels := prometheus.Labels{
				"client_app_id": beatmsg.AppID,
				"server_app_id": n.metadata.AppID,
				"client_addr":   clientID,
				"server_addr":   n.metadata.HostAddress,
			}

			delta = float64(arrivalTime.Sub(detector.lastHeartbeat) / time.Millisecond)

			phi = detector.Suspicion(arrivalTime)
			suspicionHist.With(labels).Observe(phi)

			detector.AddValue(ctx, arrivalTime)
			heartbeatIntervalHist.With(labels).Observe(delta)
		}()

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

