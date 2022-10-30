package main

import "errors"

const DepositResourceFactor = 5
const MaxDepositWithdrawPerMine = 3
const NumResourceTypes = 8

type Simulation struct {
	scenario  Scenario
	factories []SimulatedFactory
	deposits  []SimulatedDeposit
	mines     []SimulatedMine
	conveyors []SimulatedConveyor
}

type SimulatedDeposit struct {
	deposit            Deposit
	remainingResources int
}

type SimulatedFactory struct {
	factory         Factory
	resourceStorage []int
}

type SimulatedMine struct {
	mine             Mine
	resourcesIngress []int
	resourcesEgress  []int
}

type SimulatedConveyor struct {
	conveyor         Conveyor
	resourcesIngress []int
	resourcesEgress  []int
}

func (s Scenario) checkValidity(solution Solution) error {
	mines := make([]Mine, len(solution.mines))
	for i, mine := range solution.mines {
		mines[i] = mine
	}
	factories := make([]Factory, len(solution.factories))
	for i, factory := range solution.factories {
		factories[i] = factory
	}
	for i, mine := range solution.mines {
		if !s.isPositionAvailableForMine(factories, mines[:i], mine) {
			return errors.New("solution includes a mine which position is invalid, can't evaluate this solution")
		}
	}

	for i, factory := range solution.factories {
		if !s.isPositionAvailableForFactory(factories[:i], mines, factory.position) {
			return errors.New("solution includes a factory which position is invalid, can't evaluate this solution")
		}
	}
	return nil
}

func (s Scenario) evaluateSolution(solution Solution) (int, error) {
	err := s.checkValidity(solution)
	if err != nil {
		return 0, err
	}
	simulation := simulationFromScenarioAndSolution(s, solution)
	for i := 0; i < s.turns; i++ {
		simulation.simulateOneRound()
	}
	score := 0
	for _, factory := range simulation.factories {
		for i := 0; i < NumResourceTypes; i++ {
			score += factory.resourceStorage[i]
		}
	}
	return score, nil
}

func simulationFromScenarioAndSolution(scenario Scenario, solution Solution) Simulation {
	simulation := Simulation{
		scenario:  scenario,
		factories: make([]SimulatedFactory, len(solution.factories)),
		deposits:  make([]SimulatedDeposit, len(scenario.deposits)),
		mines:     make([]SimulatedMine, len(solution.mines)),
		conveyors: make([]SimulatedConveyor, len(solution.conveyors)),
	}
	for i, deposit := range scenario.deposits {
		simulation.deposits[i] = SimulatedDeposit{
			deposit:            deposit,
			remainingResources: deposit.width * deposit.height * DepositResourceFactor,
		}
	}
	for i, factory := range solution.factories {
		simulation.factories[i] = SimulatedFactory{
			factory:         factory,
			resourceStorage: []int{0, 0, 0, 0, 0, 0, 0, 0},
		}
	}
	for i, mine := range solution.mines {
		simulation.mines[i] = SimulatedMine{
			mine:             mine,
			resourcesIngress: []int{0, 0, 0, 0, 0, 0, 0, 0},
			resourcesEgress:  []int{0, 0, 0, 0, 0, 0, 0, 0},
		}
	}
	return simulation
}

func (s *Simulation) simulateOneRound() {
	// Transfer resources from mine egresses to factories
	for _, factory := range s.factories {
		for _, mine := range s.adjacentMinesToFactory(factory) {
			for i := 0; i < NumResourceTypes; i++ {
				factory.resourceStorage[i] += mine.resourcesEgress[i]
				mine.resourcesEgress[i] = 0
			}
		}
	}
	// Transfer resources from mine ingresses to mine egresses
	for _, mine := range s.mines {
		for i := 0; i < NumResourceTypes; i++ {
			mine.resourcesEgress[i] += mine.resourcesIngress[i]
			mine.resourcesIngress[i] = 0
		}
	}
	// Transfer resources from deposits to mine ingresses
	for _, deposit := range s.deposits {
		adjacentMines := s.adjacentMinesToDeposit(deposit)
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
			mine.resourcesIngress[deposit.deposit.subtype] += withdrawAmount
		}
	}
}

func (s *Simulation) adjacentMinesToFactory(factory SimulatedFactory) []*SimulatedMine {
	mines := make([]*SimulatedMine, 0)
	for _, position := range factory.factory.mineEgressPositions() {
		mine, foundMine := s.getMineWithEgressAt(position)
		if foundMine {
			mines = append(mines, mine)
		}
	}
	return mines
}

func (s *Simulation) adjacentMinesToDeposit(deposit SimulatedDeposit) []*SimulatedMine {
	mines := make([]*SimulatedMine, 0)
	for _, position := range deposit.deposit.mineIngressPositions() {
		mine, foundMine := s.getMineWithIngressAt(position)
		if foundMine {
			mines = append(mines, mine)
		}
	}
	return mines
}

func (s *Simulation) getMineWithIngressAt(position Position) (*SimulatedMine, bool) {
	for _, mine := range s.mines {
		if mine.mine.Ingress() == position {
			return &mine, true
		}
	}
	return &SimulatedMine{}, false
}

func (s *Simulation) getMineWithEgressAt(position Position) (*SimulatedMine, bool) {
	for _, mine := range s.mines {
		if mine.mine.Egress() == position {
			return &mine, true
		}
	}
	return &SimulatedMine{}, false
}
