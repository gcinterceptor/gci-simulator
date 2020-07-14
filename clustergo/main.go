package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/agoussia/godes"
	"github.com/gcinterceptor/gci-simulator/clustergo/interval"
)

var (
	duration         = flag.Duration("d", 300*time.Second, "Duration of the simulation.")
	warmup           = flag.Duration("warmup", 240*time.Second, "Server warmup duration, discarded from the input files.")
	rate             = flag.Float64("rate", 30, "Number of requests processed per second.")
	inputs           = flag.String("i", "", "Comma-separated file paths (one per server)")
	hedgingThreshold = flag.Float64("hedging-threshold", -1, "Threshold of the response to time to start hedging requests. -1 means no hedging.")
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

	lb := newLoadBalancer(servers)
	godes.AddRunner(lb)

	godes.Run()

	reqID := int64(0)
	for godes.GetSystemTime() < float64(duration.Milliseconds()) {
		arrivalQueue.Place(&request{id: reqID})
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
	fmt.Printf("PCP:%f\n", unavSum/procSum)

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
	fmt.Printf("PVN:%f\n", msUnav/procSum)
	fmt.Printf("THROUGHPUT: %f\n", float64(lb.nTerminatedSucc)/finishTime)
	fmt.Printf("PERC_HEDGED: %f\n", float64(lb.nHedged)/float64(lb.nProc))
}

type loadBalancer struct {
	*godes.Runner
	servers      *godes.FIFOQueue
	queueWaiter  *godes.BooleanControl
	isTerminated bool

	// Metrics
	nIgnored        int64
	nTerminatedSucc int64
	nTerminatedFail int64
	nHedged         int64
	nProc           int64
}

func (lb *loadBalancer) schedule(r *request) {
	if lb.servers.Len() == 0 {
		lb.nIgnored++
		return // ignoring new incoming requests when all servers are busy.
	}
	s := lb.servers.Get().(*server)
	if lb.servers.Len() == 0 {
		lb.queueWaiter.Set(false) // hedged requests must wait in queue.
	}
	s.newRequest(r)
}

func (lb *loadBalancer) reqFinished(s *server, r *request) {
	lb.servers.Place(s)      // Sending server back to the availability queue
	lb.queueWaiter.Set(true) // Needed for hedged requests.
	fmt.Printf("%d,%.1f,%d,%.4f,%d\n", r.id, r.ts, r.status, r.rt, r.sID)
	lb.nProc++
	if r.status == 200 {
		lb.nTerminatedSucc++
	} else {
		lb.nTerminatedFail++
	}
}

func (lb *loadBalancer) reqHedged(s *server, r *request, remainingTime float64) *server {
	lb.nHedged++
	r.hedged = true

	// The server always wait for the server queue cond before calling this method.
	// Furthermore, there one unfinished request, which will block the arrival queue.
	s1 := lb.servers.Get().(*server)
	dur, _ := s1.peek()

	// The duration on the new service will be greater than the previous one.
	// So, both servers only need to wait for the remaining time of the task
	// The previous service will declare the task as finished and the new one
	// only goes back to the queue.
	if dur >= remainingTime {
		godes.Advance(remainingTime)
		lb.reqFinished(s, s.req)
		lb.servers.Place(s1)     // Sending server back to the availability queue
		lb.queueWaiter.Set(true) // Needed for hedged requests.
		return nil
	}

	// Otherwise the previous service waits for the duration of the new processing.
	s1.newRequest(r)
	return s1
}

func (lb *loadBalancer) terminate() {
	arrivalCond.Set(true)
	lb.queueWaiter.Set(true)
	lb.isTerminated = true
}

func (lb *loadBalancer) Run() {
	fmt.Println("id,status,ts,rt,sID")
	for {
		arrivalCond.Wait(true)
		if lb.isTerminated {
			break
		}
		lb.schedule(arrivalQueue.Get().(*request))
		if arrivalQueue.Len() == 0 {
			arrivalCond.Set(false)
		}
	}
}

func newLoadBalancer(servers []*server) *loadBalancer {
	lb := &loadBalancer{
		Runner:       &godes.Runner{},
		isTerminated: false,
		servers:      godes.NewFIFOQueue("servers"),
		queueWaiter:  godes.NewBooleanControl(),
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
		s.req.ts = godes.GetSystemTime()
		s.req.sID = s.id

		switch {
		// If it is an unavailability signal.
		case s.req.status != 200:
			s.unavIntervals.Limits = append(s.unavIntervals.Limits, interval.Limit{Start: s.req.ts, End: s.req.ts + s.req.rt})
			godes.Advance(s.req.rt)
			s.lb.reqFinished(s, s.req)

		// If the request must be reissued.
		// There is no mean to know in advance where the request will be processed.
		// So, the replica re-issues the request and will keep itself blocked, waiting
		// for the processing of the re-issued request.
		case s.ht > 0 && s.req.rt > s.ht:
			diff := s.req.rt - s.ht
			fin := godes.GetSystemTime() + diff
			s.lb.queueWaiter.WaitAndTimeout(true, diff)

			// The request has finished in this server before a server was able to serve the request.
			if godes.GetSystemTime() == fin {
				s.lb.reqFinished(s, s.req)

			} else { // If both servers must execute the request.
				s1 := s.lb.reqHedged(s, s.req, diff)
				if s1 != nil { // wait for the other server to finish processing.
					s1.cond.Wait(false)
					s.lb.servers.Place(s)      // Sending server back to the availability queue
					s.lb.queueWaiter.Set(true) // Needed for hedged requests.
				}
			}

		// If the request must be processed by this replica without re-issuing.
		default:
			s.procIntervals.Limits = append(s.procIntervals.Limits, interval.Limit{Start: s.req.ts, End: s.req.ts + s.req.rt})
			s.procReqCount++
			s.procTime += s.req.rt
			godes.Advance(s.req.rt)
			s.lb.reqFinished(s, s.req)
		}

		s.cond.Set(false)
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
	sID    int
	ts     float64
	hedged bool
}
