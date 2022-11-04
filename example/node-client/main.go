package main

import (
	"context"

	exproto "github.com/dmw2151/go-failure/example/proto"
	failproto "github.com/dmw2151/go-failure/proto"
	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
	insecure "google.golang.org/grpc/credentials/insecure"
)

var lookAsideLoadBalancerAddr string = "localhost:52151"

func main() {

	conn, _ := grpc.Dial(
		lookAsideLoadBalancerAddr, []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}...,
	)
	defer conn.Close()

	// start the actual service client
	lalbClient := exproto.NewLBClient(conn)

	//
	healthyNodes, _ := lalbClient.HealthyNodes(context.Background(), &failproto.NodeHealthRequest{
		Limit:     8,
		Threshold: 100.0,
	})

	log.Infof("%+v", healthyNodes)

}
