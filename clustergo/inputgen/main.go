package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

var outPrefix = flag.String("pref", "input", "")
var nReplicas = flag.Int("r", 1, "")
var mu = flag.Float64("mu", 1, "")
var littleOmega = flag.Float64("littleOmega", 1, "")
var bigOmega = flag.Float64("bigOmega", 0.5, "")
var duration = flag.Float64("duration", 1000, "")
var enableCCT = flag.Bool("cct", true, "")

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	for i := 1; i <= *nReplicas; i++ {
		ev := 0
		succ := 0
		func() {
			f, err := os.Create(fmt.Sprintf("%s_%d.csv", *outPrefix, i))
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			w := bufio.NewWriter(f)
			defer w.Flush()

			ts := float64(0)
			for ts < *duration {
				rt := rand.ExpFloat64() / *mu
				st := 200
				if rand.Float64() <= *bigOmega {
					if *enableCCT {
						rt = rand.ExpFloat64() / *littleOmega
						st = 503
						ev++
					} else {
						rt += rand.ExpFloat64() / *littleOmega
						succ++
					}
				} else {
					succ++
				}
				fmt.Fprintf(w, "%.4f;%d;%.4f;%.4f\n", ts, st, rt, rt)
				ts += rt
			}
			fmt.Println("Succ:", succ)
			fmt.Println("Ev:", ev)
		}()
	}
}
