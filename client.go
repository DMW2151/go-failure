package failure

import (
	"context"
	"fmt"
	"time"

	failproto "github.com/dmw2151/go-failure/proto"

	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

// ClientOptions - options for managing what/how we send to the fd server
type ClientOptions struct {
	Interval time.Duration
}

// StartPhiAccClient - run forever
func StartPhiAccClient(ctx context.Context, c failproto.PhiAccrualClient, msg *failproto.Beat, clOpts *ClientOptions) error {

	//
	hbTicker := time.NewTicker(clOpts.Interval)
	defer hbTicker.Stop()

	for {
		select {
		case <-hbTicker.C:
			// send a heartbeat to the server...
			_, err := c.Heartbeat(ctx, msg)
			if err != nil {
				log.WithFields(log.Fields{
					"client_process_uuid": msg.Uuid,
					"tags":                msg.Tags,
					"err":                 err,
				}).Warn("error sending heartbeat")
			}

		// exit condition, at the moment)
		case <-ctx.Done():
			log.WithFields(log.Fields{
				"client_process_uuid": msg.Uuid,
				"tags":                msg.Tags,
				"err":                 ctx.Err(),
			}).Warn("client context canceled")
			return ctx.Err()
		}
	}

	return nil
}

// NewFailureDetectorClient -
func NewFailureDetectorClient(host string, port string, dialOptions []grpc.DialOption) (failproto.PhiAccrualClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", host, port), dialOptions...)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to dial fd server")
		return nil, err
	}
	return failproto.NewPhiAccrualClient(conn), nil
}
