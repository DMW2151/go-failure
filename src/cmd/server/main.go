package main

import (
	"context"
	"net"
	"net/http"
	"time"

	fail "failure"
	failproto "failure/proto"

	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	phiCalculationWindowSize      = 100
	managementInterval            = time.Duration(time.Millisecond * 100)
	failureDetectorMetricsAddress = os.Getenv("METRICS_SERVE_ADDR") // "127.0.0.1:52150"
	failureDetectorListenAddress  = os.Getenv("FAILURE_DETECTOR_LISTEN_ADDR") // "0.0.0.0:52151"
	publishMetrics                = true
)

func main() {

	// run server...
	fdServer, err := fail.NewFailureDetectorServer(&fail.DetectorOptions{
		WindowSize:         phiCalculationWindowSize,
		ManagementInterval: managementInterval,
	}, nil)

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("detector server failed to init")
	}

	//
	lis, err := net.Listen("tcp", failureDetectorListenAddress)
	if err != nil {
		log.WithFields(log.Fields{
			"addr": failureDetectorListenAddress,
			"err":  err,
		}).Fatal("detector server failed to listen on addr")
	}

	//
	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)
	failproto.RegisterPhiAccrualServer(grpcServer, fdServer)

	// in the background ->
	go fdServer.ManageLifecycle(context.Background(), publishMetrics)

	// configure logrus to use the Prometheus hook && expose Prometheus metrics via HTTP
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(failureDetectorMetricsAddress, nil)

	// run the failure detection server s.t. it can recv heartbeat msgs
	grpcServer.Serve(lis)
}
