package main

import (
	"reflect"
	"testing"

	"github.com/agoussia/godes"
	"github.com/stretchr/testify/assert"
)

func TestReceive_Success(t *testing.T) {
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

func TestTerminateAndScaleDown(t *testing.T) {
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
		{"NoAdvance", &instance{Runner: &godes.Runner{}, createdTime: 0.0, cond: godes.NewBooleanControl()}, 0.0, &triad{true, false, 0.0}},
		{"AdvanceStartedAtZero", &instance{Runner: &godes.Runner{}, createdTime: 0.0, cond: godes.NewBooleanControl()}, 1.0, &triad{true, false, 1.0}},
		{"AdvanceStartedAtFive", &instance{Runner: &godes.Runner{}, createdTime: 1.0, cond: godes.NewBooleanControl()}, 1.5, &triad{true, false, 2.5}},
	}
	check := func(d *data) {
		got := &triad{d.instance.isTerminated(), d.instance.isWorking(), d.instance.terminateTime}
		if !reflect.DeepEqual(d.want, got) {
			t.Fatalf("Want: %v, got: %v", d.want, got)
		}
	}
	for _, d := range testData {
		t.Run("Terminate"+d.desc, func(t *testing.T) {
			godes.Run()
			defer godes.Clear()

			godes.Advance(d.instance.createdTime)
			godes.AddRunner(d.instance)
			godes.Advance(d.advance)

			d.instance.terminate()
			godes.Advance(d.advance)
			d.instance.terminate()
			godes.WaitUntilDone()

			check(&d)
		})
	}
	for _, d := range testData {
		t.Run("ScaleDown"+d.desc, func(t *testing.T) {
			godes.Run()
			defer godes.Clear()

			godes.Advance(d.instance.createdTime)
			godes.AddRunner(d.instance)
			godes.Advance(d.advance)

			d.instance.scaleDown()
			godes.Advance(d.advance)
			d.instance.scaleDown()
			godes.WaitUntilDone()

			check(&d)
		})
	}
}
