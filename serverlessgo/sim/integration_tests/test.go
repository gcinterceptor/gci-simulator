// This binary is needed because the simulation can only run in the main thread.
package main

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/gcinterceptor/gci-simulator/serverless/sim"
)

type collectorListener []sim.Request

func (rs *collectorListener) RequestFinished(req *sim.Request) {
	*rs = append(*rs, *req)
}

// TODO(david): Improve integration tests
// - read from a file the simulation input/output.
func main() {
	duration := 80 * time.Millisecond
	idlenessDeadline := 6 * time.Millisecond
	input := [][]sim.InputEntry{
		{{200, 0.015}, {200, 0.008}, {503, 0.0001}},
		{{200, 0.011}},
		{{200, 0.005}, {200, 0.005}, {503, 0.0002}},
	}
	optimized := false
	var collector collectorListener
	res := sim.Run(duration, idlenessDeadline, sim.NewConstantInterArrival(0.01), input, &collector, optimized)
	
	instancesWant, instancesGot := 5, len(res.Instances)
	if !reflect.DeepEqual(instancesWant, instancesGot) {
		log.Fatalf("number of instances - want:%+v got:%+v", instancesWant, instancesGot)
	}
	requestsWant := []sim.Request{
		{ID: 0, Status: 200, ResponseTime: 0.015, Hops: []int{0}},
		{ID: 1, Status: 200, ResponseTime: 0.011, Hops: []int{1}},
		{ID: 2, Status: 200, ResponseTime: 0.008, Hops: []int{0}},
		{ID: 3, Status: 200, ResponseTime: 0.0051, Hops: []int{0, 2}},
		{ID: 4, Status: 200, ResponseTime: 0.005, Hops: []int{2}},
		{ID: 5, Status: 200, ResponseTime: 0.0152, Hops: []int{2, 3}},
		{ID: 6, Status: 200, ResponseTime: 0.011, Hops: []int{4}},
		{ID: 7, Status: 200, ResponseTime: 0.008, Hops: []int{3}},
	}
	requestGot := collector
	if len(requestsWant) != len(requestGot) {
		log.Fatalf("number of requests - want:%+v got:%+v", len(requestsWant), len(requestGot))
	}
	for i, got := range requestGot {
		if !reflect.DeepEqual(requestsWant[i], got) {
			log.Fatalf("request output - want:%+v got:%+v", requestsWant[i], got)
		}
	}
	fmt.Println("OK")
}
