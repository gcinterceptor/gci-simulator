package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/agoussia/godes"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

var (
	idlenessDeadline = flag.Duration("i", 300*time.Second, "The idleness deadline is the time that an instance may be idle until be terminated.")
	duration         = flag.Duration("d", 300*time.Second, "Duration of the simulation.")
	lambda           = flag.Float64("lambda", 140.0, "The lambda of the Poisson distribution used on workload.")
	inputs           = flag.String("inputs", "", "Comma-separated file paths (one per instance)")
)

func main() {
	// TODO(David): to abstract output via struct
	fmt.Println("id,status,response_time")
	flag.Parse()

	if len(*inputs) == 0 {
		log.Fatalf("Must have at least one file input!")
	}

	var entries [][]inputEntry
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

	lb := newLoadBalancer(*idlenessDeadline, entries)
	godes.AddRunner(lb)
	godes.Run()

	reqID := int64(0)
	poissonDist := &distuv.Poisson{
		Lambda: *lambda,
		Src:    rand.NewSource(uint64(time.Now().Nanosecond())),
	}
	for godes.GetSystemTime() < duration.Seconds() {
		lb.receiveRequest(&request{id: reqID})
		interArrivalTime := poissonDist.Rand()
		godes.Advance(interArrivalTime / 1000)
		reqID++
	}

	lb.terminate()
	godes.WaitUntilDone()
}

type request struct {
	id           int64
	responseTime float64
	status       int
}

type loadBalancer struct {
	*godes.Runner
	isTerminated     bool
	arrivalQueue     *godes.FIFOQueue
	arrivalCond      *godes.BooleanControl
	instances        []*instance
	idlenessDeadline time.Duration
	inputs           [][]inputEntry
	index            int
}

func newLoadBalancer(idlenessDeadline time.Duration, inputs [][]inputEntry) *loadBalancer {
	return &loadBalancer{
		Runner:           &godes.Runner{},
		arrivalQueue:     godes.NewFIFOQueue("arrival"),
		arrivalCond:      godes.NewBooleanControl(),
		instances:        make([]*instance, 0),
		idlenessDeadline: idlenessDeadline,
		inputs:           inputs,
	}
}

func (lb *loadBalancer) receiveRequest(r *request) {
	lb.arrivalQueue.Place(r)
	lb.arrivalCond.Set(true)
}

func (lb *loadBalancer) terminate() {
	for _, i := range lb.instances {
		i.terminate()
	}
	lb.isTerminated = true
	lb.arrivalCond.Set(true)
}

func (lb *loadBalancer) nextInstanceInputs() []inputEntry {
	input := lb.inputs[lb.index]
	lb.index = (lb.index + 1) % len(lb.inputs)
	return input
}

func (lb *loadBalancer) nextInstance() *instance {
	var selected *instance
	// sorting instances to have the most recently used ones ahead on the array
	sort.SliceStable(lb.instances, func(i, j int) bool { return lb.instances[i].getLastWorked() > lb.instances[j].getLastWorked() })
	for i := 0; i < len(lb.instances); i++ {
		instance := lb.instances[i]
		if !instance.isWorking() && !instance.isTerminated() {
			selected = instance
			break
		}
	}
	if selected == nil {
		selected = newInstance(len(lb.instances), lb.idlenessDeadline, lb.nextInstanceInputs())
		godes.AddRunner(selected)
		// inserts the instance ahead of the array
		lb.instances = append([]*instance{selected}, lb.instances...)
	}

	return selected
}

func (lb *loadBalancer) Run() {
	for {
		lb.arrivalCond.Wait(true)
		if lb.arrivalQueue.Len() > 0 {
			r := lb.arrivalQueue.Get().(*request)
			if r.status == 200 {
				fmt.Printf("%d,%d,%.1f\n", r.id, r.status, r.responseTime*1000)
			} else {
				lb.nextInstance().receiveRequest(lb, r)
			}
		}

		if lb.arrivalQueue.Len() == 0 {
			if lb.isTerminated {
				break
			}
			lb.arrivalCond.Set(false)
		}
		lb.tryScaleDown()
	}
}

func (lb *loadBalancer) tryScaleDown() {
	for _, i := range lb.instances {
		if godes.GetSystemTime()-i.getLastWorked() >= lb.idlenessDeadline.Seconds() {
			i.scaleDown()
		}
	}
}

type reqRef struct {
	lb *loadBalancer
	r  *request
}

type instance struct {
	*godes.Runner
	id               int
	terminated       bool
	cond             *godes.BooleanControl
	req              *request
	createdTime      float64
	terminateTime    float64
	lastWorked       float64
	busyTime         float64
	idlenessDeadline time.Duration
	entries          []inputEntry
	index            int
}

func newInstance(id int, idlenessDeadline time.Duration, input []inputEntry) *instance {
	return &instance{
		Runner:           &godes.Runner{},
		id:               id,
		cond:             godes.NewBooleanControl(),
		createdTime:      godes.GetSystemTime(),
		idlenessDeadline: idlenessDeadline,
		entries:          input,
	}
}

func (i *instance) receiveRequest(lb *loadBalancer, r *request) {
	if i.isWorking() == true {
		panic(fmt.Sprintf("Instances may not enqueue requests."))
	}
	i.req = &reqRef{lb, r}
	i.cond.Set(true)
}

func (i *instance) terminate() {
	i.terminateTime = godes.GetSystemTime()
	i.terminated = true
	i.cond.Set(true)
}

func (i *instance) next() (float64, int) {
	e := i.entries[i.index]
	i.index = (i.index + 1) % len(i.entries)
	return e.duration, e.status
}

func (i *instance) scaleDown() {
	i.terminate()
	i.terminateTime = i.getLastWorked() + i.idlenessDeadline.Seconds()
}

func (i *instance) Run() {
	for {
		i.cond.Wait(true)
		if i.isTerminated() {
			break
		}

		responseTime, status := i.next()
		i.req.status = status
		i.req.responseTime += responseTime
		i.busyTime += responseTime

		godes.Advance(responseTime)
		i.lastWorked = godes.GetSystemTime()
		i.req.lb.receiveRequest(i.req)

		i.cond.Set(false)
	}
}

func (i *instance) isWorking() bool {
	return i.cond.GetState() == true
}

func (i *instance) isTerminated() bool {
	return i.terminated
}

func (i *instance) getUpTime() float64 {
	return i.terminateTime - i.createdTime
}

func (i *instance) getIdleTime() float64 {
	return i.getUpTime() - i.getBusyTime()
}

func (i *instance) getBusyTime() float64 {
	return i.busyTime
}

func (i *instance) getLastWorked() float64 {
	return i.lastWorked
}

func (i *instance) getEfficiency() float64 {
	return i.getBusyTime() / i.getUpTime()
}
