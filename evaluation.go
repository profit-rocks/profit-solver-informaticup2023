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
	combiners []SimulatedCombiner
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
	mine                   Mine
	resourcesIngress       []int
	resourcesEgress        []int
	resourcesIngressUpdate []int
	resourcesEgressUpdate  []int
	connectedFactory       *SimulatedFactory
}

type SimulatedPath struct {
	conveyors     []SimulatedConveyor
	startCombiner *SimulatedCombiner
	endCombiner   *SimulatedCombiner
	startMine     *SimulatedMine
	endFactory    *SimulatedFactory
}

type SimulatedConveyor struct {
	conveyor        Conveyor
	resources       []int
	resourcesUpdate []int
}

type SimulatedCombiner struct {
	combiner        Combiner
	resources       []int
	resourcesUpdate []int
}

// TODO: Try to find a faster implementation
// Checks that all egresses are connected to a single ingress. We assume that objects don't overlap
func (s *Scenario) checkEgressesHaveSingleIngress(solution Solution) bool {
	Egress := 1
	Ingress := 2
	ingressEgressMatrix := make([][]int, s.width)
	for i := range ingressEgressMatrix {
		ingressEgressMatrix[i] = make([]int, s.height)
	}
	for _, mine := range solution.mines {
		ingressEgressMatrix[mine.Egress().x][mine.Egress().y] = Egress
		ingressEgressMatrix[mine.Ingress().x][mine.Ingress().y] = Ingress
	}
	for _, factory := range solution.factories {
		for _, position := range factory.ingressPositions() {
			ingressEgressMatrix[position.x][position.y] = Ingress
		}
	}
	for _, deposit := range s.deposits {
		for _, position := range deposit.egressPositions() {
			ingressEgressMatrix[position.x][position.y] = Egress
		}
	}
	for _, combiner := range solution.combiners {
		for _, position := range combiner.Ingresses() {
			ingressEgressMatrix[position.x][position.y] = Ingress
		}
		ingressEgressMatrix[combiner.Egress().x][combiner.Egress().y] = Egress
	}
	for _, path := range solution.paths {
		for _, conveyor := range path.conveyors {
			ingressEgressMatrix[conveyor.Egress().x][conveyor.Egress().y] = Egress
			ingressEgressMatrix[conveyor.Ingress().x][conveyor.Ingress().y] = Ingress
		}
	}
	for i := range ingressEgressMatrix {
		for j := range ingressEgressMatrix[i] {
			if ingressEgressMatrix[i][j] == Egress {
				numIngresses := 0
				for _, position := range (Position{i, j}).NeighborPositions() {
					if s.inBounds(position) {
						if ingressEgressMatrix[position.x][position.y] == Ingress {
							numIngresses += 1
						}
					}
				}
				if numIngresses > 1 {
					return false
				}
			}
		}
	}
	return true
}

func (s *Scenario) checkValidity(solution Solution) error {
	for i, mine := range solution.mines {
		if !s.positionAvailableForMine(solution.factories, solution.mines[:i], solution.combiners, solution.paths, mine) {
			return errors.New("solution includes a mine which position is invalid, can't evaluate this solution")
		}
	}

	for i, factory := range solution.factories {
		if !s.positionAvailableForFactory(solution.factories[:i], solution.mines, solution.combiners, solution.paths, factory.position) {
			return errors.New("solution includes a factory which position is invalid, can't evaluate this solution")
		}
	}

	for i, combiner := range solution.combiners {
		if !s.positionAvailableForCombiner(solution.factories, solution.mines, solution.paths, solution.combiners[:i], combiner) {
			return errors.New("solution includes a combiner which position is invalid, can't evaluate this solution")
		}
	}
	paths := make([]Path, len(solution.paths))
	for i, path := range solution.paths {
		paths = append(paths, Path{})
		for _, conveyor := range path.conveyors {
			if !s.positionAvailableForConveyor(solution.factories, solution.mines, solution.combiners, paths, conveyor) {
				return errors.New("solution includes a factory which position is invalid, can't evaluate this solution")
			}
			paths[i].conveyors = append(paths[i].conveyors, conveyor)
		}
	}
	if !s.checkEgressesHaveSingleIngress(solution) {
		return errors.New("solution includes multiple ingresses at an egress")
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
		if simulation.simulateOneRound(i) {
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
		combiners: make([]SimulatedCombiner, len(solution.combiners)),
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
			mine:                   mine,
			resourcesIngress:       []int{0, 0, 0, 0, 0, 0, 0, 0},
			resourcesEgress:        []int{0, 0, 0, 0, 0, 0, 0, 0},
			resourcesIngressUpdate: []int{0, 0, 0, 0, 0, 0, 0, 0},
			resourcesEgressUpdate:  []int{0, 0, 0, 0, 0, 0, 0, 0},
		}
		if mine.connectedFactory == nil {
			continue
		}
		for n, factory := range solution.factories {
			if factory.position == mine.connectedFactory.position {
				simulation.mines[i].connectedFactory = &simulation.factories[n]
			}
		}
	}
	for i := range scenario.deposits {
		simulation.deposits[i].mines = simulation.adjacentMinesToDeposit(simulation.deposits[i])
	}
	return simulation
}

func (s *Simulation) adjacentCombinerToConveyor(conveyor Conveyor, checkCombinerEgress bool) (*SimulatedCombiner, bool) {
	for i := range s.combiners {
		if checkCombinerEgress {
			if s.combiners[i].combiner.Egress().NextTo(conveyor.Ingress()) {
				return &s.combiners[i], true
			}
		} else {
			for _, ingress := range s.combiners[i].combiner.Ingresses() {
				if ingress.NextTo(conveyor.Egress()) {
					return &s.combiners[i], true
				}
			}
		}
	}
	return nil, false
}

func (s *Simulation) adjacentCombinerToCombiner(combiner SimulatedCombiner) (*SimulatedCombiner, bool) {
	for i := range s.combiners {
		for _, ingress := range s.combiners[i].combiner.Ingresses() {
			if ingress.NextTo(combiner.combiner.Egress()) {
				return &s.combiners[i], true
			}
		}
	}
	return nil, false
}

func (s *Simulation) adjacentFactoryToCombiner(combiner SimulatedCombiner) (*SimulatedFactory, bool) {
	for i := range s.factories {
		for _, ingress := range s.factories[i].factory.ingressPositions() {
			if ingress.NextTo(combiner.combiner.Egress()) {
				return &s.factories[i], true
			}
		}
	}
	return nil, false
}

func (s *Simulation) adjacentFactoryToMine(mine SimulatedMine) (*SimulatedFactory, bool) {
	for i := range s.factories {
		for _, ingress := range s.factories[i].factory.ingressPositions() {
			if ingress.NextTo(mine.mine.Egress()) {
				return &s.factories[i], true
			}
		}
	}
	return nil, false
}

func (s *Simulation) adjacentCombinerToMine(mine SimulatedMine) (*SimulatedCombiner, bool) {
	for i := range s.combiners {
		for _, ingress := range s.combiners[i].combiner.Ingresses() {
			if ingress.NextTo(mine.mine.Egress()) {
				return &s.combiners[i], true
			}
		}
	}
	return nil, false
}

func (s *Simulation) adjacentMineToConveyor(conveyor Conveyor) (*SimulatedMine, bool) {
	for i := range s.mines {
		if s.mines[i].mine.Egress().NextTo(conveyor.Ingress()) {
			return &s.mines[i], true
		}
	}
	return nil, false
}

func (s *Simulation) adjacentFactoryToConveyor(conveyor Conveyor) (*SimulatedFactory, bool) {
	for i := range s.factories {
		for _, egress := range s.factories[i].factory.NextToIngressPositions() {
			if egress == conveyor.Egress() {
				return &s.factories[i], true
			}
		}
	}
	return nil, false
}

func (s *Simulation) simulateOneRound(currentTurn int) bool {
	for i := range s.deposits {
		deposit := &s.deposits[i]
		if deposit.remainingResources < MaxDepositWithdrawPerMine {
			continue
		}
		for _, mine := range deposit.mines {
			if deposit.remainingResources > 0 && currentTurn+mine.mine.distance+1 < s.scenario.turns {
				minedResources := minInt(deposit.remainingResources, MaxDepositWithdrawPerMine)
				deposit.remainingResources -= minedResources
				if mine.connectedFactory != nil {
					mine.connectedFactory.resources[deposit.deposit.subtype] += minedResources
				}
			}
		}
	}
	return false
}

func (s *Simulation) adjacentMinesToFactory(factory SimulatedFactory) []*SimulatedMine {
	mines := make([]*SimulatedMine, 0)
	for _, position := range factory.factory.NextToIngressPositions() {
		mine, foundMine := s.mineWithEgress(position)
		if foundMine {
			mines = append(mines, mine)
		}
	}
	return mines
}

func (s *Simulation) adjacentMinesToDeposit(deposit SimulatedDeposit) []*SimulatedMine {
	mines := make([]*SimulatedMine, 0)
	for _, position := range deposit.deposit.nextToEgressPositions() {
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
