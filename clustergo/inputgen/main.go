package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
)

var outPrefix = flag.String("pref", "input", "")
var nReplicas = flag.Int("r", 1, "")
var mu = flag.Float64("mu", 1, "")
var littleOmega = flag.Float64("littleOmega", 1, "")
var bigOmega = flag.Float64("bigOmega", 0.5, "")
var duration = flag.Float64("duration", 1000, "")

func main() {
	/*
		ModelEntropySimple.r = replicas;
		ModelEntropySimple.omega = 0.0001;
		ModelEntropySimple.mu = 0.0036;
		ModelEntropySimple.pImpact = 0.003;
		ModelEntropySimple.lambda = 1;
	*/
	flag.Parse()
	for i := 1; i <= *nReplicas; i++ {
		ev := 0
		succ := 0
		func() {
			f, err := os.Create(fmt.Sprintf("%s_%d.csv", *outPrefix, i))
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			ts := float64(0)
			for ts < *duration {
				rt := rand.ExpFloat64() / *mu
				st := 200
				unav := rand.Float64() <= *bigOmega
				if unav {
					rt = rand.ExpFloat64() / *littleOmega
					st = 503
					ev++
				} else {
					succ++
				}
				fmt.Fprintf(f, "%.4f;%d;%.4f;%.4f\n", ts, st, rt, rt)
				ts += rt
			}
			fmt.Println("Succ:", succ)
			fmt.Println("Ev:", ev)
		}()
	}
}
