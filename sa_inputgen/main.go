package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gonum.org/v1/gonum/stat"
)

var (
	warmReqs      = flag.Int("warm-reqs", 100, "Number of first request to be ignore due to warmup")
	printProbUnav = flag.Bool("print-prob-unav", false, "Whether to print the default unav probability")
	probUnav      = flag.Float64("prob-unav", 0.00018, "New unav probability")
	nReqs         = flag.Int64("nreqs", 100, "Number of requests in the simulator input.")
	nOut          = flag.Int("nout", 1, "Number of output files.")
	rtVar         = flag.Float64("rt-var", 0, "Percentage of RT that should be added to each request. For example, -0.2 to decrease each RT request in 20%")
	unavVar       = flag.Float64("unav-var", 0, "Percentage of RT that should be added to each request. For example, -0.2 to decrease each RT request in 20%")
	gci           = flag.Bool("gci", false, "Whether the input files for RT should consider GCI")
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UTC().UnixNano())

	rt := loadRT()
	meanRT, stdRT := stat.MeanStdDev(rt, nil)
	fmt.Println("RT Data - #Entries:", len(rt), "Mean:", meanRT, "StdDev:", stdRT)

	gc := loadGC()
	meanGC, stdGC := stat.MeanStdDev(gc, nil)
	fmt.Println("GC Data - #Entries:", len(gc), "Mean:", meanGC, "StdDev:", stdGC)

	if *printProbUnav {
		fmt.Println("ProbUnav: ", float64(len(gc))/float64(len(rt)))
		return
	}

	for outID := 1; outID <= *nOut; outID++ {
		func() {
			out := fmt.Sprintf("input_%d.csv", outID)
			f, err := os.Create(out)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			w := csv.NewWriter(f)
			w.Comma = ';'
			defer w.Flush()

			row := []string{"0", "0", "0", "0"}
			for i := int64(0); i < *nReqs; i++ {
				if rand.Float64() < *probUnav { // should be unav?
					row[3] = "true"
					if *gci {
						id := rand.Intn(len(gc)) // sampling
						unav := gc[id]           // applying variation
						unav += unav * (*unavVar)
						row[1] = "503"
						row[2] = fmt.Sprintf("%.4f", unav)
					} else {
						rtID := rand.Intn(len(rt)) // sampling
						rt := rt[rtID]             // applying variation
						rt += rt * (*rtVar)

						gcID := rand.Intn(len(gc)) // sampling
						gcImpact := gc[gcID]       // applying variation
						gcImpact += gcImpact * (*unavVar)

						row[1] = "200"
						row[2] = fmt.Sprintf("%.4f", rt+gcImpact)
					}
				} else {
					id := rand.Intn(len(rt)) // sampling
					r := rt[id]              // applying variation
					r += r * (*rtVar)
					row[1] = "200"
					row[2] = fmt.Sprintf("%.4f", r)
					row[3] = "false"
				}
				w.Write(row)
			}
			fmt.Printf("%s written with %d lines\n", out, *nReqs)
		}()
	}
}

func loadGC() []float64 {
	var result []float64
	var gcFiles []string
	var err error
	if *gci {
		gcFiles, err = filepath.Glob("/home/danielfireman/tese/resultados/exp_hazelcast/results/gc_hazelcast_gci_c3_*.log")
	} else {
		gcFiles, err = filepath.Glob("/home/danielfireman/tese/resultados/exp_hazelcast/results/gc_hazelcast_nogci_c3_*.log")
	}
	if err != nil {
		log.Fatal(err)
	}
	for _, fp := range gcFiles {
		func() {
			f, err := os.Open(fp)
			if err != nil {
				log.Fatalf("[%s]%q", fp, err)
			}
			defer f.Close()
			s := bufio.NewScanner(f)
			for line := 0; s.Scan(); line++ {
				pSplit := strings.Split(s.Text(), " ")
				durStr := pSplit[len(pSplit)-1] // this includes the ms suffix
				dur, err := strconv.ParseFloat(strings.TrimSuffix(durStr, "ms"), 64)
				if err != nil {
					log.Fatalf("[%s][%d]%q", fp, line, err)
				}
				result = append(result, dur)
			}
		}()
	}
	return result
}

func loadRT() []float64 {
	var result []float64
	// files used to sample the response time (without gc.)
	// c3 are the readonly ones.
	rtFiles := []string{
		"/home/danielfireman/tese/resultados/exp_hazelcast/results/al_hazelcast_gci_c3_0.log",
		"/home/danielfireman/tese/resultados/exp_hazelcast/results/al_hazelcast_gci_c3_1.log",
		"/home/danielfireman/tese/resultados/exp_hazelcast/results/al_hazelcast_gci_c3_2.log",
		"/home/danielfireman/tese/resultados/exp_hazelcast/results/al_hazelcast_gci_c3_3.log",
		"/home/danielfireman/tese/resultados/exp_hazelcast/results/al_hazelcast_gci_c3_4.log",
	}

	// rtFiles = []string{
	// 	"/home/danielfireman/tese/resultados/exp_hazelcast/results/al_hazelcast_nogci_c3_0.log",
	// 	"/home/danielfireman/tese/resultados/exp_hazelcast/results/al_hazelcast_nogci_c3_1.log",
	// 	"/home/danielfireman/tese/resultados/exp_hazelcast/results/al_hazelcast_nogci_c3_2.log",
	// 	"/home/danielfireman/tese/resultados/exp_hazelcast/results/al_hazelcast_nogci_c3_3.log",
	// 	"/home/danielfireman/tese/resultados/exp_hazelcast/results/al_hazelcast_nogci_c3_4.log",
	// }

	for _, fp := range rtFiles {
		func() {
			f, err := os.Open(fp)
			if err != nil {
				log.Fatalf("[%s]%q", fp, err)
			}
			defer f.Close()

			s := bufio.NewScanner(f)
			for line := 0; s.Scan(); line++ {
				if line < *warmReqs { // ignore first lines until warmup
					continue
				}
				// Log format: log_format exp '$msec;$status;$request_time;$upstream_response_time;$upstream_addr;$request_method;$request';
				// Example: /home/danielfireman/tese/resultados/exp_hazelcast/nginx_16.conf
				pSplit := strings.SplitN(s.Text(), ";", 5)
				status := pSplit[1]
				// only considering successfull requests.
				// requests that lead to eviction will come from other files.
				if status != "200" && status != "204" { // NOTE: 204 means no content.
					continue
				}
				// https://stackoverflow.com/questions/37430951/why-is-request-time-much-larger-than-upstream-response-time-in-nginx-access-log
				// Getting the $upstream_response_time
				// Description: keeps time spent on receiving the response from the upstream server; the time is kept in seconds with millisecond resolution.
				// Times of several responses are separated by commas and colons like addresses in the $upstream_addr variable.
				rt := 0.
				for _, rStr := range strings.Split(strings.ReplaceAll(pSplit[3], " ", ""), ",") {
					r, err := strconv.ParseFloat(rStr, 64)
					if err != nil {
						log.Fatalf("[%s][%d]%q", fp, line, err)
					}
					rt += r
				}
				if rt == 0 {
					continue // ignoring lines with rt == 0
				}
				result = append(result, rt*1000) // converting to ms
			}
			if err := s.Err(); err != nil {
				log.Fatalf("[%s]%q", fp, err)
			}
		}()
	}
	return result
}
