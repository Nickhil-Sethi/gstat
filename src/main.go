package main

import (
	"flag"
	"fmt"
	"net/http"
	"runtime"
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
		"https://google.com/",
		"Endpoint to benchmark again")

	durationPtr := flag.Int(
		"duration",
		30,
		"duration of benchmark test in milliseconds. defaults to 30 seconds")

	// outFilePtr := flag.String(
	// 	"outfile",
	// 	"benchmark.json",
	// 	"name of file to write results to")

	runtime.GOMAXPROCS(*threadPtr)

	resultChannel := make(chan result)
	for i := 0; i < 1000; i++ {
		go requestEndpoint(*endpointPtr, resultChannel)
	}

	<-time.After(time.Duration(*durationPtr) * time.Second)
	return
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

	fmt.Println("Got status ", response.StatusCode)
	resultRow = result{response.StatusCode, latency, false}
	resultChannel <- resultRow
}
