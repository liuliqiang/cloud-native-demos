package xds

import (
	"encoding/json"
	"fmt"
)

func NewXDSOptsWithJsonBytes(bytes []byte) (*xDSOpts, error) {
	if len(bytes) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	opts := &xDSOpts{}
	if err := json.Unmarshal(bytes, opts); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return opts, nil
}

type xDSOpts struct {
	XDSServers []*xDSServerOpts `json:"xds_servers"`
	Node       *xDSNodeOpts     `json:"node"`
}

type xDSServerOpts struct {
	ServerURI      string   `json:"server_uri"`
	ServerFeatures []string `json:"server_features"`
}

type xDSNodeOpts struct {
	ID       string            `json:"id"`
	MetaData map[string]string `json:"meta_data"`
	Locality map[string]string `json:"locality"`
}
