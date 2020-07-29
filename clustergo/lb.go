package main

import (
	"fmt"

	"github.com/agoussia/godes"
)

func newLoadBalancer(servers []*server, ht float64, hc bool) *loadBalancer {
	lb := &loadBalancer{
		Runner:        &godes.Runner{},
		isTerminated:  false,
		servers:       godes.NewFIFOQueue("servers"),
		serversBusy:   godes.NewBooleanControl(),
		ht:            ht,
		hCancellation: hc,
		hedgeMap:      make(map[int64]struct{}),
		hedgedReqs:    godes.NewFIFOQueue("hedgedReqs"),
	}
	for _, s := range servers {
		s.lb = lb
		lb.servers.Place(s)
	}
	return lb
}

type loadBalancer struct {
	*godes.Runner
	servers       *godes.FIFOQueue
	serversBusy   *godes.BooleanControl
	isTerminated  bool
	ht            float64
	hCancellation bool

	// Queues and control of request processing flow.
	hedgedReqs *godes.FIFOQueue
	hedgeMap   map[int64]struct{}

	// Metrics
	nIgnored        int64
	nTerminatedSucc int64
	nTerminatedFail int64
	nHedged         int64
	nProc           int64
	hedgingWaist    float64
	nCancelled      int64
}

func (lb *loadBalancer) Run() {
	// id, status, start time, finish time, response time, server id, copy, waist, cancelled
	fmt.Println("id,status,ts,ft,rt,sID,hedge,waist,canc")
	for {
		if lb.isTerminated {
			break
		}

		// First wait for a server and process hedged requests.
		lb.serversBusy.Wait(false)
		if lb.servers.Len() == 0 { // needed due a nil pointer ref at schedule that I couldn't spot the cause.
			godes.Yield()
			continue
		}

		if lb.hedgedReqs.Len() > 0 {
			r := lb.hedgedReqs.Get().(*request)
			lb.schedule(r)
			continue
		}

		// Then wait for a request that arrived.
		arrivalCond.Wait(true)

		// Ignoring arrived requests if there is available server.
		r := arrivalQueue.Get().(*request)
		if arrivalQueue.Len() == 0 {
			arrivalCond.Set(false)
		}
		lb.schedule(r)
	}
}

func (lb *loadBalancer) schedule(r *request) {
	// The model which computes the emergent behavior of CTC (for PVN) does not account for
	// the time in queue. This is equivalent to ignore requests that arrive and all server
	// were busy
	if !r.hedged {
		r.startTime = godes.GetSystemTime()
	}
	lb.servers.Get().(*server).newRequest(r)
	if lb.servers.Len() == 0 {
		lb.serversBusy.Set(true)
	}
}

func (lb *loadBalancer) computeReqMetrics(r *request) {
	// Metrics
	lb.nProc++
	switch {
	case r.status == 200 && !r.waist:
		lb.nTerminatedSucc++
	case r.status == 503:
		lb.nTerminatedFail++
	}
	if r.hedged {
		lb.nHedged++
	}
	if r.cancel {
		lb.nCancelled++
	}
	if r.waist {
		lb.hedgingWaist += r.rt
	}
}

func print(r *request) {
	fmt.Printf("%d,%d,%.2f,%.2f,%.2f,%d,%t,%t,%t\n", r.id, r.status, r.startTime, r.finishTime, r.rt, r.sid, r.hedged, r.waist, r.cancel)
}

func (lb *loadBalancer) reqCancelled(r *request) {
	defer lb.computeReqMetrics(r)
	defer print(r)

	r.finishTime = godes.GetSystemTime()
	r.rt = r.finishTime - r.startTime
	r.finish.Set(true)
	r.status = 0
	r.cancel = true
	r.waist = true
	if r.finish != nil {
		r.finish.Set(true)
	}
}

func (lb *loadBalancer) reqFinished(r *request) {
	defer lb.computeReqMetrics(r)
	defer print(r)

	r.finishTime = godes.GetSystemTime()
	r.rt = r.finishTime - r.startTime
	r.finish.Set(true)

	// Hedging is not active.
	if lb.ht < 0 {
		return
	}

	// The cancellation policy is active.
	if lb.hCancellation {
		return
	}

	// First finished copy of this hedge request.
	if _, ok := lb.hedgeMap[r.id]; !ok {
		r.waist = false
		r.cancel = false
		lb.hedgeMap[r.id] = struct{}{}
		return
	}

	// Second finished copy of this hedge request request.
	r.cancel = false
	r.waist = true
	delete(lb.hedgeMap, r.id)
}

func (lb *loadBalancer) terminate() {
	arrivalCond.Set(true)
	lb.isTerminated = true
	lb.serversBusy.Set(false)
}
