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
		return nil, fmt.Errorf("Error trying to create the reqs output file: %q", err)
	}
	_, err = f.WriteString(header)
	if err != nil {
		return nil, fmt.Errorf("Error trying to write the csv reqs header: %q", err)
	}
	return &outputWriter{f: f}, nil
}

func (o *outputWriter) RequestFinished(r *sim.Request) {
	s := fmt.Sprintf("%d,%d,%f,%f,%v\n", r.ID, r.Status, r.CreatedTime, r.ResponseTime*1000, r.Hops)
	_, err := o.f.WriteString(s)
	if err != nil {
		// Crash the simulation binary if we can not write output.
		log.Fatalf("Error trying to write req s (%s) in file (%v+): %q", s, o.f, err)
	}
}

func (o *outputWriter) close() {
	o.f.Close()
}

func saveSimulationMetrics(path string, res sim.Results) error {
	throughput := float64(res.RequestCount) / (*duration).Seconds()
	totalCost := res.Cost
	totalEfficiency := res.Efficiency
	simulationTime := res.SimulationTime

	s := fmt.Sprintf("Throughput: %f\n", throughput)
	s += fmt.Sprintf("Total cost of instances: %.5f\n", totalCost)
	s += fmt.Sprintf("Total efficiency of instances: %.10f\n", totalEfficiency)
	s += fmt.Sprintf("time running the simulation: %d seconds\n", simulationTime)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Error trying to create the output file: %q", err)
	}
	_, err = f.WriteString(s)
	if err != nil {
		return fmt.Errorf("Error trying to write the csv metrics: %q", err)
	}
	f.Close()

	return nil
}

func saveSimulationInstances(path string, instances []sim.IInstance) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Error trying to create the output file: %q", err)
	}

	s := fmt.Sprintf("id,is_terminated,is_working,is_available,lastWorked,busyTime,up_time,idle_time,efficiency,created_time\n")
	_, err = f.WriteString(s)
	if err != nil {
		return fmt.Errorf("Error trying to write the csv instances header: %q", err)
	}
	for _, i := range instances {
		s = fmt.Sprintf(
			"%s,%t,%t,%t,%f,%f,%f,%f,%f,%f\n",
			i.GetId(), i.IsTerminated(), i.IsWorking(), i.IsAvailable(),
			i.GetLastWorked(), i.GetBusyTime(), i.GetUpTime(),
			i.GetIdleTime(), i.GetEfficiency(), i.GetCreatedTime(),
		)
		_, err = f.WriteString(s)
		if err != nil {
			return fmt.Errorf("Error trying to write the csv instances: %q", err)
		}
	}

	return nil
}
