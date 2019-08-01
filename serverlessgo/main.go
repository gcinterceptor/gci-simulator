package main

import (
	"fmt"
	"flag"
	"time"

	"github.com/agoussia/godes"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

var (
	duration = flag.Duration("d", 300*time.Second, "Duration of the simulation.")
	lambda   = flag.Float64("lambda", 140.0, "The lambda of the Poisson distribution used on workload.")
)

func main() {
	fmt.Println("Simulation Started")
	flag.Parse()
	
	arrivalQueue := godes.NewFIFOQueue("arrival")
	poissonDist := &distuv.Poisson{*lambda, rand.NewSource(uint64(time.Now().Nanosecond()))}
	reqID := int64(0)
	godes.Run()
	for godes.GetSystemTime() < duration.Seconds() {
		arrivalQueue.Place(&request{id: reqID})
		godes.Advance(poisson_dist.Rand())
		reqID++
	}

	godes.WaitUntilDone()
	fmt.Println("Simulation Ended")
}

type request struct {
	id      int64
	responseTime float64
	status  int
}
