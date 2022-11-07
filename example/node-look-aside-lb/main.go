package main

import (
	"context"
	"net"
	"net/http"
	"time"

	fail "github.com/dmw2151/go-failure"
	lalbproto "github.com/dmw2151/go-failure/example/proto/lalb"
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
	lalbproto.UnimplementedHeartBeatServer
	failureDetector *fail.Node
}

// Beat -
func (lb lookasideLoadBalancer) Beat(ctx context.Context, in *failproto.Beat) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// HealthyNodes -
func (lb lookasideLoadBalancer) HealthyNodes(ctx context.Context, in *lalbproto.NodeHealthRequest) (*lalbproto.NodeHealthResponse, error) {

	var (
		hNodes                = []*lalbproto.NodeHealthStatus{}
		ctr         int64     = 0
		arrivalTime time.Time = time.Now()
	)

	// todo: run these on own go-routines to save a few ms (prob. only worth when large #
	// of connected clients)
	for addr, detector := range lb.failureDetector.RecentClients {
		if phi := detector.Suspicion(arrivalTime); phi < in.Threshold {
			hNodes = append(hNodes, &lalbproto.NodeHealthStatus{
				Addr:      addr,
				Suspicion: phi,
			})
			ctr++
		}
		if ctr >= in.Limit {
			break
		}
	}

	// return all healthy nodes...
	return &lalbproto.NodeHealthResponse{
		Statuses: hNodes,
	}, nil
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

	lalbproto.RegisterHeartBeatServer(grpcServer, lookasideLoadBalancer{
		failureDetector: failureDetector,
	})

	log.Info("starting look-aside-lb server")
	grpcServer.Serve(lis)
}
