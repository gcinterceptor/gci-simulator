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
	fmt.Println("id,status,latency")
	flag.Parse()
	
	lb := newLoadBalancer()
	godes.AddRunner(lb)
	godes.Run()

	reqID := int64(0)
	poissonDist := &distuv.Poisson{*lambda, rand.NewSource(uint64(time.Now().Nanosecond()))}
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
	instances      []*instance
}

func newLoadBalancer() *loadBalancer {
	return &loadBalancer{&godes.Runner{}, false, godes.NewFIFOQueue("arrival"), godes.NewBooleanControl(), make([]*instance, 0)}
}

func (lb *loadBalancer) receiveRequest(r *request) {
	lb.arrivalQueue.Place(r)
	lb.arrivalCond.Set(true)
}

func (lb *loadBalancer) terminate() {
	for i := 0; i < len(lb.instances); i++ {
		lb.instances[i].terminate()
	}
	lb.isTerminated = true
}

func (lb *loadBalancer) nextInstance() *instance {
	var instance *instance
	instance = newInstance(len(lb.instances))
	lb.instances = append(lb.instances, instance)
	godes.AddRunner(instance)
	return instance
}

func (lb *loadBalancer) Run() {
	for {
		lb.arrivalCond.Wait(true)
		if lb.arrivalQueue.Len() > 0 {
			r := lb.arrivalQueue.Get().(*request)
			lb.nextInstance().receiveRequest(r)			
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

func (i *instance) isWorking() bool {
	return i.cond.GetState() == true
}

func (i *instance) receiveRequest(r *request) {
	if i.isWorking() == true {
		panic(fmt.Sprintf("Instances may not enqueue requests."))
	}
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
		
		fmt.Printf("%d,%d,%.1f\n", i.req.id, i.req.status, i.req.responseTime*1000)

		i.cond.Set(false)
	}
}

type request struct {
	id      int64
	responseTime float64
	status  int
}
