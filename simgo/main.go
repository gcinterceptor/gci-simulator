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
	rate     = flag.Int("rate", 30, "Number of requests processed per second.")
	inputs   = flag.String("i", "", "Comma-separated file paths (one per server)")
)

var arrivalQueue = godes.NewFIFOQueue("arrival")
var arrivalCond = godes.NewBooleanControl()

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
	for godes.GetSystemTime() < duration.Seconds() {
		arrivalQueue.Place(&request{id: reqID})
		arrivalCond.Set(true)
		godes.Advance(1.0 / float64(*rate))
		reqID++
	}
	lb.terminate()
	for _, s := range servers {
		s.terminate()
	}
	godes.WaitUntilDone()
}

type loadBalancer struct {
	*godes.Runner
	servers      []*server
	next         int
	isTerminated bool
}

func (lb *loadBalancer) nextServer() *server {
	lb.next = (lb.next + 1) % len(lb.servers)
	return lb.servers[lb.next]
}

func (lb *loadBalancer) terminate() {
	arrivalCond.Set(true)
	lb.isTerminated = true
}

func (lb *loadBalancer) Run() {
	fmt.Println("status,latency")
	for {
		arrivalCond.Wait(true)
		if arrivalQueue.Len() > 0 {
			r := arrivalQueue.Get().(*request)
			if r.status != 200 && len(r.hops) < len(lb.servers) {
				lb.nextServer().newRequest(r)
			} else {
				fmt.Printf("%d,%f\n", r.status, r.latency*1000)
			}
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
	return &loadBalancer{&godes.Runner{}, servers, 0, false}
}

type server struct {
	*godes.Runner
	id           int
	entries      []inputEntry
	index        int
	cond         *godes.BooleanControl
	queue        *godes.FIFOQueue
	isTerminated bool
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
			r.hops = append(r.hops, hop{serverID: s.id, duration: duration, status: status})

			// Sending updated request back to the load balancer.
			arrivalQueue.Place(r)
			arrivalCond.Set(true)
			// Advancing simulation time.
			godes.Advance(duration)
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

	first, err := strconv.ParseFloat(records[1][0], 64)
	if err != nil {
		log.Fatalf("Error parsing timestamp in row (%v): %q", records[1], err)
	}
	delta := warmup.Seconds()

	// Processing request entries after warmup period.
	var entries []inputEntry
	for _, row := range records[1:] {
		timestamp, e, err := toEntry(row)
		if err != nil {
			log.Fatal(err)
		}
		if timestamp >= first+delta {
			entries = append(entries, e)
		}
	}
	return &server{&godes.Runner{}, id, entries, 0, godes.NewBooleanControl(), godes.NewFIFOQueue(fmt.Sprintf("server%d", id)), false}, nil
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
}

type hop struct {
	serverID int
	duration float64
	status   int
}
