package main

import (
	"sort"
	"time"
	"errors"

	"github.com/agoussia/godes"
)

type ILoadBalancer interface {
	forward(r *Request) error
	response(r *Request) error
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

func (lb *LoadBalancer) forward(r *Request) error {
	if r == nil {
		return errors.New("Error while calling the LB's forward method. Request cannot be nil.")
	}
	lb.arrivalQueue.Place(r)
	lb.arrivalCond.Set(true)
	return nil
}

func (lb *LoadBalancer) response(r *Request) error {
	if r == nil {
		return errors.New("Error while calling the LB's response method. Request cannot be nil.")
	}
	if r.status == 200 {
		lb.output.record(r)
		lb.finishedReqs++
	} else {
		lb.nextInstance(r).receive(r)
	}
	return nil
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
	var reproducer IInputReproducer
	nextInstanceInput := lb.nextInstanceInputs()
	if lb.optimizedScheduler && r.status != 503 {
		reproducer = newWarmedInputReproducer(nextInstanceInput)

	} else {
		reproducer = newInputReproducer(nextInstanceInput)
	}
	newInstance := newInstance(len(lb.instances), lb, lb.idlenessDeadline, reproducer)
	godes.AddRunner(newInstance)
	// inserts the instance ahead of the array
	lb.instances = append([]IInstance{newInstance}, lb.instances...)
	return newInstance

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
