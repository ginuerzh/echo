package main

import (
    "context"
    "flag"
    "log"
    "net/http"
    gw "echo/proto"

    "google.golang.org/grpc"
    "github.com/grpc-ecosystem/grpc-gateway/runtime"
)

var (
    echoSvc = flag.String("s", "localhost:8080", "echo gRPC service")
)

func init() {
    flag.Parse()
}

func run() error {
    ctx := context.Background()
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    mux := runtime.NewServeMux()
    opts := []grpc.DialOption{grpc.WithInsecure()}
    if err := gw.RegisterEchoHandlerFromEndpoint(ctx, mux,  *echoSvc, opts); err != nil {
        return err
    }

    return http.ListenAndServe(":8000", mux)
}

func main() {
    flag.Parse()
    if err := run(); err != nil {
        log.Fatal(err)
    }
}
