package main

import (
	"math/big"
	"math/rand"
)

// LehmerRNG is a Lehmer random number generator. Its key property is that it has a fixed period, specified by maximum.
// The RNG is not perfect, but should be good enough for our purpose: Randomly iterating through possible positions in O(1) space.
type LehmerRNG struct {
	state      int
	multiplier int
	seed       int
	modulus    int
	maximum    int
}

var nextPrimeNumber = make(map[int]int)

func nextPrime(n int) int {
	for {
		n++
		if big.NewInt(int64(n)).ProbablyPrime(20) {
			return n
		}
	}
}

func NewLehmerRNG(maximum int) *LehmerRNG {
	if nextPrimeNumber[maximum] == 0 {
		nextPrimeNumber[maximum] = nextPrime(maximum)
	}
	seed := rand.Intn(maximum-1) + 1
	multiplier := rand.Intn(maximum-2) + 2
	return &LehmerRNG{
		state:      seed,
		seed:       seed,
		multiplier: multiplier,
		modulus:    nextPrimeNumber[maximum],
		maximum:    maximum,
	}
}

func (rng *LehmerRNG) Next() (int, bool) {
	rng.state = (rng.state * rng.multiplier) % rng.modulus
	for rng.state > rng.maximum {
		rng.state = (rng.state * rng.multiplier) % rng.modulus
	}
	return rng.state, rng.state == rng.seed
}
