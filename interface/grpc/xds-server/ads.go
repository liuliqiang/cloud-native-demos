package xds_server

import (
    "fmt"
    "sync"

    cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
    core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
    endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
    model "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
    route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
    httpv3mgt "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
    discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
    "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
    "github.com/golang/protobuf/ptypes"
    "github.com/golang/protobuf/ptypes/any"
    "github.com/google/uuid"
    log "github.com/liuliqiang/log4go"
    "google.golang.org/protobuf/types/known/anypb"
    "google.golang.org/protobuf/types/known/wrapperspb"
)

var (
    _ discovery.AggregatedDiscoveryServiceServer = (*aggregatedDiscoveryService)(nil)
)

type ADSOptions struct {
    BackendHostName string
    ListenerName    string
    RouteConfigName string
    ClusterName     string
    VirtualHostName string
}

type aggregatedDiscoveryService struct {
    signal chan struct{}
    mu     sync.Mutex
    opts   *ADSOptions
}

func newAggregatedDiscoveryService(opts *ADSOptions) (s *aggregatedDiscoveryService) {
    return &aggregatedDiscoveryService{
        opts: opts,
    }
}

func (s *aggregatedDiscoveryService) StreamAggregatedResources(
    stream discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer,
) error {
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

        var respResource *any.Any
        switch {
        case req.TypeUrl == resource.ListenerType:
            respResource = s.newListenerResp()
        case req.TypeUrl == resource.RouteType:
            respResource = s.newRouterResp()
        case req.TypeUrl == resource.ClusterType:
            respResource = s.newClusterResp()
        case req.TypeUrl == resource.EndpointType:
            respResource = s.newEndpointResp()
        }

        resp := &discovery.DiscoveryResponse{
            VersionInfo: "1111",
            Resources: []*any.Any{
                respResource,
            },
            TypeUrl: req.TypeUrl,
            Nonce:   uuid.New().String(),
        }
        stream.Send(resp)
    }
}

func (s *aggregatedDiscoveryService) newListenerResp() *any.Any {
    hcRds := &httpv3mgt.HttpConnectionManager_Rds{
        Rds: &httpv3mgt.Rds{
            RouteConfigName: s.opts.RouteConfigName,
            ConfigSource: &core.ConfigSource{
                ConfigSourceSpecifier: &core.ConfigSource_Ads{
                    Ads: &core.AggregatedConfigSource{},
                },
            },
        },
    }

    manager := &httpv3mgt.HttpConnectionManager{
        CodecType:      httpv3mgt.HttpConnectionManager_AUTO,
        RouteSpecifier: hcRds,
    }

    pbst, err := ptypes.MarshalAny(manager)
    if err != nil {
        panic(err)
    }

    listener := &model.Listener{
        Name: s.opts.ListenerName,
        ApiListener: &model.ApiListener{
            ApiListener: pbst,
        },
    }
    anyL, _ := anypb.New(listener)

    return anyL
}

func (s *aggregatedDiscoveryService) newRouterResp() *any.Any {
    vh := &route.VirtualHost{
        Name:    s.opts.VirtualHostName,
        Domains: []string{s.opts.ListenerName}, //******************* >> must match what is specified at xds:/// //

        Routes: []*route.Route{{
            Match: &route.RouteMatch{
                PathSpecifier: &route.RouteMatch_Prefix{
                    Prefix: "",
                },
            },
            Action: &route.Route_Route{
                Route: &route.RouteAction{
                    ClusterSpecifier: &route.RouteAction_Cluster{
                        Cluster: s.opts.ClusterName,
                    },
                },
            },
        }}}

    route := &route.RouteConfiguration{
        Name:         s.opts.RouteConfigName,
        VirtualHosts: []*route.VirtualHost{vh},
    }
    anyR, _ := anypb.New(route)

    return anyR
}

func (s *aggregatedDiscoveryService) newClusterResp() *any.Any {
    log.Info(">>>>>>>>>>>>>>>>>>> creating CLUSTER ")
    cls := &cluster.Cluster{
        Name:                 s.opts.ClusterName,
        LbPolicy:             cluster.Cluster_ROUND_ROBIN,
        ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
        EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
            EdsConfig: &core.ConfigSource{
                ConfigSourceSpecifier: &core.ConfigSource_Ads{},
            },
        },
    }

    anyC, _ := anypb.New(cls)

    return anyC
}

func (s *aggregatedDiscoveryService) newEndpointResp() *any.Any {
    log.Info(">>>>>>>>>>>>>>>>>>> creating ENDPOINT for remoteHost")

    eds := &endpoint.ClusterLoadAssignment{
        ClusterName: s.opts.ClusterName,
        Endpoints: []*endpoint.LocalityLbEndpoints{{
            Locality: &core.Locality{
                Region: "us-central1",
                Zone:   "us-central1-a",
            },
            Priority:            0,
            LoadBalancingWeight: &wrapperspb.UInt32Value{Value: uint32(1000)},
            LbEndpoints: []*endpoint.LbEndpoint{
                s.createLbEndpoint("127.0.0.1", 3000),
                s.createLbEndpoint("127.0.0.1", 5000),
                s.createLbEndpoint("127.0.0.1", 7000),
            },
        }},
    }
    anyE, _ := anypb.New(eds)

    return anyE
}

func (s *aggregatedDiscoveryService) createLbEndpoint(address string, port int) *endpoint.LbEndpoint {
    hst := &core.Address{Address: &core.Address_SocketAddress{
        SocketAddress: &core.SocketAddress{
            Address:  address,
            Protocol: core.SocketAddress_TCP,
            PortSpecifier: &core.SocketAddress_PortValue{
                PortValue: uint32(port),
            },
        },
    }}

    return &endpoint.LbEndpoint{
        HostIdentifier: &endpoint.LbEndpoint_Endpoint{
            Endpoint: &endpoint.Endpoint{
                Address: hst,
            }},
        HealthStatus: core.HealthStatus_HEALTHY,
    }
}

func (s *aggregatedDiscoveryService) DeltaAggregatedResources(
    stream discovery.AggregatedDiscoveryService_DeltaAggregatedResourcesServer,
) error {
    panic("implement me")
}
