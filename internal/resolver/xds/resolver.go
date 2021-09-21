package xds

import (
    "context"
    "errors"
    "fmt"
    "io/ioutil"
    "os"
    "sync"
    "time"

    envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
    endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
    ads "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
    "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
    itf "github.com/liuliqiang/xds-demo/internal/resolver"
    log "github.com/liuliqiang/log4go"
    "google.golang.org/grpc"
    "google.golang.org/protobuf/proto"
    "google.golang.org/protobuf/types/known/anypb"
)

var (
    _ itf.Resolver = (*resolver)(nil)
)

func NewXDSResolver(configPath string) (itf.Resolver, error) {
    opts, err := loadBootstrap(configPath)
    if err != nil {
        return nil, fmt.Errorf("load bootstrap file: %w", err)
    }

    r := &resolver{
        opts:               opts,
        waitServiceChannel: make(chan string),
        resolveTimeout:     time.Second * 5,
    }

    if err = r.initAdsConnection(); err != nil {
        return nil, fmt.Errorf("init xds client: %w", err)
    }

    go r.resolveProcess()
    go r.recvProcess()

    return r, nil
}

func loadBootstrap(bsPath string) (*xDSOpts, error) {
    if bsPath == "" {
        if bsPath = os.Getenv("HTTP_XDS_BOOTSTRAP"); bsPath == "" {
            return nil, errors.New("bootstrap file not found")
        }
    }

    bsBytes, err := ioutil.ReadFile(bsPath)
    if err != nil {
        return nil, fmt.Errorf("read bootstrap file: %w", err)
    }

    opts, err := NewXDSOptsWithJsonBytes(bsBytes)
    if err != nil {
        return nil, fmt.Errorf("parse bootstrap: %w", err)
    }

    return opts, nil
}

func (r *resolver) initAdsConnection() error {
    ctx, cancel := context.WithTimeout(context.Background(), r.resolveTimeout)
    defer cancel()

    log.Info("connect to server")
    conn, err := grpc.DialContext(ctx, r.opts.XDSServers[0].ServerURI, grpc.WithInsecure())
    if err != nil {
        return fmt.Errorf("connect to xds: %w", err)
    }

    // simplify just get clusters info
    adsCli := ads.NewAggregatedDiscoveryServiceClient(conn)
    r.adsConnection, err = adsCli.StreamAggregatedResources(context.Background())
    if err != nil {
        return fmt.Errorf("init xds client: %w", err)
    }

    return nil
}

func (r *resolver) Resolve(serviceName string) (endpoints []string, err error) {
    epsItf, exists := r.serviceEndpointsCache.Load(serviceName)
    if !exists {
        ch := make(chan bool)
        r.serviceEndpointsCache.Store(serviceName, ch)
        r.waitServiceChannel <- serviceName
        timeTick := time.NewTicker(r.resolveTimeout)
        select {
        case <-ch:
            epsItf, _ := r.serviceEndpointsCache.Load(serviceName)
            return epsItf.([]string), nil
        case <-timeTick.C:
            return nil, errors.New("resolve timeout")
        }
    } else {
        if value, ok := epsItf.([]string); ok {
            return value, nil
        }
        // what should be do when two goroutine process here?
        panic("I don't know what to do now.")
    }
}

func (r *resolver) resolveProcess() {
    for {
        serviceName := <-r.waitServiceChannel
        _, exists := r.serviceEndpointsCache.Load(serviceName)
        if !exists {
            log.Error("cache not exits")
            continue
        }

        err := r.adsConnection.Send(&ads.DiscoveryRequest{
            Node: &envoy_config_core_v3.Node{
                Id: r.opts.Node.ID,
            },
            ResourceNames: []string{serviceName},
            TypeUrl:       resource.EndpointType,
        })
        if err != nil {
            log.Error("query endpoints " + err.Error())
        }
    }
}

func (r *resolver) recvProcess() {
    for {
        resp, err := r.adsConnection.Recv()
        if err != nil {
            log.Error("failed to recv ads resp: " + err.Error())
            time.Sleep(time.Second / 20)
            continue
        }

        var result []string
        for _, res := range resp.Resources {
            var endpoints endpoint.ClusterLoadAssignment
            if err = anypb.UnmarshalTo(res, &endpoints, proto.UnmarshalOptions{}); err != nil {
                log.Error("Failed to unmarshal endpoints info")
                continue
            }
            for _, ep := range endpoints.Endpoints {
                for _, lep := range ep.LbEndpoints {
                    epAddr := lep.HostIdentifier.((*endpoint.LbEndpoint_Endpoint)).Endpoint.Address
                    addr := epAddr.Address.(*envoy_config_core_v3.Address_SocketAddress).SocketAddress.Address
                    port := epAddr.Address.(*envoy_config_core_v3.Address_SocketAddress).
                        SocketAddress.
                        PortSpecifier.(*envoy_config_core_v3.SocketAddress_PortValue).
                        PortValue
                    result = append(result, fmt.Sprintf("%s:%d", addr, port))
                }
            }

            serviceName := endpoints.ClusterName
            value, exists := r.serviceEndpointsCache.Load(serviceName)
            if !exists {
                log.Error("cache not exits")
                continue
            }
            ch, ok := value.(chan bool)
            if !ok {
                log.Error("cache not with channel")
                continue
            }
            r.serviceEndpointsCache.Store(serviceName, result)
            ch <- true
        }
    }
}

type resolver struct {
    opts                  *xDSOpts
    resolveTimeout        time.Duration
    serviceEndpointsCache sync.Map
    waitServiceChannel    chan string
    adsConnection         ads.AggregatedDiscoveryService_StreamAggregatedResourcesClient
}
