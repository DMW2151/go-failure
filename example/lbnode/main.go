package main

import (
	"context"
	"net"
	"net/http"
	"time"

	fail "github.com/dmw2151/go-failure"
	lalbproto "github.com/dmw2151/go-failure/example/proto"
	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

const (
	failureDetectorMetricsAddress string        = "localhost:52150"
	balancerListenAddresss        string        = "localhost:52151"
	nodeID                        string        = "look-aside-load-balancer"
	estimationWindowSize          int           = 100
	reapInterval                  time.Duration = time.Second * 10
)

// healthCheck -
type lookasideLoadBalancer struct {
	lalbproto.UnimplementedLBServer
	failureDetector *fail.Node
}

func main() {

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(failureDetectorMetricsAddress, nil)

	nodeOptions := fail.NodeOptions{
		EstimationWindowSize: estimationWindowSize,
		ReapInterval:         reapInterval,
	}

	nodeMetadata := fail.NodeMetadata{
		HostAddress: balancerListenAddresss,
		AppID:       nodeID,
	}

	failureDetector := fail.NewFailureDetectorNode(&nodeOptions, &nodeMetadata)
	go failureDetector.WatchConnectedClients(context.Background())

	// start new look-aside load balancer
	log.Info("starting look-aside lb server")
	lalbServer := lookasideLoadBalancer{
		failureDetector: failureDetector,
	}

	// init grpc server && listen + serve
	lis, err := net.Listen("tcp", balancerListenAddresss)
	if err != nil {
		log.Error(err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(failureDetector.FailureDetectorInterceptor()),
	)

	lalbproto.RegisterLBServer(grpcServer, lalbServer)
	grpcServer.Serve(lis)
}
