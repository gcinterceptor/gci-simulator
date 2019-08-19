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
		desc    string
		lb      *LoadBalancer
		reqs     []*Request
		want    *Want
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

type TestOutputWriter struct {
	count int
}

func (t TestOutputWriter) record(s string) error { t.count++; return nil}
func TestResponse(t *testing.T) {
	type Want struct {
		responsed    int
		reforwarded int
	}
	type TestData struct {
		desc    string
		lb      *LoadBalancer
		reqs     []*Request
		want    *Want
	}
	var testData = []TestData{
		{"OneRequestReforwared", &LoadBalancer{output: &TestOutputWriter{}}, []*Request{{status: 503}}, &Want{0, 1}},
		{"OneRequestResponsed", &LoadBalancer{output: &TestOutputWriter{}}, []*Request{{status: 200}}, &Want{1, 0}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			for _, r := range d.reqs {
				d.lb.response(r)
			}
			got := &Want{d.lb.output.(TestOutputWriter).count, len(d.lb.instances)}
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}

func TestLBTerminate(t *testing.T) {}

func TestNextInstanceInputs(t *testing.T) {}

func TestNextInstance(t *testing.T) {}

func TestTryScaleDown(t *testing.T) {}

func TestLBRun(t *testing.T) {}
