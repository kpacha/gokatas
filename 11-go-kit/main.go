package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/hashicorp/consul/api"

	gokitlog "github.com/go-kit/kit/log"
	consulsd "github.com/go-kit/kit/sd/consul"
)

const (
	BackendSVC = "backendsvc"
	ProxySVC   = "proxysvc"
)

func main() {
	rand.Seed(time.Now().Unix())

	port := flag.Int("port", 8080, "port")
	consulAddr := flag.String("consul.addr", "", "Consul agent address")
	proxy := flag.Bool("proxy", false, "proxy")
	qps := flag.Int("qps", 500, "max queries per second")
	concurrency := flag.Int("concurrency", 3, "concurrency level")
	retryMax := flag.Int("retry.max", 3, "per-request retries to different instances")
	retryTimeout := flag.Duration("retry.timeout", 1000*time.Millisecond, "per-request timeout, including retries")
	flag.Parse()

	logger := gokitlog.NewLogfmtLogger(os.Stderr)
	logger = gokitlog.NewContext(logger).With("ts", gokitlog.DefaultTimestampUTC)

	var client consulsd.Client
	consulConfig := api.DefaultConfig()
	if len(*consulAddr) > 0 {
		consulConfig.Address = *consulAddr
	}
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}
	client = consulsd.NewClient(consulClient)

	var (
		agent api.AgentServiceRegistration
		task  func() error
	)

	if *proxy {
		agent = api.AgentServiceRegistration{
			ID:   ProxySVC + strconv.Itoa(*port),
			Name: ProxySVC,
			Port: *port,
			Tags: []string{"supu", "tupu"},
		}
		settings := proxySettings{
			port:         *port,
			retryMax:     *retryMax,
			retryTimeout: *retryTimeout,
			qps:          *qps,
			concurrency:  *concurrency,
		}
		task = func() error { return newProxy(settings, client, logger) }
	} else {
		agent = api.AgentServiceRegistration{
			ID:   BackendSVC + strconv.Itoa(*port),
			Name: BackendSVC,
			Port: *port,
			Tags: []string{"foo", "bar"},
		}
		task = func() error { return newBackend(*port, logger) }
	}

	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		client.Register(&agent)
		errc <- task()
	}()

	logger.Log("exit", <-errc)
	client.Deregister(&agent)
}
