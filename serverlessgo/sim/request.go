package sim

type Request struct {
	ID           int64
	Status       int
	ResponseTime float64
	Hops         []int
}

func newRequest(id int64) *Request {
	return &Request{ID: id}
}

func (r *Request) updateHops(i int) {
	r.Hops = append(r.Hops, i)
}

func (r *Request) updateStatus(s int) {
	r.Status = s
}

func (r *Request) updateResponseTime(t float64) {
	r.ResponseTime += t
}

func (r *Request) hasBeenProcessed(id int) bool {
	for _, i := range r.Hops {
		if id == i {
			return true
		}
	}
	return false
}
