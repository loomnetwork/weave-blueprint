PKG = github.com/loomnetwork/weave-blueprint
PROTOC = protoc --plugin=./protoc-gen-gogo -Isrc -I/usr/local/include
PROTOBUF_VERSION = 3.5.1
UNAME_S := $(shell uname -s)
CURRENT_DIRECTORY = $(shell pwd)
GETH_GIT_REV = f9c06695672d0be294447272e822db164739da67


ifeq ($(UNAME_S),Linux)
	PLATFORM = linux
endif
ifeq ($(UNAME_S),Darwin)
	PLATFORM = osx
	BREW = $(shell which brew)
endif

.PHONY: all clean test lint deps proto

all: contracts cli

export GOPATH=$(CURRENT_DIRECTORY)/tmpgopath:$(CURRENT_DIRECTORY)
HASHICORP_DIR = $(CURRENT_DIRECTORY)/tmpgopath/src/github.com/hashicorp/go-plugin
GO_ETHEREUM_DIR = $(CURRENT_DIRECTORY)/tmpgopath/src/github.com/ethereum/go-ethereum
SSHA3_DIR = $(CURRENT_DIRECTORY)/tmpgopath/src/github.com/miguelmota/go-solidity-sha3

ETHEREUM_GIT_REV = f9c06695672d0be294447272e822db164739da67

$(GO_ETHEREUM_DIR):
	git clone -q https://github.com/loomnetwork/go-ethereum.git $@

$(SSHA3_DIR):
	git clone -q https://github.com/loomnetwork/go-solidity-sha3.git $@

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

deps: $(GO_ETHEREUM_DIR) $(SSHA3_DIR)
	go get \
		golang.org/x/crypto/ripemd160 \
		golang.org/x/crypto/sha3 \
		github.com/gogo/protobuf/jsonpb \
		github.com/gogo/protobuf/proto \
		github.com/gorilla/websocket \
		github.com/phonkee/go-pubsub \
		google.golang.org/grpc \
		github.com/spf13/cobra \
		github.com/hashicorp/go-plugin \
		github.com/stretchr/testify/assert \
		github.com/go-kit/kit/log \
		github.com/pkg/errors \
		github.com/loomnetwork/go-loom \
		github.com/grpc-ecosystem/go-grpc-prometheus \
		github.com/go-kit/kit/log \
		github.com/loomnetwork/yubihsm-go \
		gopkg.in/check.v1
	cd $(GO_ETHEREUM_DIR) && git checkout master && git pull && git checkout $(ETHEREUM_GIT_REV)
	cd $(HASHICORP_DIR) && git checkout f4c3476bd38585f9ec669d10ed1686abd52b9961


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
