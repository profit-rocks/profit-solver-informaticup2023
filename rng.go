package main

import (
	"math/rand"
)

// UniqueRNG is a random number generator that only outputs unique numbers. It generates numbers in [0, n).
type UniqueRNG struct {
	state  int
	n      int
	buffer []int
}

func NewUniqueRNG(n int) *UniqueRNG {
	rng := &UniqueRNG{
		state:  0,
		n:      n,
		buffer: make([]int, n),
	}
	for i := 0; i < n; i++ {
		rng.buffer[i] = i
	}
	rand.Shuffle(n, func(i, j int) {
		rng.buffer[i], rng.buffer[j] = rng.buffer[j], rng.buffer[i]
	})
	return rng
}

func (rng *UniqueRNG) Next() (int, bool) {
	num := rng.buffer[rng.state]
	rng.state++
	if rng.state == rng.n {
		return num, true
	}
	return num, false
}
