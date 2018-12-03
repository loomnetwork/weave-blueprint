package main

import (
	"fmt"
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/loomnetwork/weave-blueprint/src/types"
	"github.com/prometheus/client_golang/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stefantalpalaru/pool"
	"github.com/segmentio/ksuid"
	"gopkg.in/yaml.v2"
    "io/ioutil"
    "log"
    "net"
    "net/http"
    "os"
    "sync"
    "time"
    "path/filepath"
	"runtime"
    "flag"
	)



var defaultContract = "BluePrint"
var reg = prometheus.NewRegistry()

var requestCount metrics.Counter
var requestLatency metrics.Histogram
var requestLatencySummary metrics.Histogram

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

func work(args...interface{}) interface{} {

	c := args[0].(Servers)
	id := args[1].(int)
	data := args[2].(string)
	wg  := args[3].(*sync.WaitGroup)
    var err1 error

	defer wg.Done()
	defer func(begin time.Time) {
		//Measures time lapse when data is first seen in ith node
		lvs := []string{"method", "readpoll", "error", fmt.Sprint(err1 != nil),"server",c[1].Name}
		requestCount.With(lvs...).Add(1)
		requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
		requestLatencySummary.With(lvs...).Observe(time.Since(begin).Seconds())

	}(time.Now())

	//fmt.Printf("Worker-%d started polling \n", id)
	for {
		serverUrlRpc := fmt.Sprintf("http://%s:46658/rpc",  c[id].Extipaddress)
		serverUrlQuery := fmt.Sprintf("http://%s:46658/query",  c[id].Extipaddress)
		conns2 := map[string]*TxConn{}
		t := NewtxConn(serverUrlRpc, serverUrlQuery, defaultContract)
		conns2[c[id].Name] = t

		err1 := read(t,data, c[id].Name)

		if err1 == nil {
			//If read successfull while polling exit from go routine
			//fmt.Printf("Worker-%d finished polling \n", id)
			return nil
		}


	}
}


func main() {





	fieldKeys := []string{"method", "error", "server"}
	requestCount = kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "loomchain",
		Subsystem: "tx_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
   //Added bucket data point for histogram_quantile
	requestLatency = kitprometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Namespace: "loomchain",
		Subsystem: "tx_service",
		Name:      "request_latency",
		Help:      "Total duration of requests",
		Buckets: []float64{0.001,0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},

		}, fieldKeys)
	requestLatencySummary = kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "loomchain",
		Subsystem: "tx_service",
		Name:      "request_latency_summary",
		Help:      "Total duration of requests",
	}, fieldKeys)

	// default hostport for metrics
	var hostport = "0.0.0.0:9091"
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
	absPath, _ := filepath.Abs("weave-blueprint/src/latencymeasurementsworkerpool/nodelist.yaml")
	b, err := ioutil.ReadFile(absPath)

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

	var programCounter = 0;
	loopcontrol := flag.Int("loop", 10000, "Number of times to loop argument")
	flag.Parse()
	fmt.Println("Will Collect Following Number of Data Sets ", *loopcontrol)
	var mypool= pool.New(len(c) - 1) // number of workers
	cpus := runtime.NumCPU()
	//Multiple cores support for parallel execution
	runtime.GOMAXPROCS(cpus)
	mypool.Run()

	for programCounter <= *loopcontrol {
		var wg sync.WaitGroup
		var id= 0;
		wg.Add(len(c) - 1)

		data := ksuid.New().String()

		for _, v := range c {

			if id == 0 {

				serverUrlRpc := fmt.Sprintf("http://%s:46658/rpc", v.Extipaddress)

				serverUrlQuery := fmt.Sprintf("http://%s:46658/query", v.Extipaddress)
				conns1 := map[string]*TxConn{}

				t := NewtxConn(serverUrlRpc, serverUrlQuery, defaultContract)
				conns1[v.Name] = t

				err := write(t, data, v.Name)
				if err != nil {
					fmt.Printf("write error -%s\n", err.Error())
				}
				//Writing Key Value to First Node in File
			} else {
				//Submit task to worker
				mypool.Add(work, c, id, data, &wg)
			}

			id ++
		}
		//Exit when polling on all nodes is completes
		wg.Wait()
		programCounter++
		//prometheus.yaml is configured to scrape metrics from this application
		fmt.Printf("sleeping for final prometheus metrics\n")
		time.Sleep(5 * time.Second)
	}


	mypool.Stop()
	}

func read(t *TxConn, data, name string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "read", "error", fmt.Sprint(err != nil), "server", name}
		requestCount.With(lvs...).Add(1)
		requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
		requestLatencySummary.With(lvs...).Observe(time.Since(begin).Seconds())
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
		requestLatencySummary.With(lvs...).Observe(time.Since(begin).Seconds())

		}(time.Now())

	fmt.Printf("writing to %v\n", t)

	params := &types.MapEntry{
		Key:   "key",
		Value: data,
	}

	return t.CallContract("SetMsg", params, nil)
}
