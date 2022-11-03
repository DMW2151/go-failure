package main

import (
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

// lookasideLoadBalancer -
type lookasideLoadBalancer struct {
	lalbproto.UnimplementedLBServer
	failureDetector *fail.Node
}

func main() {

	log.Info("starting look-aside-lb metrics server")
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(failureDetectorMetricsAddress, nil)

	failureDetector := fail.NewFailureDetectorNode(&fail.NodeOptions{
		EstimationWindowSize: estimationWindowSize,
		ReapInterval:         reapInterval,
	}, &fail.NodeMetadata{
		HostAddress: balancerListenAddresss,
		AppID:       nodeID,
	})

	// init grpc server && listen + serve
	lis, err := net.Listen("tcp", balancerListenAddresss)
	if err != nil {
		log.WithFields(log.Fields{
			"balancerListenAddresss": balancerListenAddresss, 
			"err": err,
		}).Error("failed to start look-aside-lb; failed to listen on address")
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(failureDetector.FailureDetectorInterceptor()),
	)

	lalbproto.RegisterLBServer(grpcServer, lookasideLoadBalancer{
		failureDetector: failureDetector,
	})
	grpcServer.Serve(lis)
}
