package main

type Request struct {
	id           int64
	status       int
	responseTime float64
	hops         []int
}

func newRequest(id int64) *Request {
	return &Request{id: id}
}

func (r *Request) updateHops(i int) {
	r.hops = append(r.hops, i)
}

func (r *Request) updateStatus(s int) {
	r.status = s
}

func (r *Request) updateResponseTime(t float64) {
	r.responseTime += t
}

func (r *Request) hasBeenProcessed(id int) bool {
	for _, i := range r.hops {
		if id == i {
			return true
		}
	}
	return false
}
