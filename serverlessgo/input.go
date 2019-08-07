package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

type inputEntry struct {
	status   int
	duration float64
}

func buildEntryArray(records [][]string) ([]inputEntry, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("Must have at least one file input!")
	}
	var entries []inputEntry
	for _, row := range records {
		entry, err := toEntry(row)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func readRecords(f io.Reader, p string) ([][]string, error) {
	r := csv.NewReader(f)
	r.Comma = ','

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("Error parsing csv (%s): %q", p, err)
	}
	if len(records) <= 1 {
		return nil, fmt.Errorf("Can not create a server with no requests (empty or header-only input file): %s", p)
	}
	return records[1:], nil
}

func toEntry(row []string) (inputEntry, error) {
	// Row format: status;request_time
	status, err := strconv.Atoi(row[0])
	if err != nil {
		return inputEntry{}, fmt.Errorf("Error parsing status in row (%v): %q", row, err)
	}
	duration, err := strconv.ParseFloat(row[1], 64)
	if err != nil {
		return inputEntry{}, fmt.Errorf("Error parsing duration in row (%v): %q", row, err)
	}
	return inputEntry{status, duration}, nil
}
