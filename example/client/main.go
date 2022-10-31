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

const testEchoServerAddress string = "localhost:12151"

func main() {

	conn, err := grpc.Dial(
		testEchoServerAddress, []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}...,
	)
	if err != nil {
		log.Error(err)
	}

	clientUUID := uuid.New().String()

	// start the hb client
	beatClient := failproto.NewHealthClient(conn)
	go func() {
		for {
			dur := time.Duration(rand.Intn(1000)) * time.Millisecond
			time.Sleep(dur)

			log.Infof("sending heartbeat message...")
			beatClient.Heartbeat(context.Background(), &failproto.Beat{
				Uuid: clientUUID,
				Tags: map[string]string{
					"a-very-long-tag": "a-very-long-tag",
				},
			})
		}
	}()

	// start the actual service client
	echoClient := exproto.NewEchoClient(conn)

	t := time.NewTicker(time.Second * 1)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			log.Infof("sending a application message...")
			echoClient.Echo(context.Background(), &exproto.EchoRequest{Body: "hello-world"})
		}
	}
}
