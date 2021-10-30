package container

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/liuliqiang/cloud-native-demo/internal/model"
)

type ContainerOper interface {
	ListInstanceWithLabel(key, value string) ([]model.Instance, error)
}

func NewDockerOper(host string) ContainerOper {
	oper := &dockerOper{
		currHost: host,
	}

	return oper
}

type dockerOper struct {
	currHost string
}

func (o *dockerOper) ListInstanceWithLabel(key, value string) ([]model.Instance, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	listFilters := types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "label",
				Value: key + "=" + value,
			},
		),
	}
	containers, err := cli.ContainerList(ctx, listFilters)
	if err != nil {
		panic(err)
	}

	var rst []model.Instance
	for _, container := range containers {
		rst = append(rst, model.Instance{
			Name:   container.Names[0],
			Health: container.Status == "running",
			Host:   o.currHost,
			Count:  0,
		})
	}

	return rst, nil
}
