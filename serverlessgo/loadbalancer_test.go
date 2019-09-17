package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/agoussia/godes"
)

func TestFoward(t *testing.T) {
	lb := &LoadBalancer{
		arrivalQueue: godes.NewFIFOQueue("arrival"),
		arrivalCond:  godes.NewBooleanControl(),
	}
	type Want struct {
		queueSize         int
		arrivalCondBefore bool
		arrivalCondAfter  bool
		expectedError     bool
	}
	data := []struct{
		desc string
		req *Request
		want *Want
	 }{
	   {"Nil request", nil, &Want{0, false, false, true}},
	   {"First request", &Request{}, &Want{1, false, true, false}},
	   {"Following request", &Request{}, &Want{2, true, true, false}},
	 }
	 for _, d := range data {
		t.Run(d.desc, func(t *testing.T){
			arrivalCondBefore := lb.arrivalCond.GetState()
			err := lb.forward(d.req)
			expectedError := err != nil
			arrivalCondAfter := lb.arrivalCond.GetState()
			queueSized := lb.arrivalQueue.Len()
			got := &Want{queueSized, arrivalCondBefore, arrivalCondAfter, expectedError}
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	 }
}

type TestOutputWriter struct{}

func (t TestOutputWriter) record(r *Request) error { return nil }

func TestResponse(t *testing.T) {
	lb := &LoadBalancer{
		inputs: [][]inputEntry{{{200, 0.5}}},
		output: TestOutputWriter{},
	}
	type Want struct {
		responsed     int
		reforwarded   int
		expectedError bool
	}
	data := []struct{
		desc string
		req *Request
		want *Want
	 }{
	   {"Nil request", nil, &Want{0, 0, true}},
	   {"Success", &Request{status: 200}, &Want{1, 0, false}},
	   {"Unavailable", &Request{status: 503}, &Want{1, 1, false}},
	 }
	 for _, d := range data {
		t.Run(d.desc, func(t *testing.T){
			err := lb.response(d.req)
			expectedError := err != nil
			responsed := lb.finishedReqs
			reforwarded := len(lb.instances)
			got := &Want{responsed, reforwarded, expectedError}
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	 }
}

func TestLBTerminate(t *testing.T) {
	type TestData struct {
		desc string
		lb   *LoadBalancer
		want bool
	}
	var testData = []TestData{
		{"NoInstance", &LoadBalancer{
			arrivalCond: godes.NewBooleanControl(),
			instances:   make([]IInstance, 0),
		}, true},
		{"OneInstance", &LoadBalancer{
			arrivalCond: godes.NewBooleanControl(),
			instances:   []IInstance{&Instance{id: 0, cond: godes.NewBooleanControl()}},
		}, true},
		{"ManyInstances", &LoadBalancer{
			arrivalCond: godes.NewBooleanControl(),
			instances: []IInstance{
				&Instance{id: 1, cond: godes.NewBooleanControl()},
				&Instance{id: 2, cond: godes.NewBooleanControl()},
				&Instance{id: 3, cond: godes.NewBooleanControl()},
			},
		}, true},
	}
	checkFunc := func(want, got bool) {
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("Want: %v, got: %v", want, got)
		}
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
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

func TestNextInstance_HopedRequest(t *testing.T) {
	lb := &LoadBalancer{
		Runner:      &godes.Runner{},
		arrivalCond: godes.NewBooleanControl(),
		inputs:      [][]inputEntry{{{200, 0.5}}},
		instances: []IInstance{
			&Instance{id: 0, terminated: false, cond: godes.NewBooleanControl()},
			&Instance{id: 1, terminated: false, cond: godes.NewBooleanControl()},
			&Instance{id: 2, terminated: true, cond: godes.NewBooleanControl()},
			&Instance{id: 3, terminated: false, cond: godes.NewBooleanControl()},
		},
	}
	data := []struct{
		desc string
		req *Request
		want int
	 }{
	   {"Free Instance", &Request{}, 0},
	   {"Busy Instance", &Request{}, 1},
	   {"Terminated Instance", &Request{hops: []int{0, 1}}, 3},
	   {"New Instance Required", &Request{hops: []int{0, 1, 2, 3}}, 4},
	 }
	 for _, d := range data {
		t.Run(d.desc, func(t *testing.T){
			nextInstance := lb.nextInstance(d.req)
			nextInstance.receive(d.req)
			got := nextInstance.getId()
			if d.want != got {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	 }
}

type TestInstance struct {
	*Instance
	id         int
	terminated bool
	lastWorked float64
}

func (t *TestInstance) terminate()             { t.terminated = true }
func (t *TestInstance) scaleDown()             { t.terminated = true }
func (t *TestInstance) isTerminated() bool     { return t.terminated }
func (t *TestInstance) getLastWorked() float64 { return t.lastWorked }
func (t *TestInstance) getId() int             { return t.id }

func TestTryScaleDown(t *testing.T) {
	idleness, _ := time.ParseDuration("5s")
	type TestData struct {
		desc string
		lb   *LoadBalancer
		want []bool
	}
	var testData = []TestData{
		{"NoInstances", &LoadBalancer{
			idlenessDeadline: idleness,
			instances:        make([]IInstance, 0),
		}, make([]bool, 0)},
		{"OneInstance", &LoadBalancer{
			idlenessDeadline: idleness,
			instances:        []IInstance{&TestInstance{id: 0, terminated: false, lastWorked: -5.0}},
		}, []bool{true}},
		{"ManyInstances", &LoadBalancer{
			idlenessDeadline: idleness,
			instances: []IInstance{
				&TestInstance{id: 0, terminated: false, lastWorked: -5.0},
				&TestInstance{id: 1, terminated: false, lastWorked: 0.0},
				&TestInstance{id: 2, terminated: false, lastWorked: -5.0},
				&TestInstance{id: 3, terminated: false, lastWorked: -1.0},
				&TestInstance{id: 4, terminated: false, lastWorked: -8.0},
			},
		}, []bool{true, false, true, false, true}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			d.lb.tryScaleDown()
			got := make([]bool, 0)
			for _, i := range d.lb.instances {
				got = append(got, i.isTerminated())
			}
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}
