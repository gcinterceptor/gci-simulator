package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/agoussia/godes"
	"github.com/stretchr/testify/assert"
)

func TestReceive_Hops(t *testing.T) {
	var testData = []struct {
		desc     string
		instance *instance
		request  *request
		want     []int
	}{
		{"UpdateEmptyHop", &instance{id: 1, cond: godes.NewBooleanControl()}, &request{hops: []int{}}, []int{1}},
		{"UpdateNotEmptyHop", &instance{id: 2, cond: godes.NewBooleanControl()}, &request{hops: []int{0, 1}}, []int{0, 1, 2}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			d.instance.receive(d.request)
			got := d.instance.req.hops
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}

func TestReceive_Panic(t *testing.T) {
	i := &instance{id: 1, cond: godes.NewBooleanControl()}
	i.cond.Set(true)
	assert.Panics(t, func() { i.receive(&request{}) }, "Already working instance did not panic receiving the request")
}

func TestTerminate(t *testing.T) {
	type triad struct {
		isTerminated, isWorking, terminatedTime interface{}
	}
	type data struct {
		desc     string
		instance *instance
		advance  float64
		want     *triad
	}
	var testData = []data{
		{"NoAdvance", &instance{
			Runner:      &godes.Runner{},
			createdTime: 0.0,
			cond:        godes.NewBooleanControl(),
		}, 0.0, &triad{true, false, 0.0}},
		{"AdvanceStartedAtZero", &instance{
			Runner:      &godes.Runner{},
			createdTime: 0.0,
			cond:        godes.NewBooleanControl(),
		}, 1.0, &triad{true, false, 1.0}},
		{"AdvanceStartedAtFive", &instance{
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
		instance *instance
		advance  float64
		want     *triad
	}
	idleness, _ := time.ParseDuration("5m")
	var testData = []data{
		{"NoAdvance", &instance{
			Runner:           &godes.Runner{},
			cond:             godes.NewBooleanControl(),
			createdTime:      0.0,
			lastWorked:       godes.GetSystemTime(),
			idlenessDeadline: idleness,
		}, 0.0, &triad{true, false, 300.0}},
		{"AdvanceStartedAtZero", &instance{
			Runner:           &godes.Runner{},
			cond:             godes.NewBooleanControl(),
			createdTime:      0.0,
			lastWorked:       godes.GetSystemTime(),
			idlenessDeadline: idleness,
		}, 1.0, &triad{true, false, 300.0}},
		{"AdvanceStartedAtFive", &instance{
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
		instance *instance
		want     []inputEntry
	}
	var testData = []data{
		{"RemovingWithOneEntry", &instance{entries: []inputEntry{{200, 0.2}}}, []inputEntry{{200, 0.2}}},
		{"RemovingWithManyEntries", &instance{entries: []inputEntry{{200, 0.3}, {200, 0.2}, {200, 0.1}}}, []inputEntry{{200, 0.2}, {200, 0.1}}},
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
		&instance{entries: []inputEntry{{200, 0.3}, {200, 0.2}, {200, 0.1}}},
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
