package xds_server

import (
    "context"
    "fmt"

    model "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
    v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
    envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
    httpv3mgt "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
    cds "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
    ads "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
    "github.com/golang/protobuf/ptypes/any"
    log "github.com/liuliqiang/log4go"
    "google.golang.org/grpc/peer"
    "google.golang.org/protobuf/types/known/anypb"
    "google.golang.org/protobuf/types/known/durationpb"
)

var (
    _ cds.ClusterDiscoveryServiceServer = (*clusterDiscoveryService)(nil)
)

func newClusterDiscoveryService() (s cds.ClusterDiscoveryServiceServer) {
    return &clusterDiscoveryService{}
}

type clusterDiscoveryService struct {
}

func (c *clusterDiscoveryService) StreamClusters(stream cds.ClusterDiscoveryService_StreamClustersServer) error {
    log.Info("stream clusters invoke!")

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
            log.Info("node: <null>")
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

        cluster := newCluster()
        clusterResource, _ := anypb.New(cluster)
        resp := &ads.DiscoveryResponse{
            VersionInfo: "2222",
            Resources: []*any.Any{
                clusterResource,
            },
            TypeUrl: req.TypeUrl,
            Nonce:   "123dfasjdlfjas",
        }
        _ = stream.Send(resp)
    }
}

func newCluster() *model.Cluster {
    return &model.Cluster{
        Name: "some_service",
        ConnectTimeout: &durationpb.Duration{
            Seconds: 1,
        },
        ClusterDiscoveryType: &model.Cluster_Type{
            Type: model.Cluster_STATIC,
        },
        LbPolicy: model.Cluster_ROUND_ROBIN,
        TypedExtensionProtocolOptions: map[string]*any.Any{
            "envoy.extensions.upstreams.http.v3.HttpProtocolOptions": newHttp2Option(),
        },
        LoadAssignment: &envoy_config_endpoint_v3.ClusterLoadAssignment{
            ClusterName: "some_service",
            Endpoints: []*envoy_config_endpoint_v3.LocalityLbEndpoints{
                {
                    LbEndpoints: []*envoy_config_endpoint_v3.LbEndpoint{
                        newLBEndpoint("127.0.0.1", 3000),
                        newLBEndpoint("127.0.0.1", 5000),
                        newLBEndpoint("127.0.0.1", 7000),
                    },
                },
            },
        },
    }
}

func newHttp2Option() *any.Any {
    opts := &httpv3mgt.HttpProtocolOptions{
        UpstreamProtocolOptions: &httpv3mgt.HttpProtocolOptions_ExplicitHttpConfig_{
            ExplicitHttpConfig: &httpv3mgt.HttpProtocolOptions_ExplicitHttpConfig{
                ProtocolConfig: &httpv3mgt.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
                    Http2ProtocolOptions: &v3.Http2ProtocolOptions{},
                },
            },
        },
    }

    rtn, _ := anypb.New(opts)
    return rtn
}

func newLBEndpoint(ip string, port int) *envoy_config_endpoint_v3.LbEndpoint {
    return &envoy_config_endpoint_v3.LbEndpoint{
        HostIdentifier: &envoy_config_endpoint_v3.LbEndpoint_Endpoint{
            Endpoint: &envoy_config_endpoint_v3.Endpoint{
                Address: &v3.Address{
                    Address: &v3.Address_SocketAddress{
                        SocketAddress: &v3.SocketAddress{
                            Address: ip,
                            PortSpecifier: &v3.SocketAddress_PortValue{
                                PortValue: uint32(port),
                            },
                        },
                    },
                },
            },
        },
    }
}

func (c *clusterDiscoveryService) DeltaClusters(server cds.ClusterDiscoveryService_DeltaClustersServer) error {
    log.Info("delta clusters invoke!")

    return nil
}

func (c *clusterDiscoveryService) FetchClusters(
    ctx context.Context,
    request *ads.DiscoveryRequest,
) (
    *ads.DiscoveryResponse,
    error,
) {

    log.Info("fetch clusters invoke!")

    return nil, nil
}
