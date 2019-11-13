//go:generate protoc -I../proto --go_out=plugins=grpc:../proto ../proto/echo.proto
//go:generate protoc -I../proto --swagger_out=logtostderr=true:../proto ../proto/echo.proto

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/ginuerzh/echo/proto"
	svc1 "github.com/ginuerzh/svc1/proto"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	status "google.golang.org/grpc/status"
)

var (
	addr     = ":8080"
	svc1Addr = "svc1.echo:8080" // k8s service name for svc1
)

func init() {
	flag.StringVar(&addr, "l", ":8080", "grpc server address")
	flag.Parse()
}

type echoServer struct {
	svc1Client svc1.Svc1Client
}

func (s *echoServer) Echo(ctx context.Context, in *pb.EchoRequest) (*pb.EchoReply, error) {
	r, err := s.svc1Client.Serve(ctx, &svc1.Svc1Request{
		Request: in.Request,
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	b := bytes.Buffer{}
	fmt.Fprintf(&b, "[echo] from %s, ", hostname)
	b.WriteString(r.GetReply() + ", ")

	return &pb.EchoReply{
		Reply: b.String(),
	}, nil
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
	ctx := context.Background()
	svc1Conn, err := grpc.DialContext(ctx, svc1Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial svc1: %v", err)
	}

	srv := &echoServer{
		svc1Client: svc1.NewSvc1Client(svc1Conn),
	}

	s := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			unaryServerRecoveryInterceptor(),
			// unaryServerOpenTracingInterceptor(tracer),
			// unaryServerAuthInterceptor(),
			unaryServerLoggingInterceptor(),
		),
	)
	pb.RegisterEchoServer(s, srv)
	grpc_health_v1.RegisterHealthServer(s, srv)
	reflection.Register(s)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("server listen on", addr)

	if err := s.Serve(ln); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
