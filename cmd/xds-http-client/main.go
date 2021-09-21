package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/liuliqiang/log4go"
	"github.com/liuliqiang/xds-demo/internal/resolver"
	"github.com/liuliqiang/xds-demo/internal/resolver/xds"
)

var (
	_                http.RoundTripper = (*customTransport)(nil)
	defaultTransport                   = &http.Transport{}
	xdsBootstrap                       = "~/.grpc/bootstrap"
)

func main() {
	flag.StringVar(&xdsBootstrap, "xds.bootstrap", xdsBootstrap, "xds bootstrap file path")
	flag.Parse()

	t := &http.Transport{}
	xdsTransport, err := NewCustomTransport(xdsBootstrap)
	if err != nil {
		panic(fmt.Errorf("new cusomer transport: %w", err))
	}

	t.RegisterProtocol("xds", xdsTransport)
	t.RegisterProtocol("xdss", xdsTransport)
	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: t,
	}

	resp, err := httpClient.Get("xds:///cluster_name/index")
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	println(string(data))
}

type customTransport struct {
	rawTrans    http.RoundTripper
	xdsResolver resolver.Resolver
}

func NewCustomTransport(configPaths ...string) (*customTransport, error) {
	var configPath string
	if len(configPaths) > 0 {
		configPath = configPaths[0]
	}

	rsv, err := xds.NewXDSResolver(configPath)
	if err != nil {
		return nil, fmt.Errorf("init xds xdsResolver: %w", err)
	}

	return &customTransport{
		xdsResolver: rsv,
	}, nil
}

func (c *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme == "xds" {
		req.URL.Scheme = "http"
	} else {
		req.URL.Scheme = "https"
	}

	if req.URL.Path[0] != '/' {
		return nil, errors.New("Invalid host: " + req.URL.Path)
	}
	endIndex := len(req.URL.Path)
	for i := 1; i < endIndex; i++ {
		if req.URL.Path[i] == '/' {
			endIndex = i
			break
		}
	}
	log.Info("begin resolve")
	hosts, err := c.xdsResolver.Resolve(req.URL.Path[1:endIndex])
	if err != nil {
		log.Error("resolved fail")
		return nil, fmt.Errorf("resolve host: %w", err)
	}

	if len(hosts) == 0 {
		log.Error("resolve empty hosts")
		return nil, errors.New("host not found")
	}

	log.Info("resolve success")
	req.URL.Host = hosts[0]
	return defaultTransport.RoundTrip(req)
}

type xdsResolver struct {
}
