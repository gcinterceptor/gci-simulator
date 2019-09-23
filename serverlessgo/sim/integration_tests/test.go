// This binary is needed because the simulation can only run in the main thread.
package main

import (
	"fmt"
	"log"
	"math"
	"reflect"
	"time"

	"github.com/gcinterceptor/gci-simulator/serverless/sim"
)

type collectorListener []sim.Request

func (rs *collectorListener) RequestFinished(req *sim.Request) {
	*rs = append(*rs, *req)
}

func main() {
	duration := 80 * time.Millisecond
	idlenessDeadline := 6 * time.Millisecond
	input := [][]sim.InputEntry{
		{{200, 0.015}, {200, 0.008}, {503, 0.0001}},
		{{200, 0.011}},
		{{200, 0.005}, {200, 0.005}, {503, 0.0002}},
	}
	optimized := false
	var simulatedRequests collectorListener
	res := sim.Run(duration, idlenessDeadline, sim.NewConstantInterArrival(0.01), input, &simulatedRequests, optimized)

	if len(res.Instances) != 5 {
		log.Fatalf("number of instances - want:5 got:%+v", len(res.Instances))
	}
	expectedRequests := []sim.Request{
		{ID: 0, Status: 200, ResponseTime: 0.015, Hops: []int{0}},     // response time from {200, 0.015} of instance 0
		{ID: 1, Status: 200, ResponseTime: 0.011, Hops: []int{1}},     // response time from {200, 0.011} of instance 1
		{ID: 2, Status: 200, ResponseTime: 0.008, Hops: []int{0}},     // response time from {200, 0.008} of instance 0
		{ID: 3, Status: 200, ResponseTime: 0.0051, Hops: []int{0, 2}}, // response time from {503, 0.0001} of instance 0 plus {200, 0.005} of instance 2
		{ID: 4, Status: 200, ResponseTime: 0.005, Hops: []int{2}},     // response time from {200, 0.005} of instance 2
		{ID: 5, Status: 200, ResponseTime: 0.0152, Hops: []int{2, 3}}, // response time from {503, 0.0002} of instance 0 plus {200, 0.015} of instance 3
		{ID: 6, Status: 200, ResponseTime: 0.011, Hops: []int{4}},     // response time from {200, 0.011} of instance 4
		{ID: 7, Status: 200, ResponseTime: 0.008, Hops: []int{3}},     // response time from {200, 0.008} of instance 3
	}
	if len(expectedRequests) != len(simulatedRequests) {
		log.Fatalf("number of requests - want:%+v got:%+v", len(expectedRequests), len(simulatedRequests))
	}
	for i, got := range simulatedRequests {
		if !reflect.DeepEqual(expectedRequests[i], got) {
			log.Fatalf("request output - want:%+v got:%+v", expectedRequests[i], got)
		}
	}
	if math.Abs(1000*res.Cost-126.3) > 0.5 { // 36.1 + 17 + 26.2 + 30 + 17 = 126.3
		// where 36.1, 17, 26.2, 30 and 17 are the uptime of instances 0, 1, 2, 3 and 4, respectively
		log.Fatalf("instances cost - want:%+v got:%+v", 126.3, 1000*res.Cost)
	}
	if math.Abs(res.Efficiency-0.61866396416) > 0.001 { // (23.1/36.1 + 11/17 + 10.2/26.2 + 23.1/30 + 11/17) / 5 = 0.61866396416
		// where 23.1/36.1, 11/17, 10.2/26.2,  23.1/30 and 11/17 are the uptime of instances 0, 1, 2, 3 and 4, respectively
		log.Fatalf("instances efficiency - want:%+v got:%+v", 0.61866396416, res.Efficiency)
	}
	fmt.Println("OK")
}
