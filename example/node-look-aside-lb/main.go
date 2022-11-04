package main

import (
	"context"
	"net"
	"net/http"
	"time"

	fail "github.com/dmw2151/go-failure"
	lalbproto "github.com/dmw2151/go-failure/example/proto"
	failproto "github.com/dmw2151/go-failure/proto"

	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

var (
	failureDetectorMetricsAddress = "localhost:52150"
	balancerListenAddresss        = "localhost:52151"

	nOpts = fail.NodeOptions{
		EstimationWindowSize: 100,
		ReapInterval:         time.Second * 10,
	}

	nMetadata = fail.NodeMetadata{
		HostAddress: balancerListenAddresss,
		AppID:       "look-aside-load-balancer",
	}
)

// lookasideLoadBalancer -
type lookasideLoadBalancer struct {
	lalbproto.UnimplementedLBServer
	failureDetector *fail.Node
}

func (lb lookasideLoadBalancer) Beat(ctx context.Context, in *failproto.Beat) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (lb lookasideLoadBalancer) HealthyNodes(ctx context.Context, in *failproto.NodeHealthRequest) (*failproto.NodeHealthResponse, error) {

	var (
		hNodes                = []*failproto.NodeHealthStatus{}
		ctr         int64     = 0
		arrivalTime time.Time = time.Now()
		phi         float64
	)

	for addr, detector := range lb.failureDetector.RecentClients {
		phi = detector.Suspicion(arrivalTime)
		if detector.Suspicion(arrivalTime) < in.Threshold {
			hNodes = append(hNodes, &failproto.NodeHealthStatus{
				Addr:      addr,
				Suspicion: phi,
			})
			ctr++
		}

		if ctr >= in.Limit {
			break
		}
	}

	resp := failproto.NodeHealthResponse{
		Statuses: hNodes,
	}
	return &resp, nil
}

// startPromMetricsEndPoint -
func startPromMetricsEndPoint(addr string) {
	log.Info("starting look-aside-lb metrics server")
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(addr, nil)
}

func main() {

	// begin serving a metrics endpoint at localhost:52150
	go startPromMetricsEndPoint(failureDetectorMetricsAddress)

	// init failure detector node & begin monitoring the status of all clients
	failureDetector := fail.NewFailureDetectorNode(&nOpts, &nMetadata)
	go failureDetector.WatchConnectedNodes(context.Background())

	// init grpc server && listen + serve
	lis, err := net.Listen("tcp", balancerListenAddresss)
	if err != nil {
		log.WithFields(log.Fields{
			"balancerListenAddresss": balancerListenAddresss,
			"err":                    err,
		}).Error("failed to start look-aside-lb; failed to listen on address")
	}

	// start look-aside load balancer
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(failureDetector.FailureDetectorInterceptor()),
	)

	lalbproto.RegisterLBServer(grpcServer, lookasideLoadBalancer{
		failureDetector: failureDetector,
	})

	log.Info("starting look-aside-lb server")
	grpcServer.Serve(lis)
}
