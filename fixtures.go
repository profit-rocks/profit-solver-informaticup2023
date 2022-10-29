package main

func largeEmptyScenario() Scenario {
	return Scenario{
		width:     20,
		height:    20,
		deposits:  []*Deposit{},
		obstacles: []*Obstacle{},
		turns:     100,
	}
}

func invalidSolutionForLargeEmptyScenario() Solution {
	return Solution{
		factories: []*Factory{{
			position:        Position{0, 0},
			product:         0,
			resourceStorage: nil,
		}, {
			position:        Position{0, 0},
			product:         0,
			resourceStorage: nil,
		}},
		mines:     nil,
		conveyors: nil,
	}
}

func smallEmptyScenario() Scenario {
	return Scenario{
		width:     4,
		height:    4,
		deposits:  []*Deposit{},
		obstacles: []*Obstacle{},
		turns:     100,
	}
}

func scenarioWithObstacle() Scenario {
	return Scenario{
		width:     10,
		height:    10,
		deposits:  []*Deposit{},
		obstacles: []*Obstacle{{Position{4, 4}, 2, 2}}, turns: 100,
	}
}

func scenarioWithDeposit() Scenario {
	return Scenario{
		width:     15,
		height:    15,
		deposits:  []*Deposit{{position: Position{5, 5}, width: 5, height: 5, subtype: 0}},
		obstacles: []*Obstacle{},
		products: []*Product{{
			subtype:   0,
			points:    10,
			resources: []int{1, 0, 0, 0, 0, 0, 0, 0},
		}},
		turns: 10,
	}
}

func largeScenarioWithDeposit() Scenario {
	return Scenario{
		width:     20,
		height:    20,
		deposits:  []*Deposit{{position: Position{0, 0}, width: 5, height: 5, subtype: 0}},
		obstacles: []*Obstacle{},
		products: []*Product{{
			subtype:   0,
			points:    10,
			resources: []int{1, 0, 0, 0, 0, 0, 0, 0},
		}},
		turns: 10,
	}
}

func solutionForLargeScenarioWithDeposit() Solution {
	return Solution{
		factories: []*Factory{{
			position: Position{
				x: 9,
				y: 0,
			},
			product:         0,
			resourceStorage: nil,
		}},
		mines: []*Mine{{
			position: Position{
				x: 6,
				y: 1,
			},
			direction:        0,
			resourcesIngress: nil,
			resourcesEgress:  nil,
		}},
		conveyors: nil,
	}
}

func solutionWithSingleMineForLargeEmptyScenario() Solution {
	return Solution{
		factories: []*Factory{},
		mines: []*Mine{{
			position: Position{
				x: 6,
				y: 1,
			},
			direction:        0,
			resourcesIngress: nil,
			resourcesEgress:  nil,
		}},
		conveyors: nil,
	}
}
