package main

import (
	"time"

	"github.com/agoussia/godes"
)

type IInstance interface {
	receive(r *Request)
	terminate()
	scaleDown()
	isWorking() bool
	isTerminated() bool
	getId() int
	getLastWorked() float64
	getUpTime() float64
	getEfficiency() float64
}

type Instance struct {
	*godes.Runner
	id               int
	lb               ILoadBalancer
	terminated       bool
	cond             *godes.BooleanControl
	req              *Request
	createdTime      float64
	terminateTime    float64
	lastWorked       float64
	busyTime         float64
	idlenessDeadline time.Duration
	reproducer       IInputReproducer
	index            int
}

func newInstance(id int, lb ILoadBalancer, idlenessDeadline time.Duration, reproducer IInputReproducer) *Instance {
	return &Instance{
		Runner:           &godes.Runner{},
		lb:               lb,
		id:               id,
		cond:             godes.NewBooleanControl(),
		createdTime:      godes.GetSystemTime(),
		lastWorked:       godes.GetSystemTime(),
		idlenessDeadline: idlenessDeadline,
		reproducer:          reproducer,
	}
}

func (i *Instance) receive(r *Request) {
	i.req = r
	i.req.updateHops(i.id)
	i.cond.Set(true)
}

func (i *Instance) terminate() {
	if !i.terminated {
		i.terminateTime = godes.GetSystemTime()
		i.terminated = true
		i.cond.Set(true)
	}
}

func (i *Instance) scaleDown() {
	if !i.terminated {
		i.terminate()
		i.terminateTime = i.getLastWorked() + i.idlenessDeadline.Seconds()
	}
}

func (i *Instance) next() (int, float64) {
	return i.reproducer.next()
}

func (i *Instance) Run() {
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

func (i *Instance) isWorking() bool {
	return i.cond.GetState()
}

func (i *Instance) isTerminated() bool {
	return i.terminated
}

func (i *Instance) getId() int {
	return i.id
}

func (i *Instance) getUpTime() float64 {
	return i.terminateTime - i.createdTime
}

func (i *Instance) getIdleTime() float64 {
	return i.getUpTime() - i.getBusyTime()
}

func (i *Instance) getBusyTime() float64 {
	return i.busyTime
}

func (i *Instance) getLastWorked() float64 {
	return i.lastWorked
}

func (i *Instance) getEfficiency() float64 {
	return i.getBusyTime() / i.getUpTime()
}
