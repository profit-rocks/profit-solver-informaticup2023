package main

import "testing"

func largeEmptyScenario() Scenario {
	return Scenario{
		width:        20,
		height:       20,
		deposits:     []Deposit{},
		obstacles:    []Obstacle{},
		turns:        100,
		numFactories: 4,
	}
}

func TestLargeEmptyScenarioIsAvailable(t *testing.T) {
	scenario := largeEmptyScenario()
	for x := 0; x <= scenario.width-FactoryWidth; x++ {
		for y := 0; y <= scenario.height-FactoryHeight; y++ {
			if !isPositionAvailableForFactory(scenario, Chromosome{}, Position{
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
			if isPositionAvailableForFactory(scenario, Chromosome{}, Position{
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
			if isPositionAvailableForFactory(scenario, Chromosome{}, Position{
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
		turns:        100,
		numFactories: 4,
	}
}

func TestSmallEmptyScenario(t *testing.T) {
	scenario := smallEmptyScenario()
	for x := 0; x < scenario.width; x++ {
		for y := 0; y < scenario.height; y++ {
			pos := Position{x, y}
			if isPositionAvailableForFactory(scenario, Chromosome{}, pos) {
				t.Errorf("Position %v should not be available", pos)
			}
		}
	}
}

func scenarioWithObstacle() Scenario {
	return Scenario{
		width:    10,
		height:   10,
		deposits: []Deposit{},
		obstacles: []Obstacle{
			{
				position: Position{
					x: 4,
					y: 4,
				},
				width:  2,
				height: 2,
			},
		},
		turns:        100,
		numFactories: 4,
	}
}

func TestScenarioWithObstacles(t *testing.T) {
	scenario := scenarioWithObstacle()

	for x := 0; x < scenario.width; x++ {
		for y := 0; y < scenario.height; y++ {
			pos := Position{x, y}
			if isPositionAvailableForFactory(scenario, Chromosome{}, pos) {
				t.Errorf("Position %v should not be available", pos)
			}
		}
	}

}

func scenarioWithFactory() (Scenario, Chromosome) {
	return Scenario{
			width:        15,
			height:       15,
			deposits:     []Deposit{},
			obstacles:    []Obstacle{},
			turns:        100,
			numFactories: 4,
		}, Chromosome{
			factories: []Factory{
				{
					position: Position{x: 5, y: 5},
					product:  0,
				},
			},
		}
}

func TestScenarioWithFactory(t *testing.T) {
	scenario, chromosome := scenarioWithFactory()

	for x := 0; x <= scenario.width-FactoryWidth; x++ {
		for y := 0; y <= scenario.height-FactoryHeight; y++ {
			pos := Position{x, y}
			if x > 5-FactoryWidth && x < 5+FactoryWidth && y > 5-FactoryHeight && y < 5+FactoryWidth {
				if isPositionAvailableForFactory(scenario, chromosome, pos) {
					t.Errorf("Position %v should not be available", pos)
				}
			} else {
				if !isPositionAvailableForFactory(scenario, chromosome, pos) {
					t.Errorf("Position %v should be available", pos)
				}
			}
		}
	}
}

func scenarioWithDeposit() Scenario {
	return Scenario{
		width:     15,
		height:    15,
		deposits:  []Deposit{{position: Position{x: 5, y: 5}, width: 5, height: 5, subtype: 0}},
		obstacles: []Obstacle{},
		turns:     100,
	}
}

func TestScenarioWithDeposit(t *testing.T) {
	scenario := scenarioWithDeposit()

	for x := 0; x <= scenario.width-FactoryWidth; x++ {
		for y := 0; y <= scenario.height-FactoryHeight; y++ {
			pos := Position{x, y}
			if x == 0 && y == 0 || x == 0 && y == 10 || x == 10 && y == 0 || x == 10 && y == 10 {
				if !isPositionAvailableForFactory(scenario, Chromosome{}, pos) {
					t.Errorf("Position %v should be available", pos)
				}
			} else {
				if isPositionAvailableForFactory(scenario, Chromosome{}, pos) {
					t.Errorf("Position %v should not be available", pos)
				}
			}
		}
	}
}
