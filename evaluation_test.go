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

type EvaluationTestConfig struct {
	pathToFixture string
	expectedScore int
	expectedTurns int
}

func TestEvaluationOfSolutions(t *testing.T) {
	configs := []EvaluationTestConfig{
		{"fixtures/solutionWithCombiner.json", 60, 15},
		{"fixtures/solutionWithMineCombinerPath.json", 20, 10},
		{"fixtures/solutionWithMultipleCombinersPath.json", 80, 19},
		{"fixtures/solutionWithAdjacentCombiners.json", 70, 20},
		{"fixtures/solutionWithCombinerNextToFactory.json", 80, 20},
		{"fixtures/solutionWithCombiningCombiner.json", 100, 39},
		{"fixtures/solutionWithDisconnectedMines.json", 300, 31}, // on profit.phinau this only needs 30 turns to reach score 300. This is due to random distribution of last few resources
	}

	for _, config := range configs {
		t.Run(fmt.Sprintf("Testing_%s", config.pathToFixture), func(t *testing.T) {
			scenario, solution, err := importFromProfitJson(config.pathToFixture)
			if err != nil {
				t.Errorf("import of fixture failed")
			}
			score, turns, err := scenario.evaluateSolution(solution)
			if score != config.expectedScore {
				t.Errorf("score should be %d and not %d", config.expectedScore, score)
			}
			if turns != config.expectedTurns {
				t.Errorf("turns should be %d and not %d", config.expectedTurns, turns)
			}
		})
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
