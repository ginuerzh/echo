package main

import (
    "context"
    "os"
    "log"
    "net"
    pb "echo/proto"

    "google.golang.org/grpc"
    codes "google.golang.org/grpc/codes"
    status "google.golang.org/grpc/status"
    "google.golang.org/grpc/health/grpc_health_v1"
)


var (
   addr = ":8080"
)

type echoServer struct{}

func (s *echoServer) Echo(ctx context.Context, in *pb.EchoRequest) (*pb.EchoReply, error) {
    hostname, err := os.Hostname()
    if err != nil {
            return nil, err
    }

    return &pb.EchoReply{
        Reply: in.Request + " from " + hostname,
    },  nil
}

func (s *echoServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
    return &grpc_health_v1.HealthCheckResponse{
        Status: grpc_health_v1.HealthCheckResponse_SERVING,
    }, nil
}

func (s *echoServer) Watch(req *grpc_health_v1.HealthCheckRequest, srv grpc_health_v1.Health_WatchServer) error {
    return status.Errorf(codes.Unimplemented, "method Watch not implemented")
}

func main() {
    srv := &echoServer{}
    s := grpc.NewServer()
    pb.RegisterEchoServer(s, srv)
    grpc_health_v1.RegisterHealthServer(s, srv)

    ln, err := net.Listen("tcp", addr)
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    log.Println("server listen on", addr)

    if err := s.Serve(ln); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
