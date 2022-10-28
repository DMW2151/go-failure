package main

import (
	"context"
	"fmt"
	"os"
	"time"

	fail "github.com/dmw2151/go-failure"
	failproto "github.com/dmw2151/go-failure/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	failureDetectorServerHost = os.Getenv("FAILURE_DETECTOR_SERVER_HOST") // TBD
	failureDetectorServerPort = os.Getenv("FAILURE_DETECTOR_SERVER_PORT") // "52151"
	clientHeartBeatInterval   = 300 * time.Millisecond
)

func main() {

	// init NewFailureDetectorClient...
	fdc, _ := fail.NewFailureDetectorClient(
		failureDetectorServerHost, failureDetectorServerPort,
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	)

	// define message to send as heartbeat
	// todo: `StartPhiAccClient` could send full message once & (only) ID on subsequent beats
	host, _ := os.Hostname()
	msg := failproto.Beat{
		Uuid: uuid.New().String(),
		Tags: map[string]string{
			"client_pid":     fmt.Sprintf("%d", os.Getpid()),
			"client_host_id": host,
			"client_region":  os.Getenv("DIGITALOCEAN_REGION"),
		},
	}

	// send a heartbeat messsage every N milliseconds
	fail.StartPhiAccClient(context.Background(), fdc, &msg, &fail.ClientOptions{
		Interval: clientHeartBeatInterval,
	})

}
