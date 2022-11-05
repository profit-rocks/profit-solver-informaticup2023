package main

import (
	"errors"
	"golang.org/x/exp/slices"
)

const DepositResourceFactor = 5
const MaxDepositWithdrawPerMine = 3
const NumResourceTypes = 8

type Simulation struct {
	scenario  *Scenario
	factories []SimulatedFactory
	deposits  []SimulatedDeposit
	mines     []SimulatedMine
	paths     []SimulatedPath
}

type SimulatedDeposit struct {
	deposit            Deposit
	remainingResources int
	mines              []*SimulatedMine
}

type SimulatedFactory struct {
	factory         Factory
	resourceStorage []int
	mines           []*SimulatedMine
}

type SimulatedMine struct {
	mine             Mine
	resourcesIngress []int
	resourcesEgress  []int
}

type SimulatedPath struct {
	conveyors  []SimulatedConveyor
	startMine  *SimulatedMine
	endFactory *SimulatedFactory
}

type SimulatedConveyor struct {
	conveyor         Conveyor
	resourcesIngress []int
	resourcesEgress  []int
}

func (s *Scenario) checkValidity(solution Solution) error {
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

	// TODO: Check validity of conveyors
	return nil
}

func (s *Scenario) evaluateSolution(solution Solution) (int, error) {
	err := s.checkValidity(solution)
	if err != nil {
		return 0, err
	}
	simulation := simulationFromScenarioAndSolution(s, solution)
	for i := 0; i < s.turns; i++ {
		if simulation.simulateOneRound() {
			break
		}
	}
	score := 0
	for _, factory := range simulation.factories {
		for i := 0; i < NumResourceTypes; i++ {
			score += factory.resourceStorage[i]
		}
	}
	return score, nil
}

func simulationFromScenarioAndSolution(scenario *Scenario, solution Solution) Simulation {
	simulation := Simulation{
		scenario:  scenario,
		factories: make([]SimulatedFactory, len(solution.factories)),
		deposits:  make([]SimulatedDeposit, len(scenario.deposits)),
		mines:     make([]SimulatedMine, len(solution.mines)),
		paths:     make([]SimulatedPath, len(solution.paths)),
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

	for i, path := range solution.paths {
		if len(path) > 0 {
			simulatedPath := SimulatedPath{
				conveyors:  make([]SimulatedConveyor, len(path)),
				startMine:  simulation.adjacentMineToConveyor(path[0]),
				endFactory: simulation.adjacentFactoryToConveyor(path[len(path)-1]),
			}
			for j, conveyor := range path {
				simulatedPath.conveyors[j] = SimulatedConveyor{
					conveyor:         conveyor,
					resourcesIngress: []int{0, 0, 0, 0, 0, 0, 0, 0},
					resourcesEgress:  []int{0, 0, 0, 0, 0, 0, 0, 0},
				}
			}
			simulation.paths[i] = simulatedPath
		}
	}

	for i := range scenario.deposits {
		simulation.deposits[i].mines = simulation.adjacentMinesToDeposit(simulation.deposits[i])
	}
	for i := range solution.factories {
		simulation.factories[i].mines = simulation.adjacentMinesToFactory(simulation.factories[i])
	}
	return simulation
}

func (s *Simulation) adjacentMineToConveyor(conveyor Conveyor) *SimulatedMine {
	for i := range s.mines {
		if s.mines[i].mine.Egress().NextTo(conveyor.Ingress()) {
			return &s.mines[i]
		}
	}
	return nil
}

func (s *Simulation) adjacentFactoryToConveyor(conveyor Conveyor) *SimulatedFactory {
	for i := range s.factories {
		if slices.Contains(s.factories[i].factory.nextToIngressPositions(), conveyor.Egress()) {
			return &s.factories[i]
		}
	}
	return nil
}

func (s *Simulation) simulateOneRound() bool {
	finished := true
	// Transfer resources from mine egresses to factories
	for i := range s.factories {
		// value is copied if used in range
		factory := &s.factories[i]
		for _, mine := range factory.mines {
			for j := 0; j < NumResourceTypes; j++ {
				finished = finished && mine.resourcesEgress[j] == 0
				factory.resourceStorage[j] += mine.resourcesEgress[j]
				mine.resourcesEgress[j] = 0
			}
		}
	}
	// Transfer resources from mine ingresses to mine egresses
	for i := range s.mines {
		// value is copied if used in range
		mine := &s.mines[i]
		for j := 0; j < NumResourceTypes; j++ {
			finished = finished && mine.resourcesIngress[j] == 0
			mine.resourcesEgress[j] += mine.resourcesIngress[j]
			mine.resourcesIngress[j] = 0
		}
	}
	// Transfer resources from deposits to mine ingresses
	for i := range s.deposits {
		// value is copied if used in range
		deposit := &s.deposits[i]
		//TODO: mix mines
		for _, mine := range deposit.mines {
			withdrawAmount := 0
			if deposit.remainingResources >= MaxDepositWithdrawPerMine {
				//withdrawAmount = rand.Intn(3) + 1
				withdrawAmount = MaxDepositWithdrawPerMine
			} else {
				//TODO: randomize remaining amount
				withdrawAmount = deposit.remainingResources
			}
			finished = finished && withdrawAmount == 0

			deposit.remainingResources -= withdrawAmount
			mine.resourcesIngress[deposit.deposit.subtype] += withdrawAmount
		}
	}
	return finished
}

func (s *Simulation) adjacentMinesToFactory(factory SimulatedFactory) []*SimulatedMine {
	mines := make([]*SimulatedMine, 0)
	for _, position := range factory.factory.nextToIngressPositions() {
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
	for i := range s.mines {
		if s.mines[i].mine.Ingress() == position {
			return &s.mines[i], true
		}
	}
	return nil, false
}

func (s *Simulation) getMineWithEgressAt(position Position) (*SimulatedMine, bool) {
	for i := range s.mines {
		if s.mines[i].mine.Egress() == position {
			return &s.mines[i], true
		}
	}
	return nil, false
}
