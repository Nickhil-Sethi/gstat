package main

import (
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"runtime"
	"sync"
	"text/tabwriter"
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

	flag.Parse()

	if *endpointPtr == "" {
		fmt.Print(*endpointPtr)
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("Running %ds benchmark test @ %s\n", *durationPtr, *endpointPtr)

	runtime.GOMAXPROCS(*threadPtr)

	resultChannel := make(chan requestResult)
	var wg sync.WaitGroup

	wg.Add(1)
	// background goroutine
	// compiles request
	// responses and computes
	// descriptive status
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

	var count int64
	var minLatency time.Duration
	var maxLatency time.Duration
	minLatency, maxLatency = time.Duration(290*time.Millisecond), time.Duration(0)
	var totalLatency time.Duration
	var secondMoment int64
	count = 0

	// w := csv.NewWriter(os.Stdout)

	for res := range resultChannel {
		if res.latency < minLatency {
			minLatency = res.latency
		}
		if res.latency > maxLatency {
			maxLatency = res.latency
		}
		totalLatency += res.latency
		intLatency := int64(res.latency / time.Millisecond)
		secondMoment += intLatency * intLatency
		count++
	}
	if count == 0 {
		fmt.Print("Error: Launched 0 requests. Something is wrong.")
		os.Exit(1)
	}
	ms := int64(totalLatency / time.Millisecond)
	max := int64(maxLatency / time.Millisecond)
	min := int64(minLatency / time.Millisecond)
	avg := ms / count
	variance := secondMoment/count - avg*avg
	stddev := math.Sqrt(float64(variance))

	w := tabwriter.NewWriter(os.Stdout, 2, 5, 1, ' ', tabwriter.AlignRight)
	fmt.Fprintf(w, "\t%d concurrent requests / %d threads\n", count, runtime.GOMAXPROCS(-1))
	fmt.Fprintf(w, "\tLatency stats\n")
	fmt.Fprintf(w, "\t\tMax\tMin\tAvg\tStDev\t\n")
	fmt.Fprintf(w, "\t\t%d\t%d\t%d\t%0.2f\t\n", max, min, avg, stddev)
	w.Flush()
	// TODO(nickhil) : write results to file
	wg.Done()
}

func streamRequests(
	endpoint string,
	resultChannel chan requestResult,
	timeout <-chan time.Time) {
	for keepGoing := true; keepGoing; {
		select {
		case <-timeout:
			keepGoing = false
		default:
		}
		go requestEndpoint(endpoint, resultChannel)
		time.Sleep(10 * time.Millisecond)
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
