package main

import (
	"context"
	"net"
	"time"

	fail "github.com/dmw2151/go-failure"
	exproto "github.com/dmw2151/go-failure/example/proto"
	failproto "github.com/dmw2151/go-failure/proto"

	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

const testListenAddresss string = "localhost:12151"

// echoServer -
type echoServer struct {
	exproto.UnimplementedEchoServer
	failproto.UnimplementedHealthServer
}

func (e echoServer) Echo(ctx context.Context, req *exproto.EchoRequest) (*exproto.EchoResponse, error) {
	log.Infof("got a message: %s", req.Body)
	return &exproto.EchoResponse{
		Body: req.Body,
	}, nil
}

func main() {

	failureDetector, _ := fail.NewFailureDetectorNode(&fail.NodeOptions{
		EstimationWindowSize: 100,
		ReapInterval:         time.Second * 10,
	})
	go failureDetector.WatchNodeStatus(context.Background())

	// start new echo server
	log.Info("starting echo server...")
	exampleEchoServer := echoServer{}

	// init grpc server && listen + serve
	lis, err := net.Listen("tcp", testListenAddresss)
	if err != nil {
		log.Error(err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(failureDetector.PhiAccrualInterceptor()),
	)

	exproto.RegisterEchoServer(grpcServer, exampleEchoServer)
	failproto.RegisterHealthServer(grpcServer, exampleEchoServer)
	grpcServer.Serve(lis)
}
