package main

type request struct {
	id           int64
	responseTime float64
	status       int
	hops         []int
}

func newRequest(id int64) *request {
	return &request{id: id}
}

func (r *request) updateHops(i int) {
	r.hops = append(r.hops, i)
}

func (r *request) updateStatus(s int) {
	r.status = s
}

func (r *request) updateResponseTime(t float64) {
	r.responseTime += t
}

func (r *request) hasBeenProcessed(id int) bool {
	for _, i := range r.hops {
		if id == i {
			return true
		}
	}
	return false
}
