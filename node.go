package failure

import (
	"context"
	"math"
	"time"

	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"

	failproto "github.com/dmw2151/go-failure/proto"
)

// Node - maintains collection of health-detectors + metadata for each neighbor of this node
type Node struct {
	connectedNodes map[string]*PhiAccrualDetector
	Opts           *NodeOptions
}

// NodeOptions - config options to control failure detection estimation window, etc...
type NodeOptions struct {
	EstimationWindowSize int
	ReapInterval         time.Duration
}

// NewFailureDetectorNode - creates new Node
func NewFailureDetectorNode(nOpts *NodeOptions) (*Node, error) {
	return &Node{
		connectedNodes: make(map[string]*PhiAccrualDetector),
		Opts:           nOpts,
	}, nil
}

// Heartbeat - required to implement `failproto.PhiAccrualServer`, handler for recv'ing a heartbeat. On recv.,
// either create a new detector, or update an existing
func (n *Node) ReceiveHeartbeat(ctx context.Context, arrivalTime time.Time, hb *failproto.Beat) (error) {

	var detector *PhiAccrualDetector

	// lookup a client process by UUID && check if exists/DNE
	detector, ok := n.connectedNodes[hb.Uuid]

	// if client process DNE -> create a new entry in the registry of tracked clients
	if !ok {
		detector = NewPhiAccrualDetector(n.Opts.EstimationWindowSize)
		detector.lastHeartbeat = arrivalTime
		detector.Tags = hb.Tags
		n.connectedNodes[hb.Uuid] = detector
		return nil
	}

	// if client process (by UUID) already exists -> calculate dela since last event
	// add to interval, update stats, update last arrival time, etc.
	hbDelta := float64(arrivalTime.Sub(detector.lastHeartbeat) / time.Microsecond)
	detector.AddValue(hbDelta)
	detector.lastHeartbeat = arrivalTime

	return nil
}

// PhiAccrualInterceptor
func (n *Node) PhiAccrualInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		// foo..
		if beat, ok := req.(*failproto.Beat); ok {
			if err := n.ReceiveHeartbeat(ctx, time.Now(), beat); err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("error adding...")
			}
		}

		// main
		h, err := handler(ctx, req)
		return h, err
	}
}

// WatchNeighborNodes - calc phi && de-registers expired procs on set interval + publishes
func (n *Node) WatchNodeStatus(ctx context.Context) {

	// start ticker
	logTicker := time.NewTicker(n.Opts.ReapInterval)
	defer logTicker.Stop()

	for {
		select {
		case <-logTicker.C:
			err := n.PurgeNeighbors(ctx, time.Now())
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("error purging suspected procs")
			}
		}
	}
}

// PurgeNeighbors - Calculates phi and removes processes that have been marked suspicious (using
// +inf as suspicion threshold)
func (n *Node) PurgeNeighbors(ctx context.Context, calcTimestamp time.Time) error {

	var deadProcs []string

	// calulate phi from (last heartbeat, present), this can be called at any time, so 0.0 suspicion
	// is a very common outcome when called at random (esp. if low variance on arrival times)
	//
	// When suspicion is +Inf (distribution collapses to +Inf around mean + 12 stdev),
	// then mark node failed and remove.
	for pUuid, detector := range n.connectedNodes {
		if phi := detector.CurrentSuspicion(); phi == math.Inf(1) {
			deadProcs = append(deadProcs, pUuid)
		}
	}

	// remove nodes marked for deletion in prev. step
	for _, pUuid := range deadProcs {
		delete(n.connectedNodes, pUuid)
		log.WithFields(log.Fields{
			"client_process_uuid": pUuid,
		}).Info("removed client process, suspected crash")
	}
	return nil
}
