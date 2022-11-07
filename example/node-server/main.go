package main

import (
	"context"
	"math/rand"
	"net"
	"time"

	lalbproto "github.com/dmw2151/go-failure/example/proto/lalb"
	orcaproto "github.com/dmw2151/go-failure/example/proto/orca"
	failproto "github.com/dmw2151/go-failure/proto"

	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
	insecure "google.golang.org/grpc/credentials/insecure"
)

const (
	lookAsideLoadBalancerAddr string = "localhost:52151"
	orcaListenAddr            string = "0.0.0.0:52152"
)

type orcaServer struct {
	orcaproto.UnimplementedORCAServer
	heartBeatClient            lalbproto.HeartBeatClient
	heartBeatPublishIntervalMs int
}

func (orca orcaServer) Orca(ctx context.Context, req *orcaproto.ORCARequest) (*orcaproto.ORCAResponse, error) {
	return &orcaproto.ORCAResponse{Name: "Sus Scrofa Linnaeus"}, nil
}

// publishHeartBeat -
func (orca *orcaServer) publishHeartBeat(ctx context.Context, msg *failproto.Beat) error {
	for {
		log.Info("sending heartbeat message...")
		dur := time.Duration(rand.Intn(orca.heartBeatPublishIntervalMs)) * time.Millisecond
		time.Sleep(dur)
		if _, err := orca.heartBeatClient.Beat(ctx, msg); err != nil {
			log.Error(err)
		}
	}
	return ctx.Err()
}

func main() {

	conn, _ := grpc.Dial(
		lookAsideLoadBalancerAddr, []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}...,
	)
	defer conn.Close()

	// start the load-balancer heartbeat client
	orca := orcaServer{
		heartBeatClient:            lalbproto.NewHeartBeatClient(conn),
		heartBeatPublishIntervalMs: 2000,
	}

	msg := failproto.Beat{
		ClientID: "worker",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go orca.publishHeartBeat(ctx, &msg)

	// bind to addr && start ORCA server...
	lis, err := net.Listen("tcp", orcaListenAddr)
	if err != nil {
		log.WithFields(log.Fields{
			"orcaListenAddr": orcaListenAddr,
			"err":            err,
		}).Error("failed to start orca server; failed to listen on address")
	}

	grpcServer := grpc.NewServer()
	orcaproto.RegisterORCAServer(grpcServer, orca)

	log.Info("starting orca server")
	grpcServer.Serve(lis)

}
