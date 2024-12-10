/* Task 1:
 * 1) Generate slice (number) in N threads
 * 2) Summarize slice in N threads
 */

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func timeIt(f func()) time.Duration {
	t := time.Now()
	f()
	return time.Since(t)
}

func reducer(input chan int, output chan int) {
	defer close(output)
	sum := 0
	for item := range input {
		sum += item
	}
	output <- sum
}

func summarize(input []int8, output chan int, signal *sync.WaitGroup) {
	defer signal.Done()
	sum := 0
	for _, val := range input {
		sum += int(val)
	}
	output <- sum
}

func generateN(nums []int8) {
	for i := range nums {
		nums[i] = int8(rand.Intn(127))
	}
	// fmt.Printf("Generate slice: %v  Len: %d\n", nums, len(nums))
}

func generator(input chan []int8, output chan []int8, signal *sync.WaitGroup) {
	defer signal.Done()
	slice := <-input
	generateN(slice)
	output <- slice
}

func main() {
	size := 10_000
	treads := 4
	batch_size := size / treads

	number := make([]int8, size)
	var signalGeneration sync.WaitGroup
	var signalSummary sync.WaitGroup

	inputSlice := make(chan []int8)
	outputSlice := make(chan []int8, treads)
	outputSum := make(chan int, treads)
	result := make(chan int)

	var dur time.Duration

	dur = timeIt(func() {
		cursor := 0
		defer close(inputSlice)
		for i := 0; i < treads; i++ {
			signalGeneration.Add(1)
			go generator(inputSlice, outputSlice, &signalGeneration)
		}
		for i := 0; i < treads; i++ {
			if i == treads-1 {
				inputSlice <- number[cursor:size]
			} else {
				inputSlice <- number[cursor : cursor+batch_size]
				cursor += batch_size
			}
		}
		go func() {
			signalGeneration.Wait()
			close(outputSlice)
		}()
		// fmt.Printf("Biuld slice: %v  Len: %d\n", number, len(number))
	})
	fmt.Println("Time generation result:", dur)

	dur = timeIt(func() {
		for slice := range outputSlice {
			signalSummary.Add(1)
			go summarize(slice, outputSum, &signalSummary)
		}
		go func() {
			signalSummary.Wait()
			close(outputSum)
		}()

		go reducer(outputSum, result)

		sum := <-result
		fmt.Println("Sum = ", sum)
	})
	fmt.Println("Time summarize result:", dur)
}
