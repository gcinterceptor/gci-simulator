package sim

import (
	"time"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

type InterArrival interface {
	next() float64
}

type poissonInterArrival struct {
	p *distuv.Poisson
}

func (pia *poissonInterArrival) next() float64 {
	return pia.p.Rand() / 1000
}

func NewPoissonInterArrival(lambda float64) InterArrival {
	return &poissonInterArrival{
		&distuv.Poisson{
			Lambda: lambda,
			Src:    rand.NewSource(uint64(time.Now().Nanosecond())),
		}}
}
