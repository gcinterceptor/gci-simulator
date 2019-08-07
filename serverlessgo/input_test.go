package main

import (
	"reflect"
	"strconv"
	"testing"
)

const (
	input_test = "test.csv"
)

func Test_buildEntryArray(t *testing.T) {
	records, err := getRecords(input_test)
	if err != nil {
		t.Fatalf("%q", err)
	}

	var expectedEntries []inputEntry
	for _, row := range records[1:] {
		expectedEntry, err := toEntry(row)
		if err != nil {
			t.Fatalf("Error while using toEntry function: %q", err)
		}
		expectedEntries = append(expectedEntries, expectedEntry)
	}

	entries, err := buildEntryArray(input_test)
	if err != nil {
		t.Fatalf("Error while using buildEntryArray function: %q", err)
	}
	if !reflect.DeepEqual(expectedEntries, entries) {
		t.Fatalf("Expected entries: %v+\nreceived entries: %v+", expectedEntries, entries)
	}

}

func Test_toEntry(t *testing.T) {
	records, err := getRecords(input_test)
	if err != nil {
		t.Fatalf("%q", err)
	}
	for _, row := range records[1:] {
		entry, err := toEntry(row)
		if err != nil {
			t.Fatalf("Error while using toEntry function: %q", err)
		}

		status, _ := strconv.Atoi(row[1])
		duration, _ := strconv.ParseFloat(row[2], 64)
		if entry.status != status || entry.duration != duration {
			expectedEntry := inputEntry{duration, status}
			t.Fatalf("Expected values: %v+, received: %v+", expectedEntry, entry)
		}
	}
}
