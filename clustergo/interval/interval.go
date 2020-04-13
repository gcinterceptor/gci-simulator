package interval

import (
	"fmt"
	"math"
	"sort"
)

type Limit struct {
	Start, End float64
}

type limits []Limit

func (in limits) Len() int           { return len(in) }
func (in limits) Less(i, j int) bool { return in[i].Start < in[j].Start }
func (in limits) Swap(i, j int)      { in[i], in[j] = in[j], in[i] }

type LimitSet struct {
	ID     int
	Limits []Limit
}

func (in LimitSet) Len() int           { return len(in.Limits) }
func (in LimitSet) Less(i, j int) bool { return in.Limits[i].Start < in.Limits[j].Start }
func (in LimitSet) Swap(i, j int)      { in.Limits[i], in.Limits[j] = in.Limits[j], in.Limits[i] }

type Intersection struct {
	id           int
	Limits       []Limit
	Participants map[int]struct{} // set of intervals which generated the intersection
}

func (in Intersection) Len() int           { return len(in.Limits) }
func (in Intersection) Less(i, j int) bool { return in.Limits[i].Start < in.Limits[j].Start }
func (in Intersection) Swap(i, j int)      { in.Limits[i], in.Limits[j] = in.Limits[j], in.Limits[i] }

type Union struct {
	Limits       []Limit
	Participants map[int]struct{}
}

type intersections []Intersection

func (in intersections) Len() int           { return len(in) }
func (in intersections) Less(i, j int) bool { return len(in[i].Participants) < len(in[j].Participants) }
func (in intersections) Swap(i, j int)      { in[i], in[j] = in[j], in[i] }

// Unite calcula a união de todos os intervalos passados como parâmetro.
func Unite(limits ...LimitSet) Union {
	if len(limits) == 0 {
		return Union{}
	}
	if len(limits) == 1 {
		return Union{Limits: limits[0].Limits, Participants: map[int]struct{}{limits[0].ID: {}}}
	}

	var u Union
	for _, ls := range limits {
		if len(ls.Limits) == 0 {
			continue
		}
		u = u.unite(Union{
			Participants: map[int]struct{}{ls.ID: {}},
			Limits:       uniteL(ls.Limits...),
		})
	}
	return u
}

func (u Union) unite(u1 Union) Union {
	// Juntando os participantes
	p := u.Participants
	if p == nil {
		p = make(map[int]struct{})
	}
	for k := range u1.Participants {
		p[k] = struct{}{}
	}
	// Unindo tudo.
	return Union{
		Participants: p,
		Limits:       uniteL(append(u.Limits, u1.Limits...)...),
	}
}

func uniteL(ls ...Limit) []Limit {
	if len(ls) == 0 {
		return []Limit{}
	}
	if len(ls) == 1 {
		return ls
	}
	var res []Limit
	sort.Sort(limits(ls)) // essa ordenação é importante!
	last := ls[0]
	for _, l := range ls[1:] {
		// Os intervalos são disjuntos
		if l.Start > last.End {
			res = append(res, last)
			last = l
		} else { // Fusão de intervalos
			last = Limit{math.Min(last.Start, l.Start), math.Max(last.End, l.End)}
		}
	}
	return append(res, last)
}

// Intersect calcula a interseção entre todas as combinações de conjuntos de intervalos passados como parâmetro.
// Primeiro calcula as intersecções dos pares de conjuntos 2 a 2, depois 3 a 3 e assim sucessivamente.
func Intersect(ints ...LimitSet) []Intersection {
	// pair wise.
	ret := make(map[int]Intersection)
	if len(ints) < 2 {
		return []Intersection{}
	}
	// Calcula os pares.
	for i := 0; i < len(ints); i++ {
		for j := i + 1; j < len(ints); j++ {
			if _, ok := ret[idLimits(ints[i], ints[j])]; !ok { // do not calculate intersections twice.
				aux := pairL(ints[i], ints[j])
				if len(aux.Limits) > 0 {
					ret[aux.id] = aux
				}
			}
		}
	}
	// Calcula as demais combinações.
	for i := 2; i < len(ints); i++ {
		for _, currIntersection := range ret {
			for _, currInterval := range ints {
				if _, ok := ret[idLimitIntersection(currInterval, currIntersection)]; !ok { // do not calculate intersections twice.
					aux := pairI(currInterval, currIntersection)
					if len(aux.Limits) > 0 {
						ret[aux.id] = aux
					}
				}
			}
		}
	}
	var s []Intersection
	for _, v := range ret {
		s = append(s, v)
	}
	sort.Sort(intersections(s))
	return s
}

func pairI(a LimitSet, b Intersection) Intersection {
	if _, ok := b.Participants[a.ID]; ok {
		return b
	}
	p := map[int]struct{}{
		a.ID: struct{}{},
	}
	for k, v := range b.Participants {
		p[k] = v
	}
	// Sorting
	sort.Sort(a)
	sort.Sort(b)
	return Intersection{
		id:           idLimitIntersection(a, b),
		Participants: p,
		Limits:       inter(a.Limits, b.Limits),
	}
}

func pairL(a LimitSet, b LimitSet) Intersection {
	// Sorting
	sort.Sort(a)
	sort.Sort(b)
	return Intersection{
		id: idLimits(a, b),
		Participants: map[int]struct{}{
			a.ID: struct{}{},
			b.ID: struct{}{},
		},
		Limits: inter(a.Limits, b.Limits),
	}
}

// identificadores tem que ser determinísticos e basear em operações comutativas.
func idLimits(a LimitSet, b LimitSet) int {
	return a.ID + b.ID
}

// identificadores tem ser determinísticos e se basear em operações comutativas.
func idLimitIntersection(a LimitSet, b Intersection) int {
	return a.ID + b.id
}

// Heavily inspired by: https://leetcode.com/problems/interval-list-intersections/discuss/242413/Golang-100-faster-and-100-smaller
func inter(a []Limit, b []Limit) []Limit {
	res := make(map[string]Limit)
	var currA *Limit
	var currB *Limit
	for {
		if currA == nil && len(a) > 0 {
			currA = &a[0]
			a = a[1:]
		}

		if currB == nil && len(b) > 0 {
			currB = &b[0]
			b = b[1:]
		}

		if currA == nil || currB == nil {
			// nothing else to compare
			break
		}

		// [    A    ]
		//              [   B   ]
		// [    B    ]
		//              [   A   ]
		// No intersection
		if currA.End < currB.Start {
			currA = nil
			continue
		} else if currB.End < currA.Start {
			currB = nil
			continue
		}

		// [    A    ]
		//     [    B    ]
		// [    B    ]
		//     [    A    ]
		//         [    A    ]
		//     [    B    ]
		//         [    B    ]
		//     [    A    ]
		// [    B    ]
		//     [ A ]
		// [    A    ]
		//     [ B ]
		// There's an intersection, so find the intersection
		start := math.Max(currA.Start, currB.Start)
		end := math.Min(currA.End, currB.End)
		if start != end {
			res[fmt.Sprintf("%f,%f", start, end)] = Limit{
				Start: start,
				End:   end,
			}
		}

		if currA.End < currB.End {
			currA = nil
		} else if currB.End < currA.End {
			currB = nil
		} else {
			currA = nil
			currB = nil
		}
	}
	var r []Limit
	for _, v := range res {
		r = append(r, v)
	}
	sort.Sort(limits(r))
	return r
}
