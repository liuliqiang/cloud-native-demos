package xds_server

import (
    "context"

    ads "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
    eds "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
    log "github.com/liuliqiang/log4go"
)

var (
    _ eds.EndpointDiscoveryServiceServer = (*endpointDiscoveryService)(nil)
)

func newEndpointDiscoveryService() (s eds.EndpointDiscoveryServiceServer) {
    return &endpointDiscoveryService{}
}

type endpointDiscoveryService struct {
}

func (e *endpointDiscoveryService) StreamEndpoints(server eds.EndpointDiscoveryService_StreamEndpointsServer) error {
    log.Info("stream endpoints invoked!")

    return nil
}

func (e *endpointDiscoveryService) DeltaEndpoints(server eds.EndpointDiscoveryService_DeltaEndpointsServer) error {
    log.Info("delta endpoints invoked!")

    return nil
}

func (e *endpointDiscoveryService) FetchEndpoints(
    ctx context.Context,
    request *ads.DiscoveryRequest,
) (
    *ads.DiscoveryResponse,
    error,
) {

    log.Info("fetch endpoints invoked!")

    return nil, nil
}
