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
	flag.IntVar(&batchSize, "b", 1, "Slice size")
	flag.IntVar(&size, "sN", 1, "Batch Size")
}

func reducer(input chan int, output chan int) {
	defer close(output)
	sum := 0
	for item := range input {
		sum += item
	}
	output <- sum
}

func reduce(input chan int) <-chan int {
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

func summarize(input chan []int8, output chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for slice := range input {
		output <- summary(slice)
	}
}

func generateN(slice []int8) []int8 {
	for i := range slice {
		slice[i] = int8(rand.Intn(127))
	}
	return slice
}

func generator(input chan []int8, output chan []int8, signal *sync.WaitGroup) {
	defer signal.Done()
	for slice := range input {
		output <- generateN(slice)
	}
}

func guard(wg *sync.WaitGroup, closeChan any) {
	wg.Wait()

	switch closeChan := closeChan.(type) {
	case chan []int8:
		close(closeChan)
	case chan int:
		close(closeChan)
	default:
		panic("wtf don't what to do")
	}
}

// Generic guard
func gg[T chan int | chan []int8](wg *sync.WaitGroup, closeChan T) {
	wg.Wait()
	close(closeChan)
}

func pusher(v []int8, bs int) (output chan []int8) {
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

func pool[I any, O any](input chan I, n int, execute func(msg I) O) (output chan O) {
	wg := &sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
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

	var wgGen sync.WaitGroup
	var wgSum sync.WaitGroup

	numbers := make([]int8, size)

	input := pusher(numbers, batchSize)
	bridge := make(chan []int8)

	for i := 0; i < threadsGen; i++ {
		wgGen.Add(1)
		go generator(input, bridge, &wgGen)
	}

	go gg(&wgGen, bridge)

	output := make(chan int)

	for i := 0; i < threadsSum; i++ {
		wgSum.Add(1)
		go summarize(bridge, output, &wgSum)
	}

	go gg(&wgSum, output)

	sum := <-reduce(output)

	fmt.Println("Result sum: ", sum)
	fmt.Println("Len: ", len(numbers))
}
