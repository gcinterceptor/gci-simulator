package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

type inputEntry struct {
	duration float64
	status   int
}

func buildEntryArray(p string) ([]inputEntry, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("Error opening the file (%s), %q", p, err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = ';'

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("Error reading input file (%s): %q", p, err)
	}
	if len(records) <= 1 {
		return nil, fmt.Errorf("Can not create a server with no requests (empty or header-only input file): %s", p)
	}

	var entries []inputEntry
	for _, row := range records[1:] {
		entry, err := toEntry(row)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func toEntry(row []string) (inputEntry, error) {
	// Row format: timestamp;status;request_time;upstream_response_time
	state, err := strconv.Atoi(row[1])
	if err != nil {
		return inputEntry{}, fmt.Errorf("Error parsing state in row (%v): %q", row, err)
	}
	duration, err := strconv.ParseFloat(row[2], 64)
	if err != nil {
		return inputEntry{}, fmt.Errorf("Error parsing duration in row (%v): %q", row, err)
	}
	return inputEntry{duration, state}, nil
}
