package main

import (
	"sort"
	"time"

	"github.com/agoussia/godes"
)

type ILoadBalancer interface {
	response(r *Request)
}

type LoadBalancer struct {
	*godes.Runner
	isTerminated       bool
	arrivalQueue       *godes.FIFOQueue
	arrivalCond        *godes.BooleanControl
	instances          []IInstance
	idlenessDeadline   time.Duration
	inputs             [][]inputEntry
	index              int
	output             IOutputWriter
	finishedReqs       int
	optimizedScheduler bool
}

func newLoadBalancer(idlenessDeadline time.Duration, inputs [][]inputEntry, output IOutputWriter, optimized bool) *LoadBalancer {
	return &LoadBalancer{
		Runner:             &godes.Runner{},
		arrivalQueue:       godes.NewFIFOQueue("arrival"),
		arrivalCond:        godes.NewBooleanControl(),
		instances:          make([]IInstance, 0),
		idlenessDeadline:   idlenessDeadline,
		inputs:             inputs,
		output:             output,
		optimizedScheduler: optimized,
	}
}

func (lb *LoadBalancer) foward(r *Request) {
	lb.arrivalQueue.Place(r)
	lb.arrivalCond.Set(true)
}

func (lb *LoadBalancer) response(r *Request) {
	if r.status == 200 {
		lb.output.record(r)
		lb.finishedReqs++
	} else {
		lb.nextInstance(r).receive(r)
	}
}

func (lb *LoadBalancer) terminate() {
	if !lb.isTerminated {
		for _, i := range lb.instances {
			i.terminate()
		}
		lb.isTerminated = true
		lb.arrivalCond.Set(true)
	}
}

func (lb *LoadBalancer) nextInstanceInputs() []inputEntry {
	input := lb.inputs[lb.index]
	lb.index = (lb.index + 1) % len(lb.inputs)
	return input
}

func (lb *LoadBalancer) nextInstance(r *Request) IInstance {
	var selected IInstance
	// sorting instances to have the most recently used ones ahead on the array
	sort.SliceStable(lb.instances, func(i, j int) bool { return lb.instances[i].getLastWorked() > lb.instances[j].getLastWorked() })
	for i := 0; i < len(lb.instances); i++ {
		instance := lb.instances[i]
		if !instance.isWorking() && !instance.isTerminated() && !r.hasBeenProcessed(instance.getId()) {
			selected = instance
			break
		}
	}
	if selected == nil {
		selected = lb.newInstance(r)
	}
	return selected
}

func (lb *LoadBalancer) newInstance(r *Request) IInstance {
	var newIInstance IInstance
	if lb.optimizedScheduler {
		nextInstanceInput := lb.nextInstanceInputs()
		if r.status != 503 { // GCI receives cold start impacts even on the optimized scheduler
			nextInstanceInput = nextInstanceInput[1:]
		}
		newInstance := newInstance(len(lb.instances), lb, lb.idlenessDeadline, nextInstanceInput)
		godes.AddRunner(newInstance)
		// inserts the instance ahead of the array
		lb.instances = append([]IInstance{newInstance}, lb.instances...)
		newIInstance = newInstance

	} else {
		nextInstanceInput := lb.nextInstanceInputs()
		newInstance := newInstance(len(lb.instances), lb, lb.idlenessDeadline, nextInstanceInput)
		godes.AddRunner(newInstance)
		// inserts the instance ahead of the array
		lb.instances = append([]IInstance{newInstance}, lb.instances...)
		newIInstance = newInstance
	}
	return newIInstance

}

func (lb *LoadBalancer) Run() {
	for {
		lb.arrivalCond.Wait(true)
		if lb.arrivalQueue.Len() > 0 {
			r := lb.arrivalQueue.Get().(*Request)
			lb.nextInstance(r).receive(r)
		} else {
			lb.arrivalCond.Set(false)
			if lb.isTerminated {
				break
			}
		}
		lb.tryScaleDown()
	}
}

func (lb *LoadBalancer) tryScaleDown() {
	for _, i := range lb.instances {
		if godes.GetSystemTime()-i.getLastWorked() >= lb.idlenessDeadline.Seconds() {
			i.scaleDown()
		}
	}
}

func (lb *LoadBalancer) getFinishedReqs() int {
	return lb.finishedReqs
}

func (lb *LoadBalancer) getTotalCost() float64 {
	var totalCost float64
	for _, i := range lb.instances {
		totalCost += i.getUpTime()
	}
	return totalCost
}

func (lb *LoadBalancer) getTotalEfficiency() float64 {
	var totalEfficiency float64
	for _, i := range lb.instances {
		totalEfficiency += i.getEfficiency()
	}
	return totalEfficiency / float64(len(lb.instances))
}
