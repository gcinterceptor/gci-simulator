package main

type request struct {
	id           int64
	responseTime float64
	status       int
	hops		 []int
}
