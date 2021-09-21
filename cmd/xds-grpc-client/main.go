package main

import (
    "context"
    "flag"
    "fmt"
    "time"

    log "github.com/liuliqiang/log4go"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    _ "google.golang.org/grpc/xds" // To install the xds resolvers and balancers.

    "github.com/liuliqiang/xds-demo/interface/grpc/gen/echo/v1"
)

var (
    target string
    info   string
)

func init() {
    flag.StringVar(&target, "target", "xds:///listener_0", "xds address")
    flag.StringVar(&info, "info", "hi", "message sent to echo-server")
    flag.Parse()
}

func main() {
    creds := insecure.NewCredentials()

    log.Info("Ready to to connect to xds")
    conn, err := grpc.Dial(target, grpc.WithTransportCredentials(creds))
    if err != nil {
        log.Error(fmt.Sprintf("grpc.Dial(%s) failed: %v", target, err))
        return
    }
    defer conn.Close()

    log.Info("Connect xds-server success.")

    cli := echo.NewEchoClient(conn)
    sendCtx, cancel2 := context.WithTimeout(context.Background(), time.Second*30)
    defer cancel2()

    for i := 0; i < 10; i++ {

        resp, err := cli.Ping(sendCtx, &echo.Message{
            Say: info,
        })
        if err != nil {
            log.Error(fmt.Sprintf("send message to echo-server '%s': %v", target, err))
            return
        }
        log.Info(resp.Say)
        time.Sleep(time.Second)
    }

}
