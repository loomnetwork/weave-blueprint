# weave-blueprint [![Build Status](https://travis-ci.org/loomnetwork/weave-blueprint.svg?branch=master)](https://travis-ci.org/loomnetwork/weave-blueprint)

Sample Loom Blockchain project *GO Language*

For more info please checkout the docs page for the [Loom SDK](https://loomx.io/developers/docs/en/prereqs.html)

To Build
```
export GOPATH=$GOPATH:`pwd`

make deps
make
```


To Run (Requires Loom Dappchain engine binary)
```
cd build
./loom init
cp ../genesis.example.json genesis.json
./loom run
```
