package xds_server

import (
    "fmt"
    "net"

    cds "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
    ads "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
    eds "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
    lds "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
    rds "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
    log "github.com/liuliqiang/log4go"
    "google.golang.org/grpc"
)

const (
    ConcurrentStreamCount = 10
    ListenProtocol        = "tcp"
    AddressFormat         = "%s:%d"
)

func NewXDSServer(opts *Opts) (s *xdsServer) {
    s = &xdsServer{
        opts: opts,
        ads:  newAggregatedDiscoveryService(opts.AdsOption),
        lds:  newListenerDiscoveryService(),
        cds:  newClusterDiscoveryService(),
        rds:  newRouteDiscoveryService(),
        eds:  newEndpointDiscoveryService(),
    }

    return s
}

func (x *xdsServer) Run() (err error) {
    var grpcOptions []grpc.ServerOption
    grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(ConcurrentStreamCount))
    grpcServer := grpc.NewServer(grpcOptions...)

    lis, err := net.Listen(ListenProtocol, x.opts.Addr())
    if err != nil {
        log.Error("Addr: " + x.opts.Addr())
    }

    if x.opts.AdsOption != nil {
        ads.RegisterAggregatedDiscoveryServiceServer(grpcServer, x.ads)
    } else {
        lds.RegisterListenerDiscoveryServiceServer(grpcServer, x.lds)
        rds.RegisterRouteDiscoveryServiceServer(grpcServer, x.rds)
        cds.RegisterClusterDiscoveryServiceServer(grpcServer, x.cds)
        eds.RegisterEndpointDiscoveryServiceServer(grpcServer, x.eds)
    }

    if err = grpcServer.Serve(lis); err != nil {
        log.Error("serve")
    }

    return nil
}

type Opts struct {
    Host      string
    Port      int
    AdsOption *ADSOptions
}

func (o *Opts) Addr() string {
    return fmt.Sprintf(AddressFormat, o.Host, o.Port)
}

type xdsServer struct {
    opts *Opts
    ads  ads.AggregatedDiscoveryServiceServer
    lds  lds.ListenerDiscoveryServiceServer
    cds  cds.ClusterDiscoveryServiceServer
    rds  rds.RouteDiscoveryServiceServer
    eds  eds.EndpointDiscoveryServiceServer
}
