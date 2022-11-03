package main

import (
	"context"
	"math/rand"
	"time"

	exproto "github.com/dmw2151/go-failure/example/proto"
	failproto "github.com/dmw2151/go-failure/proto"

	uuid "github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
	insecure "google.golang.org/grpc/credentials/insecure"
)

const testEchoServerAddress string = "localhost:52151"

var clientUUID = uuid.New().String()

func main() {

	conn, _ := grpc.Dial(
		testEchoServerAddress, []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}...,
	)
	defer conn.Close()

	// start the actual service client
	echoClient := exproto.NewLBClient(conn)

	for {
		dur := time.Duration(rand.Intn(1000)) * time.Millisecond
		time.Sleep(dur)

		log.Infof("sending heartbeat message...")
		echoClient.Beat(context.Background(), &failproto.Beat{
			Uuid:  clientUUID,
			AppID: "worker-1",
		})
	}
	
}
