package main

import (
	"math"
	"net/http"
)

type requeststream struct {
	endpoint        string
	responseChannel chan HTTPResponse
}

func (*requeststream) streamRequest() (
	timeout <-chan time.Time) {

	endpoint := requeststream.endpoint
	resopnseChannel := requeststream.responseChannel

	for timedOut := false; !timedOut; {
		select {
		case <-timeout:
			timedOut = true
		default:
			go requestEndpoint(endpoint, responseChannel)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (*requeststream) requestEndpoint() {

	defer func() {
		recover()
	}()

	endpoint := requeststream.endpoint
	responseChannel := requeststream.responseChannel

	var resultRow HTTPResponse
	before := time.Now()
	response, err := http.Get(endpoint)
	after := time.Now()

	latency := after.Sub(before)
	if err != nil {
		resultRow = HTTPResponse{-1, latency, true}
		responseChannel <- resultRow
		return
	}

	defer response.Body.Close()

	resultRow = HTTPResponse{response.StatusCode, latency, false}
	responseChannel <- resultRow
}
