package interval

import (
	"testing"

	"github.com/matryer/is"
)

func TestIntersect_Pair(t *testing.T) {
	is := is.New(t)
	a := LimitSet{1, []Limit{{1, 2}, {0.3, 0.6}}}     // desordenado
	b := LimitSet{2, []Limit{{0.5, 1.5}, {0.5, 1.5}}} // intervalo repetido, deve ser ignorado
	i := Intersect(a, b)
	is.Equal(1, len(i))                 // quantidade de intersecções entre conjuntos 1+2
	is.Equal(2, len(i[0].Limits))       // quantidade de intersecções entre os dois conjuntos
	is.Equal(0.5, i[0].Limits[0].Start) // 1. inicio
	is.Equal(0.6, i[0].Limits[0].End)   // 1. fim
	is.Equal(1., i[0].Limits[1].Start)  // 2. inicio
	is.Equal(1.5, i[0].Limits[1].End)   // 2. fim
}

func TestIntersect_Many(t *testing.T) {
	is := is.New(t)
	a := LimitSet{1, []Limit{{1, 2}, {0.3, 0.6}}}
	b := LimitSet{2, []Limit{{0.5, 1.5}}}
	c := LimitSet{3, []Limit{{0.2, 0.3}}}  // sem intersecção
	d := LimitSet{4, []Limit{{0.55, 1.2}}} // intersecta o 1 e o 2
	i := Intersect(a, b, c, d)
	is.Equal(4, len(i)) // quantidade de combinações entre conjuntos 1+2, 1+4, 2+4, 1+2+4
	// Intersecções entre 1 e 2
	is.Equal(2, len(i[0].Limits))       // quantidade de intersecções entre 1 e 2
	is.Equal(0.5, i[0].Limits[0].Start) // 1+2 1. inicio
	is.Equal(0.6, i[0].Limits[0].End)   // 1+2 1. fim
	is.Equal(1., i[0].Limits[1].Start)  // 1+2 2. inicio
	is.Equal(1.5, i[0].Limits[1].End)   // 1+2 2. fim

	// Intersecções entre 1 e 4
	is.Equal(2, len(i[1].Limits))        // quantidade de intersecções entre 1 e 4
	is.Equal(0.55, i[1].Limits[0].Start) // 1+4 1. inicio
	is.Equal(0.6, i[1].Limits[0].End)    // 1+4 1. fim
	is.Equal(1., i[1].Limits[1].Start)   // 1+4 2. inicio
	is.Equal(1.2, i[1].Limits[1].End)    // 1+4 2. fim

	// Intersecções entre 2 e 4
	is.Equal(1, len(i[2].Limits))        // quantidade de intersecções entre 2 e 4
	is.Equal(0.55, i[2].Limits[0].Start) // 2+4 1. inicio
	is.Equal(1.2, i[2].Limits[0].End)    // 2+4 1. fim

	// Intersecções entre 1, 2 e 4
	is.Equal(2, len(i[3].Limits))        // quantidade de intersecções entre 1, 2 e 4
	is.Equal(0.55, i[3].Limits[0].Start) // 1+2+4 1. inicio
	is.Equal(0.6, i[3].Limits[0].End)    // 1+2+4 1. fim
	is.Equal(1., i[3].Limits[1].Start)   // 1+2+4 1. inicio
	is.Equal(1.2, i[3].Limits[1].End)    // 1+2+4 1. fim
}

func TestUnite(t *testing.T) {
	is := is.New(t)
	is.Equal(0, len(Unite().Limits)) // união vazia

	a := LimitSet{1, []Limit{{1, 2}}}
	is.Equal(map[int]struct{}{a.ID: {}}, Unite(a).Participants) // união unitária
	is.Equal(a.Limits, Unite(a).Limits)                         // união unitária

	is.Equal(map[int]struct{}{a.ID: {}}, Unite(a, a).Participants) // união identitária
	is.Equal(a.Limits, Unite(a, a).Limits)                         // união identitária

	b := LimitSet{2, []Limit{{0, 0.2}, {3, 4}}}
	is.Equal(map[int]struct{}{a.ID: {}, b.ID: {}}, Unite(a, b).Participants) // intervalos disjuntos
	is.Equal([]Limit{{0, 0.2}, {1, 2}, {3, 4}}, Unite(a, b).Limits)          // intervalos disjuntos

	c := LimitSet{2, []Limit{{0.3, 0.5}, {2, 3}, {3, 4}}}
	is.Equal(map[int]struct{}{a.ID: {}, b.ID: {}, c.ID: {}}, Unite(a, b, c).Participants) // intervalos disjuntos + fusão
	is.Equal([]Limit{{0, 0.2}, {0.3, 0.5}, {1, 4}}, Unite(a, b, c).Limits)                // // intervalos disjuntos + fusão
}
