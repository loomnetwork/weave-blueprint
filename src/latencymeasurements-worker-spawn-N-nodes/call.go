package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	loom "github.com/loomnetwork/go-loom"
	"github.com/loomnetwork/go-loom/auth"
	"github.com/loomnetwork/go-loom/cli"
	"github.com/loomnetwork/go-loom/client"
	//"path/filepath"
	"path/filepath"
)

type TxConn struct {
	WriteURI     string
	ReadURI      string
	ContractAddr string
	ChainID      string
	PrivFile     string
	contract     *client.Contract
	signer       auth.Signer
}

func (n *TxConn) ResolveAddress(s string) (loom.Address, error) {
	rpcClient := client.NewDAppChainRPCClient(n.ChainID, n.WriteURI, n.ReadURI)
	contractAddr, err := cli.ParseAddress(s)
	if err != nil {
		// if address invalid, try to resolve it using registry
		contractAddr, err = rpcClient.Resolve(s)
		if err != nil {
			return loom.Address{}, err
		}
	}

	return contractAddr, nil
}

func (n *TxConn) InitContract(contractAddrStr string) error {
	if contractAddrStr == "" {
		return errors.New("contract address or name required")
	}

	fmt.Printf("n.ReadURI-%s\n", n.ReadURI)
	fmt.Printf("n.WriteURI-%s\n", n.WriteURI)

	fmt.Printf("trying to find %s\n", contractAddrStr)
	contractAddr, err := n.ResolveAddress(contractAddrStr)
	if err != nil {
		return err
	}

	if n.PrivFile == "" {
		return errors.New("private key required to call contract")
	}

	privKeyB64, err := ioutil.ReadFile(n.PrivFile)
	if err != nil {
		return err
	}

	privKey, err := base64.StdEncoding.DecodeString(string(privKeyB64))
	if err != nil {
		return err
	}

	n.signer = auth.NewEd25519Signer(privKey)

	// create rpc client
	rpcClient := client.NewDAppChainRPCClient(n.ChainID, n.WriteURI, n.ReadURI)
	// create contract
	n.contract = client.NewContract(rpcClient, contractAddr.Local)
	return nil
}

func NewtxConn(wurl, rurl, contract string) *TxConn {

	absPath, _ := filepath.Abs("weave-blueprint/build/key")

	t := &TxConn{
		WriteURI:     wurl,
		ReadURI:      rurl,
		ContractAddr: contract,
		ChainID:      "default",
		PrivFile:     absPath,
	}
	err := t.InitContract(t.ContractAddr)
	if err != nil {
		panic(err)
	}
	return t
}

func (n *TxConn) CallContract(method string, params proto.Message, result interface{}) error {
	_, err := n.contract.Call(method, params, n.signer, result)
	return err
}

func (n *TxConn) StaticCallContract(method string, params proto.Message, result interface{}) error {
	_, err := n.contract.StaticCall(method, params, loom.RootAddress(n.ChainID), result)
	return err
}
