# weave-blueprint [![Build Status](https://travis-ci.org/loomnetwork/weave-blueprint.svg?branch=master)](https://travis-ci.org/loomnetwork/weave-blueprint)

Sample Loom Blockchain project *GO Language*

For more info please checkout the docs page for the [Loom SDK](https://loomx.io/developers/docs/en/prereqs.html)

## To Build
```bash
export GOPATH=$GOPATH:`pwd`

make deps
make
```


## To Run (Requires Loom Dappchain engine binary)
```bash
cd build
./loom init
cp ../genesis.example.json genesis.json
./loom run
```

After running loom, open new terminal tab and and blueprint as follows.

## Generate private key
```bash
cd build
./blueprint genkey 
```
this will generate private key file named "key" to be further

## Create user account
```bash
./blueprint call create-acct -u <account-name> -p key
```

## Set value for user
```bash
./blueprint call create-acct -u <account-name> -v <any-integer> -p key
```

## Get value for user
```bash
./blueprint call create-acct -u <account-name> -p key
```
