package main

import (
	"errors"
	"math"
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
	factory   Factory
	resources []int
	mines     []*SimulatedMine
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
	conveyor  Conveyor
	resources []int
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
	paths := make([]Path, len(solution.paths))
	for i, path := range solution.paths {
		paths[i] = path
	}
	for i, mine := range solution.mines {
		if !s.positionAvailableForMine(factories, mines[:i], paths, mine) {
			return errors.New("solution includes a mine which position is invalid, can't evaluate this solution")
		}
	}

	for i, factory := range solution.factories {
		if !s.positionAvailableForFactory(factories[:i], mines, paths, factory.position) {
			return errors.New("solution includes a factory which position is invalid, can't evaluate this solution")
		}
	}

	for i, path := range solution.paths {
		for _, conveyor := range path.conveyors {
			if !s.positionAvailableForConveyor(factories, mines, paths[:i], conveyor) {
				return errors.New("solution includes a factory which position is invalid, can't evaluate this solution")
			}
		}
	}
	return nil
}

func (s *Scenario) evaluateSolution(solution Solution) (int, error) {
	// TODO: remove validity check
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
		units := math.MaxInt32
		// TODO: efficiency can be improved by precomputing a subtype -> product map
		for _, product := range s.products {
			if product.subtype == factory.factory.product {
				for i, resource := range product.resources {
					if resource != 0 {
						units = minInt(units, factory.resources[i]/resource)
					}
				}
				score += units * product.points
				break
			}
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
		paths:     []SimulatedPath{},
	}
	for i, deposit := range scenario.deposits {
		simulation.deposits[i] = SimulatedDeposit{
			deposit:            deposit,
			remainingResources: deposit.width * deposit.height * DepositResourceFactor,
		}
	}
	for i, factory := range solution.factories {
		simulation.factories[i] = SimulatedFactory{
			factory:   factory,
			resources: []int{0, 0, 0, 0, 0, 0, 0, 0},
		}
	}
	for i, mine := range solution.mines {
		simulation.mines[i] = SimulatedMine{
			mine:             mine,
			resourcesIngress: []int{0, 0, 0, 0, 0, 0, 0, 0},
			resourcesEgress:  []int{0, 0, 0, 0, 0, 0, 0, 0},
		}
	}

	for _, path := range solution.paths {
		if len(path.conveyors) > 0 {
			startMine, err := simulation.adjacentMineToConveyor(path.conveyors[0])
			if err != nil {
				//fmt.Println("No adjacent mine, skipping path")
				continue
			}
			endFactory, err := simulation.adjacentFactoryToConveyor(path.conveyors[len(path.conveyors)-1])
			if err != nil {
				//fmt.Println("No adjacent factory, skipping path")
				continue
			}
			simulatedPath := SimulatedPath{
				conveyors:  make([]SimulatedConveyor, len(path.conveyors)),
				startMine:  startMine,
				endFactory: endFactory,
			}
			for j, conveyor := range path.conveyors {
				simulatedPath.conveyors[j] = SimulatedConveyor{
					conveyor:  conveyor,
					resources: []int{0, 0, 0, 0, 0, 0, 0, 0},
				}
			}
			simulation.paths = append(simulation.paths, simulatedPath)
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

func (s *Simulation) adjacentMineToConveyor(conveyor Conveyor) (*SimulatedMine, error) {
	for i := range s.mines {
		if s.mines[i].mine.Egress().NextTo(conveyor.Ingress()) {
			return &s.mines[i], nil
		}
	}
	return nil, errors.New("conveyor has no adjacent mine")
}

func (s *Simulation) adjacentFactoryToConveyor(conveyor Conveyor) (*SimulatedFactory, error) {
	for i := range s.factories {
		for _, egress := range s.factories[i].factory.nextToIngressPositions() {
			if egress == conveyor.Egress() {
				return &s.factories[i], nil
			}
		}
	}
	return nil, errors.New("conveyor has no adjacent factory")
}

func (s *Simulation) simulateOneRound() bool {
	finished := true

	// Transfer resources from end of path to factories
	for i := range s.paths {
		// value is copied if used in range
		path := &s.paths[i]
		lastConveyor := path.conveyors[len(path.conveyors)-1]
		for j := 0; j < NumResourceTypes; j++ {
			finished = finished && lastConveyor.resources[j] == 0
			path.endFactory.resources[j] += lastConveyor.resources[j]
			lastConveyor.resources[j] = 0
		}
	}
	// Transfer resources from mine egresses to factories
	for i := range s.factories {
		// value is copied if used in range
		factory := &s.factories[i]
		for _, mine := range factory.mines {
			for j := 0; j < NumResourceTypes; j++ {
				finished = finished && mine.resourcesEgress[j] == 0
				factory.resources[j] += mine.resourcesEgress[j]
				mine.resourcesEgress[j] = 0
			}
		}
	}
	// Transfer resources along the paths
	for i := range s.paths {
		// value is copied if used in range
		path := &s.paths[i]
		for j := range path.conveyors {
			// value is copied if used in range
			conveyor := &path.conveyors[len(path.conveyors)-1-j]
			// Transfer resources from previous conveyor (if present) to current conveyor
			if 0 <= len(path.conveyors)-1-j-1 {
				previousConveyor := &path.conveyors[len(path.conveyors)-1-j-1]
				for k := 0; k < NumResourceTypes; k++ {
					finished = finished && previousConveyor.resources[k] == 0
					conveyor.resources[k] += previousConveyor.resources[k]
					previousConveyor.resources[k] = 0
				}
			}
		}
	}

	// Transfer resources from mine to first conveyor of path
	for i := range s.paths {
		// value is copied if used in range
		path := &s.paths[i]
		firstConveyor := &path.conveyors[0]
		for j := 0; j < NumResourceTypes; j++ {
			finished = finished && path.startMine.resourcesEgress[j] == 0
			firstConveyor.resources[j] += path.startMine.resourcesEgress[j]
			path.startMine.resourcesEgress[j] = 0
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
		mine, foundMine := s.mineWithEgress(position)
		if foundMine {
			mines = append(mines, mine)
		}
	}
	return mines
}

func (s *Simulation) adjacentMinesToDeposit(deposit SimulatedDeposit) []*SimulatedMine {
	mines := make([]*SimulatedMine, 0)
	for _, position := range deposit.deposit.mineIngressPositions() {
		mine, foundMine := s.mineWithIngress(position)
		if foundMine {
			mines = append(mines, mine)
		}
	}
	return mines
}

func (s *Simulation) mineWithIngress(position Position) (*SimulatedMine, bool) {
	for i := range s.mines {
		if s.mines[i].mine.Ingress() == position {
			return &s.mines[i], true
		}
	}
	return nil, false
}

func (s *Simulation) mineWithEgress(position Position) (*SimulatedMine, bool) {
	for i := range s.mines {
		if s.mines[i].mine.Egress() == position {
			return &s.mines[i], true
		}
	}
	return nil, false
}
