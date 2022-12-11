package main

import "testing"

func TestAllNumbers(t *testing.T) {
	const maximum = 1000
	rng := NewLehmerRNG(maximum)
	seen := make(map[int]bool)
	for i := 0; i < maximum; i++ {
		n, done := rng.Next()
		if done && i != maximum-1 {
			t.Errorf("RNG returned to seed after %d numbers", i)
		}
		if seen[n] {
			t.Errorf("duplicate number %d", n)
		}
		seen[n] = true
	}
}
