package sim

import (
	"reflect"
	"testing"
	"time"

	"github.com/agoussia/godes"
)

func TestReceive(t *testing.T) {
	instance := &instance{id: 2, cond: godes.NewBooleanControl()}

	workingBefore := instance.isWorking()
	if workingBefore {
		t.Fatalf("Want: %v, got: %v", workingBefore, !workingBefore)
	}
	req := &Request{Hops: []int{0, 1}}
	instance.receive(req)

	workingAfter := instance.isWorking()
	if !workingAfter {
		t.Fatalf("Want: %v, got: %v", !workingAfter, workingAfter)
	}

	wHops := []int{0, 1, 2}
	if !reflect.DeepEqual(wHops, req.Hops) {
		t.Fatalf("Want: %v, got: %v", wHops, req.Hops)
	}
}

func TestInstanceTerminate(t *testing.T) {
	instance := &instance{
		Runner:      &godes.Runner{},
		createdTime: 0.0,
		cond:        godes.NewBooleanControl(),
	}
	type Want struct {
		isTerminated  bool
		terminateTime float64
	}
	got := &Want{isTerminated: instance.isTerminated()}
	want := &Want{isTerminated: false}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("Before terminate - Want: %v, got: %v", want, got)
	}

	instance.terminate()

	got = &Want{isTerminated: instance.isTerminated(), terminateTime: instance.terminateTime}
	want = &Want{isTerminated: true, terminateTime: 0.0}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("After terminate - Want: %v, got: %v", want, got)
	}
}

func TestScaleDown(t *testing.T) {
	idleness, _ := time.ParseDuration("5m")
	instance := &instance{
		Runner:           &godes.Runner{},
		createdTime:      0.0,
		cond:             godes.NewBooleanControl(),
		lastWorked:       godes.GetSystemTime(),
		idlenessDeadline: idleness,
	}
	type Want struct {
		isTerminated  bool
		terminateTime float64
	}
	got := &Want{isTerminated: instance.isTerminated()}
	want := &Want{isTerminated: false}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("Before terminate - Want: %v, got: %v", want, got)
	}

	instance.scaleDown()

	got = &Want{isTerminated: instance.isTerminated(), terminateTime: instance.terminateTime}
	want = &Want{isTerminated: true, terminateTime: 300.0}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("After terminate - Want: %v, got: %v", want, got)
	}
}

type TestLoadBalancer struct{ req *Request }

func (lb *TestLoadBalancer) forward(r *Request) error  { return nil }
func (lb *TestLoadBalancer) response(r *Request) error { lb.req = r; return nil }
func TestInstanceRun(t *testing.T) {
	instance := &instance{
		Runner:     &godes.Runner{},
		cond:       godes.NewBooleanControl(),
		reproducer: newInputReproducer([]InputEntry{{200, 0.8}, {200, 0.1}, {200, 0.2}}),
		lb:         &TestLoadBalancer{},
	}

	godes.Run()
	godes.AddRunner(instance)
	godes.Advance(0.8)

	req := &Request{ID: 1}
	instance.receive(req)
	godes.WaitUntilDone()

	want := req.ID
	got := instance.lb.(*TestLoadBalancer).req.ID
	if want != got {
		t.Fatalf("Want: %v, got: %v", want, got)
	}
}
