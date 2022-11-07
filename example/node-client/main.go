package main

import (
	"context"

	lalbproto "github.com/dmw2151/go-failure/example/proto/lalb"
	orcaproto "github.com/dmw2151/go-failure/example/proto/orca"

	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
	insecure "google.golang.org/grpc/credentials/insecure"
)

var (
	lookAsideLoadBalancerAddr string  = "localhost:52151"
	svcLabel                  string  = "worker"
	maxAllowedSuspicion       float64 = 0.8
	numNodesRequested         int64   = 1
)

func main() {

	conn, _ := grpc.Dial(
		lookAsideLoadBalancerAddr, []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}...,
	)
	defer conn.Close()

	// start the client - this will talk to the look-aside loadbalancing server / discovery service
	// then one of the downstream application servers
	lalbClient := lalbproto.NewHeartBeatClient(conn)

	// requesting a healthy node from look-aside-loadbalancer...
	healthyNodes, _ := lalbClient.HealthyNodes(context.Background(), &lalbproto.NodeHealthRequest{
		Limit:        numNodesRequested,
		Threshold:    maxAllowedSuspicion,
		ServiceLabel: svcLabel,
	})

	// note: bug here -> the call to healthy nodes returns the *client* address, not the listening
	// address of the orca server...
	if len(healthyNodes.Statuses) > 0 {
		connToUse := healthyNodes.Statuses[0]
		orcaconn, _ := grpc.Dial(
			"127.0.0.1:52152", []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			}...,
		)
		defer orcaconn.Close()

		orcaClient := orcaproto.NewORCAClient(orcaconn)
		if resp, err := orcaClient.Orca(context.Background(), &orcaproto.ORCARequest{}); err == nil {
			log.WithFields(log.Fields{
				"an-orielly-animal": resp.Name,
				"connection":        connToUse.Addr,
			}).Info("got an o'reilly anima")
		}
	}
}
