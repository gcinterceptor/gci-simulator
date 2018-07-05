package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/agoussia/godes"
)

var (
	duration = flag.Duration("d", 300*time.Second, "Duration of the simulation.")
	warmup   = flag.Duration("warmup", 240*time.Second, "Server warmup duration, discarded from the input files.")
	rate     = flag.Int("rate", 30, "Number of requests processed per second.")
	inputs   = flag.String("i", "", "Comma-separated file paths (one per server)")
)

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
		servers = append(servers, s)
	}
	lb := newLoadBalancer(servers)

	fmt.Println("status,latency")
	godes.Run()
	reqID := int64(0)
	for godes.GetSystemTime() < duration.Seconds() {
		godes.AddRunner(&trigger{&godes.Runner{}, lb, reqID})
		godes.Advance(1.0 / float64(*rate))
		reqID++
	}
	godes.WaitUntilDone()
}

type loadBalancer struct {
	sync.Mutex
	servers []*server
	next    int
}

func (lb *loadBalancer) dispatch(r *request) {
	for attempts := 0; attempts < len(lb.servers); attempts++ {
		status := lb.nextUpstream().process(r)
		if status == 200 {
			break
		}
	}
	fmt.Printf("%d,%f\n", r.status, r.latency)
}

func (lb *loadBalancer) nextUpstream() *server {
	lb.Lock()
	defer lb.Unlock()
	lb.next = (lb.next + 1) % len(lb.servers)
	return lb.servers[lb.next]
}

func newLoadBalancer(servers []*server) *loadBalancer {
	return &loadBalancer{sync.Mutex{}, servers, 0}
}

type server struct {
	sync.Mutex
	id      int
	entries []inputEntry
	index   int
	busy    bool
}

func (s *server) process(r *request) int {
	e := s.next()
	r.latency += e.latency
	r.status = e.status
	r.hops = append(r.hops, hop{serverID: s.id, duration: e.latency, status: e.status})
	return r.status
}

func (s *server) next() inputEntry {
	s.Lock()
	defer s.Unlock()
	i := s.entries[s.index]
	s.index = (s.index + 1) % len(s.entries)
	return i
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
	return &server{id: id, entries: entries}, nil
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
	latency, err := strconv.ParseFloat(row[2], 64)
	if err != nil {
		return 0, inputEntry{}, fmt.Errorf("Error parsing latency in row (%v): %q", row, err)
	}
	return timestamp, inputEntry{latency * 1000, state}, nil
}

type inputEntry struct {
	latency float64
	status  int
}

type trigger struct {
	*godes.Runner
	lb *loadBalancer
	id int64
}

func (t *trigger) Run() {
	t.lb.dispatch(&request{id: t.id})
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
