package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/agoussia/godes"
)

var (
	duration = flag.Duration("d", 300*time.Second, "Duration of the simulation.")
	warmup   = flag.Duration("warmup", 240*time.Second, "Server warmup duration, discarded from the input files.")
	rate     = flag.Float64("rate", 30, "Number of requests processed per second.")
	inputs   = flag.String("i", "", "Comma-separated file paths (one per server)")
)

var arrivalQueue = godes.NewFIFOQueue("arrival")
var arrivalCond = godes.NewBooleanControl()
var arrivalDist = godes.NewExpDistr(false)

func main() {
	flag.Parse()
	if *inputs == "" {
		log.Fatal("At least one server description should be passed. Have you set the --i flag?")
	}
	var servers []*server
	for i, f := range strings.Split(*inputs, ",") {
		s, err := newServer(f, i)
		if err != nil {
			log.Fatalf("Error loading file \"%s\":%q", f, err)
		}
		godes.AddRunner(s)
		servers = append(servers, s)
	}

	lb := newLoadBalancer(servers)
	godes.AddRunner(lb)

	reqID := int64(0)
	godes.Run()
	for godes.GetSystemTime() < float64(duration.Milliseconds()) {
		arrivalQueue.Place(&request{id: reqID})
		arrivalCond.Set(true)
		godes.Advance(arrivalDist.Get(*rate))
		reqID++
	}
	lb.terminate()
	for _, s := range servers {
		s.terminate()
	}
	godes.WaitUntilDone()
	fmt.Printf("NSERVERS: %d\n", lb.nServers)
	fmt.Printf("Proc: %f\n", lb.nRequests)
	fmt.Printf("EvI: %f\n", lb.nReqEvictedByInstance)
	fmt.Printf("EvS: %f\n", lb.nReqEvictedByService)
	fmt.Printf("Succ: %f\n", lb.nSucc)
	fmt.Printf("Succ Ratio: %.4f\n", lb.nSucc/lb.nRequests)
	fmt.Printf("PVN: %.4f\n", lb.nReqEvictedByService/lb.nRequests)
	fmt.Printf("PCP: %.4f\n", lb.nReqEvictedByInstance/lb.nRequests)
}

type loadBalancer struct {
	*godes.Runner
	replicaQueue          *godes.FIFOQueue
	next                  int
	nServers              int
	isTerminated          bool
	nRequests             float64
	nSucc                 float64
	nReqEvictedByInstance float64
	nReqEvictedByService  float64
}

func (lb *loadBalancer) schedule(r *request) {
	// Ignore requests when there is no available replica
	if lb.replicaQueue.Len() == 0 {
		return
	}
	lb.replicaQueue.Get().(*server).newRequest(r)
}

func (lb *loadBalancer) reqFinished(s *server, r *request) {
	// Sending server back to the availability queue.
	switch {
	case r.status == 200:
		fmt.Printf("%d,%d,%.1f,%.4f,%d,%v\n", r.id, r.status, r.latency, r.ts, len(r.hops), r.hops)
		lb.nSucc++
	case r.status == 503:
		lb.nReqEvictedByInstance++
		if len(r.hops) < lb.nServers {
			lb.schedule(r)
		} else {
			lb.nReqEvictedByService++
			fmt.Printf("%d,%d,%.1f,%.4f,%d,%v\n", r.id, r.status, r.latency, r.ts, len(r.hops), r.hops)
		}
	default:
		// Stop simulation when we don't what to do.
		panic(fmt.Sprintf("I don't know what to do with this request:%+v server:%+v", *r, *s))
	}
}

func (lb *loadBalancer) terminate() {
	arrivalCond.Set(true)
	lb.isTerminated = true
}

func (lb *loadBalancer) Run() {
	fmt.Println("id,status,latency,ts,nhops,hops")
	for {
		arrivalCond.Wait(true)
		if arrivalQueue.Len() > 0 {
			lb.nRequests++
			lb.schedule(arrivalQueue.Get().(*request))
		}
		if lb.isTerminated && arrivalQueue.Len() == 0 {
			break
		}
		if arrivalQueue.Len() == 0 {
			arrivalCond.Set(false)
		}
	}
}

func newLoadBalancer(servers []*server) *loadBalancer {
	lb := &loadBalancer{
		Runner:                &godes.Runner{},
		next:                  0,
		isTerminated:          false,
		nRequests:             0,
		nReqEvictedByService:  0,
		nSucc:                 0,
		nReqEvictedByInstance: 0}
	q := godes.NewFIFOQueue("serverQueue")
	for _, s := range servers {
		s.lb = lb
		q.Place(s)
	}
	lb.replicaQueue = q
	lb.nServers = q.Len()
	return lb
}

type server struct {
	*godes.Runner
	id           int
	entries      []inputEntry
	index        int
	cond         *godes.BooleanControl
	queue        *godes.FIFOQueue
	lb           *loadBalancer
	isTerminated bool
	unavailable  *godes.BooleanControl
}

func (s *server) Run() {
	for {
		s.cond.Wait(true)
		if s.queue.Len() > 0 {
			// Processing request.

			duration, status := s.next()
			r := s.queue.Get().(*request)
			r.latency += duration
			r.status = status
			r.ts = godes.GetSystemTime()
			r.hops = append(r.hops, hop{serverID: s.id, duration: duration, status: status})

			// Advancing simulation time.
			godes.Advance(duration)

			if r.status == 200 {
				s.lb.replicaQueue.Place(s)
				s.lb.reqFinished(s, r) // Sending the request back to the loadbalancer.
			} else {
				s.lb.reqFinished(s, r) // Sending the request back to the loadbalancer.
				fmt.Println("starting unavailability", godes.GetSystemTime())
				s.unavailable.WaitAndTimeout(true, r.latency) // Only comes back to the queue after the duration period.
				s.lb.replicaQueue.Place(s)
				fmt.Println("backing to the queue", godes.GetSystemTime())
			}

		}
		if s.isTerminated && arrivalQueue.Len() == 0 {
			break
		}
		if s.queue.Len() == 0 {
			s.cond.Set(false)
		}
	}
}

func (s *server) newRequest(r *request) {
	s.queue.Place(r)
	s.cond.Set(true)
}

func (s *server) next() (float64, int) {
	i := s.entries[s.index]
	s.index = (s.index + 1) % len(s.entries)
	return i.duration, i.status
}

func (s *server) terminate() {
	s.cond.Set(true)
	s.isTerminated = true
}

func newServer(p string, id int) (*server, error) {
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
		Runner:       &godes.Runner{},
		id:           id,
		entries:      entries,
		index:        0,
		cond:         godes.NewBooleanControl(),
		queue:        godes.NewFIFOQueue(fmt.Sprintf("server%d", id)),
		unavailable:  godes.NewBooleanControl(),
		isTerminated: false}, nil
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
	id      int64
	latency float64
	status  int
	hops    []hop
	ts      float64
}

type hop struct {
	serverID int
	duration float64
	status   int
}
