package main

import (
	"testing"
)

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
	expectedScore := 240
	if score != expectedScore {
		t.Errorf("score should be %d and not %d", expectedScore, score)
	}
}

func TestSolutionWithPathForLargeScenarioWithDepositEvaluation(t *testing.T) {
	scenario := largeScenarioWithDeposit()
	solution := solutionWithPathForLargeScenarioWithDeposit()
	score, err := scenario.evaluateSolution(solution)
	if err != nil {
		t.Errorf("evaluating empty solution should not return an error %v", err)
	}
	expectedScore := 180
	if score != expectedScore {
		t.Errorf("score should be %d and not %d", expectedScore, score)
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

func TestSolutionWithOverlappingFactoriesInvalid(t *testing.T) {
	scenario := largeEmptyScenario()
	solution := Solution{
		factories: []Factory{{
			position: Position{0, 0},
			product:  0,
		}, {
			position: Position{0, 0},
			product:  0,
		}},
	}
	err := scenario.checkValidity(solution)
	if err == nil {
		t.Errorf("two factories at same position should not be valid")
	}
}

func TestSolutionWithMultipleIngressesAtEgressInvalid(t *testing.T) {
	scenario, solution, err := importFromProfitJson("fixtures/solutionMultipleIngressesAtEgress.json")
	if err != nil {
		t.Errorf("solution should be importable")
	}
	err = scenario.checkValidity(solution)
	if err == nil {
		t.Errorf("solution with two ingresses at an egress should not be valid")
	}
}

func TestSolutionWithOverlappingConveyorsIsValid(t *testing.T) {
	scenario, solution, err := importFromProfitJson("fixtures/solutionWithOverlappingConveyors.json")
	if err != nil {
		t.Errorf("solution should be importable")
	}
	err = scenario.checkValidity(solution)
	if err != nil {
		t.Errorf("solution with overlapping conveyorrs should be valid")
	}
}

func TestSolutionWithOverlappingConveyorsIsInvalid(t *testing.T) {
	scenario, solution, err := importFromProfitJson("fixtures/invalidSolutionWithOverlappingConveyors.json")
	if err != nil {
		t.Errorf("solution should be importable")
	}
	err = scenario.checkValidity(solution)
	if err == nil {
		t.Errorf("invalid solution with overlapping conveyorrs should be invalid")
	}
}

func TestSolutionWithOverlappingConveyorsInSamePathIsInvalid(t *testing.T) {
	scenario := largeEmptyScenario()
	solution := Solution{paths: []Path{{
		conveyors: []Conveyor{{position: Position{x: 3, y: 3}, direction: Right, length: Long}, {position: Position{x: 3, y: 3}, direction: Left, length: Long}},
	}}}
	err := scenario.checkValidity(solution)
	if err == nil {
		t.Errorf("invalid solution with overlapping conveyorrs should be invalid")
	}
}
