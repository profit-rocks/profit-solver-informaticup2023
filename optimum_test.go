package main

import "testing"

func TestOptimumTask1(t *testing.T) {
	scenario, _, err := ImportScenario("official_tasks/002.task.json")
	if err != nil {
		t.Errorf("failed to import fixture: %e", err)
	}
	optimum, err := TheoreticalOptimum(scenario)
	if optimum != 120 {
		t.Errorf("optimum should be 120")
	}
}
