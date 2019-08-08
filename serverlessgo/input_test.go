package main

import (
	"reflect"
	"strings"
	"testing"
)

const (
	input_test = "test.csv"
)

func TestBuildEntryArray_Sucess(t *testing.T) {
	var testData = []struct {
		desc string
		row  [][]string
		want []inputEntry
	}{
		{"OneEntry", [][]string{{"503", "0.250"}}, []inputEntry{{503, 0.250}}},
		{"ManyEntries", [][]string{{"200", "0.019"}, {"503", "0.250"}}, []inputEntry{{200, 0.019}, {503, 0.250}}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			got, err := buildEntryArray(d.row)
			if err != nil {
				t.Fatalf("Error while using toEntry function: %q", err)
			}
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v+, got: %v+", d.want, got)
			}
		})
	}
}

func TestBuildEntryArray_Error(t *testing.T) {
	var testData = []struct {
		desc string
		row  [][]string
	}{
		{"EmptyEntry", [][]string{}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			_, err := buildEntryArray(d.row)
			if err == nil {
				t.Fatal("Error expected")
			}
		})
	}
}

func TestReadRecords_Sucess(t *testing.T) {
	in := `status,request_time
200,0.019
200,0.023
503,0.001`

	want := [][]string{{"200", "0.019"}, {"200", "0.023"}, {"503", "0.001"}}
	got, err := readRecords(strings.NewReader(in), "test.csv")
	if err != nil {
		t.Fatalf("Error not expected: %q", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("Want: %v+, got: %v+", want, got)
	}
}

func TestToEntry_Sucess(t *testing.T) {
	var testData = []struct {
		desc string
		row  []string
		want inputEntry
	}{
		{"Success", []string{"200", "0.019"}, inputEntry{200, 0.019}},
		{"Error", []string{"503", "0.250"}, inputEntry{503, 0.250}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			got, err := toEntry(d.row)
			if err != nil {
				t.Fatalf("Error while using toEntry function: %q", err)
			}
			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v+, got: %v+", d.want, got)
			}
		})
	}
}

func TestToEntry_Error(t *testing.T) {
	var testData = []struct {
		desc string
		row  []string
	}{
		{"StatusString", []string{"string", "0.019"}},
		{"DurationString", []string{"200", "string"}},
		{"StatusFloat", []string{"0.200", "0.200"}},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			_, err := toEntry(d.row)
			if err == nil {
				t.Fatal("Error expected")
			}
		})
	}
}