package main

import (
	"fmt"
	"errors"
	"github.com/aybabtme/uniplot/histogram"
	"math"
	"os"
	"runtime"
	"sync"
	"text/tabwriter"
	"time"
)

type StatManager struct {
	responseChannel chan HTTPResponse
	wg              *sync.WaitGroup
}

func (s *StatManager) WriteHistogram(latencyData []float64) {
	latencyHist := histogram.Hist(5, latencyData)
	histogram.Fprint(os.Stdout, latencyHist, histogram.Linear(5))
}

func (s *StatManager) compileResults(
	responseChannel chan HTTPResponse,
	wg *sync.WaitGroup) {

	defer wg.Done()

	compiledResults := make([]HTTPResponse, 1)
	for HTTPResponse := range responseChannel {
		compiledResults = append(compiledResults, HTTPResponse)
	}
	s.WriteResults(compiledResults)
}

func (s *StatManager) WriteResults(
	compiledResults []HTTPResponse) error {

	if len(compiledResults) == 0 {
		return errors.New("Received empty results array.")
	}

	var count int64
	var minLatency time.Duration
	var maxLatency time.Duration

	var latencyData []float64

	// TODO(nickhil) : this is a hack to compute
	// the maxLatency initial value. is there something
	// equivalent to np.inf here?
	minLatency, maxLatency = time.Duration(
		290*time.Millisecond), time.Duration(0)
	var totalLatency time.Duration
	var secondMoment int64
	count = 0

	// w := csv.NewWriter(os.Stdout)

	for _, res := range compiledResults {
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

		latencyData = append(latencyData, float64(res.latency))
	}

	ms := int64(totalLatency / time.Millisecond)
	max := int64(maxLatency / time.Millisecond)
	min := int64(minLatency / time.Millisecond)
	avg := ms / count
	variance := secondMoment/count - avg*avg
	stddev := math.Sqrt(float64(variance))

	w := tabwriter.NewWriter(os.Stdout, 2, 5, 1, ' ', tabwriter.AlignRight)
	fmt.Fprintf(w, "\t%d concurrent requests / %d threads\n", count, runtime.GOMAXPROCS(-1))
	fmt.Fprintf(w, "\tLatency stats (ms)\n")
	fmt.Fprintf(w, "\t\tMax\tMin\tAvg\t+/- StDev\t\n")
	fmt.Fprintf(w, "\t\t%d\t%d\t%d\t%0.2f\t\n", max, min, avg, stddev)
	w.Flush()

	s.WriteHistogram(latencyData)
	return nil
}
