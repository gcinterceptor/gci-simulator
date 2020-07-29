package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/agoussia/godes"
	"github.com/gcinterceptor/gci-simulator/clustergo/interval"
)

var (
	duration            = flag.Duration("d", 300*time.Second, "Duration of the simulation.")
	warmup              = flag.Duration("warmup", 240*time.Second, "Server warmup duration, discarded from the input files.")
	rate                = flag.Float64("rate", 30, "Number of requests processed per second.")
	inputs              = flag.String("i", "", "Comma-separated file paths (one per server)")
	hedgingThreshold    = flag.Float64("ht", -1, "Threshold of the response to time to start hedging requests. -1 means no hedging.")
	hedgingCancellation = flag.Bool("hedge-cancellation", false, "Whether to apply cancelation on hedging. Must have the ht flag set.")
	enableCCT           = flag.Bool("cct", false, "Wheter CTC should be enabled.")
)

var arrivalQueue = godes.NewFIFOQueue("arrival")
var arrivalCond = godes.NewBooleanControl()
var arrivalDist = godes.NewExpDistr(false)

func main() {
	flag.Parse()
	if *inputs == "" {
		log.Fatal("At least one server description should be passed. Have you set the --i flag?")
	}
	rand.Seed(time.Now().UnixNano())
	var servers []*server
	for i, f := range strings.Split(*inputs, ",") {
		if strings.Trim(f, " ,;") != "" {
			s, err := newServer(f, i, *hedgingThreshold, *hedgingCancellation)
			if err != nil {
				log.Fatalf("Error loading file \"%s\":%q", f, err)
			}
			godes.AddRunner(s)
			servers = append(servers, s)
		}
	}

	lb := newLoadBalancer(servers, *hedgingThreshold, *hedgingCancellation)
	godes.AddRunner(lb)

	godes.Run()

	reqID := int64(0)
	for godes.GetSystemTime() < float64(duration.Milliseconds()) {
		arrivalQueue.Place(&request{
			id:     reqID,
			finish: godes.NewBooleanControl(),
		})
		arrivalCond.Set(true)
		godes.Advance(arrivalDist.Get(1 / *rate))
		reqID++
	}
	//fmt.Println("terminating simulation", godes.GetSystemTime())
	lb.terminate()
	for _, s := range servers {
		s.terminate()
	}
	godes.WaitUntilDone()

	finishTime := godes.GetSystemTime()
	fmt.Printf("NSERVERS: %d\n", len(servers))
	fmt.Printf("FINISH TIME: %f\n", finishTime)
	fmt.Printf("NIGNORED: %d\n", lb.nIgnored)

	var nProc int64
	var unav []interval.LimitSet
	var proc []interval.LimitSet
	var procSum float64
	for _, s := range servers {
		unav = append(unav, s.unavIntervals)
		proc = append(proc, s.procIntervals)
		nProc += s.procReqCount
		procSum += s.procTime
		fmt.Printf("SERVER: %d UPTIME:%f NPROC:%d PROCTIME:%f\n", s.id, s.uptime, s.procReqCount, s.procTime)
	}
	procTime := float64(0)
	procUnion := interval.Unite(proc...)
	for _, i := range procUnion.Limits {
		procTime += i.End - i.Start
	}
	fmt.Printf("NPROC: %d PROC_UNION:%f PROC_SUM:%f\n", nProc, procTime, procSum)

	unavSum := float64(0)
	for _, u := range unav {
		for _, l := range u.Limits {
			unavSum += l.End - l.Start
		}
	}

	var msUnav float64
	if len(servers) == 1 {
		for _, u := range unav {
			for _, l := range u.Limits {
				msUnav += l.End - l.Start
			}
		}
	} else {
		intersect := interval.Intersect(unav...)
		for _, i := range intersect {
			if len(i.Participants) == len(servers) {
				for _, l := range i.Limits {
					msUnav += l.End - l.Start
				}
			}
		}
	}

	// Grouped metrics. When changing any of those, please also change
	// the run_exp.sh script.
	fmt.Printf("PCP:%f\n", unavSum/procSum)
	fmt.Printf("PVN:%f\n", msUnav/procSum)
	fmt.Printf("NUM_PROC_SUCC:%d\n", lb.nTerminatedSucc)
	fmt.Printf("NUM_PROC_FAILED:%d\n", lb.nTerminatedFail)
	fmt.Printf("DURATION:%f\n", finishTime)
	fmt.Printf("HEDGED:%d\n", lb.nHedged)
	fmt.Printf("HEDGE_WAIST:%f\n", lb.hedgingWaist)
}

type request struct {
	id         int64
	rt         float64
	status     int
	sid        int
	startTime  float64
	finishTime float64
	hedged     bool
	cancel     bool
	waist      bool
	finish     *godes.BooleanControl
	hedge      *request
}
