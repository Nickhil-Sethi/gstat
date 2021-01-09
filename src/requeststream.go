package main

import (
	"net/http"
	"time"
)

type RequestStream struct {
	Endpoint        string
	ResponseChannel chan HTTPResponse
}

func (r *RequestStream) StreamRequests(
	timeout <-chan time.Time) {

	for timedOut := false; !timedOut; {
		select {
		case <-timeout:
			timedOut = true
		default:
			go r.RequestEndpoint()
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (r *RequestStream) RequestEndpoint() {

	defer func() {
		recover()
	}()

	var Endpoint string = r.Endpoint
	var responseChannel chan HTTPResponse = r.ResponseChannel

	var resultRow HTTPResponse
	before := time.Now()
	response, err := http.Get(Endpoint)
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
