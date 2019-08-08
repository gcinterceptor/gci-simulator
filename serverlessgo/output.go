package main

import (
    "fmt"
    "os"
)

type outputWriter struct {
	f    *os.File
}

func newOutputWriter(p string) (*outputWriter, error) {
	f, err := os.Create(p)
	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("File to output already exists: %q", err)
		} else {
			return nil, fmt.Errorf("Error trying to create the output file: %q", err)
		}
	}
	_, err = f.WriteString("id,status,response_time\n")
    if err != nil {
        return nil, fmt.Errorf("Error trying to write the csv header: %q", err)
    }
	return &outputWriter{f: f}, nil
}

func (o *outputWriter) record(s string) error {
	_, err := o.f.WriteString(s)
    if err != nil {
        return fmt.Errorf("Error trying to write s (%s) in file (%v+): %q", s, o.f, err)
	}
	return nil
}

func (o *outputWriter) close() {
	o.f.Close()
}
