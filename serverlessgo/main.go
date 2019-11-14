package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gcinterceptor/gci-simulator/serverless/sim"
)

var (
	idlenessDeadline = flag.Duration("idleness", 300*time.Second, "The idleness deadline is the time that an instance may be idle until be terminated.")
	duration         = flag.Duration("duration", 36000*time.Second, "Duration of the simulation.") // default value is 10 hours
	lambda           = flag.Float64("lambda", 150.0, "The lambda of the Poisson distribution used on workload.")
	inputs           = flag.String("inputs", "default.csv", "Comma-separated file paths (one per instance)")
	outputPath       = flag.String("output", "./output/", "file path to output results")
	filename         = flag.String("filename", "default-filename", "file name of output results")
	optimized        = flag.Bool("optimized", false, "Define if the simulation must use the optimized scheduler")
	warmUp           = flag.Int("warmup", 500, "The Warm Up value to remove , default value is 500")
)

func main() {
	flag.Parse()

	if len(*inputs) == 0 {
		log.Fatalf("Must have at least one file input!")
	}

	var entries [][]sim.InputEntry
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

	outputPathAndFileName := *outputPath + *filename
	outputReqsFilePath := outputPathAndFileName + "-reqs.csv"
	header := "id,status,created_time,response_time,hops\n"
	reqsOutputWriter, err := newOutputWriter(outputReqsFilePath, header)
	defer reqsOutputWriter.close()
	if err != nil {
		log.Fatalf("Error creating LB's reqsOutputWriter: %q", err)
	}

	res := sim.Run(*duration, *idlenessDeadline, sim.NewPoissonInterArrival(*lambda), entries, reqsOutputWriter, *optimized, *warmUp)
	
	outputMetricsFilePath := outputPathAndFileName + "-metrics.log"
	err = saveSimulationMetrics(outputMetricsFilePath, res)
	if err != nil {
		log.Fatalf("Error when save metrics. Error: %q", err)
	}

	outputInstancesFilePath := outputPathAndFileName + "-instances.csv"
	err = saveSimulationInstances(outputInstancesFilePath, res.Instances)
	if err != nil {
		log.Fatalf("Error when save metrics. Error: %q", err)
	}
}

func buildEntryArray(records [][]string) ([]sim.InputEntry, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("Must have at least one file input!")
	}
	var entries []sim.InputEntry
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

func toEntry(row []string) (sim.InputEntry, error) {
	// Row format: id,status,response_time,body,tsbefore,tsafter
	status, err := strconv.Atoi(row[1])
	if err != nil {
		return sim.InputEntry{}, fmt.Errorf("Error parsing status in row (%v): %q", row, err)
	}
	responseTime, err := strconv.ParseFloat(row[2], 64)
	if err != nil {
		return sim.InputEntry{}, fmt.Errorf("Error parsing response_time in row (%v): %q", row, err)
	}
	body := row[3]
	tsbefore, err := strconv.ParseFloat(row[4], 64)
	if err != nil {
		return sim.InputEntry{}, fmt.Errorf("Error parsing tsbefore in row (%v): %q", row, err)
	}
	tsafter, err := strconv.ParseFloat(row[5], 64)
	if err != nil {
		return sim.InputEntry{}, fmt.Errorf("Error parsing tsafter in row (%v): %q", row, err)
	}
	return sim.InputEntry{Status: status, ResponseTime: responseTime, Body: body, TsBefore: tsbefore, TsAfter: tsafter}, nil
}
