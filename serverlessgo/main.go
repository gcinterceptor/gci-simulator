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

var arrivalQueue = godes.NewFIFOQueue("arrival")
var arrivalCond = godes.NewBooleanControl()

func main() {
	fmt.Println("Simulation Started")
	flag.Parse()
	
	lb := newLoadBalancer()
	godes.AddRunner(lb)

	poissonDist := &distuv.Poisson{*lambda, rand.NewSource(uint64(time.Now().Nanosecond()))}
	reqID := int64(0)
	godes.Run()
	for godes.GetSystemTime() < duration.Seconds() {
		arrivalQueue.Place(&request{id: reqID})
		arrivalCond.Set(true)
		interArrivalTime := poissonDist.Rand()
		godes.Advance(interArrivalTime)
		reqID++
	}

	lb.terminate()
	godes.WaitUntilDone()
	fmt.Println("Simulation Ended")
}

type loadBalancer struct {
	*godes.Runner
	isTerminated bool
}

func newLoadBalancer() *loadBalancer {
	return &loadBalancer{&godes.Runner{}, false}
}

func (lb *loadBalancer) terminate() {
	arrivalCond.Set(true)
	lb.isTerminated = true
}

func (lb *loadBalancer) Run() {
	fmt.Println("id,status,latency")
	for {
		arrivalCond.Wait(true)
		if arrivalQueue.Len() > 0 {
			r := arrivalQueue.Get().(*request)
			if r.status != 200 {
				r.status = 200
				r.responseTime = *lambda // temporary value
				arrivalQueue.Place(r)
			} else {
				fmt.Printf("%d,%d,%.1f\n", r.id, r.status, r.responseTime*1000)
			}
		}

		if arrivalQueue.Len() == 0 {
			if lb.isTerminated {
				break
			}
			arrivalCond.Set(false)
		}
	}
}

type request struct {
	id      int64
	responseTime float64
	status  int
}
