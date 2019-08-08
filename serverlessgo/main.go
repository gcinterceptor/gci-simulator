package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/agoussia/godes"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

var (
	idlenessDeadline = flag.Duration("idleness", 300*time.Second, "The idleness deadline is the time that an instance may be idle until be terminated.")
	duration         = flag.Duration("duration", 12000*time.Second, "Duration of the simulation.")
	lambda           = flag.Float64("lambda", 140.0, "The lambda of the Poisson distribution used on workload.")
	inputs           = flag.String("inputs", "", "Comma-separated file paths (one per instance)")
	output           = flag.String("output", "", "file paths to output without extension")
)

func main() {
	before := time.Now()
	flag.Parse()

	if len(*inputs) == 0 {
		log.Fatalf("Must have at least one file input!")
	}
	var entries [][]inputEntry
	for _, p := range strings.Split(*inputs, ",") {
		func() {
			f, err := os.Open(p)
			if err != nil {
				log.Fatalf("Error opening the file (%s), %q", p, err)
			}
			defer f.Close()

			records, err := readRecords(f, p)
			if err != nil {
				log.Fatalf("Error reading records: %q", err)
			}
			e, err := buildEntryArray(records)
			if err != nil {
				log.Fatalf("Error building entries %s. Error: %q", p, err)
			}
			entries = append(entries, e)
		}()
	}
	header := "id,status,response_time\n"
	outputWriter, err := newOutputWriter(*output, header)
	defer outputWriter.close()
	if err != nil {
		log.Fatalf("Error creating LB's outputWriter: %q", err)
	}
	lb := newLoadBalancer(*idlenessDeadline, entries, outputWriter)

	godes.AddRunner(lb)
	godes.Run()

	reqID := int64(0)
	poissonDist := &distuv.Poisson{
		Lambda: *lambda,
		Src:    rand.NewSource(uint64(time.Now().Nanosecond())),
	}
	for godes.GetSystemTime() < duration.Seconds() {
		lb.foward(&request{id: reqID})
		interArrivalTime := poissonDist.Rand()
		godes.Advance(interArrivalTime / 1000)
		reqID++
	}
	lb.terminate()
	godes.WaitUntilDone()

	throughput := lb.getFinishedReqs()
	totalCost := lb.getTotalCost()
	totalEfficiency := lb.getTotalEfficiency()
	fmt.Println("throughput,totalCost,totalEfficiency")
	fmt.Printf("%d,%.5f,%.10f\n", throughput, totalCost, totalEfficiency)

	fmt.Printf("time running the simulation: %v seconds\n", time.Since(before).Nanoseconds()/1000000000)
}
