package main

import (
	"reflect"
	"testing"
)

func testHasBeenProcessed(t *testing.T) {
	var testData = []struct {
		desc      string
		request   *request
		instances []int
		want      []bool
	}{
		{"EmptyHop", &request{}, []int{1}, []bool{true}},
		{"OneHopTrue", &request{hops: []int{1}}, []int{1}, []bool{true}},
		{"OneHopFalse", &request{hops: []int{1}}, []int{1}, []bool{true}},
		{"ManyHopsTrue", &request{hops: []int{1, 2, 3}}, []int{1, 2, 3}, []bool{true, true, true}},
		{"ManyHopsFalse", &request{hops: []int{1, 2, 3}}, []int{5, 6, 7}, []bool{false, false, false}},
		{"ManyHopsMixed", &request{hops: []int{1, 2}}, []int{1, 3, 2, 4}, []bool{true, false, true, false}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			for i, w := range d.want {
				got := d.request.hasBeenProcessed(d.instances[i])
				if !reflect.DeepEqual(w, got) {
					t.Fatalf("Want: %v+, got: %v+", w, got)
				}
			}
		})
	}
}
