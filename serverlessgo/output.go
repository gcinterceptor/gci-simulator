package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gcinterceptor/gci-simulator/serverless/sim"
)

type outputWriter struct {
	f *os.File
}

func newOutputWriter(path, header string) (*outputWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("Error trying to create the output file: %q", err)
	}
	_, err = f.WriteString(header)
	if err != nil {
		return nil, fmt.Errorf("Error trying to write the csv header: %q", err)
	}
	return &outputWriter{f: f}, nil
}

func (o *outputWriter) RequestFinished(r *sim.Request) {
	s := fmt.Sprintf("%d,%d,%.1f\n", r.ID, r.Status, r.ResponseTime*1000)
	_, err := o.f.WriteString(s)
	if err != nil {
		// Crash the simulation binary if we can not write output.
		log.Fatalf("Error trying to write s (%s) in file (%v+): %q", s, o.f, err)
	}
}

func (o *outputWriter) close() {
	o.f.Close()
}

func printSimulationMetrics(throughput float64, totalCost, totalEfficiency float64, simulationTime int64) {
	fmt.Printf("Throughput: %f\n", throughput)
	fmt.Printf("Total cost of instances: %.5f\n", totalCost)
	fmt.Printf("Total efficiency of instances: %.10f\n", totalEfficiency)
	fmt.Printf("time running the simulation: %d seconds\n", simulationTime)
}
