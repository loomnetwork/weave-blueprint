PKG = github.com/loomnetwork/weave-blueprint
PROTOC = protoc --plugin=./protoc-gen-gogo -Isrc -I/usr/local/include
PROTOBUF_VERSION = 3.5.1
UNAME_S := $(shell uname -s)
CURRENT_DIRECTORY = $(shell pwd)

ifeq ($(UNAME_S),Linux)
	PLATFORM = linux
endif
ifeq ($(UNAME_S),Darwin)
	PLATFORM = osx
	BREW = $(shell which brew)
endif

export GOPATH=$(CURRENT_DIRECTORY)/tmpgopath:$(CURRENT_DIRECTORY)

.PHONY: all clean test lint deps proto

all: contracts cli

contracts: build/contracts/blueprint.0.0.1

build/contracts/blueprint.0.0.1: proto
	mkdir -p build/contracts
	go build -o $@ src/blueprint.go

protoc-gen-gogo:
	go build github.com/gogo/protobuf/protoc-gen-gogo

cli:
	go build -o build/blueprint src/cli/main.go

%.pb.go: %.proto protoc-gen-gogo
	$(PROTOC) --gogo_out=src $<

proto: src/types/types.proto src/types/types.pb.go

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
		src/types/types.pb.go \
		testdata/test.pb.go \

protobuf-install:
ifeq ($(BREW),)
	curl -OL https://github.com/google/protobuf/releases/download/v$(PROTOBUF_VERSION)/protoc-${PROTOBUF_VERSION}-$(PLATFORM)-x86_64.zip \
	&& sudo unzip protoc-$(PROTOBUF_VERSION)-$(PLATFORM)-x86_64.zip -d /usr/local && sudo chmod 755 /usr/local/bin/protoc \
	&& sudo find /usr/local/include/google -type d -exec chmod 755 -- {} + && sudo find /usr/local/include/google -type f -exec chmod 644 -- {} + \
	&& rm protoc-$(PROTOBUF_VERSION)-$(PLATFORM)-x86_64.zip
else
	$(BREW) install protobuf
endif
