package main

import (
	"math"
	"sort"
)

type interval struct {
	start, end float64
}

type intervals struct {
	set []interval
	id  int
}

func (in intervals) Len() int           { return len(in.set) }
func (in intervals) Less(i, j int) bool { return in.set[i].start < in.set[j].start }
func (in intervals) Swap(i, j int)      { in.set[i], in.set[j] = in.set[j], in.set[i] }

type intersection struct {
	id   int
	set  []interval
	comb map[int]struct{} // set of intervals which generated the intersection
}

func (in intersection) Len() int           { return len(in.set) }
func (in intersection) Less(i, j int) bool { return in.set[i].start < in.set[j].start }
func (in intersection) Swap(i, j int)      { in.set[i], in.set[j] = in.set[j], in.set[i] }

type intersections []intersection

func (in intersections) Len() int           { return len(in) }
func (in intersections) Less(i, j int) bool { return len(in[i].comb) < len(in[j].comb) }
func (in intersections) Swap(i, j int)      { in[i], in[j] = in[j], in[i] }

func combinations(ints ...intervals) []intersection {
	// pair wise.
	ret := make(map[int]intersection)
	if len(ints) < 2 {
		return []intersection{}
	}
	// Calcula os pares.
	for i := 0; i < len(ints); i++ {
		for j := i + 1; j < len(ints); j++ {
			if _, ok := ret[ints[i].id+ints[j].id]; !ok { // do not calculate intersections twice.
				aux := pair(ints[i], ints[j])
				if len(aux.set) > 0 {
					ret[aux.id] = aux
				}
			}
		}
	}
	// Calcula as demais combinações.
	for i := 1; i < len(ints); i++ {
		for _, currIntersection := range ret {
			for _, currInterval := range ints {
				if _, ok := ret[currInterval.id+currInterval.id]; !ok { // do not calculate intersections twice.
					aux := pairI(currInterval, currIntersection)
					if len(aux.set) > 0 {
						ret[aux.id] = aux
					}
				}
			}
		}
	}

	var s []intersection
	for _, v := range ret {
		s = append(s, v)
	}
	sort.Sort(intersections(s))
	return s
}

func pairI(a intervals, b intersection) intersection {
	if _, ok := b.comb[a.id]; ok {
		return b
	}
	comb := map[int]struct{}{
		a.id: struct{}{},
	}
	for k, v := range b.comb {
		comb[k] = v
	}
	// Sorting
	sort.Sort(a)
	sort.Sort(b)
	return intersection{
		id:   a.id + b.id, // id can be any comutative operation.
		comb: comb,
		set:  inter(a.set, b.set),
	}
}

func pair(a intervals, b intervals) intersection {
	// Sorting
	sort.Sort(a)
	sort.Sort(b)
	return intersection{
		id: a.id + b.id, // id can be any comutative operation.
		comb: map[int]struct{}{
			a.id: struct{}{},
			b.id: struct{}{},
		},
		set: inter(a.set, b.set),
	}
}

func inter(a []interval, b []interval) []interval {
	var currA *interval
	var currB *interval
	var res []interval
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
		if currA.end < currB.start {
			currA = nil
			continue
		} else if currB.end < currA.start {
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
		start := math.Max(currA.start, currB.start)
		end := math.Min(currA.end, currB.end)
		if start != end {
			res = append(res, interval{
				start: start,
				end:   end,
			})
		}

		if currA.end < currB.end {
			currA = nil
		} else if currB.end < currA.end {
			currB = nil
		} else {
			currA = nil
			currB = nil
		}
	}
	return res
}
