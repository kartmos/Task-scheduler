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
func summary(input chan int, output chan int) {
	sum := 0
	defer close(output)
	for item := range input {
		sum += item
	}
	output <- sum

}
func summarize(input chan []int8, output chan int, signal *sync.WaitGroup) {
	defer signal.Done()
	sum := 0
	for item := range input {
		slice := item
		for _, val := range slice {
			sum += int(val)
		}
	}
	output <- sum
}

func generateN(nums []int8) {
	for i := range nums {
		nums[i] = int8(rand.Intn(127))
	}
}

func generator(input chan []int8, output chan []int8, signal *sync.WaitGroup) {
	defer signal.Done()
	slice := <-input
	generateN(slice)
	output <- slice
}

func main() {
	size := 100
	treads := 3
	batch_size := size / treads

	number := make([]int8, size)
	var signal sync.WaitGroup
	inputSlice := make(chan []int8)
	outputSlice := make(chan []int8, treads)
	inputSum := make(chan []int8, treads)
	outputSum := make(chan int, treads)
	resault := make(chan int)

	var dur time.Duration

	dur = timeIt(func() {
		cursor := 0
		defer close(inputSlice)
		for i := 0; i < treads; i++ {
			signal.Add(1)
			go generator(inputSlice, outputSlice, &signal)
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
			signal.Wait()
			close(outputSlice)
		}()
		index := 0
		for item := range outputSlice {
			for _, val := range item {
				number[index] = val
				index++
			}
		}
		fmt.Printf("Biuld slice: %d  Len: %d\n", number, len(number))
	})
	fmt.Println("Time generation result:", dur)

	dur = timeIt(func() {
		defer close(inputSum)
		for slice := range outputSlice {
			inputSum <- slice
		}
		for i := 0; i < treads; i++ {
			signal.Add(1)
			go summarize(inputSum, outputSum, &signal)
		}
		go func() {
			signal.Wait()
			close(outputSum)
		}()
		summary(outputSum, resault)
		sum := <-resault
		fmt.Println("Sum = ", sum)

	})
}
