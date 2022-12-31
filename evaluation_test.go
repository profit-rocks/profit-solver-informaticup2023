package main

import (
	"fmt"
	"testing"
)

func TestEmptySolutionEvaluation(t *testing.T) {
	scenario := largeEmptyScenario()
	solution := Solution{}
	score, _, err := scenario.evaluateSolution(solution)
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
	score, _, err := scenario.evaluateSolution(solution)
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
	score, _, err := scenario.evaluateSolution(solution)
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
	score, _, err := scenario.evaluateSolution(solution)
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

// TODO: Check out subtests to remove code duplication in the following tests (https://go.dev/blog/subtests)

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

func TestEvaluationOfSolutionWithCombiner(t *testing.T) {
	scenario, solution, err := importFromProfitJson("fixtures/solutionWithCombiner.json")
	if err != nil {
		t.Errorf("import of fixture failed")
	}

	score, err := scenario.evaluateSolution(solution)
	expectedScore := 60
	if score != expectedScore {
		t.Errorf("score should be %d and not %d", expectedScore, score)
	}
}

func TestEvaluationOfSolutionWithMineCombinerPath(t *testing.T) {
	scenario, solution, err := importFromProfitJson("fixtures/solutionWithMineCombinerPath.json")
	if err != nil {
		t.Errorf("import of fixture failed")
	}

	score, err := scenario.evaluateSolution(solution)
	expectedScore := 20
	if score != expectedScore {
		t.Errorf("score should be %d and not %d", expectedScore, score)
	}
}

func TestEvaluationOfSolutionWithMultipleCombinersPath(t *testing.T) {
	scenario, solution, err := importFromProfitJson("fixtures/solutionWithMultipleCombinersPath.json")
	if err != nil {
		t.Errorf("import of fixture failed")
	}

	score, err := scenario.evaluateSolution(solution)
	expectedScore := 80
	if score != expectedScore {
		t.Errorf("score should be %d and not %d", expectedScore, score)
	}
}

func TestEvaluationOfSolutionWithAdjacentCombiners(t *testing.T) {
	scenario, solution, err := importFromProfitJson("fixtures/solutionWithAdjacentCombiners.json")
	if err != nil {
		t.Errorf("import of fixture failed")
	}

	score, err := scenario.evaluateSolution(solution)
	expectedScore := 70
	if score != expectedScore {
		t.Errorf("score should be %d and not %d", expectedScore, score)
	}
}

func TestEvaluationOfSolutionWithCombinerNextToFactory(t *testing.T) {
	scenario, solution, err := importFromProfitJson("fixtures/solutionWithCombinerNextToFactory.json")
	if err != nil {
		t.Errorf("import of fixture failed")
	}

	score, err := scenario.evaluateSolution(solution)
	expectedScore := 80
	if score != expectedScore {
		t.Errorf("score should be %d and not %d", expectedScore, score)
	}
}

func TestEvaluationOfSolutionWithCombiningCombiner(t *testing.T) {
	scenario, solution, err := importFromProfitJson("fixtures/solutionWithCombiningCombiner.json")
	if err != nil {
		t.Errorf("import of fixture failed")
	}

	score, err := scenario.evaluateSolution(solution)
	expectedScore := 100
	if score != expectedScore {
		t.Errorf("score should be %d and not %d", expectedScore, score)
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
		t.Errorf("invalid solution with overlapping conveyors should be invalid")
	}
}

func TestSolutionWithOverlappingConveyorsInSamePathIsInvalid(t *testing.T) {
	scenario := largeEmptyScenario()
	solution := Solution{paths: []Path{{
		conveyors: []Conveyor{{position: Position{x: 3, y: 3}, direction: Right, length: Long}, {position: Position{x: 3, y: 3}, direction: Left, length: Long}},
	}}}
	err := scenario.checkValidity(solution)
	if err == nil {
		t.Errorf("invalid solution with overlapping conveyors should be invalid")
	}
}

func TestEvaluationOfSolutionWithDisconnectedMines(t *testing.T) {
	scenario, solution, err := importFromProfitJson("fixtures/solutionWithDisconnectedMines.json")
	if err != nil {
		t.Errorf("import of fixture failed")
	}

	score, err := scenario.evaluateSolution(solution)
	expectedScore := 300
	if score != expectedScore {
		t.Errorf("score should be %d and not %d", expectedScore, score)
	}
}
