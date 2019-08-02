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
	flag.Parse()
	
	lb := newLoadBalancer()
	godes.AddRunner(lb)

	poissonDist := &distuv.Poisson{*lambda, rand.NewSource(uint64(time.Now().Nanosecond()))}
	reqID := int64(0)
	godes.Run()
	for godes.GetSystemTime() < duration.Seconds() {
		lb.receiveRequest(&request{id: reqID})
		interArrivalTime := poissonDist.Rand()
		godes.Advance(interArrivalTime)
		reqID++
	}

	lb.terminate()
	godes.WaitUntilDone()
}

type loadBalancer struct {
	*godes.Runner
	isTerminated bool
	arrivalQueue *godes.FIFOQueue
	arrivalCond *godes.BooleanControl
}

func newLoadBalancer() *loadBalancer {
	return &loadBalancer{&godes.Runner{}, false, godes.NewFIFOQueue("arrival"), godes.NewBooleanControl()}
}

func (lb *loadBalancer) receiveRequest(r *request) {
	lb.arrivalQueue.Place(r)
	lb.arrivalCond.Set(true)
}

func (lb *loadBalancer) terminate() {
	lb.isTerminated = true
}

func (lb *loadBalancer) Run() {
	fmt.Println("id,status,latency")
	for {
		lb.arrivalCond.Wait(true)
		if lb.arrivalQueue.Len() > 0 {
			r := lb.arrivalQueue.Get().(*request)
			r.status = 200
			r.responseTime += *lambda // temporary value
			godes.Advance(*lambda)
			fmt.Printf("%d,%d,%.1f\n", r.id, r.status, r.responseTime*1000)
		}

		if lb.arrivalQueue.Len() == 0 {
			if lb.isTerminated {
				break
			}
			lb.arrivalCond.Set(false)
		}
	}
}

type instance struct {
	*godes.Runner
	id           int
	isTerminated bool
	cond  *godes.BooleanControl
	req   *request
}

func newInstance(id int) *instance {
	return &instance{&godes.Runner{}, id, false, godes.NewBooleanControl(), nil}
}

func (i *instance) receiveRequest(r *request) {
	i.req = r
	i.cond.Set(true)
}

func (i *instance) terminate() {
	i.isTerminated = true
}

func (i *instance) Run() {
	for {
		if i.isTerminated {
			break
		}

		i.cond.Wait(true)
		
		i.req.status = 200
		i.req.responseTime += *lambda // temporary value
		godes.Advance(*lambda)
		
		si.cond.Set(false)
	}
}

type request struct {
	id      int64
	responseTime float64
	status  int
}
