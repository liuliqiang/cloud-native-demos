export GO111MODULE=on

ifndef GOOS
	GOOS=linux
endif
ifndef GOARCH
	GOARCH=amd64
endif

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

run-http-server:
	go run cmd/http-server/*.go \
	    -static $(shell pwd)/interface/http/assets \
	    -store /tmp
