export GO111MODULE=on
export image_name=lukelau/basic_grpc_client-server
export version=v0.0.1
export docker_image=$(image_name):$(version)

ifndef GOOS
	GOOS=linux
endif
ifndef GOARCH
	GOARCH=amd64
endif

.PHONY: protoc clean

clean:
	rm -rf .build

protoc:
	@protoc -I interface/grpc/proto echo.proto --go_out=plugins=grpc:.

setup:
	mkdir -p .build

server: setup
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -gcflags="all=-N -l" -o .build/server cmd/server/*.go

client: setup
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o .build/client cmd/client/*.go

docker:
	docker build --tag $(docker_image) -f ./build/docker/Dockerfile .

up:
	docker run --rm \
		-p 9000:9000 \
		-p 2345:2345 \
		--name grpc-server -d \
		$(docker_image)

down:
	docker stop grpc-server

run-client:
	go run cmd/client/main.go
