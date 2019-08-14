package main

import (
	"reflect"
	"testing"
)

func TestHasBeenProcessed(t *testing.T) {
	var testData = []struct {
		desc      string
		request   *request
		instances []int
		want      []bool
	}{
		{"EmptyHop", &request{}, []int{1}, []bool{false}},
		{"OneHopTrue", &request{hops: []int{1}}, []int{1}, []bool{true}},
		{"OneHopFalse", &request{hops: []int{1}}, []int{2}, []bool{false}},
		{"ManyHopsTrue", &request{hops: []int{1, 2, 3}}, []int{1, 2, 3}, []bool{true, true, true}},
		{"ManyHopsFalse", &request{hops: []int{1, 2, 3}}, []int{5, 6, 7}, []bool{false, false, false}},
		{"ManyHopsMixed", &request{hops: []int{1, 2}}, []int{1, 3, 2, 4}, []bool{true, false, true, false}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			for i, w := range d.want {
				got := d.request.hasBeenProcessed(d.instances[i])
				if !reflect.DeepEqual(w, got) {
					t.Fatalf("Want: %v, got: %v", w, got)
				}
			}
		})
	}
}

func TestUpdateHops(t *testing.T) {
	type data struct {
		desc    string
		request *request
		value   int
		want    []int
	}
	var updateHop = []data{
		{"UpdateEmptyHop", &request{}, 1, []int{1}},
		{"UpdateNotEmptyHop", &request{hops: []int{0, 5}}, 2, []int{0, 5, 2}},
	}
	for _, d := range updateHop {
		t.Run(d.desc, func(t *testing.T) {
			d.request.updateHops(d.value)
			got := d.request.hops
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}

func TestUpdateStatus(t *testing.T) {
	type data struct {
		desc    string
		request *request
		value   int
		want    int
	}
	var updateHop = []data{
		{"DefaultStatus", &request{}, 503, 503},
		{"NotDefaultStatus", &request{status: 503}, 200, 200},
	}
	for _, d := range updateHop {
		t.Run(d.desc, func(t *testing.T) {
			d.request.updateStatus(d.value)
			got := d.request.status
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}

func TestUpdateResponseTime(t *testing.T) {
	type data struct {
		desc    string
		request *request
		value   float64
		want    float64
	}
	var updateHop = []data{
		{"DefaultResponse", &request{}, 0.5, 0.5},
		{"NotDefaultResponse", &request{responseTime: 1.5}, 0.8, 2.3},
	}
	for _, d := range updateHop {
		t.Run(d.desc, func(t *testing.T) {
			d.request.updateResponseTime(d.value)
			got := d.request.responseTime
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}
