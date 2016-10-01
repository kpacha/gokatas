gokatas
====

simple katas for golang

## TOC

### 01 - simple parser

A quick introduction to the language.

Goal: First taste of golang packages, goroutines and channels.

### 02 - flaky backend

Build an http service with not good enough performance and high error rate.

Goal: Introduction to http services

### 03 - simple client

Adapt the simple parser so it takes the XML to unmarshall from the flaky http service.

Goal: Basic knowledge of native http package

### 04 - proxy

Expose the results of the parsing process in another http service.

Goal: Explore the language features for concurrency

### 05 - proxy with multiple backends

Distribute the load across multiple backends and use the fastest one to serve the response

Goal: Build subscribers, generators and timeouts with goroutines and channels

### 06 - balanced proxy

Create a balancer in order to encapsulate the distribution of the load against the backend services

Goal: Encapsulate the components in structs. Goroutine reutilization.

### 07 - proxy with shared channels

After a quick profiling of the proxy, we can see it has too many channels, so it should be fixed

Goal: Share channels between multiple goroutines

### 08 - naive backoff

What will happen if a backend goes down? How could it be addressed?

Goal: Abstract the membership from the balancing strategy and add a very simple backoff strategy

### 09 - cluster aware proxy

Add a new endpoint to the proxy service so the backends could register themselves to the proxies

Goal: Be able to adapt the number of goroutines depending on the cluster status

### 10 - cluster client

Extend the simple client so it registers itself periodically to the proxy

Goal: Create a simple scheduled task

## Slides

The `slides` folder contains several go presentations with a complete explanation of the katas

Run them with the native `present` tool

	$ cd slides
	$ present
	2016/10/01 11:48:29 Open your web browser and visit http://127.0.0.1:3999
