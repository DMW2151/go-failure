package main

import (
	"context"
	"fmt"
	"os"
	"time"

	fail "failure"
	failproto "failure/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	failureDetectorServerHost = os.Getenv("FAILURE_DETECTOR_SERVER_HOST") // localhost | TBD
	failureDetectorServerPort = os.Getenv("FAILURE_DETECTOR_SERVER_PORT") // "52151"
	clientHeartBeatInterval   = 500 * time.Millisecond
)

func main() {

	//
	fdc, _ := fail.NewFailureDetectorClient(
		failureDetectorServerHost, failureDetectorServerPort,
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	)

	//
	host, _ := os.Hostname()
	msg := failproto.Beat{
		Uuid: uuid.New().String(),
		Tags: map[string]string{
			"client_pid":     fmt.Sprintf("%d", os.Getpid()),
			"client_host_id": host,
		},
	}

	// send a heartbeat messsage every N milliseconds
	fail.StartPhiAccClient(context.Background(), fdc, &msg, &fail.ClientOptions{
		Interval: clientHeartBeatInterval,
	})

}
