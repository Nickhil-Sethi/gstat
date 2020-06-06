package main

import (
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type result struct {
	status  int
	latency time.Duration
	errored bool
}

func main() {
	threadPtr := flag.Int(
		"threads",
		-1,
		"Number of threads to use. Defaults to number of cores on machine.")

	endpointPtr := flag.String(
		"endpoint",
		"http://nickhil-sethi.com/",
		"Endpoint to benchmark against")

	durationPtr := flag.Int(
		"duration",
		30,
		"duration of benchmark test in milliseconds. defaults to 30 seconds")

	outFilePtr := flag.String(
		"outfile",
		"benchmark.json",
		"name of file to write results to")

	runtime.GOMAXPROCS(*threadPtr)

	resultChannel := make(chan result)
	var wg sync.WaitGroup
	wg.Add(1)
	go compileResults(resultChannel, *outFilePtr, &wg)

	timeout := time.After(
		time.Duration(*durationPtr) * time.Second)

	for keepGoing, i := true, 0; keepGoing && i < 5; i++ {
		select {
		case <-timeout:
			keepGoing = false
		default:
		}
		go requestEndpoint(*endpointPtr, resultChannel)
		// keepGoing = false
	}

	// close(resultChannel)
	wg.Wait()
	return
}

func compileResults(resultChannel chan result, filename string, wg *sync.WaitGroup) {
	for res := range resultChannel {
		fmt.Println(res)
	}
	wg.Done()
}

func requestEndpoint(endpoint string, resultChannel chan result) {

	before := time.Now()
	response, err := http.Get(endpoint)
	after := time.Now()

	latency := after.Sub(before)
	var resultRow result
	if err != nil {
		resultRow = result{-1, latency, true}
		resultChannel <- resultRow
		return
	}

	defer response.Body.Close()

	resultRow = result{response.StatusCode, latency, false}
	resultChannel <- resultRow
}
