package main

import (
	"fmt"
	"flag"
	"time"
	"sort"
	"strings"

	"github.com/agoussia/godes"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

var (
	idlenessDeadline = flag.Duration("i", 300*time.Second, "The idleness deadline is the time that an instance may be idle until be terminated.")
	duration = flag.Duration("d", 300*time.Second, "Duration of the simulation.")
	lambda   = flag.Float64("lambda", 140.0, "The lambda of the Poisson distribution used on workload.")
	inputs   = flag.String("inputs", "test.csv", "Comma-separated file paths (one per instance)")
)

func main() {
	// TODO(David): to abstract output via struct
	fmt.Println("id,status,response_time")
	flag.Parse()
	
	lb := newLoadBalancer(*idlenessDeadline, *inputs)
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
	idlenessDeadline time.Duration
	inputs   []string
	index    int
}

func newLoadBalancer(idlenessDeadline time.Duration, inputs string) *loadBalancer {
	return &loadBalancer{&godes.Runner{}, false, godes.NewFIFOQueue("arrival"), godes.NewBooleanControl(), make([]*instance, 0), idlenessDeadline, strings.Split(inputs, ","), 0}
}

func (lb *loadBalancer) receiveRequest(r *request) {
	lb.arrivalQueue.Place(r)
	lb.arrivalCond.Set(true)
}

func (lb *loadBalancer) terminate() {
	for _, i := range lb.instances {
		i.terminate()
	}
	lb.isTerminated = true
	lb.arrivalCond.Set(true)
}

func (lb *loadBalancer) nextInstanceInputFile() string {
	input := lb.inputs[lb.index]
	lb.index = (lb.index + 1) % len(lb.inputs)
	return input
}

func (lb *loadBalancer) nextInstance() *instance {
	var selected *instance
	// sorting instances to have the most recently used ones ahead on the array
	sort.SliceStable(lb.instances, func(i, j int) bool { return lb.instances[i].getLastWorked() > lb.instances[j].getLastWorked() })
	for i := 0; i < len(lb.instances); i++ {
		instance := lb.instances[i]
		if !instance.isWorking() && !instance.isTerminated() {
			selected = instance
			break
		}
	}
	
	if selected == nil {
		selected = newInstance(len(lb.instances), lb.idlenessDeadline, lb.nextInstanceInputFile())
		godes.AddRunner(selected)
		// inserts the instance ahead of the array
		lb.instances = append([]*instance{selected}, lb.instances...)
	}

	return selected
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
		lb.tryScaleDown()
	}
}

func (lb *loadBalancer) tryScaleDown() {
	for _, i := range lb.instances {
		if godes.GetSystemTime() - i.getLastWorked() >= lb.idlenessDeadline.Seconds() {
			i.scaleDown()
		}
	}
}

type instance struct {
	*godes.Runner
	id           int
	terminated   bool
	cond  *godes.BooleanControl
	req   *request
	createdTime   float64
	terminateTime float64
	lastWorked     float64
	busyTime      float64
	idlenessDeadline time.Duration
    entries       []inputEntry
	index 		  int
}

func newInstance(id int, idlenessDeadline time.Duration, input string) *instance {
	return &instance{&godes.Runner{}, id, false, godes.NewBooleanControl(), nil, godes.GetSystemTime(), 0, 0, 0, idlenessDeadline, buildEntryArray(input), 0}
}

func (i *instance) receiveRequest(r *request) {
	if i.isWorking() == true {
		panic(fmt.Sprintf("Instances may not enqueue requests."))
	}
	i.req = r
	i.cond.Set(true)
}

func (i *instance) terminate() {
	i.terminateTime = godes.GetSystemTime()
	i.terminated = true
	i.cond.Set(true)
}

func (i *instance) next() (float64, int) {
	e := i.entries[i.index]
	i.index = (i.index + 1) % len(i.entries)
	return e.duration, e.status
}

func (i *instance) scaleDown() {
	i.terminate()
	i.terminateTime = i.getLastWorked() + i.idlenessDeadline.Seconds()
}

func (i *instance) Run() {
	for {
		i.cond.Wait(true)
		if i.isTerminated() {
			break
		}

		responseTime, status := i.next()
		i.req.status = status
		i.req.responseTime += responseTime 
		i.busyTime += responseTime
		godes.Advance(responseTime)
		i.lastWorked = godes.GetSystemTime()
		
		fmt.Printf("%d,%d,%.1f\n", i.req.id, i.req.status, i.req.responseTime*1000)

		i.cond.Set(false)
	}
}

func (i *instance) isWorking() bool {
	return i.cond.GetState() == true
}

func (i *instance) isTerminated() bool {
	return i.terminated
}

func (i *instance) getUpTime() float64 {
	return i.terminateTime - i.createdTime
}

func (i *instance) getIdleTime() float64 {
	return i.getUpTime() - i.getBusyTime()
}

func (i *instance) getBusyTime() float64 {
	return i.busyTime
}

func (i *instance) getLastWorked() float64 {
	return i.lastWorked
}

func (i *instance) getEfficiency() float64 {
	return i.getBusyTime() / i.getUpTime()
}

type request struct {
	id      int64
	responseTime float64
	status  int
}
