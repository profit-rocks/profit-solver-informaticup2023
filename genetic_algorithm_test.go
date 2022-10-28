package main

import "testing"

func largeEmptyScenario() Scenario {
	return Scenario{
		width:     20,
		height:    20,
		deposits:  []Deposit{},
		obstacles: []Obstacle{},
		turns:     100,
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
				t.Errorf("position %v should be available", Position{
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
				t.Errorf("position %v should not be available", Position{
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
				t.Errorf("position %v should not be available", Position{
					x: x,
					y: y,
				})
			}
		}
	}
}

func smallEmptyScenario() Scenario {
	return Scenario{
		width:     4,
		height:    4,
		deposits:  []Deposit{},
		obstacles: []Obstacle{},
		turns:     100,
	}
}

func TestSmallEmptyScenario(t *testing.T) {
	scenario := smallEmptyScenario()
	for x := 0; x < scenario.width; x++ {
		for y := 0; y < scenario.height; y++ {
			pos := Position{x, y}
			if isPositionAvailableForFactory(scenario, Chromosome{}, pos) {
				t.Errorf("position %v should not be available", pos)
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
		turns: 100,
	}
}

func TestScenarioWithObstacles(t *testing.T) {
	scenario := scenarioWithObstacle()

	for x := 0; x < scenario.width; x++ {
		for y := 0; y < scenario.height; y++ {
			pos := Position{x, y}
			if isPositionAvailableForFactory(scenario, Chromosome{}, pos) {
				t.Errorf("position %v should not be available", pos)
			}
		}
	}

}

func scenarioWithFactory() (Scenario, Chromosome) {
	return Scenario{
			width:     15,
			height:    15,
			deposits:  []Deposit{},
			obstacles: []Obstacle{},
			turns:     100,
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
					t.Errorf("position %v should not be available", pos)
				}
			} else {
				if !isPositionAvailableForFactory(scenario, chromosome, pos) {
					t.Errorf("position %v should be available", pos)
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
					t.Errorf("position %v should be available", pos)
				}
			} else {
				if isPositionAvailableForFactory(scenario, Chromosome{}, pos) {
					t.Errorf("position %v should not be available", pos)
				}
			}
		}
	}
}

func TestAvailableMinePositions(t *testing.T) {

	validMines := []Mine{
		{Position{6, 3}, Right},
		{Position{7, 3}, Right},
		{Position{8, 3}, Right},
		{Position{9, 3}, Right},
		{Position{10, 3}, Right},
		{Position{11, 4}, Right},
		{Position{11, 5}, Right},
		{Position{11, 6}, Right},
		{Position{11, 7}, Right},
		{Position{11, 8}, Right},
		{Position{10, 9}, Right},
		{Position{4, 10}, Bottom},
		{Position{5, 11}, Bottom},
		{Position{6, 11}, Bottom},
		{Position{7, 11}, Bottom},
		{Position{8, 11}, Bottom},
		{Position{9, 11}, Bottom},
		{Position{10, 10}, Bottom},
		{Position{10, 10}, Bottom},
		{Position{10, 9}, Bottom},
		{Position{10, 8}, Bottom},
		{Position{10, 7}, Bottom},
		{Position{10, 6}, Bottom},
		{Position{3, 4}, Left},
		{Position{2, 5}, Left},
		{Position{2, 6}, Left},
		{Position{2, 7}, Left},
		{Position{2, 8}, Left},
		{Position{2, 9}, Left},
		{Position{3, 10}, Left},
		{Position{4, 10}, Left},
		{Position{5, 10}, Left},
		{Position{6, 10}, Left},
		{Position{7, 10}, Left},
		{Position{9, 3}, Top},
		{Position{8, 2}, Top},
		{Position{7, 2}, Top},
		{Position{6, 2}, Top},
		{Position{5, 2}, Top},
		{Position{4, 2}, Top},
		{Position{3, 3}, Top},
		{Position{3, 4}, Top},
		{Position{3, 5}, Top},
		{Position{3, 6}, Top},
		{Position{3, 7}, Top},
	}

	scenario := scenarioWithDeposit()
	mines := minesAroundDeposit(scenario, scenario.deposits[0], Chromosome{})

	for _, mine := range mines {
		found := false
		for _, validMine := range validMines {
			if mine.position == validMine.position && mine.direction == validMine.direction {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("mine %v is not valid", mine)
		}
	}

	for _, validMine := range validMines {
		found := false
		for _, mine := range mines {
			if mine.position == validMine.position && mine.direction == validMine.direction {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("mine %v is not found", validMine)
		}
	}
}
