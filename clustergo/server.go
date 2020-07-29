package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/agoussia/godes"
	"github.com/gcinterceptor/gci-simulator/clustergo/interval"
)

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
	hCancellation bool
}

func (s *server) Run() {
	for {
		s.cond.Wait(true)
		if s.isTerminated {
			break
		}

		// Updating request info.
		duration, status := s.next()
		s.req.startTime = godes.GetSystemTime()
		s.req.status = status
		s.req.sid = s.id

		switch {
		// If it is an unavailability signal and CCT is enabled.
		case *enableCCT && s.req.status == 503:
			s.unavIntervals.Limits = append(s.unavIntervals.Limits, interval.Limit{Start: s.req.startTime, End: s.req.startTime + s.req.rt})
			godes.Advance(duration)
			s.lb.reqFinished(s.req)

		// The processing request must be reissued.
		case s.ht >= 0 && duration > s.ht && !s.req.hedged: // We only accept one hedge.
			godes.Advance(s.ht)

			if s.hCancellation {
				// If the request must be reissued and the cancellation policy is active.
				// There is no mean to know in advance where the request will be processed.
				// So, the replica re-issues the request and will keep itself blocked, waiting
				// for the processing of the re-issued request.
				hedgeReq := &request{
					id:     s.req.id,
					hedged: true,
					hedge:  s.req,
					finish: godes.NewBooleanControl(),
				}
				s.lb.hedgedReqs.Place(hedgeReq)

				// Wait for the other server, but timeout if this very own request finishes first.
				remaining := duration - s.ht
				ts := godes.GetSystemTime()
				hedgeReq.finish.WaitAndTimeout(true, remaining)
				waitTime := godes.GetSystemTime() - ts

				// If the other finishes first, this request will be cancelled (and finished). The timeout won't be reached.

				if remaining-waitTime < 0.05 { // same as waitTime == remaining, but accounting for float64 comparisons.
					s.lb.reqFinished(s.req)
				} else { // The other server finished first
					s.lb.reqCancelled(s.req)
				}

			} else {
				// When the cancellation polity is not active, simply trigger a new hedge request and finish the processing one.
				hedgeReq := &request{
					id:     s.req.id,
					hedged: true,
					hedge:  s.req,
					finish: godes.NewBooleanControl(),
				}
				s.req.hedge = hedgeReq
				s.lb.hedgedReqs.Place(hedgeReq)

				// Finish processing.
				godes.Advance(duration - s.ht)
				s.lb.reqFinished(s.req)
			}

		// The processing request has been reissued. and the cancellation policy is active.
		case s.ht >= 0 && s.hCancellation && s.req.hedged:
			// Wait for the other server, but timeout if this very own request finishes first.
			ts := godes.GetSystemTime()
			s.req.hedge.finish.WaitAndTimeout(true, duration)
			waitTime := godes.GetSystemTime() - ts

			// If the other finishes first, this request will be cancelled (and finished). The timeout won't be reached.
			// waitTime == 0 means the other server has finished the request while this one was being enqueue.
			if waitTime > 0 && waitTime-duration < 0.05 { // same as waitTime == duration, but accounting for float64 comparisons.
				s.lb.reqFinished(s.req)
			} else { // The other server finished first
				s.lb.reqCancelled(s.req)
			}

		// If the request must be processed by this replica without re-issuing.
		default:
			s.procIntervals.Limits = append(s.procIntervals.Limits, interval.Limit{Start: s.req.startTime, End: s.req.startTime + s.req.rt})
			s.procReqCount++
			s.procTime += duration
			godes.Advance(duration)
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

func newServer(p string, id int, ht float64, hCancellation bool) (*server, error) {
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
		ht:            ht,
		hCancellation: hCancellation}, nil
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
