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
		slice := make([]int, 0)
		for n := rand.Intn(10); n < math.MaxInt && limit > 0; n += rand.Intn(10)+1 {
			limit--
			output <- n
			slice = append(slice, n)
			dur := int(rand.Intn(100)) * int(time.Microsecond)
			time.Sleep(time.Duration(dur))
		}
		fmt.Printf("Slice: %d\n", slice)
	}()

	return output
}

func unique(ch1, ch2 chan int) (out chan int) {
	out = make(chan int)
	go func() {
		defer close(out)

		var val1, val2 int
		var ok1, ok2 bool

		val1, ok1 = <-ch1
		val2, ok2 = <-ch2

		for ok1 || ok2 {
			switch {
			case !ok1:
				out <- val2
				val2, ok2 = <-ch2
			case !ok2:
				out <- val1
				val1, ok1 = <-ch1
			case val1 > val2:
				out <- val2
				val2, ok2 = <-ch2
			case val1 < val2:
				out <- val1
				val1, ok1 = <-ch1
			default:
				val1, ok1 = <-ch1
				val2, ok2 = <-ch2
			}
		}
	}()
	return
}

func main() {
	ch1 := generateN(rand.Intn(20))
	ch2 := generateN(rand.Intn(20))

	fmt.Println("start")

	for v := range unique(ch1, ch2) {
		fmt.Println(v)
	}
}
