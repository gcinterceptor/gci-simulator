package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
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
	hedgingCancellation = flag.Bool("hedge-cancellation", true, "Whether to apply cancelation on hedging. Must have the ht flag set.")
	enableCCT           = flag.Bool("cct", true, "Wheter CTC should be enabled.")
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
			s, err := newServer(f, i, *hedgingThreshold)
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
		arrivalQueue.Place(&request{id: reqID, ts: godes.GetSystemTime()})
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

type loadBalancer struct {
	*godes.Runner
	servers       *godes.FIFOQueue
	serversBusy   *godes.BooleanControl
	isTerminated  bool
	ht            float64
	hCancellation bool

	// Queues and controll of request processing flow.
	hedgedReqs    *godes.FIFOQueue
	hedgeMap      map[int64]struct{}
	finishedQueue *godes.FIFOQueue

	// Metrics
	nIgnored        int64
	nTerminatedSucc int64
	nTerminatedFail int64
	nHedged         int64
	nProc           int64
	hedgingWaist    float64
	nCancelled      int64
}

func (lb *loadBalancer) Run() {
	fmt.Println("id,status,ts,rt,sID,hedge,waist,canc")
	for {
		if lb.isTerminated {
			break
		}

		if lb.finishedQueue.Len() > 0 {
			r := lb.finishedQueue.Get().(*request)
			lb.computeReqMetrics(r)
			fmt.Printf("%d,%d,%.1f,%.4f,%d,%t,%t,%t\n", r.id, r.status, r.ts, r.rt, r.sid, r.hedged, r.waist, r.cancel)
			continue
		}

		// First wait for a server and process hedged requests.
		lb.serversBusy.Wait(false)

		if lb.hedgedReqs.Len() > 0 {
			r := lb.hedgedReqs.Get().(*request)
			if lb.servers.Len() == 0 {
				lb.hedgedReqs.Place(r)
				godes.Yield()
			} else {
				lb.schedule(r)
			}
			continue
		}

		// Then wait for a request that arrived.
		arrivalCond.Wait(true)
		r := arrivalQueue.Get().(*request)
		if arrivalQueue.Len() == 0 {
			arrivalCond.Set(false)
		}
		if lb.servers.Len() == 0 {
			godes.Yield()
			continue
		}
		lb.schedule(r)
	}
}

func (lb *loadBalancer) schedule(r *request) {
	lb.servers.Get().(*server).newRequest(r)
	if lb.servers.Len() == 0 {
		lb.serversBusy.Set(true)
	}
}

func (lb *loadBalancer) computeReqMetrics(r *request) {
	// Metrics
	lb.nProc++
	switch r.status {
	case 200:
		lb.nTerminatedSucc++
	case 503:
		lb.nTerminatedFail++
	}
	if r.hedged {
		lb.nHedged++
	}
	if r.cancel {
		lb.nCancelled++
	}
	if r.waist {
		lb.hedgingWaist += r.rt
	}
}

func (lb *loadBalancer) reqCancelled(r *request) {
	defer lb.finishedQueue.Place(r)

	r.rt = godes.GetSystemTime() - r.ts
	r.status = 0
	r.cancel = true
	r.waist = true
	if r.finish != nil {
		r.finish.Set(true)
	}
}

func (lb *loadBalancer) reqFinished(r *request) {
	defer lb.finishedQueue.Place(r)

	if r.finish != nil {
		r.finish.Set(true)
	}

	// If the policy is no-hedging or hedging-cancellation
	if lb.ht <= 0 || lb.hCancellation {
		return
	}

	// Policy is hedging, but no cancellation.
	// First finished version of this hedge request. Used when cancellation is not active.
	if _, ok := lb.hedgeMap[r.id]; !ok {
		r.waist = false
		r.cancel = false
		lb.hedgeMap[r.id] = struct{}{}
		return
	}

	// Second finished copy of this hedge request request. Happens when cancelation is not active.
	r.cancel = false
	r.waist = true
	delete(lb.hedgeMap, r.id)
}

func (lb *loadBalancer) terminate() {
	arrivalCond.Set(true)
	lb.isTerminated = true
	lb.serversBusy.Set(false)
}

func newLoadBalancer(servers []*server, ht float64, hc bool) *loadBalancer {
	lb := &loadBalancer{
		Runner:        &godes.Runner{},
		isTerminated:  false,
		servers:       godes.NewFIFOQueue("servers"),
		serversBusy:   godes.NewBooleanControl(),
		ht:            ht,
		hCancellation: hc,
		hedgeMap:      make(map[int64]struct{}),
		hedgedReqs:    godes.NewFIFOQueue("hedgedReqs"),
		finishedQueue: godes.NewFIFOQueue("fininshedReqs"),
	}
	for _, s := range servers {
		s.lb = lb
		lb.servers.Place(s)
	}
	return lb
}

type server struct {
	*godes.Runner
	id            int
	entries       []inputEntry
	index         int
	cond          *godes.BooleanControl
	lb            *loadBalancer
	req           *request
	isTerminated  bool
	unavIntervals interval.LimitSet
	unavCond      *godes.BooleanControl
	uptime        float64
	procTime      float64
	procIntervals interval.LimitSet
	procReqCount  int64
	ht            float64
}

func (s *server) Run() {
	for {
		s.cond.Wait(true)
		if s.isTerminated {
			break
		}

		// Updating request info.
		duration, status := s.next()
		s.req.rt = duration
		s.req.status = status
		s.req.sid = s.id

		switch {
		// If it is an unavailability signal and CCT is enabled.
		case *enableCCT && s.req.status == 503:
			s.unavIntervals.Limits = append(s.unavIntervals.Limits, interval.Limit{Start: s.req.ts, End: s.req.ts + s.req.rt})
			godes.Advance(s.req.rt)
			s.lb.reqFinished(s.req)

		// The processing request must be reissued.
		case s.ht > 0 && s.req.rt > s.ht && !s.req.hedged: // We only accept one hedge.
			godes.Advance(s.ht)
			if s.lb.hCancellation {
				// If the request must be reissued and the cancellation policy is active.
				// There is no mean to know in advance where the request will be processed.
				// So, the replica re-issues the request and will keep itself blocked, waiting
				// for the processing of the re-issued request.
				s.req.finish = godes.NewBooleanControl()
				hedgeReq := &request{id: s.req.id, ts: godes.GetSystemTime(), hedged: true, hedge: s.req}
				s.lb.hedgedReqs.Place(hedgeReq)
				s.req.finish.Wait(true)
			} else {

				// When the cancellation polity is not active, simply trigger a new hedge request and finish the processing one.
				hedgeReq := &request{id: s.req.id, ts: godes.GetSystemTime(), hedged: true}
				s.lb.hedgedReqs.Place(hedgeReq)
				godes.Advance(s.req.rt)
				s.lb.reqFinished(s.req)
			}

		// The processing request has been reissued. and the cancellation policy is active.
		case s.ht > 0 && s.lb.hCancellation && s.req.hedged:
			// The request will finish on the other server first.
			otherServerFinishTime := s.req.hedge.ts + s.req.hedge.rt
			if otherServerFinishTime <= (godes.GetSystemTime() + s.req.rt) {
				godes.Advance(math.Max(0, otherServerFinishTime-godes.GetSystemTime()))
				s.lb.reqFinished(s.req.hedge)
				s.lb.reqCancelled(s.req)
			} else {
				godes.Advance(s.req.rt)
				s.lb.reqFinished(s.req)
				s.lb.reqCancelled(s.req.hedge)
			}

		// If the request must be processed by this replica without re-issuing.
		default:
			s.procIntervals.Limits = append(s.procIntervals.Limits, interval.Limit{Start: s.req.ts, End: s.req.ts + s.req.rt})
			s.procReqCount++
			s.procTime += s.req.rt
			godes.Advance(s.req.rt)
			s.lb.reqFinished(s.req)
		}

		s.cond.Set(false)
		s.lb.servers.Place(s)
		s.lb.serversBusy.Set(false)
	}
}

func (s *server) newRequest(r *request) {
	s.req = r
	s.cond.Set(true)
}

func (s *server) next() (float64, int) {
	i := s.entries[s.index]
	s.index = (s.index + 1) % len(s.entries)
	return i.duration, i.status
}

// peerk returns information about the next request processing.
// The difference from next() is that peek() does not advance the
// incremental pointer of the request processing history.
func (s *server) peek() (float64, int) {
	i := s.entries[s.index]
	return i.duration, i.status
}

func (s *server) terminate() {
	s.uptime = godes.GetSystemTime()
	s.isTerminated = true
	s.cond.Set(true)
}

func newServer(p string, id int, ht float64) (*server, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
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
	delta := float64(warmup.Milliseconds())

	// Processing request entries after warmup period.
	var entries []inputEntry
	for _, row := range records {
		timestamp, e, err := toEntry(row)
		if err != nil {
			log.Fatal(err)
		}
		if timestamp >= delta {
			entries = append(entries, e)
		}
	}
	return &server{
		Runner:        &godes.Runner{},
		id:            id,
		entries:       entries,
		index:         0,
		cond:          godes.NewBooleanControl(),
		unavCond:      godes.NewBooleanControl(),
		isTerminated:  false,
		unavIntervals: interval.LimitSet{ID: id},
		procIntervals: interval.LimitSet{ID: id},
		ht:            ht}, nil
}

func toEntry(row []string) (float64, inputEntry, error) {
	// Row format: timestamp;status;request_time;upstream_response_time
	timestamp, err := strconv.ParseFloat(row[0], 64)
	if err != nil {
		log.Fatalf("Error parsing timestamp in row (%v): %q", row, err)
	}
	state, err := strconv.Atoi(row[1])
	if err != nil {
		return 0, inputEntry{}, fmt.Errorf("Error parsing state in row (%v): %q", row, err)
	}
	duration, err := strconv.ParseFloat(row[2], 64)
	if err != nil {
		return 0, inputEntry{}, fmt.Errorf("Error parsing duration in row (%v): %q", row, err)
	}
	return timestamp, inputEntry{duration, state}, nil
}

type inputEntry struct {
	duration float64
	status   int
}

type request struct {
	id     int64
	rt     float64
	status int
	sid    int
	ts     float64
	hedged bool
	cancel bool
	waist  bool
	finish *godes.BooleanControl
	hedge  *request
}
