package xds_test

import (
	"context"
	"testing"
	"time"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	ads "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/grpc"
)

func TestNewXDSResolver(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "127.0.0.1:1234", grpc.WithInsecure())
	if err != nil {
		t.Errorf("connection context error: %v", err)
	}

	// simplify just get clusters info
	adsCli := ads.NewAggregatedDiscoveryServiceClient(conn)
	adsConnection, err := adsCli.StreamAggregatedResources(context.Background())
	if err != nil {
		t.Errorf("create bidirec error: %v", err)
	}

	err = adsConnection.Send(&ads.DiscoveryRequest{
		VersionInfo: "",
		Node: &envoy_config_core_v3.Node{
			Id: "node-id",
		},
		ResourceNames: []string{"listener_name"},
		TypeUrl:       resource.ListenerType,
	})
	if err != nil {
		t.Errorf("send request error: %v", err)
	}
}
