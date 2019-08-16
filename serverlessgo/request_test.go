package main

import (
	"reflect"
	"testing"
)

func TestHasBeenProcessed(t *testing.T) {
	var testData = []struct {
		desc      string
		req       *Request
		instances []int
		want      []bool
	}{
		{"EmptyHop", &Request{}, []int{1}, []bool{false}},
		{"OneHopTrue", &Request{hops: []int{1}}, []int{1}, []bool{true}},
		{"OneHopFalse", &Request{hops: []int{1}}, []int{2}, []bool{false}},
		{"ManyHopsTrue", &Request{hops: []int{1, 2, 3}}, []int{1, 2, 3}, []bool{true, true, true}},
		{"ManyHopsFalse", &Request{hops: []int{1, 2, 3}}, []int{5, 6, 7}, []bool{false, false, false}},
		{"ManyHopsMixed", &Request{hops: []int{1, 2}}, []int{1, 3, 2, 4}, []bool{true, false, true, false}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			for i, w := range d.want {
				got := d.req.hasBeenProcessed(d.instances[i])
				if !reflect.DeepEqual(w, got) {
					t.Fatalf("Want: %v, got: %v", w, got)
				}
			}
		})
	}
}

func TestUpdateHops(t *testing.T) {
	type data struct {
		desc  string
		req   *Request
		value int
		want  []int
	}
	var updateHop = []data{
		{"UpdateEmptyHop", &Request{}, 1, []int{1}},
		{"UpdateNotEmptyHop", &Request{hops: []int{0, 5}}, 2, []int{0, 5, 2}},
	}
	for _, d := range updateHop {
		t.Run(d.desc, func(t *testing.T) {
			d.req.updateHops(d.value)
			got := d.req.hops
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}

func TestUpdateStatus(t *testing.T) {
	type data struct {
		desc  string
		req   *Request
		value int
		want  int
	}
	var updateHop = []data{
		{"DefaultStatus", &Request{}, 503, 503},
		{"NotDefaultStatus", &Request{status: 503}, 200, 200},
	}
	for _, d := range updateHop {
		t.Run(d.desc, func(t *testing.T) {
			d.req.updateStatus(d.value)
			got := d.req.status
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}

func TestUpdateResponseTime(t *testing.T) {
	type data struct {
		desc  string
		req   *Request
		value float64
		want  float64
	}
	var updateHop = []data{
		{"DefaultResponse", &Request{}, 0.5, 0.5},
		{"NotDefaultResponse", &Request{responseTime: 1.5}, 0.8, 2.3},
	}
	for _, d := range updateHop {
		t.Run(d.desc, func(t *testing.T) {
			d.req.updateResponseTime(d.value)
			got := d.req.responseTime
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}
