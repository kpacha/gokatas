# 'Final' Implementation

Recreate the cluster with more 'production-ready' implementations

### Components:

* consul
* go-kit/kit
* juju/ratelimit
* sony/gobreaker

### Install

Install the required dependencies:

	$ go get github.com/go-kit/kit
	$ go get github.com/juju/ratelimit
	$ go get github.com/sony/gobreaker
	$ go get github.com/hashicorp/consul/api

now, the dependencies of the framework

	$ go get github.com/afex/hystrix-go/hystrix
	$ go get github.com/go-logfmt/logfmt
	$ go get github.com/go-stack/stack
	$ go get github.com/streadway/handy/breaker

and you're ready to build it

	$ go build

Start a consul agent

	$ consul agent -dev -ui

And on different terminals, start the backends and the proy

	$ ./11-go-kit -port 8081 -consul.addr 127.0.0.1:8500
	$ ./11-go-kit -port 8082 -consul.addr 127.0.0.1:8500
	$ ./11-go-kit -port 8083 -consul.addr 127.0.0.1:8500
	$ ./11-go-kit -proxy -consul.addr 127.0.0.1:8500