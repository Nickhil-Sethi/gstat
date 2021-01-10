package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/alexflint/go-arg"
)

type HTTPResponse struct {
	status  int
	latency time.Duration
	errored bool
}

type csvRow struct {
	status  string
	latency string
	errored string
}

func main() {
	var benchmarkargs struct {
		Endpoint string `arg:"required"`
		Duration time.Duration `default:"30s"`
		NumGoroutines int `default:"-1"`
	}

	arg.MustParse(&benchmarkargs)

	responseChannel := make(chan HTTPResponse)
	var wg sync.WaitGroup

	sm := StatManager{
		responseChannel,
		&wg}

	duration := benchmarktask.Duration
	endpoint := benchmarktask.Endpoint
	numGoroutines := runtime.GOMAXPROCS(
		benchmarktask.NumGoroutines)
	timeout := time.After(duration)

	fmt.Printf(
		"Running %s benchmark test @ %s over %d goroutines\n",
		duration,
		endpoint,
		numGoroutines)
	return
	wg.Add(1)
	go sm.compileResults(responseChannel, &wg)

	for i := 0; i < runtime.GOMAXPROCS(numGoroutines); i++ {
		stream := RequestStream{
			endpoint,
			responseChannel}
		go stream.StreamRequests(timeout)
	}

	<-timeout
	close(responseChannel)

	wg.Wait()
	// sm.writeResults()
}
