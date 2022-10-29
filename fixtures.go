package main

func largeEmptyScenario() Scenario {
	return Scenario{
		width:     20,
		height:    20,
		deposits:  []Deposit{},
		obstacles: []Obstacle{},
		turns:     100,
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

func scenarioWithObstacle() Scenario {
	return Scenario{
		width:     10,
		height:    10,
		deposits:  []Deposit{},
		obstacles: []Obstacle{{Position{4, 4}, 2, 2}}, turns: 100,
	}
}

func scenarioWithDeposit() Scenario {
	return Scenario{
		width:     15,
		height:    15,
		deposits:  []Deposit{{position: Position{5, 5}, width: 5, height: 5, subtype: 0}},
		obstacles: []Obstacle{},
		turns:     100,
	}
}
