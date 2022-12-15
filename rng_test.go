package main

import "testing"

func TestAllNumbers(t *testing.T) {
	for _, maximum := range []int{100, 1000, 10000} {
		rng := NewUniqueRNG(maximum)
		seen := make(map[int]bool)
		for i := 0; i < maximum; i++ {
			n, done := rng.Next()
			if done && i != maximum-1 {
				t.Errorf("RNG done too early after %d numbers", i)
			}
			if n >= maximum {
				t.Errorf("RNG returned number %d >= maximum %d", n, maximum)
			}
			if seen[n] {
				t.Errorf("duplicate number %d", n)
			}
			seen[n] = true
		}
	}
}
