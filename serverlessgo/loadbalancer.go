package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/agoussia/godes"
)

type loadBalancer struct {
	*godes.Runner
	isTerminated     bool
	arrivalQueue     *godes.FIFOQueue
	arrivalCond      *godes.BooleanControl
	instances        []*instance
	idlenessDeadline time.Duration
	inputs           [][]inputEntry
	index            int
}

func newLoadBalancer(idlenessDeadline time.Duration, inputs [][]inputEntry) *loadBalancer {
	return &loadBalancer{
		Runner:           &godes.Runner{},
		arrivalQueue:     godes.NewFIFOQueue("arrival"),
		arrivalCond:      godes.NewBooleanControl(),
		instances:        make([]*instance, 0),
		idlenessDeadline: idlenessDeadline,
		inputs:           inputs,
	}
}

func (lb *loadBalancer) foward(r *request) {
	lb.arrivalQueue.Place(r)
	lb.arrivalCond.Set(true)
}

func (lb *loadBalancer) response(r *request) {
	if r.status == 200 {
		fmt.Printf("%d,%d,%.1f\n", r.id, r.status, r.responseTime*1000)
	} else {
		lb.nextInstance(r).receive(lb, r)
	}
}

func (lb *loadBalancer) terminate() {
	for _, i := range lb.instances {
		i.terminate()
	}
	lb.isTerminated = true
	lb.arrivalCond.Set(true)
}

func (lb *loadBalancer) nextInstanceInputs() []inputEntry {
	input := lb.inputs[lb.index]
	lb.index = (lb.index + 1) % len(lb.inputs)
	return input
}

func contains(ids []int, id int) bool {
	for _, i := range ids {
		if id == i {
			return true
		}
	}
	return false
}

func (lb *loadBalancer) nextInstance(r *request) *instance {
	var selected *instance
	// sorting instances to have the most recently used ones ahead on the array
	sort.SliceStable(lb.instances, func(i, j int) bool { return lb.instances[i].getLastWorked() > lb.instances[j].getLastWorked() })
	for i := 0; i < len(lb.instances); i++ {
		instance := lb.instances[i]
		if !instance.isWorking() && !instance.isTerminated() && !contains(r.hops, instance.id) {
			selected = instance
			break
		}
	}
	if selected == nil {
		selected = newInstance(len(lb.instances), lb, lb.idlenessDeadline, lb.nextInstanceInputs())
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
			lb.nextInstance(r).receive(r)
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
		if godes.GetSystemTime()-i.getLastWorked() >= lb.idlenessDeadline.Seconds() {
			i.scaleDown()
		}
	}
}
