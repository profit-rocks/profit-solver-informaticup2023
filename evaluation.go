package main

import (
	"errors"
	"math"
)

const DepositResourceFactor = 5
const MaxDepositWithdrawPerMine = 3
const NumResourceTypes = 8

type Simulation struct {
	scenario    *Scenario
	factories   []SimulatedFactory
	deposits    []SimulatedDeposit
	mines       []SimulatedMine
	maxDistance int
}

type SimulatedDeposit struct {
	deposit            Deposit
	remainingResources int
	mines              []*SimulatedMine
}

type SimulatedFactory struct {
	factory         Factory
	resources       []int
	resourceUpdates [][]int
	mines           []*SimulatedMine
}

type SimulatedMine struct {
	mine             Mine
	connectedFactory *SimulatedFactory
}

// TODO: Try to find a faster implementation
// Checks that all egresses are connected to a single ingress. We assume that objects don't overlap
func (s *Scenario) checkEgressesHaveSingleIngress(c Chromosome) bool {
	Egress := 1
	Ingress := 2
	ingressEgressMatrix := make([][]int, s.width)
	for i := range ingressEgressMatrix {
		ingressEgressMatrix[i] = make([]int, s.height)
	}
	for _, mine := range c.mines {
		ingressEgressMatrix[mine.Egress().x][mine.Egress().y] = Egress
		ingressEgressMatrix[mine.Ingress().x][mine.Ingress().y] = Ingress
	}
	for _, factory := range c.factories {
		for _, position := range factory.ingressPositions() {
			ingressEgressMatrix[position.x][position.y] = Ingress
		}
	}
	for _, deposit := range s.deposits {
		for _, position := range deposit.egressPositions() {
			ingressEgressMatrix[position.x][position.y] = Egress
		}
	}
	for _, combiner := range c.combiners {
		for _, position := range combiner.Ingresses() {
			ingressEgressMatrix[position.x][position.y] = Ingress
		}
		ingressEgressMatrix[combiner.Egress().x][combiner.Egress().y] = Egress
	}
	for _, path := range c.paths {
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

func (s *Scenario) checkValidity(c Chromosome) error {
	for i, mine := range c.mines {
		if !s.positionAvailableForMine(c.factories, c.mines[:i], c.combiners, c.paths, mine) {
			return errors.New("chromosome includes a mine which position is invalid, can't evaluate this chromosome")
		}
	}

	for i, factory := range c.factories {
		if !s.positionAvailableForFactory(c.factories[:i], c.mines, c.combiners, c.paths, factory.position) {
			return errors.New("chromosome includes a factory which position is invalid, can't evaluate this chromosome")
		}
	}

	for i, combiner := range c.combiners {
		if !s.positionAvailableForCombiner(c.factories, c.mines, c.paths, c.combiners[:i], combiner) {
			return errors.New("chromosome includes a combiner which position is invalid, can't evaluate this chromosome")
		}
	}
	paths := make([]Path, len(c.paths))
	for i, path := range c.paths {
		paths = append(paths, Path{})
		for _, conveyor := range path.conveyors {
			if !s.positionAvailableForConveyor(c.factories, c.mines, c.combiners, paths, conveyor) {
				return errors.New("chromosome includes a conveyor which position is invalid, can't evaluate this chromosome")
			}
			paths[i].conveyors = append(paths[i].conveyors, conveyor)
		}
	}
	if !s.checkEgressesHaveSingleIngress(c) {
		return errors.New("chromosome includes multiple ingresses at an egress")
	}
	return nil
}

func (s *Scenario) evaluateChromosome(c Chromosome) (int, int, error) {
	// TODO: remove validity check
	err := s.checkValidity(c)
	if err != nil {
		return 0, s.turns, err
	}
	if len(c.mines) == 0 || len(c.mines) == 0 {
		return 0, s.turns, nil
	}

	simulation := simulationFromScenarioAndChromosome(s, c)
	neededTurns := 0
	finalScore := 0
	products := make(map[int]Product)
	for _, product := range s.products {
		products[product.subtype] = product
	}
	maxDistance := 0
	for i := range c.mines {
		mine := &c.mines[i]
		if mine.distance > maxDistance {
			maxDistance = mine.distance
		}
	}
	// add 1 since we need one more round to mine resources from deposits
	maxDistance += 1

	for i := range simulation.factories {
		factory := &simulation.factories[i]
		factory.resourceUpdates = make([][]int, maxDistance)
		for j := 0; j < maxDistance; j++ {
			factory.resourceUpdates[j] = make([]int, NumResourceTypes)
		}
	}
	simulation.maxDistance = maxDistance

	for i := 0; i < s.turns; i++ {
		simulation.simulateOneTurn(i)
		score := 0
		for _, factory := range simulation.factories {
			units := math.MaxInt32
			product := products[factory.factory.product]
			for j, resource := range product.resources {
				if resource != 0 {
					units = minInt(units, factory.resources[j]/resource)
				}
			}
			score += units * product.points
		}
		if score > finalScore {
			finalScore = score
			neededTurns = i
		}
	}
	return finalScore, neededTurns + 1, nil
}

func simulationFromScenarioAndChromosome(scenario *Scenario, c Chromosome) Simulation {
	simulation := Simulation{
		scenario:  scenario,
		factories: make([]SimulatedFactory, len(c.factories)),
		deposits:  make([]SimulatedDeposit, len(scenario.deposits)),
		mines:     make([]SimulatedMine, len(c.mines)),
	}
	for i, deposit := range scenario.deposits {
		simulation.deposits[i] = SimulatedDeposit{
			deposit:            deposit,
			remainingResources: deposit.width * deposit.height * DepositResourceFactor,
		}
	}
	for i, factory := range c.factories {
		simulation.factories[i] = SimulatedFactory{
			factory:   factory,
			resources: []int{0, 0, 0, 0, 0, 0, 0, 0},
		}
	}
	for i, mine := range c.mines {
		simulation.mines[i] = SimulatedMine{
			mine: mine,
		}
		if mine.connectedFactory == nil {
			continue
		}
		for n, factory := range c.factories {
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

func (s *Simulation) simulateOneTurn(currentTurn int) {
	// deliver resources that arrive in this turn to factories
	for i := range s.factories {
		factory := &s.factories[i]
		updateIndex := currentTurn % s.maxDistance
		for j := range factory.resourceUpdates[updateIndex] {
			factory.resources[j] += factory.resourceUpdates[updateIndex][j]
			factory.resourceUpdates[updateIndex][j] = 0
		}
	}
	// mine new resources from deposits
	for i := range s.deposits {
		deposit := &s.deposits[i]
		for _, mine := range deposit.mines {
			if deposit.remainingResources > 0 && currentTurn < s.scenario.turns {
				minedResources := minInt(deposit.remainingResources, MaxDepositWithdrawPerMine)
				deposit.remainingResources -= minedResources
				if mine.connectedFactory != nil {
					updateIndex := (currentTurn + mine.mine.distance + 1) % s.maxDistance
					mine.connectedFactory.resourceUpdates[updateIndex][deposit.deposit.subtype] += minedResources
				}
			}
		}
	}
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
