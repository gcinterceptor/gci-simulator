package main

import (
	"fmt"
	"time"

	"github.com/agoussia/godes"
)

type instance struct {
	*godes.Runner
	id               int
	lb               *loadBalancer
	terminated       bool
	cond             *godes.BooleanControl
	req              *request
	createdTime      float64
	terminateTime    float64
	lastWorked       float64
	busyTime         float64
	idlenessDeadline time.Duration
	entries          []inputEntry
	index            int
	warmed           bool
}

func newInstance(id int, lb *loadBalancer, idlenessDeadline time.Duration, input []inputEntry) *instance {
	return &instance{
		Runner:           &godes.Runner{},
		lb:               lb,
		id:               id,
		cond:             godes.NewBooleanControl(),
		createdTime:      godes.GetSystemTime(),
		lastWorked:       godes.GetSystemTime(),
		idlenessDeadline: idlenessDeadline,
		entries:          input,
	}
}

func (i *instance) receive(r *request) {
	if i.isWorking() == true {
		panic(fmt.Sprintf("Instances may not enqueue requests."))
	}
	i.req = r
	i.req.updateHops(i.id)
	i.cond.Set(true)
}

func (i *instance) terminate() {
	if !i.terminated {
		i.terminateTime = godes.GetSystemTime()
		i.terminated = true
		i.cond.Set(true)
	}
}

func (i *instance) scaleDown() {
	if !i.terminated {
		i.terminate()
		i.terminateTime = i.getLastWorked() + i.idlenessDeadline.Seconds()
	}
}

func (i *instance) next() (int, float64) {
	e := i.entries[i.index]
	i.index = (i.index + 1) % len(i.entries)
	if !i.isWarm() {
		i.warmed = true
		if len(i.entries) > 1 {
			i.entries = i.entries[1:] // remove first entry
			i.index = 0
		}
	}
	return e.status, e.duration
}

func (i *instance) Run() {
	for {
		i.cond.Wait(true)
		if i.isTerminated() {
			i.cond.Set(false)
			break
		}
		status, responseTime := i.next()
		i.req.updateStatus(status)
		i.req.updateResponseTime(responseTime)
		i.busyTime += responseTime

		godes.Advance(responseTime)
		i.lastWorked = godes.GetSystemTime()
		i.lb.response(i.req)

		i.cond.Set(false)
	}
}

func (i *instance) isWorking() bool {
	return i.cond.GetState() == true
}

func (i *instance) isTerminated() bool {
	return i.terminated
}

func (i *instance) isWarm() bool {
	return i.warmed
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
