package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

type IInputReproducer interface {
	next() (int, float64)
}

type InputReproducer struct {
	index   int
	warmed  bool
	entries []inputEntry
}

type WarmedInputReproducer struct {
	index   int
	entries []inputEntry
}

func newInputReproducer(input []inputEntry) IInputReproducer {
	return &InputReproducer{entries: input}
}

func newWarmedInputReproducer(input []inputEntry) IInputReproducer {
	if len(input) > 1 {input = input[1:]}
	return &WarmedInputReproducer{entries: input}
}

func (r *InputReproducer) next() (int, float64) {
	e := r.entries[r.index]
	r.index = (r.index + 1) % len(r.entries)
	r.setWarm()
	return e.status, e.duration
}

func (r *InputReproducer) setWarm() {
	if !r.warmed {
		r.warmed = true
		if len(r.entries) > 1 {
			r.entries = r.entries[1:] // remove first entry
			r.index = 0
		}
	}
}

func (r *WarmedInputReproducer) next() (int, float64) {
	e := r.entries[r.index]
	r.index = (r.index + 1) % len(r.entries)
	return e.status, e.duration
}

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
