PKG = github.com/loomnetwork/weave-blueprint
PROTOC = protoc --plugin=./protoc-gen-gogo -Ivendor -Isrc -I/usr/local/include

.PHONY: all clean test lint deps proto

all: blueprint-plugin helloworld-plugin
cli: blueprint-cli

blueprint-plugin: build/contracts/blueprint.0.0.1
helloworld-plugin: build/contracts/helloworld.1.0.0

build/contracts/blueprint.0.0.1: proto
	mkdir -p build/contracts
	go build -o $@ src/blueprint/plugin.go

build/contracts/helloworld.1.0.0: proto
	mkdir -p build/contracts
	go build -o $@ src/helloworld/plugin.go

protoc-gen-gogo:
	go build github.com/gogo/protobuf/protoc-gen-gogo

blueprint-cli:
	go build -o build/blueprint src/blueprint/cli/main.go

%.pb.go: %.proto protoc-gen-gogo
	$(PROTOC) --gogo_out=src $<

proto: \
	src/blueprint/types/types.pb.go \
	src/helloworld/types/types.pb.go

test: proto
	go test $(PKG)/...

lint:
	golint ./...

deps:
	go get \
		github.com/gogo/protobuf/jsonpb \
		github.com/gogo/protobuf/proto \
		github.com/spf13/cobra \
		github.com/gomodule/redigo/redis \
		github.com/loomnetwork/go-loom \
		github.com/hashicorp/go-plugin \
		github.com/pkg/errors \
		github.com/grpc-ecosystem/go-grpc-prometheus \
		github.com/go-kit/kit/log

clean:
	go clean
	rm -f \
		protoc-gen-gogo \
		src/blueprint/types/types.pb.go \
		src/helloworld/types/types.pb.go

