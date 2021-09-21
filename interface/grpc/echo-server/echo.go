package echo_server

import (
    "context"
    "fmt"

    "github.com/liuliqiang/xds-demo/interface/grpc/gen/echo/v1"
)

var (
    _ echo.EchoServer = (*server)(nil)
)

func NewEchoServer(iden string) *server {
    return &server{
        identity: iden,
    }
}

type server struct {
    identity string
}

func (s *server) Ping(ctx context.Context, message *echo.Message) (*echo.Message, error) {
    return &echo.Message{
        Say: fmt.Sprintf("from %s: %s", s.identity, message),
    }, nil
}
