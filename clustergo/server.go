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
		duration, status, impact := s.next()
		if !s.req.hedged {
			s.req.startTime = godes.GetSystemTime()
		}
		s.req.status = status
		s.req.sid = s.id
		s.req.impact = impact

		switch {
		// If it is an unavailability signal and CCT is enabled.
		case *enableCCT && s.req.status == 503:
			s.unavIntervals.Limits = append(s.unavIntervals.Limits, interval.Limit{Start: s.req.startTime, End: s.req.startTime + duration})
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
				remaining := duration - s.ht
				hedgeReq := &request{
					id:            s.req.id,
					hedged:        true,
					hedge:         s.req,
					finish:        godes.NewBooleanControl(),
					startTime:     godes.GetSystemTime(),
					remainingTime: remaining,
				}
				s.lb.hedgedReqs.Place(hedgeReq)

				// Wait for the other server, but timeout if this very own request finishes first.
				hedgeReq.finish.WaitAndTimeout(true, remaining)
				if hedgeReq.finish.GetState() {
					s.lb.reqCancelled(s.req)
				} else {
					s.procReqCount++
					s.procTime += duration
					s.lb.reqFinished(s.req)
				}
			} else {
				// When the cancellation polity is not active, simply trigger a new hedge request and finish the processing one.
				hedgeReq := &request{
					id:        s.req.id,
					hedged:    true,
					hedge:     s.req,
					finish:    godes.NewBooleanControl(),
					startTime: godes.GetSystemTime(),
				}
				s.req.hedge = hedgeReq
				s.lb.hedgedReqs.Place(hedgeReq)

				// Finish processing.
				godes.Advance(duration - s.ht)
				s.lb.reqFinished(s.req)
				s.procReqCount++
				s.procTime += duration
			}

		// The processing request has been reissued. and the cancellation policy is active.
		case s.ht >= 0 && s.hCancellation && s.req.hedged:
			//           remaining time            - (queuing time)
			remaining := s.req.hedge.remainingTime - (godes.GetSystemTime() - s.req.startTime)
			if remaining > 0 && duration > remaining {
				godes.Advance(remaining)
				s.lb.reqCancelled(s.req)
			} else {
				godes.Advance(duration)
				s.lb.reqFinished(s.req)
				s.procReqCount++
				s.procTime += duration
			}

		// If the request must be processed by this replica without re-issuing.
		default:
			s.procIntervals.Limits = append(s.procIntervals.Limits, interval.Limit{Start: s.req.startTime, End: s.req.startTime + duration})
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

func (s *server) next() (float64, int, bool) {
	i := s.entries[s.index]
	s.index = (s.index + 1) % len(s.entries)
	return i.duration, i.status, i.impact
}

// peerk returns information about the next request processing.
// The difference from next() is that peek() does not advance the
// incremental pointer of the request processing history.
func (s *server) peek() (float64, int, bool) {
	i := s.entries[s.index]
	return i.duration, i.status, i.impact
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
			if e.duration > 0 {
				entries = append(entries, e)
			}
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
	impact, err := strconv.ParseBool(row[3])
	if err != nil {
		return 0, inputEntry{}, fmt.Errorf("Error parsing impact in row (%v): %q", row, err)
	}
	return timestamp, inputEntry{duration, state, impact}, nil
}

type inputEntry struct {
	duration float64
	status   int
	impact   bool
}
