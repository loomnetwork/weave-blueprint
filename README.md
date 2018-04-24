# weave-blueprint

Sample Loom Blockchain project *GO Language*

To Build
```
export GOPATH=$GOPATH:`pwd`

go build -buildmode=plugin -o out/cmds/create-tx.so src/cmd-plugins/create-tx/main.go
```

To rebuild the protobuf files
```
go build github.com/gogo/protobuf/protoc-gen-gogo
