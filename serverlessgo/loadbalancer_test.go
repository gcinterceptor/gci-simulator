package main

import (
	"reflect"
	"testing"

	"github.com/agoussia/godes"
)

func TestFoward(t *testing.T) {
	type Want struct {
		queueSize         int
		arrivalCondBefore bool
		arrivalCondAfter  bool
	}
	type TestData struct {
		desc string
		lb   *LoadBalancer
		reqs []*Request
		want *Want
	}
	var testData = []TestData{
		{"OneRequest", &LoadBalancer{
			arrivalQueue: godes.NewFIFOQueue("arrival"),
			arrivalCond:  godes.NewBooleanControl(),
		}, []*Request{{id: 0}}, &Want{1, false, true}},
		{"ManyRequests", &LoadBalancer{
			arrivalQueue: godes.NewFIFOQueue("arrival"),
			arrivalCond:  godes.NewBooleanControl(),
		}, []*Request{{id: 0}, {id: 1}, {id: 2}}, &Want{3, false, true}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			arrivalCondBefore := d.lb.arrivalCond.GetState()
			for _, r := range d.reqs {
				d.lb.foward(r)
			}
			got := &Want{d.lb.arrivalQueue.Len(), arrivalCondBefore, d.lb.arrivalCond.GetState()}
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}

type TestOutputWriter struct{}

func (t TestOutputWriter) record(s string) error { return nil }
func TestResponse(t *testing.T) {
	type Want struct {
		responsed   int
		reforwarded int
	}
	type TestData struct {
		desc string
		lb   *LoadBalancer
		reqs []*Request
		want *Want
	}
	var testData = []TestData{
		{"OneRequestReforwared", &LoadBalancer{
			inputs: [][]inputEntry{{{200, 0.5}}},
			output: TestOutputWriter{},
		}, []*Request{{status: 503}}, &Want{0, 1}},
		{"OneRequestResponsed", &LoadBalancer{
			inputs: [][]inputEntry{{{200, 0.5}}},
			output: TestOutputWriter{},
		}, []*Request{{status: 200}}, &Want{1, 0}},
		{"ManyRequestsReforwared", &LoadBalancer{
			inputs: [][]inputEntry{{{200, 0.5}}},
			output: TestOutputWriter{},
		}, []*Request{{status: 503}, {status: 503}, {status: 503}}, &Want{0, 3}},
		{"ManyRequestResponsed", &LoadBalancer{
			inputs: [][]inputEntry{{{200, 0.5}}},
			output: TestOutputWriter{},
		}, []*Request{{status: 200}, {status: 200}, {status: 200}}, &Want{3, 0}},
		{"ManyRequestMixed", &LoadBalancer{
			inputs: [][]inputEntry{{{200, 0.5}}},
			output: TestOutputWriter{},
		}, []*Request{{status: 200}, {status: 503}, {status: 503}, {status: 200}, {status: 200}}, &Want{3, 2}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			for _, r := range d.reqs {
				d.lb.response(r)
			}
			got := &Want{d.lb.finishedReqs, len(d.lb.instances)}
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}

func TestLBTerminate(t *testing.T) {
	type TestData struct {
		desc    string
		lb      *LoadBalancer
		advance float64
		want    bool
	}
	var testData = []TestData{
		{"NoInstance", &LoadBalancer{
			Runner:      &godes.Runner{},
			arrivalCond: godes.NewBooleanControl(),
			instances:   make([]IInstance, 0),
		}, 0.1, true},
		{"OneInstance", &LoadBalancer{
			Runner:      &godes.Runner{},
			arrivalCond: godes.NewBooleanControl(),
			instances:   []IInstance{&Instance{id: 0, cond: godes.NewBooleanControl()}},
		}, 0.1, true},
		{"ManyInstances", &LoadBalancer{
			Runner:      &godes.Runner{},
			arrivalCond: godes.NewBooleanControl(),
			instances: []IInstance{
				&Instance{id: 1, cond: godes.NewBooleanControl()},
				&Instance{id: 2, cond: godes.NewBooleanControl()},
				&Instance{id: 3, cond: godes.NewBooleanControl()},
			},
		}, 0.1, true},
	}
	checkFunc := func(want, got bool) {
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("Want: %v, got: %v", want, got)
		}
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			godes.AddRunner(d.lb)
			godes.Run()
			defer godes.Clear()

			d.lb.terminate()
			var got bool
			for _, i := range d.lb.instances {
				got = i.isTerminated()
				checkFunc(d.want, got)
			}
			got = d.lb.isTerminated
			checkFunc(d.want, got)
		})
	}
}

func TestNextInstanceInputs(t *testing.T) {
	type TestData struct {
		desc      string
		lb        *LoadBalancer
		nextCalls int
		want      [][]inputEntry
	}
	var testData = []TestData{
		{"OneInputEntry", &LoadBalancer{
			inputs: [][]inputEntry{{{200, 0.5}}},
		}, 2, [][]inputEntry{{{200, 0.5}}, {{200, 0.5}}}},
		{"ManyInputEntry", &LoadBalancer{
			inputs: [][]inputEntry{{{200, 0.5}, {503, 0.5}}, {}, {{200, 0.5}}},
		}, 5, [][]inputEntry{{{200, 0.5}, {503, 0.5}}, {}, {{200, 0.5}}, {{200, 0.5}, {503, 0.5}}, {}}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			for i := 0; i < d.nextCalls; i++ {
				got := d.lb.nextInstanceInputs()
				if !reflect.DeepEqual(d.want[i], got) {
					t.Fatalf("Want: %v, got: %v", d.want[i], got)
				}
			}
		})
	}
}

func TestNextInstance_NoHopedRequest(t *testing.T) {
	type Want struct{ firstID, secondID, thirdID int }
	type TestData struct {
		desc string
		lb   *LoadBalancer
		want *Want
	}
	var testData = []TestData{
		{"NoInstances", &LoadBalancer{
			Runner:      &godes.Runner{},
			arrivalCond: godes.NewBooleanControl(),
			inputs:      [][]inputEntry{{{200, 0.5}}},
			instances:   make([]IInstance, 0),
		}, &Want{0, 0, 1}},
		{"OneInstance", &LoadBalancer{
			Runner:      &godes.Runner{},
			arrivalCond: godes.NewBooleanControl(),
			inputs:      [][]inputEntry{{{200, 0.5}}},
			instances:   []IInstance{&Instance{id: 4, cond: godes.NewBooleanControl()}},
		}, &Want{4, 4, 1}},
		{"ManyAvailableInstances", &LoadBalancer{
			Runner:      &godes.Runner{},
			arrivalCond: godes.NewBooleanControl(),
			inputs:      [][]inputEntry{{{200, 0.5}}},
			instances: []IInstance{
				&Instance{id: 5, cond: godes.NewBooleanControl()},
				&Instance{id: 6, cond: godes.NewBooleanControl()},
				&Instance{id: 7, cond: godes.NewBooleanControl()},
			},
		}, &Want{5, 5, 6}},
		{"ManyTerminatedInstances", &LoadBalancer{
			Runner:      &godes.Runner{},
			arrivalCond: godes.NewBooleanControl(),
			inputs:      [][]inputEntry{{{200, 0.5}}},
			instances: []IInstance{
				&Instance{id: 0, terminated: true, cond: godes.NewBooleanControl()},
				&Instance{id: 1, terminated: true, cond: godes.NewBooleanControl()},
				&Instance{id: 2, terminated: false, cond: godes.NewBooleanControl()},
				&Instance{id: 3, terminated: true, cond: godes.NewBooleanControl()},
			},
		}, &Want{2, 2, 4}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			req := &Request{}
			first := d.lb.nextInstance(req)
			second := d.lb.nextInstance(req)
			first.receive(req)

			third := d.lb.nextInstance(req)
			got := &Want{first.getId(), second.getId(), third.getId()}
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}

func TestNextInstance_HopedRequest(t *testing.T) {
	type Want struct{ firstID, secondID, thirdID int }
	type TestData struct {
		desc string
		lb   *LoadBalancer
		req  *Request
		want *Want
	}
	var testData = []TestData{
		{"OneInstance", &LoadBalancer{
			Runner:      &godes.Runner{},
			arrivalCond: godes.NewBooleanControl(),
			inputs:      [][]inputEntry{{{200, 0.5}}},
			instances:   []IInstance{&Instance{id: 0, cond: godes.NewBooleanControl()}},
		}, &Request{hops: []int{0}}, &Want{1, 1, 2}},
		{"ManyInstances", &LoadBalancer{
			Runner:      &godes.Runner{},
			arrivalCond: godes.NewBooleanControl(),
			inputs:      [][]inputEntry{{{200, 0.5}}},
			instances: []IInstance{
				&Instance{id: 0, cond: godes.NewBooleanControl()},
				&Instance{id: 1, cond: godes.NewBooleanControl()},
				&Instance{id: 2, cond: godes.NewBooleanControl()},
				&Instance{id: 3, cond: godes.NewBooleanControl()},
				&Instance{id: 4, cond: godes.NewBooleanControl()},
			},
		}, &Request{hops: []int{0, 1, 3, 4}}, &Want{2, 2, 5}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			first := d.lb.nextInstance(d.req)
			second := d.lb.nextInstance(d.req)
			first.receive(d.req)

			third := d.lb.nextInstance(d.req)
			got := &Want{first.getId(), second.getId(), third.getId()}
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}

func TestTryScaleDown(t *testing.T) {}

func TestLBRun(t *testing.T) {}
