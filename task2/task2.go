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

var treadsSum int
var treadsGen int
var sizeNumber int
var batchSize int

func init() {
	flag.IntVar(&treadsSum, "s", 1, "Count treads Sum")
	flag.IntVar(&treadsGen, "g", 1, "Count treads Gen")
	flag.IntVar(&batchSize, "b", 1, "Slice size")
	flag.IntVar(&sizeNumber, "sN", 1, "Batch Size")
}

func reducer(input chan int, output chan int, signal *sync.WaitGroup) {
	defer signal.Done()
	sum := 0
	for item := range input {
		sum += item
	}
	output <- sum

}

func summary(slice []int8) int {
	sum := 0
	for _, val := range slice {
		sum += int(val)
	}
	return sum
}

func summarize(input chan []int8, output chan int, signal *sync.WaitGroup) {
	defer signal.Done()
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

func main() {
	flag.Parse()

	var signalGeneration sync.WaitGroup
	var signalSummarize sync.WaitGroup
	var signalReduce sync.WaitGroup

	batchCount := sizeNumber / batchSize
	number := make([]int8, sizeNumber)
	sum := 0

	inputSlice := make(chan []int8)
	outputSlice := make(chan []int8)
	inputSum := make(chan int)
	result := make(chan int)

	go func() {
		defer close(inputSlice)
		cursor := 0
		for i := 0; i < batchCount; i++ {
			if i == batchCount-1 {
				inputSlice <- number[cursor:sizeNumber]
			} else {
				inputSlice <- number[cursor : cursor+batchSize]
				cursor += batchSize
			}
		}
	}()

	for i := 0; i < treadsGen; i++ {
		signalGeneration.Add(1)
		go generator(inputSlice, outputSlice, &signalGeneration)
	}

	go func() {
		signalGeneration.Wait()
		close(outputSlice)
	}()

	for i := 0; i < treadsSum; i++ {
		signalSummarize.Add(1)
		go summarize(outputSlice, inputSum, &signalSummarize)
	}
	go func() {
		signalSummarize.Wait()
		close(inputSum)
	}()

	signalReduce.Add(1)
	go reducer(inputSum, result, &signalReduce)
	sum = <-result

	go func() {
		signalReduce.Wait()
		close(result)
	}()
	fmt.Println("Result sum: ", sum)
	fmt.Println("Len: ", len(number))
}
