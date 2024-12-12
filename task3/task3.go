package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

func generateN(limit int) chan int {
	output := make(chan int)

	go func() {
		defer close(output)
		for n := rand.Intn(10); n < math.MaxInt && limit > 0; n += rand.Intn(10) {
			limit--
			output <- n
			dur := int(rand.Intn(100)) * int(time.Microsecond)
			time.Sleep(time.Duration(dur))
		}
	}()

	return output
}

func unique(ch1, ch2 chan int) (out chan int) {
	out = make(chan int)

	go func() {
		defer close(out)

		// TODO: rewrite the following:

		for v := range ch1 {
			out <- v
		}

		for v := range ch2 {
			out <- v
		}
	}()

	return
}

func main() {
	ch1 := generateN(rand.Intn(1))
	ch2 := generateN(rand.Intn(50))

	for v := range unique(ch1, ch2) {
		fmt.Println(v)
	}
}
