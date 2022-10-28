package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	fail "github.com/dmw2151/go-failure"
	failproto "github.com/dmw2151/go-failure/proto"

	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

var (
	failureDetectorMetricsAddress = os.Getenv("METRICS_SERVE_ADDR")           // "127.0.0.1:52150"
	failureDetectorListenAddress  = os.Getenv("FAILURE_DETECTOR_LISTEN_ADDR") // "0.0.0.0:52151"
	phiCalculationWindowSize      = 100
	managementInterval            = time.Duration(time.Second * 30)
	publishMetrics                = true
)

func main() {

	// expose Prometheus metrics on :XXXX/metrics ...
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(failureDetectorMetricsAddress, nil)

	// start new phi-acc server && begin updating stat
	fdServer, err := fail.NewFailureDetectorServer(&fail.DetectorOptions{
		WindowSize:             phiCalculationWindowSize,
		ManagementInterval:     managementInterval,
		PurgeAllSuspectedProcs: true,
	}, nil)

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("detector server failed to init")
	}
	go fdServer.ManageLifecycle(context.Background(), publishMetrics)

	// init grpc server && listen + serve
	lis, err := net.Listen("tcp", failureDetectorListenAddress)
	if err != nil {
		log.WithFields(log.Fields{
			"addr": failureDetectorListenAddress,
			"err":  err,
		}).Fatal("detector server failed to listen on addr")
	}

	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)
	failproto.RegisterPhiAccrualServer(grpcServer, fdServer)
	grpcServer.Serve(lis)
}
