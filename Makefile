export GO111MODULE=on

ifndef GOOS
	GOOS=linux
endif
ifndef GOARCH
	GOARCH=amd64
endif

.PHONY: protoc
protoc:
	@protoc -I interface/grpc/proto echo.proto --go_out=plugins=grpc:.

setup:
	mkdir -p .build

server: setup
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o .build/server cmd/server/*.go

xds: setup
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o .build/xds cmd/xds/*.go

docker: server xds
	docker build --tag lukelau/demo:v0.0.1 -f ./build/docker/Dockerfile .

docker-http: http-server
	docker build --tag lukelau/http-server:v0.0.3 -f ./build/docker/Dockerfile-common .

up:
	docker-compose -f deploy/compose/21-grpc-servers.yml up -d

down:
	docker-compose -f deploy/compose/21-grpc-servers.yml down

run-ads:
	go run cmd/xds/main.go \
		-host 127.0.0.1 \
		-ads

run-xds-grpc:
	GRPC_GO_LOG_VERBOSITY_LEVEL=100 \
	GRPC_GO_LOG_SEVERITY_LEVEL=WARN \
	GRPC_XDS_BOOTSTRAP=$(shell pwd)/deploy/grpc/bootstrap.json \
	go run cmd/xds-grpc-client/main.go
