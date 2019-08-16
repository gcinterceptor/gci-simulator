package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/agoussia/godes"
	"github.com/stretchr/testify/assert"
)

func TestReceive(t *testing.T) {
	var testData = []struct {
		desc     string
		instance *Instance
		req      *Request
		want     []int
	}{
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

func TestReceive_Panic(t *testing.T) {
	i := &Instance{id: 1, cond: godes.NewBooleanControl()}
	i.cond.Set(true)
	assert.Panics(t, func() { i.receive(&Request{}) }, "Already working instance did not panic receiving the request")
}

func TestInstanceTerminate(t *testing.T) {
	type triad struct {
		isTerminated, isWorking, terminatedTime interface{}
	}
	type data struct {
		desc     string
		instance *Instance
		advance  float64
		want     *triad
	}
	var testData = []data{
		{"NoAdvance", &Instance{
			Runner:      &godes.Runner{},
			createdTime: 0.0,
			cond:        godes.NewBooleanControl(),
		}, 0.0, &triad{true, false, 0.0}},
		{"AdvanceStartedAtZero", &Instance{
			Runner:      &godes.Runner{},
			createdTime: 0.0,
			cond:        godes.NewBooleanControl(),
		}, 1.0, &triad{true, false, 1.0}},
		{"AdvanceStartedAtFive", &Instance{
			Runner:      &godes.Runner{},
			createdTime: 1.0,
			cond:        godes.NewBooleanControl(),
		}, 1.5, &triad{true, false, 2.5}},
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

			got := &triad{d.instance.isTerminated(), d.instance.isWorking(), d.instance.terminateTime}
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}

func TestScaleDown(t *testing.T) {
	type triad struct {
		isTerminated, isWorking, terminatedTime interface{}
	}
	type data struct {
		desc     string
		instance *Instance
		advance  float64
		want     *triad
	}
	idleness, _ := time.ParseDuration("5m")
	var testData = []data{
		{"NoAdvance", &Instance{
			Runner:           &godes.Runner{},
			cond:             godes.NewBooleanControl(),
			createdTime:      0.0,
			lastWorked:       godes.GetSystemTime(),
			idlenessDeadline: idleness,
		}, 0.0, &triad{true, false, 300.0}},
		{"AdvanceStartedAtZero", &Instance{
			Runner:           &godes.Runner{},
			cond:             godes.NewBooleanControl(),
			createdTime:      0.0,
			lastWorked:       godes.GetSystemTime(),
			idlenessDeadline: idleness,
		}, 1.0, &triad{true, false, 300.0}},
		{"AdvanceStartedAtFive", &Instance{
			Runner:           &godes.Runner{},
			cond:             godes.NewBooleanControl(),
			createdTime:      1.0,
			lastWorked:       godes.GetSystemTime(),
			idlenessDeadline: idleness,
		}, 1.5, &triad{true, false, 300.0}},
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

			got := &triad{d.instance.isTerminated(), d.instance.isWorking(), d.instance.terminateTime}
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}

func TestNext(t *testing.T) {
	type data struct {
		desc     string
		instance *Instance
		want     []inputEntry
	}
	var testData = []data{
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
	d := data{
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

func TestInstanceRun(t *testing.T) {

}
