package sim

import (
	"strconv"
	"strings"
	"time"

	"github.com/agoussia/godes"
)

type IInstance interface {
	receive(r *Request)
	terminate()
	isWorking() bool
	isTerminated() bool
	isAvailable() bool
	getLastWorked() float64
	GetId() int
	GetUpTime() float64
	GetEfficiency() float64
	GetCreatedTime() float64
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
	tsAvailableAt    float64   // TimeStamp when the instance becomes available
	shedRT           []float64 // RT, Response Time
	shedRTIndex      int
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
	if !i.isTerminated() {
		if i.getLastWorked()+i.idlenessDeadline.Seconds() > godes.GetSystemTime() {
			i.terminateTime = godes.GetSystemTime()
		} else {
			i.terminateTime = i.getLastWorked() + i.idlenessDeadline.Seconds()
		}
		i.terminated = true
		i.cond.Set(true)
	}
}

func (i *instance) next() (int, float64, string, float64, float64) {
	return i.reproducer.next()
}

func (i *instance) nextShed() (int, float64) {
	status := 503
	ResponseTime := i.shedRT[i.shedRTIndex] / 1000000000
	i.shedRTIndex = (i.shedRTIndex + 1) % len(i.shedRT)
	return status, ResponseTime
}

func (i *instance) Run() {
	for {
		i.cond.Wait(true)
		if i.isTerminated() {
			i.cond.Set(false)
			break
		}
		var status int
		var responseTime float64
		if i.isAvailable() {
			var body string
			var tsbefore, tsafter float64
			status, responseTime, body, tsbefore, tsafter = i.next()
			if status == 503 {
				unavailableTime := tsafter - tsbefore
				i.tsAvailableAt = godes.GetSystemTime() + unavailableTime
				i.shedRT = append(i.shedRT, responseTime)
				rts := strings.Split(body, ":")
				for j := 1; j < len(rts); j++ {
					rtFloat64, err := strconv.ParseFloat(rts[j], 64)
					if err == nil {
						i.shedRT = append(i.shedRT, rtFloat64)
					}
				}
			}
		} else {
			status, responseTime = i.nextShed()
		}
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

func (i *instance) isAvailable() bool {
	return godes.GetSystemTime() >= i.tsAvailableAt
}

func (i *instance) GetId() int {
	return i.id
}

func (i *instance) GetUpTime() float64 {
	return i.terminateTime - i.createdTime
}

func (i *instance) getIdleTime() float64 {
	return i.GetUpTime() - i.getBusyTime()
}

func (i *instance) getBusyTime() float64 {
	return i.busyTime
}

func (i *instance) getLastWorked() float64 {
	return i.lastWorked
}

func (i *instance) GetEfficiency() float64 {
	return i.getBusyTime() / i.GetUpTime()
}

func (i *instance) GetCreatedTime() float64 {
	return i.createdTime
}
