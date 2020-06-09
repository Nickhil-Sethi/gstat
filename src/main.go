package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

type requestResult struct {
	status  int
	latency time.Duration
	errored bool
}

func main() {
	endpointPtr := flag.String(
		"e",
		"",
		"Endpoint to benchmark against")

	durationPtr := flag.Int(
		"d",
		5,
		"duration of benchmark test in milliseconds. defaults to 30 seconds")

	threadPtr := flag.Int(
		"t",
		-1,
		"Number of threads to use. Defaults to number of cores on machine.")

	outFilePtr := flag.String(
		"o",
		"benchmark.json",
		"name of file to write results to")

	if *endpointPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	runtime.GOMAXPROCS(*threadPtr)

	resultChannel := make(chan requestResult)
	var wg sync.WaitGroup

	wg.Add(1)
	go compileResults(
		resultChannel,
		*outFilePtr,
		&wg)

	duration := time.Duration(*durationPtr)
	timeout := time.After(duration * time.Second)

	for i := 0; i < runtime.GOMAXPROCS(-1); i++ {
		go streamRequests(
			*endpointPtr,
			resultChannel,
			timeout)
	}

	select {
	case <-timeout:
		close(resultChannel)
	}

	wg.Wait()
	return
}

func compileResults(
	resultChannel chan requestResult,
	filename string,
	wg *sync.WaitGroup) {
	count := 0
	var totalLatency time.Duration
	for res := range resultChannel {
		totalLatency += res.latency
		count++
	}
	avgLatency := int(totalLatency) / count
	fmt.Printf("Launched %d requests: average latency %d\n", count, avgLatency)
	// TODO(nickhil) : write results to file
	wg.Done()
}

func streamRequests(
	endpoint string,
	resultChannel chan requestResult,
	timeout <-chan time.Time) {
	for keepGoing, i := true, 0; keepGoing && i < 1000; i++ {
		select {
		case <-timeout:
			keepGoing = false
		default:
		}
		go requestEndpoint(endpoint, resultChannel)
		time.Sleep(time.Millisecond)
	}
}
func requestEndpoint(endpoint string, resultChannel chan requestResult) {

	defer func() {
		recover()
	}()

	before := time.Now()
	response, err := http.Get(endpoint)
	after := time.Now()

	latency := after.Sub(before)
	var resultRow requestResult
	if err != nil {
		resultRow = requestResult{-1, latency, true}
		resultChannel <- resultRow
		return
	}

	defer response.Body.Close()

	resultRow = requestResult{response.StatusCode, latency, false}
	resultChannel <- resultRow
}
