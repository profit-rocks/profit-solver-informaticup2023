package main

const DepositResourceFactor = 5
const MaxDepositWithdrawPerMine = 3
const NumResourceTypes = 8

type SimulationState struct {
	scenario *Scenario
	solution *Solution
}

func (s *Scenario) evaluate(solution Solution) (int, error) {
	// TODO: Check for invalid scenario
	simulationState := simulationStateFromScenarioAndSolution(s, &solution)
	for i := 0; i < s.turns; i++ {
		simulationState.simulateOneRound()
	}
	score := 0
	for _, factory := range solution.factories {
		for i := 0; i < NumResourceTypes; i++ {
			score += factory.resourceStorage[i]
		}
	}
	return score, nil
}

func simulationStateFromScenarioAndSolution(scenario *Scenario, solution *Solution) SimulationState {
	for _, deposit := range scenario.deposits {
		deposit.remainingResources = deposit.width * deposit.height * DepositResourceFactor
	}
	for _, factory := range solution.factories {
		factory.resourceStorage = []int{0, 0, 0, 0, 0, 0, 0, 0}
	}
	for _, mine := range solution.mines {
		mine.resourcesEgress = []int{0, 0, 0, 0, 0, 0, 0, 0}
		mine.resourcesIngress = []int{0, 0, 0, 0, 0, 0, 0, 0}
	}
	return SimulationState{scenario: scenario, solution: solution}
}

func (state *SimulationState) simulateOneRound() {
	// Transfer resources from mine egresses to factories
	for _, factory := range state.solution.factories {
		for _, mine := range factory.getAdjacentMines(state.solution) {
			for i := 0; i < NumResourceTypes; i++ {
				factory.resourceStorage[i] += mine.resourcesEgress[i]
				mine.resourcesEgress[i] = 0
			}
		}
	}
	// Transfer resources from mine ingresses to mine egresses
	for _, mine := range state.solution.mines {
		for i := 0; i < NumResourceTypes; i++ {
			mine.resourcesEgress[i] += mine.resourcesIngress[i]
			mine.resourcesIngress[i] = 0
		}
	}
	// Transfer resources from deposits to mine ingresses
	for _, deposit := range state.scenario.deposits {
		adjacentMines := deposit.getAdjacentMines(state.solution)
		//TODO: mix mines
		for _, mine := range adjacentMines {
			withdrawAmount := 0
			if deposit.remainingResources >= MaxDepositWithdrawPerMine {
				//withdrawAmount = rand.Intn(3) + 1
				withdrawAmount = MaxDepositWithdrawPerMine
			} else {
				//TODO: randomize remaining amount
				withdrawAmount = deposit.remainingResources
			}
			deposit.remainingResources -= withdrawAmount
			mine.resourcesIngress[deposit.subtype] += withdrawAmount
		}
	}
}

func (f *Factory) getAdjacentMines(solution *Solution) []*Mine {
	mines := make([]*Mine, 0)
	for _, position := range f.mineEgressPositions() {
		mine, foundMine := getMineWithEgressAt(solution, position)
		if foundMine {
			mines = append(mines, mine)
		}
	}
	return mines
}

func (d *Deposit) getAdjacentMines(solution *Solution) []*Mine {
	mines := make([]*Mine, 0)
	for _, position := range d.mineIngressPositions() {
		mine, foundMine := getMineWithIngressAt(solution, position)
		if foundMine {
			mines = append(mines, mine)
		}
	}
	return mines
}

func (f *Factory) mineEgressPositions() []Position {
	positions := make([]Position, 0)
	for i := 0; i < FactoryWidth; i++ {
		positions = append(positions, Position{
			x: f.position.x + i,
			y: f.position.y - 1,
		})
		positions = append(positions, Position{
			x: f.position.x + i,
			y: f.position.y + FactoryHeight,
		})
	}
	for i := 0; i < FactoryHeight; i++ {
		positions = append(positions, Position{
			x: f.position.x - 1,
			y: f.position.y + i,
		})
		positions = append(positions, Position{
			x: f.position.x + FactoryWidth,
			y: f.position.y + i,
		})
	}
	return positions
}

func (d *Deposit) mineIngressPositions() []Position {
	positions := make([]Position, 0)
	for i := 0; i < d.width; i++ {
		positions = append(positions, Position{
			x: d.position.x + i,
			y: d.position.y - 1,
		})
		positions = append(positions, Position{
			x: d.position.x + i,
			y: d.position.y + d.height,
		})
	}
	for i := 0; i < d.height; i++ {
		positions = append(positions, Position{
			x: d.position.x - 1,
			y: d.position.y + i,
		})
		positions = append(positions, Position{
			x: d.position.x + d.width,
			y: d.position.y + i,
		})
	}
	return positions
}

func getMineWithIngressAt(solution *Solution, position Position) (*Mine, bool) {
	for _, mine := range solution.mines {
		if mine.Ingress() == position {
			return mine, true
		}
	}
	return &Mine{}, false
}

func getMineWithEgressAt(solution *Solution, position Position) (*Mine, bool) {
	for _, mine := range solution.mines {
		if mine.Egress() == position {
			return mine, true
		}
	}
	return &Mine{}, false
}
