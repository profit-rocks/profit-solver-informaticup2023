package main

import "testing"

func TestEmptySolutionEvaluation(t *testing.T) {
	scenario := largeEmptyScenario()
	solution := Solution{}
	score, err := scenario.evaluate(solution)
	if err != nil {
		t.Errorf("evaluating empty solution should not return an error #{err}")
	}
	if score != 0 {
		t.Errorf("score should be 0 and not #{score}")
	}
}
