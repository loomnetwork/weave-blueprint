package main

import (
	"github.com/loomnetwork/go-loom/plugin"
	"github.com/loomnetwork/weave-blueprint/src/blueprint"
)

var Contract = blueprint.Contract

func main() {
	plugin.Serve(Contract)
}
