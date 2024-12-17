package main

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
)

type Accumulator[T any] struct {
	input  chan T
	memory []T
	output chan T
}

func (a *Accumulator[T]) run() {
	defer close(a.output)
	for v := range a.input {
		a.memory = append(a.memory, v)
		a.output <- v
	}
}

func NewAccumulator[T any](input chan T) *Accumulator[T] {
	output := make(chan T)

	acc := &Accumulator[T]{
		input:  input,
		memory: []T{},
		output: output,
	}

	go acc.run()
	return acc
}

func fromSlice[T any](v []T) chan T {
	output := make(chan T)

	go func() {
		defer close(output)
		for _, x := range v {
			output <- x
		}
	}()

	return output
}

func buildExpected[T int](a1 *Accumulator[T], a2 *Accumulator[T]) []T {
	r := []T{}

	n1, n2 := len(a1.memory), len(a2.memory)
	i1, i2 := 0, 0

	for i1 < n1 && i2 < n2 {
		switch {
		case a1.memory[i1] < a2.memory[i2]:
			r = append(r, a1.memory[i1])
			i1++
		case a1.memory[i1] == a2.memory[i2]:
			i1++
			i2++
		case a1.memory[i1] > a2.memory[i2]:
			r = append(r, a2.memory[i2])
			i2++
		}
	}

	// copy leftovers in a1 or a2
	// NB: One of them is empty, so the order doesn't matter
	r = append(r, a1.memory[i1:]...)
	r = append(r, a2.memory[i2:]...)
	return r
}

func testAccumulator(a1, a2 *Accumulator[int]) bool {
	result := []int{}
	for v := range unique(a1.output, a2.output) {
		result = append(result, v)
	}

	golden := buildExpected(a1, a2)
	var msg string

	if len(result) != len(golden) {
		msg = fmt.Sprintf("Malformed result length (expected to be %d, got %d)", len(golden), len(result))
		goto fail
	}

	for i := 0; i < len(result); i++ {
		if golden[i] != result[i] {
			msg = fmt.Sprintf("result[%d] expected to be %d, got %d", i, golden[i], result[i])
			goto fail
		}
	}

	return true

fail:
	log.Printf("ch1: %+#v", a1.memory)
	log.Printf("ch2: %+#v", a2.memory)
	log.Printf("expected: %+#v", golden)
	log.Printf("result: %+#v", result)
	log.Println(msg)
	return false
}

func testIt(n1, n2 int) bool {
	a1 := NewAccumulator(generateN(n1))
	a2 := NewAccumulator(generateN(n2))

	return testAccumulator(a1, a2)
}

func TestRandom(t *testing.T) {
	plan := 1000
	oks := 0
	for i := 0; i < plan; i++ {
		if testIt(rand.Intn(100), rand.Intn(100)) {
			oks++
		}
	}

	if oks != plan {
		t.Fatalf("You failed %d / %d tests\n", plan-oks, plan)
	}
}

func TestFixed1(t *testing.T) {
	a1 := NewAccumulator(fromSlice([]int{}))
	a2 := NewAccumulator(fromSlice([]int{1, 2, 3, 4, 5}))

	if !testAccumulator(a1, a2) {
		t.FailNow()
	}
}
func TestFixed2(t *testing.T) {
	a1 := NewAccumulator(fromSlice([]int{1, 2, 3, 4, 5}))
	a2 := NewAccumulator(fromSlice([]int{}))

	if !testAccumulator(a1, a2) {
		t.FailNow()
	}
}
