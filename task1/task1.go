/* Task 1:
 * 1) Generate slice (number) in N threads
 * 2) Summarize slice in N threads
 */

package main

import (
	"fmt"
	"math/rand"
	"time"
)

func timeIt(f func()) time.Duration {
	t := time.Now()
	f()
	return time.Since(t)
}

func executeSum(nums []int8, output chan int) {
	sum := 0
	for _, r := range nums {
		sum += int(r)
	}
	output <- sum
}

func generateN(nums []int8) {
	for i := range nums {
		nums[i] = int8(rand.Intn(127))
	}
}

func generator(lenghtSlice int, output chan []int8, waitSignal chan bool) {
	defer func() { waitSignal <- true }()
	slice := make([]int8, lenghtSlice)
	generateN(slice)
	output <- slice
	fmt.Println("generation slice:", slice, len(slice), output)

}

func main() {
	size := 100
	chanel := 3
	batch_size := size / chanel

	number := []int8{}
	outputChan := make(chan []int8)
	waitSignal := make(chan bool)
	sumChan := make(chan int)

	var dur time.Duration

	dur = timeIt(func() {
		cursor := 0
		lenghtSlice := 0
		for i := 0; i < chanel; i++ {
			if i == chanel-1 {
				lenghtSlice = batch_size + size%chanel
				go generator(lenghtSlice, outputChan, waitSignal)
			} else {
				lenghtSlice = batch_size
				go generator(lenghtSlice, outputChan, waitSignal)
				cursor += batch_size
			}
		}
		go func() {
			defer close(outputChan)
			for i := 0; i < chanel; i++ {
				<-waitSignal
			}
		}()
		for item := range outputChan {
			number = append(number, item...)
		}
		fmt.Printf("Biuld slice: %d  Len: %d\n", number, len(number))
	})
	fmt.Println("Time generation result:", dur)

	sum := 0

	dur = timeIt(func() {
		cursor := 0
		for i := 0; i < chanel; i++ {
			if i == chanel-1 {
				go executeSum(number[cursor:size], sumChan)
			} else {
				go executeSum(number[cursor:cursor+batch_size], sumChan)
				cursor += batch_size
			}
		}
		for i := 0; i < chanel; i++ {
			sum += <-sumChan
		}
	})
	fmt.Println("Summarize result: ", sum, dur)
}
