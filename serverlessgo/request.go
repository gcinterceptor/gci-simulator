package main

type request struct {
	id           int64
	responseTime float64
	status       int
	hops		 []int
}

func (r *request) hasPassedInstance(id int) bool {
	for _, i := range r.hops {
		if id == i {
			return true
		}
	}
	return false
}