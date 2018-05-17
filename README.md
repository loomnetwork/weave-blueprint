# weave-blueprint [![Build Status](https://travis-ci.org/loomnetwork/weave-blueprint.svg?branch=master)](https://travis-ci.org/loomnetwork/weave-blueprint)

Sample Loom Blockchain project *GO Language*

To Build
```
export GOPATH=$GOPATH:`pwd`

make deps
make
```


To Run (Requires Loom Dappchain engine binary)
```
cd run
cp ../gensis.example.json .
./loom init
./loom run
```
