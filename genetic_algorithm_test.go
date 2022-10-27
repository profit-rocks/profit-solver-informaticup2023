package main

import "testing"

func largeEmptyScenario() Scenario {
	return Scenario{
		width:        20,
		height:       20,
		deposits:     []Deposit{},
		obstacles:    []Obstacle{},
		factories:    []Factory{},
		turns:        100,
		numFactories: 4,
	}
}

func TestLargeEmptyScenarioIsAvailable(t *testing.T) {
	scenario := largeEmptyScenario()
	for x := 0; x <= scenario.width-FactoryWidth; x++ {
		for y := 0; y <= scenario.height-FactoryHeight; y++ {
			if !isPositionAvailableForFactory(scenario, []Factory{}, Position{
				x: x,
				y: y,
			}) {
				t.Errorf("Position %v should be available", Position{
					x: x,
					y: y,
				})
			}
		}
	}
}

func TestLargeEmptyScenarioBorders(t *testing.T) {
	scenario := largeEmptyScenario()
	for x := 0; x <= scenario.width-FactoryWidth; x++ {
		for y := scenario.height - FactoryHeight + 1; y < scenario.width; y++ {
			if isPositionAvailableForFactory(scenario, []Factory{}, Position{
				x: x,
				y: y,
			}) {
				t.Errorf("Position %v should not be available", Position{
					x: x,
					y: y,
				})
			}
		}
	}
	for x := scenario.width - FactoryWidth + 1; x < scenario.width; x++ {
		for y := 0; y <= scenario.height-FactoryHeight; y++ {
			if isPositionAvailableForFactory(scenario, []Factory{}, Position{
				x: x,
				y: y,
			}) {
				t.Errorf("Position %v should not be available", Position{
					x: x,
					y: y,
				})
			}
		}
	}
}

func smallEmptyScenario() Scenario {
	return Scenario{
		width:        4,
		height:       4,
		deposits:     []Deposit{},
		obstacles:    []Obstacle{},
		factories:    []Factory{},
		turns:        100,
		numFactories: 4,
	}
}

func TestSmallEmptyScenario(t *testing.T) {
	scenario := smallEmptyScenario()
	for x := 0; x < scenario.width; x++ {
		for y := 0; y < scenario.height; y++ {
			if isPositionAvailableForFactory(scenario, []Factory{}, Position{
				x: x,
				y: y,
			}) {
				t.Errorf("Position %v should not be available", Position{
					x: x,
					y: y,
				})
			}
		}
	}
}
