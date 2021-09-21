package xds_server

import (
    "context"

    ads "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
    rds "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
    log "github.com/liuliqiang/log4go"
)

var (
    _ rds.RouteDiscoveryServiceServer = (*routeDiscoveryService)(nil)
)

func newRouteDiscoveryService() (s rds.RouteDiscoveryServiceServer) {
    return &routeDiscoveryService{}
}

type routeDiscoveryService struct {
}

func (r *routeDiscoveryService) StreamRoutes(server rds.RouteDiscoveryService_StreamRoutesServer) error {
    log.Info("stream routes invoked!")

    return nil
}

func (r *routeDiscoveryService) DeltaRoutes(server rds.RouteDiscoveryService_DeltaRoutesServer) error {
    log.Info("delta routes invoked!")

    return nil
}

func (r *routeDiscoveryService) FetchRoutes(
    ctx context.Context,
    request *ads.DiscoveryRequest,
) (
    *ads.DiscoveryResponse,
    error,
) {

    log.Info("fetch routes invoked!")

    return nil, nil
}
