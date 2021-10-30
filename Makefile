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

http-server: setup
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o .build/main cmd/http-server/*.go

docker-http: http-server
	docker build --tag lukelau/http-server:v0.0.3 -f ./build/docker/Dockerfile-common .

up:
	docker-compose -f deploy/compose/docker-compose.yml up -d

down:
	docker-compose -f deploy/compose/docker-compose.yml down
