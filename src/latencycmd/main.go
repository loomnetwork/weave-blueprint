package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/loomnetwork/weave-blueprint/src/types"
	"gopkg.in/yaml.v2"
)

var user = "loom"
var defaultContract = "BluePrint"

type Server struct {
	Name         string `yaml:name`
	Extipaddress string `yaml:extipaddress`
	Region       string `yaml:region`
	Subnet       string `yaml:subnet`
}

/*
type Config struct {
	Servers []Server `yaml:, innerflow`
}
*/
type Servers []Server

type MessageData struct {
	Value int
}

func read(t *TxConn, data string) error {
	fmt.Printf("reading from %v\n", t)
	params := &types.MapEntry{
		Key:   "key",
		Value: data,
	}
	var result types.MapEntry
	err := t.StaticCallContract("GetMsg", params, &result)
	if err != nil {
		return err
	}
	log.Printf("{ key: \"%s\", value: \"%s\" }\n", result.Key, result.Value)

	return nil
}

func write(t *TxConn, data string) error {
	fmt.Printf("writing to %v\n", t)

	params := &types.MapEntry{
		Key:   "key",
		Value: data,
	}

	return t.CallContract("SetMsg", params, nil)
}

type attack func(int) error

//todo make a lambda
func measure(attackFn attack) {
	t := time.Now()

	attackFn(0)

	elasped := time.Now().Sub(t)
	fmt.Printf("elapsed seconds -%f\n", elasped.Seconds())
}

func main() {
	//ar c Config
	var c Servers
	conns := map[string]*TxConn{}

	b, err := ioutil.ReadFile("nodelist.yaml")
	if err != nil {
		fmt.Print(err)
		return
	}

	//load server configs
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("configdata - %v -%s\n", c, c[0].Extipaddress)

	for _, v := range c {
		serverUrlRpc := fmt.Sprintf("http://%s:46658/rpc", v.Extipaddress)
		serverUrlQuery := fmt.Sprintf("http://%s:46658/query", v.Extipaddress)

		t := NewTxConn(serverUrlRpc, serverUrlQuery, defaultContract)
		conns[v.Name] = t

		err = write(t, "1")
		if err != nil {
			fmt.Printf("write error -%s\n", err.Error())
		}
		err = read(t, "1")
	}
	//measure(read)
	//measure(write)
}
