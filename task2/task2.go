/* Task 2:
 * N threads (generate slice, ≈ 1024 items)
 * Send to summarizers
 */

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
)

var threadsSum int
var threadsGen int
var size int
var batchSize int

func init() {
	flag.IntVar(&threadsSum, "s", 1, "Count treads Sum")
	flag.IntVar(&threadsGen, "g", 1, "Count treads Gen")
	flag.IntVar(&batchSize, "bs", 1, "Batch Size")
	flag.IntVar(&size, "n", 1, "Size")
}

func reducer(input chan int, output chan int) {
	defer close(output)
	sum := 0
	for item := range input {
		sum += item
	}
	output <- sum
}

func Reduce(input chan int) <-chan int {
	oneShot := make(chan int)
	go reducer(input, oneShot)
	return oneShot
}

func summary(slice []int8) (sum int) {
	for _, val := range slice {
		sum += int(val)
	}
	return
}

func generateN(slice []int8) []int8 {
	for i := range slice {
		slice[i] = int8(rand.Intn(127))
	}
	return slice
}

func pusher(v []int8, bs int) (output chan []int8) {
	output = make(chan []int8)
	size := len(v)
	go func() {
		defer close(output)
		for c := 0; c < size; c += bs {
			output <- v[c:min(c+bs, size)]
		}
	}()
	return
}

func work[I any, O any](input chan I, output chan O, wg *sync.WaitGroup, execute func(msg I) O) {
	defer wg.Done()
	for msg := range input {
		output <- execute(msg)
	}
}

func Map[I any, O any](input chan I, n int, execute func(msg I) O) (output chan O) {
	output = make(chan O)
	wg := &sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go work(input, output, wg, execute)
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return
}

func main() {
	flag.Parse()

	numbers := make([]int8, size)

	values := pusher(numbers, batchSize)
	bridge := Map(values, threadsGen, generateN)
	result := Map(bridge, threadsSum, summary)

	sum := <-Reduce(result)

	fmt.Println("Result sum: ", sum)
	fmt.Println("Len: ", len(numbers))
}
