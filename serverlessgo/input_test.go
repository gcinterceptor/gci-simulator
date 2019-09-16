package main

import (
	"reflect"
	"strings"
	"testing"
)

const (
	input_test = "test.csv"
)

func TestInputIReproducer(t *testing.T) {
	var testData = []struct {
		desc              string
		reproduce         IInputReproducer
		numberOfNextCalls int
		want              []inputEntry
	}{
		{"OneEntry", newInputReproducer([]inputEntry{{200, 0.2}}), 3, []inputEntry{
				{200, 0.2}, {200, 0.2}, {200, 0.2}},
			},
		{"ManyEntry", newInputReproducer([]inputEntry{{200, 0.8}, {200, 0.2}, {200, 0.3}}), 5, []inputEntry{
			{200, 0.8}, {200, 0.2}, {200, 0.3}, {200, 0.2}, {200, 0.3}},
		},
		{"WarmedOneEntry", newWarmedInputReproducer([]inputEntry{{200, 0.2}}), 3, []inputEntry{
				{200, 0.2}, {200, 0.2}, {200, 0.2}},
			},
		{"WarmedManyEntry", newWarmedInputReproducer([]inputEntry{{200, 0.8}, {200, 0.2}, {200, 0.3}}), 5, []inputEntry{
			{200, 0.2}, {200, 0.3}, {200, 0.2}, {200, 0.3}, {200, 0.2}},
		},
	}
	for _, d := range testData {
		t.Run(d.desc, func(t *testing.T) {
			var got []inputEntry
			for i := 0; i < d.numberOfNextCalls; i++ {
				status, duration := d.reproduce.next()
				got = append(got, inputEntry{status, duration})
			}

			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("Want: %v, got: %v", d.want, got)
			}
		})
	}
}

func TestBuildEntryArray_Success(t *testing.T) {
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
				t.Fatalf("Want: %v, got: %v", d.want, got)
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

func TestReadRecords_Success(t *testing.T) {
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
		t.Fatalf("Want: %v, got: %v", want, got)
	}
}

func TestToEntry_Success(t *testing.T) {
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
				t.Fatalf("Want: %v, got: %v", d.want, got)
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
