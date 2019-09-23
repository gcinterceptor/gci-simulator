package sim

import (
	"errors"
	"sort"
	"time"

	"github.com/agoussia/godes"
)

type iLoadBalancer interface {
	forward(r *Request) error
	response(r *Request) error
}

type loadBalancer struct {
	*godes.Runner
	isTerminated       bool
	arrivalQueue       *godes.FIFOQueue
	arrivalCond        *godes.BooleanControl
	instances          []iInstance
	idlenessDeadline   time.Duration
	inputs             [][]InputEntry
	index              int
	listener           Listener
	finishedReqs       int
	optimizedScheduler bool
}

func newLoadBalancer(idlenessDeadline time.Duration, inputs [][]InputEntry, listener Listener, optimized bool) *loadBalancer {
	return &loadBalancer{
		Runner:             &godes.Runner{},
		arrivalQueue:       godes.NewFIFOQueue("arrival"),
		arrivalCond:        godes.NewBooleanControl(),
		instances:          make([]iInstance, 0),
		idlenessDeadline:   idlenessDeadline,
		inputs:             inputs,
		listener:           listener,
		optimizedScheduler: optimized,
	}
}

func (lb *loadBalancer) forward(r *Request) error {
	if r == nil {
		return errors.New("Error while calling the LB's forward method. Request cannot be nil.")
	}
	lb.arrivalQueue.Place(r)
	lb.arrivalCond.Set(true)
	return nil
}

func (lb *loadBalancer) response(r *Request) error {
	if r == nil {
		return errors.New("Error while calling the LB's response method. Request cannot be nil.")
	}
	if r.Status == 200 {
		lb.listener.RequestFinished(r)
		lb.finishedReqs++
	} else {
		lb.nextInstance(r).receive(r)
	}
	return nil
}

func (lb *loadBalancer) terminate() {
	if !lb.isTerminated {
		for _, i := range lb.instances {
			i.terminate()
		}
		lb.isTerminated = true
		lb.arrivalCond.Set(true)
	}
}

func (lb *loadBalancer) nextInstanceInputs() []InputEntry {
	input := lb.inputs[lb.index]
	lb.index = (lb.index + 1) % len(lb.inputs)
	return input
}

func (lb *loadBalancer) nextInstance(r *Request) iInstance {
	var selected iInstance
	// sorting instances to have the most recently used ones ahead on the array
	sort.SliceStable(lb.instances, func(i, j int) bool { return lb.instances[i].getLastWorked() > lb.instances[j].getLastWorked() })
	for _, i := range lb.instances {
		if !i.isWorking() && !i.isTerminated() && !r.hasBeenProcessed(i.getId()) {
			selected = i
			break
		}
	}
	if selected == nil {
		selected = lb.newInstance(r)
	}
	return selected
}

func (lb *loadBalancer) newInstance(r *Request) iInstance {
	var reproducer iInputReproducer
	nextInstanceInput := lb.nextInstanceInputs()
	if lb.optimizedScheduler && r.Status != 503 {
		reproducer = newWarmedInputReproducer(nextInstanceInput)

	} else {
		reproducer = newInputReproducer(nextInstanceInput)
	}
	newInstance := newInstance(len(lb.instances), lb, lb.idlenessDeadline, reproducer)
	godes.AddRunner(newInstance)
	// inserts the instance ahead of the array
	lb.instances = append([]iInstance{newInstance}, lb.instances...)
	return newInstance

}

func (lb *loadBalancer) Run() {
	for {
		lb.arrivalCond.Wait(true)
		lb.tryScaleDown()
		if lb.arrivalQueue.Len() > 0 {
			r := lb.arrivalQueue.Get().(*Request)
			lb.nextInstance(r).receive(r)
		} else {
			lb.arrivalCond.Set(false)
			if lb.isTerminated {
				break
			}
		}
	}
}

func (lb *loadBalancer) tryScaleDown() {
	for _, i := range lb.instances {
		if !i.isWorking() && godes.GetSystemTime()-i.getLastWorked() >= lb.idlenessDeadline.Seconds() {
			i.terminate()
		}
	}
}

func (lb *loadBalancer) getFinishedReqs() int {
	return lb.finishedReqs
}

func (lb *loadBalancer) getTotalCost() float64 {
	var totalCost float64
	for _, i := range lb.instances {
		totalCost += i.getUpTime()
	}
	return totalCost
}

func (lb *loadBalancer) getTotalEfficiency() float64 {
	var totalEfficiency float64
	for _, i := range lb.instances {
		totalEfficiency += i.getEfficiency()
	}
	return totalEfficiency / float64(len(lb.instances))
}
