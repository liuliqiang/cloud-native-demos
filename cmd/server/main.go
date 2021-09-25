package main

import (
    "flag"
    "fmt"
    "log"
    "net"

    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
	
    "github.com/liuliqiang/xds-demo/interface/grpc/echo-server"
    "github.com/liuliqiang/xds-demo/interface/grpc/gen/echo/v1"
)

var (
    host string
    port int
    iden string
)

func init() {
    flag.StringVar(&host, "host", "0.0.0.0", "echo-server listen interface ip")
    flag.IntVar(&port, "port", 9000, "echo-server listen port")
    flag.StringVar(&iden, "iden", "111111", "identity for this echo echo-server")
    flag.Parse()
}

func main() {
    addr := fmt.Sprintf("%s:%d", host, port)
    lis, err := net.Listen("tcp", addr)
    if err != nil {
        log.Fatalf("failed to listen '%s': %v", addr, err)
    }

    grpcServer := grpc.NewServer()
    echoServer := echo_server.NewEchoServer(iden)

    echo.RegisterEchoServer(grpcServer, echoServer)

    log.Printf("run echo-server at '%s', identity: %s\n", addr, iden)

    reflection.Register(grpcServer)
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve echo: %v", err)
    }
}
