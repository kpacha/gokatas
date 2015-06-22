package main

// Code stolen from Peter Teichman ("http://blog.gopheracademy.com/advent-2014/backoff/")

import (
	"math/rand"
	"time"
)

type BackoffPolicy struct {
	Millis     []int
	MaxRetries int
}

func (b BackoffPolicy) Duration(n int) time.Duration {
	if n >= len(b.Millis) {
		n = len(b.Millis) - 1
	}

	return time.Duration(jitter(b.Millis[n])) * time.Millisecond
}

func jitter(millis int) int {
	if millis == 0 {
		return 0
	}

	return millis/2 + rand.Intn(millis)
}
