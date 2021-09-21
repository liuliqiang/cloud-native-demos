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

http-server: setup
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o .build/main cmd/http-server/*.go

client: setup
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o .build/client cmd/client/*.go

xds:
	mkdir -p .build
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o .build/xds cmd/xds/*.go

docker: server client
	docker build --tag lukelau/demo:0.0.1 -f ./build/docker/Dockerfile .

docker-http: http-server
	docker build --tag lukelau/http-server:v0.0.3 -f ./build/docker/Dockerfile-common .

up:
	docker-compose -f deploy/compose/22-primary-secondary.yml up -d

down:
	docker-compose -f deploy/compose/22-primary-secondary.yml down

run-xds-grpc:
	GRPC_GO_LOG_VERBOSITY_LEVEL=100 \
	GRPC_GO_LOG_SEVERITY_LEVEL=WARN \
	GRPC_XDS_BOOTSTRAP=$(shell pwd)/deploy/grpc/bootstrap.json \
	go run cmd/xds-grpc-client/main.go

run-http-server:
	go run cmd/http-server/*.go \
	    -static $(shell pwd)/interface/http/assets \
	    -store /tmp
