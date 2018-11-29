package main

import (
	"fmt"
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/loomnetwork/weave-blueprint/src/types"
	"github.com/prometheus/client_golang/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
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

/*
type MessageData struct {
	Value int
}
*/

func main() {

	var wg sync.WaitGroup
    //Set to two as two nodes are to be polled
	wg.Add(2)

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
	var hostport = "127.0.0.1:9091"
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
	//var s1 Servers
	//conns := map[string]*TxConn{}

	b, err := ioutil.ReadFile("/root/weaveblueprint/weave-blueprint/src/latencycmd/nodelist.yaml")

	if err != nil {
		fmt.Print(err)
		return
	}

	//load server configs
	err = yaml.Unmarshal(b, &c)
	fmt.Println(c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("configdata - %v -%s\n", c, c[0].Extipaddress)


//Writing Key Value to FrankFurt Node
	serverUrlRpc := fmt.Sprintf("http://%s:46658/rpc", "18.184.73.130")
	serverUrlQuery := fmt.Sprintf("http://%s:46658/query", "18.184.73.130")
	conns1 := map[string]*TxConn{}

	t := NewtxConn(serverUrlRpc, serverUrlQuery, defaultContract)
	conns1["perftest-suhas-frankfurt-0"] = t

	err = write(t, "20", "perftest-suhas-frankfurt-0")
	if err != nil {
		fmt.Printf("write error -%s\n", err.Error())
	}


//This Go routine polls North CA node
	go func() {
		defer wg.Done()
		defer func(begin time.Time) {
                //Measures time lapse when data is first seen in North CA node
			lvs := []string{"method", "readpoll", "error", fmt.Sprint(err != nil), "server", "perftest-suhas-north_ca-0"}
			requestCount.With(lvs...).Add(1)
			requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
		}(time.Now())

			for {
				serverUrlRpc := fmt.Sprintf("http://%s:46658/rpc", "54.183.107.171")
				serverUrlQuery := fmt.Sprintf("http://%s:46658/query", "54.183.107.171")
				conns2 := map[string]*TxConn{}
				t := NewtxConn(serverUrlRpc, serverUrlQuery, defaultContract)
				conns2[" perftest-suhas-north_ca-0"] = t

				err1 := read(t, "20", "perftest-suhas-north_ca-0")

				if err1 == nil {
                //If read successfull while polling exit from go routine
					return
				}

			}
      
	}()

//This Go Routine polls Tokyo Node
	go func() {
		defer wg.Done()
		defer func(begin time.Time) {
                //Measures time lapse when data is first seen in Tokyo node
			lvs := []string{"method", "readpoll", "error", fmt.Sprint(err != nil), "server", "perftest-suhas-tokyo-0"}
			requestCount.With(lvs...).Add(1)
			requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
		}(time.Now())
			for {
				serverUrlRpc := fmt.Sprintf("http://%s:46658/rpc", "52.198.14.57")
				serverUrlQuery := fmt.Sprintf("http://%s:46658/query", "52.198.14.57")
				conns3 := map[string]*TxConn{}
				t := NewtxConn(serverUrlRpc, serverUrlQuery, defaultContract)
				conns3["perftest-suhas-tokyo-0"] = t
				err2 := read(t, "20", "perftest-suhas-tokyo-0")
				if err2 == nil {
					//If read successfull while polling exit from go routine
					return
				}

			}

			
	}()

	//Exit when polling on all nodes is completes
	wg.Wait()
        //prometheus.yaml is configured to scrape metrics from this application 
	fmt.Printf("sleeping for final prometheus metrics\n")
	time.Sleep(10 * time.Second)
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

func write(t *TxConn, data, name string) (err error) {

	defer func(begin time.Time) {
		lvs := []string{"method", "write", "error", fmt.Sprint(err != nil), "server", name}
		requestCount.With(lvs...).Add(1)
		requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	fmt.Printf("writing to %v\n", t)

	params := &types.MapEntry{
		Key:   "key",
		Value: data,
	}

	return t.CallContract("SetMsg", params, nil)
}
