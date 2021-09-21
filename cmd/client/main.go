package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/liuliqiang/xds-demo/interface/grpc/gen/echo/v1"
)

var (
	host string
	port int
	info string
)

func init() {
	flag.StringVar(&host, "host", "0.0.0.0", "connect to echo-server's ip")
	flag.IntVar(&port, "port", 9000, "connect to echo-server's port")
	flag.StringVar(&info, "info", "hi", "message sent to echo-server")
	flag.Parse()
}

func main() {
	os.Setenv("GRPC_VERBOSITY", "DEBUG")
	os.Setenv("GRPC_TRACE", "all")
	target := fmt.Sprintf("%s:%d", host, port)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("connect to echo-server '%s': %v", target, err)
	}
	log.Println("Connect echo-server success.")

	cli := echo.NewEchoClient(conn)

	sendCtx, cancel2 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel2()

	resp, err := cli.Ping(sendCtx, &echo.Message{
		Say: info,
	})
	if err != nil {
		log.Fatalf("send message to echo-server '%s': %v", target, err)
	}
	log.Println(resp.Say)
}
