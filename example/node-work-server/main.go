package main

import (
	"context"
	"math/rand"
	"time"

	exproto "github.com/dmw2151/go-failure/example/proto"
	failproto "github.com/dmw2151/go-failure/proto"

	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
	insecure "google.golang.org/grpc/credentials/insecure"
)

const lookAsideLoadBalancerAddr string = "localhost:52151"

func main() {

	conn, _ := grpc.Dial(
		lookAsideLoadBalancerAddr, []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}...,
	)
	defer conn.Close()

	// start the actual service client
	beatClient := exproto.NewLBClient(conn)
	beatMsg := failproto.Beat{
		ClientID: "worker-1",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		dur := time.Duration(rand.Intn(1000)) * time.Millisecond
		time.Sleep(dur)

		log.Infof("sending heartbeat message...")
		if _, err := beatClient.Beat(ctx, &beatMsg); err != nil {
			log.Error(err)
		}
	}
}
