package main

import "testing"

func TestEmptySolutionEvaluation(t *testing.T) {
	scenario := largeEmptyScenario()
	solution := Solution{}
	score, err := scenario.evaluateSolution(solution)
	if err != nil {
		t.Errorf("evaluating empty solution should not return an error %v", err)
	}
	if score != 0 {
		t.Errorf("score should be 0 and not %d", score)
	}
}

func TestSolutionForLargeScenarioWithDepositEvaluation(t *testing.T) {
	scenario := largeScenarioWithDeposit()
	solution := solutionForLargeScenarioWithDeposit()
	score, err := scenario.evaluateSolution(solution)
	if err != nil {
		t.Errorf("evaluating empty solution should not return an error %v", err)
	}
	if score != 24 {
		t.Errorf("score should be 24 and not %d", score)
	}
}

func TestSolutionWithPathForLargeScenarioWithDepositEvaluation(t *testing.T) {
	scenario := largeScenarioWithDeposit()
	solution := solutionWithPathForLargeScenarioWithDeposit()
	score, err := scenario.evaluateSolution(solution)
	if err != nil {
		t.Errorf("evaluating empty solution should not return an error %v", err)
	}
	if score != 18 {
		t.Errorf("score should be 18 and not %d", score)
	}
}

func TestInvalidSolutionEvaluation(t *testing.T) {
	scenario := largeEmptyScenario()
	solution := invalidSolutionForLargeEmptyScenario()
	score, err := scenario.evaluateSolution(solution)
	if err == nil {
		t.Errorf("evaluating empty solution should throw an error")
	}
	if score != 0 {
		t.Errorf("score of invalid solution should be 0 and not %d", score)
	}
}
