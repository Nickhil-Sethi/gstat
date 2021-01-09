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
	endpointPtr := flag.String(
		"e",
		"",
		"Endpoint to benchmark against.")

	defaultDuration := 5 * 1000 * 1000 * 1000
	durationPtr := flag.Duration(
		"d",
		time.Duration(defaultDuration),
		"duration of benchmark test in milliseconds.")

	threadPtr := flag.Int(
		"t",
		-1,
		"Number of threads to use.")

	outFilePtr := flag.String(
		"o",
		"benchmark.csv",
		"name of file to write results to")

	flag.Parse()

	// TODO(nickhil): all inputs necessary
	// should have their own message
	if *endpointPtr == "" {
		flag.Usage()
		os.Exit(1)
	}

	responseChannel := make(chan HTTPResponse)
	var wg sync.WaitGroup

	sm := StatManager{
		responseChannel,
		&wg}

	duration := *durationPtr
	endpoint := *endpointPtr
	numGoroutines := runtime.GOMAXPROCS(*threadPtr)
	timeout := time.After(duration)

	fmt.Printf(
		"Running %s benchmark test @ %s over %d goroutines\n",
		duration,
		endpoint,
		numGoroutines)

	wg.Add(1)
	go sm.compileResults()

	for i := 0; i < runtime.GOMAXPROCS(numGoroutines); i++ {
		stream := requeststream{
			endpoint,
			responseChannel}
		go stream.streamrequests(timeout)
	}

	<-timeout
	close(responseChannel)

	wg.Wait()
	sm.writeResults()
}
