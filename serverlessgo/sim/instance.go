package sim

import (
	"time"

	"github.com/agoussia/godes"
)

type iInstance interface {
	receive(r *Request)
	terminate()
	isWorking() bool
	isTerminated() bool
	getId() int
	getLastWorked() float64
	getUpTime() float64
	getEfficiency() float64
}

type instance struct {
	*godes.Runner
	id               int
	lb               iLoadBalancer
	terminated       bool
	cond             *godes.BooleanControl
	req              *Request
	createdTime      float64
	terminateTime    float64
	lastWorked       float64
	busyTime         float64
	idlenessDeadline time.Duration
	reproducer       iInputReproducer
	index            int
}

func newInstance(id int, lb iLoadBalancer, idlenessDeadline time.Duration, reproducer iInputReproducer) *instance {
	return &instance{
		Runner:           &godes.Runner{},
		lb:               lb,
		id:               id,
		cond:             godes.NewBooleanControl(),
		createdTime:      godes.GetSystemTime(),
		lastWorked:       godes.GetSystemTime(),
		idlenessDeadline: idlenessDeadline,
		reproducer:       reproducer,
	}
}

func (i *instance) receive(r *Request) {
	i.req = r
	i.req.updateHops(i.id)
	i.cond.Set(true)
}

func (i *instance) terminate() {
	if !i.terminated {
		if i.getLastWorked()+i.idlenessDeadline.Seconds() > godes.GetSystemTime() {
			i.terminateTime = godes.GetSystemTime()
		} else {
			i.terminateTime = i.getLastWorked() + i.idlenessDeadline.Seconds()
		}
		i.terminated = true
		i.cond.Set(true)
	}
}

func (i *instance) next() (int, float64) {
	return i.reproducer.next()
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
	return i.cond.GetState()
}

func (i *instance) isTerminated() bool {
	return i.terminated
}

func (i *instance) getId() int {
	return i.id
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
