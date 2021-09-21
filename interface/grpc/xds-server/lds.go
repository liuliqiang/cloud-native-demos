package xds_server

import (
    "context"
    "fmt"

    envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
    model "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
    envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
    httpv3mgt "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
    envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
    lds "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
    "github.com/golang/protobuf/ptypes/any"
    "github.com/google/uuid"
    log "github.com/liuliqiang/log4go"
    "google.golang.org/grpc/peer"
    "google.golang.org/protobuf/types/known/anypb"
)

var (
    _ lds.ListenerDiscoveryServiceServer = (*listenerDiscoveryService)(nil)
)

func newListenerDiscoveryService() (s lds.ListenerDiscoveryServiceServer) {
    return &listenerDiscoveryService{}
}

type listenerDiscoveryService struct {
}

func (l *listenerDiscoveryService) DeltaListeners(stream lds.ListenerDiscoveryService_DeltaListenersServer) error {
    log.Info("delta listeners invoked")

    return nil
}

func (l *listenerDiscoveryService) StreamListeners(stream lds.ListenerDiscoveryService_StreamListenersServer) error {
    log.Info("stream listeners invoked")

    clientPeer, ok := peer.FromContext(stream.Context())
    clientIP := "unknown"
    if ok {
        clientIP = clientPeer.Addr.String()
    }

    log.Info("client ip " + clientIP)

    for {
        req, err := stream.Recv()
        if err != nil {
            log.Info("recv")
            return err
        }

        if req.Node != nil {
            log.Info(fmt.Sprintf("node: %v", req.Node.Id))
        } else {
            log.Info("node: <nil>")
        }
        log.Info(fmt.Sprintf("resource name: %v", req.ResourceNames))
        log.Info(fmt.Sprintf("resource nonce: %v", req.ResponseNonce))
        log.Info(fmt.Sprintf("type url: %v", req.TypeUrl))
        log.Info(fmt.Sprintf("version info: %v", req.VersionInfo))

        if req.ResponseNonce != "" {
            if req.ErrorDetail != nil {
                log.Error("Client resp NACK: " + req.ErrorDetail.Message)
            }
            continue
        }
        l := &model.Listener{
            Name: "listener_0",
            Address: &envoy_config_core_v3.Address{
                Address: &envoy_config_core_v3.Address_SocketAddress{
                    SocketAddress: &envoy_config_core_v3.SocketAddress{
                        Address: "0.0.0.0",
                        PortSpecifier: &envoy_config_core_v3.SocketAddress_PortValue{
                            PortValue: 10000,
                        },
                    },
                },
            },
            FilterChains: newFilterChain(),
        }
        anyL, _ := anypb.New(l)
        resp := &envoy_service_discovery_v3.DiscoveryResponse{
            VersionInfo: "1111",
            Resources: []*any.Any{
                anyL,
            },
            TypeUrl: req.TypeUrl,
            Nonce:   uuid.New().String(),
        }
        stream.Send(resp)
    }
}

func newFilterChain() []*model.FilterChain {
    return []*model.FilterChain{
        {
            Filters: []*model.Filter{
                {
                    Name: "envoy.filters.network.http_connection_manager",
                    ConfigType: &model.Filter_TypedConfig{
                        TypedConfig: newHttpFilter(),
                    },
                },
            },
        },
    }
}

func newHttpFilter() *any.Any {
    hcm := &httpv3mgt.HttpConnectionManager{
        StatPrefix: "ingress_http",
        CodecType:  httpv3mgt.HttpConnectionManager_AUTO,
        HttpFilters: []*httpv3mgt.HttpFilter{
            {
                Name: "envoy.filters.http.router",
            },
        },
        RouteSpecifier: &httpv3mgt.HttpConnectionManager_RouteConfig{
            RouteConfig: &envoy_config_route_v3.RouteConfiguration{
                Name: "local_route",
                VirtualHosts: []*envoy_config_route_v3.VirtualHost{
                    {
                        Name:    "local_service",
                        Domains: []string{"*"},
                        Routes: []*envoy_config_route_v3.Route{
                            {
                                Match: &envoy_config_route_v3.RouteMatch{
                                    PathSpecifier: &envoy_config_route_v3.RouteMatch_Prefix{
                                        Prefix: "/",
                                    },
                                    Grpc: &envoy_config_route_v3.RouteMatch_GrpcRouteMatchOptions{},
                                },
                                Action: &envoy_config_route_v3.Route_Route{
                                    Route: &envoy_config_route_v3.RouteAction{
                                        ClusterSpecifier: &envoy_config_route_v3.RouteAction_Cluster{
                                            Cluster: "some_service",
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    rtn, _ := anypb.New(hcm)
    return rtn
}

func IsValidSubscriber(req *envoy_service_discovery_v3.DiscoveryRequest) bool {
    return (len(req.Node.Cluster) > 0) && (len(req.Node.Id) > 0)
}

func newListenerInfo() {

}

func (l *listenerDiscoveryService) FetchListeners(
    ctx context.Context,
    request *envoy_service_discovery_v3.DiscoveryRequest,
) (
    *envoy_service_discovery_v3.DiscoveryResponse,
    error,
) {

    log.Info("fetch listeners invoked")

    return nil, nil
}
