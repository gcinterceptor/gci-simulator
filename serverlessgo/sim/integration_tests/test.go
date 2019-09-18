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
func main() {
	duration := 10 * time.Millisecond
	idlenessDeadline := 10 * time.Millisecond
	input := [][]sim.InputEntry{{{200, 10}}}
	optimized := false
	var collector collectorListener
	sim.Run(duration, idlenessDeadline, sim.NewPoissonInterArrival(1), input, &collector, optimized)
	got := collector[0]
	want := sim.Request{ID: 0, Status: 200, ResponseTime: 10, Hops: []int{0}}
	if !reflect.DeepEqual(got, want) {
		log.Fatalf("want:%+v got:%+v", got, want)
	}
	fmt.Println("OK")
}
