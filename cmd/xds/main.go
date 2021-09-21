package main

import (
    "flag"

    log "github.com/liuliqiang/log4go"

    xds_server "github.com/liuliqiang/xds-demo/interface/grpc/xds-server"
)

var (
    host      = "0.0.0.0"
    port      = 1234
    ads       = false
    source    = "static" // static/docker
    adsOption = &xds_server.ADSOptions{
        BackendHostName: "backend_host_name",
        ListenerName:    "listener_0",
        RouteConfigName: "route_config_name",
        ClusterName:     "cluster_name",
        VirtualHostName: "virtual_host_name",
    }
)

func main() {
    flag.StringVar(&host, "host", host, "ip address to run xds server")
    flag.IntVar(&port, "port", port, "port to run xds server")
    flag.BoolVar(&ads, "ads", ads, "is it ads server")
    flag.StringVar(&source, "source", source, "static/docker")
    flag.StringVar(
        &adsOption.ListenerName,
        "ads.listener.name",
        adsOption.ListenerName,
        "ads listener name'",
    )
    flag.Parse()

    opts := &xds_server.Opts{
        Host:      host,
        Port:      port,
        AdsOption: adsOption,
    }

    if !ads {
        log.Info("Ready to serve xds server at: " + opts.Addr())
        opts.AdsOption = nil
    } else {
        log.Info("Ready to serve ads server at: " + opts.Addr())
    }

    if err := xds_server.NewXDSServer(opts).Run(); err != nil {
        log.Error("run xds server")
    }

    log.Info("I will exit now.")
}
