package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/loomnetwork/weave-blueprint/src/types"
	"github.com/prometheus/client_golang/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
)

var defaultContract = "BluePrint"
var reg = prometheus.NewRegistry()

var requestCount metrics.Counter
var requestLatency metrics.Histogram

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

func read(t *TxConn, data, name string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "read", "error", fmt.Sprint(err != nil), "server", name}
		requestCount.With(lvs...).Add(1)
		requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	fmt.Printf("reading from %v\n", t)
	params := &types.MapEntry{
		Key:   "key",
		Value: data,
	}
	var result types.MapEntry
	err = t.StaticCallContract("GetMsg", params, &result)
	if err != nil {
		return err
	}
	log.Printf("{ key: \"%s\", value: \"%s\" }\n", result.Key, result.Value)

	return nil
}

func write(t *TxConn, data, name string) error {
	fmt.Printf("writing to %v\n", t)

	params := &types.MapEntry{
		Key:   "key",
		Value: data,
	}

	return t.CallContract("SetMsg", params, nil)
}

func main() {

	fieldKeys := []string{"method", "error", "server"}
	requestCount = kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "loomchain",
		Subsystem: "tx_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	requestLatency = kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "loomchain",
		Subsystem: "tx_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)

	// default hostport for metrics
	var hostport = "127.0.0.1:9095"
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid metric address: %s", err)
		os.Exit(1)
	}
	// Serve promtheus http server
	httpServer := &http.Server{
		//Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
		Handler: promhttp.Handler(),
		Addr:    net.JoinHostPort(host, port),
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, "unable to start http server: %v", err)
			os.Exit(1)
		}
	}()
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

		err = write(t, "1", v.Name)
		if err != nil {
			fmt.Printf("write error -%s\n", err.Error())
		}
		err = read(t, "1", v.Name)
	}

	fmt.Printf("sleeping for final prometheus metrics\n")
	time.Sleep(10 * time.Second)
}
