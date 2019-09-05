package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/agoussia/godes"
)

func TestReceive(t *testing.T) {
	instance := &Instance{id: 2, cond: godes.NewBooleanControl()}

	workingBefore := instance.isWorking()
	if workingBefore {
		t.Fatalf("Want: %v, got: %v", workingBefore, !workingBefore)
	}
	req := &Request{hops: []int{0, 1}}
	instance.receive(req)

	workingAfter := instance.isWorking()
	if !workingAfter {
		t.Fatalf("Want: %v, got: %v", !workingAfter, workingAfter)
	}

	wHops := []int{0, 1, 2}
	if !reflect.DeepEqual(wHops, req.hops) {
		t.Fatalf("Want: %v, got: %v", wHops, req.hops)
	}

	type TestData struct {
		desc     string
		instance *Instance
		req      *Request
		want     []int
	}
	var testData = []TestData{
		{"UpdateEmptyHop", &Instance{id: 1, cond: godes.NewBooleanControl()}, &Request{hops: []int{}}, []int{1}},
		{"UpdateNotEmptyHop", &Instance{id: 2, cond: godes.NewBooleanControl()}, &Request{hops: []int{0, 1}}, []int{0, 1, 2}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			flagBeforeWanted, flagBeforeGot := false, d.instance.isWorking()
			d.instance.receive(d.req)
			flagBeforeWanted, flagBeforeGot = true, d.instance.isWorking()

			want := struct {
				hops                  []int
				flagBefore, flagAfter bool
			}{d.want, flagBeforeWanted, flagBeforeWanted}
			got := struct {
				hops                  []int
				flagBefore, flagAfter bool
			}{d.instance.req.hops, flagBeforeGot, flagBeforeGot}
			if !reflect.DeepEqual(want, got) {
				t.Fatalf("Want: %v, got: %v", want, got)
			}
		})
	}
}

func TestInstanceTerminate(t *testing.T) {
	type Want struct {
		isTerminated, instance bool
		terminateTime          float64
	}
	type TestData struct {
		desc     string
		instance *Instance
		advance  float64
		want     *Want
	}
	var testData = []TestData{
		{"NoAdvance", &Instance{
			Runner:      &godes.Runner{},
			createdTime: 0.0,
			cond:        godes.NewBooleanControl(),
		}, 0.0, &Want{true, false, 0.0}},
		{"AdvanceStartedAtZero", &Instance{
			Runner:      &godes.Runner{},
			createdTime: 0.0,
			cond:        godes.NewBooleanControl(),
		}, 1.0, &Want{true, false, 1.0}},
		{"AdvanceStartedAfterZero", &Instance{
			Runner:      &godes.Runner{},
			createdTime: 1.0,
			cond:        godes.NewBooleanControl(),
		}, 1.5, &Want{true, false, 2.5}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			godes.Run()
			defer godes.Clear()

			godes.Advance(d.instance.createdTime)
			godes.AddRunner(d.instance)
			godes.Advance(d.advance)
			d.instance.terminate()

			godes.Advance(d.advance)
			d.instance.terminate()
			d.instance.scaleDown()
			godes.WaitUntilDone()

			got := &Want{d.instance.isTerminated(), d.instance.isWorking(), d.instance.terminateTime}
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}

func TestScaleDown(t *testing.T) {
	type Want struct {
		isTerminated, instance bool
		terminateTime          float64
	}
	type TestData struct {
		desc     string
		instance *Instance
		advance  float64
		want     *Want
	}
	idleness, _ := time.ParseDuration("5m")
	var testData = []TestData{
		{"NoAdvance", &Instance{
			Runner:           &godes.Runner{},
			cond:             godes.NewBooleanControl(),
			createdTime:      0.0,
			lastWorked:       godes.GetSystemTime(),
			idlenessDeadline: idleness,
		}, 0.0, &Want{true, false, 300.0}},
		{"AdvanceStartedAtZero", &Instance{
			Runner:           &godes.Runner{},
			cond:             godes.NewBooleanControl(),
			createdTime:      0.0,
			lastWorked:       godes.GetSystemTime(),
			idlenessDeadline: idleness,
		}, 1.0, &Want{true, false, 300.0}},
		{"AdvanceStarteAfterZero", &Instance{
			Runner:           &godes.Runner{},
			cond:             godes.NewBooleanControl(),
			createdTime:      1.0,
			lastWorked:       godes.GetSystemTime(),
			idlenessDeadline: idleness,
		}, 1.5, &Want{true, false, 300.0}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			godes.Run()
			defer godes.Clear()

			godes.Advance(d.instance.createdTime)
			godes.AddRunner(d.instance)
			godes.Advance(d.advance)
			d.instance.scaleDown()

			godes.Advance(d.advance)
			d.instance.scaleDown()
			d.instance.terminate()
			godes.WaitUntilDone()

			got := &Want{d.instance.isTerminated(), d.instance.isWorking(), d.instance.terminateTime}
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}

func TestNext(t *testing.T) {
	type TestData struct {
		desc     string
		instance *Instance
		want     []inputEntry
	}
	var testData = []TestData{
		{"RemovingWithOneEntry", &Instance{entries: []inputEntry{{200, 0.2}}}, []inputEntry{{200, 0.2}}},
		{"RemovingWithManyEntries", &Instance{entries: []inputEntry{{200, 0.3}, {200, 0.2}, {200, 0.1}}}, []inputEntry{{200, 0.2}, {200, 0.1}}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			for i := 0; i < 10; i++ {
				d.instance.next()
			}
			got := d.instance.entries
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
	d := TestData{
		"EntrySequenceSelection",
		&Instance{entries: []inputEntry{{200, 0.3}, {200, 0.2}, {200, 0.1}}},
		[]inputEntry{{200, 0.3}, {200, 0.2}, {200, 0.1}, {200, 0.2}, {200, 0.1}},
	}
	t.Run(d.desc, func(t *testing.T) {
		for _, w := range d.want {
			status, duration := d.instance.next()
			got := inputEntry{status, duration}
			if !reflect.DeepEqual(w, got) {
				t.Fatalf("Want: %v, got: %v", w, got)
			}
		}
	})
}

type TestLoadBalancer struct{ reqsResponsed int }

func (lb *TestLoadBalancer) response(r *Request) { lb.reqsResponsed++ }
func TestInstanceRun(t *testing.T) {
	type Want struct {
		count int
		reqs  []*Request
	}
	type TestData struct {
		desc     string
		instance *Instance
		advance  float64
		reqs     []*Request
		want     *Want
	}
	var testData = []TestData{
		{"OneEntry", &Instance{
			Runner:  &godes.Runner{},
			cond:    godes.NewBooleanControl(),
			entries: []inputEntry{{200, 0.3}},
			lb:      &TestLoadBalancer{},
		}, 0.5, []*Request{{id: 1}, {id: 2}, {id: 3}},
			&Want{3, []*Request{
				{1, 200, 0.3, []int{0}},
				{2, 200, 0.3, []int{0}},
				{3, 200, 0.3, []int{0}}}}},
		{"ManyEntries", &Instance{
			Runner:  &godes.Runner{},
			cond:    godes.NewBooleanControl(),
			entries: []inputEntry{{200, 0.8}, {200, 0.1}, {200, 0.2}, {200, 0.3}},
			lb:      &TestLoadBalancer{},
		}, 0.9, []*Request{{id: 1}, {id: 2}, {id: 3}, {id: 4}, {id: 5}, {id: 6}, {id: 7}, {id: 8}},
			&Want{8, []*Request{
				{1, 200, 0.8, []int{0}}, {2, 200, 0.1, []int{0}},
				{3, 200, 0.2, []int{0}}, {4, 200, 0.3, []int{0}},
				{5, 200, 0.1, []int{0}}, {6, 200, 0.2, []int{0}},
				{7, 200, 0.3, []int{0}}, {8, 200, 0.1, []int{0}}}}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			godes.Run()
			godes.AddRunner(d.instance)
			for _, r := range d.reqs {
				godes.Advance(d.advance)
				d.instance.receive(r)
			}
			godes.WaitUntilDone()
			godes.Clear()

			got := &Want{d.instance.lb.(*TestLoadBalancer).reqsResponsed, d.reqs}
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}
