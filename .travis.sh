#!/bin/bash

set -euxo pipefail

# Get loom and put into PATH
curl -OL https://storage.googleapis.com/private.delegatecall.com/loom/linux/build-${BUILD_NUMBER}/loom
chmod +x loom
sudo mv loom /usr/local/bin/loom

export GOPATH=$GOPATH:`pwd`

# Run the build
make deps
make
make cli

cd build

loom init

cp ../genesis.example.json genesis.json

loom run > run.log 2>&1 &

sleep 10

# Run sample transactions to test

cd ..

loom genkey -k priv_key -a pub_key
./build/blueprint create-acct -k priv_key
./build/blueprint set -v 1 -k priv_key
./build/blueprint get

pkill -f loom

cat build/run.log
