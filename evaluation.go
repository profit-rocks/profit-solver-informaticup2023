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
}

// PathSubtype specifies start and end of a path
type PathSubtype int

const (
	MineToFactory      PathSubtype = iota
	MineToCombiner     PathSubtype = iota
	CombinerToCombiner PathSubtype = iota
	CombinerToFactory  PathSubtype = iota
)

type SimulatedPath struct {
	conveyors     []SimulatedConveyor
	startCombiner *SimulatedCombiner
	endCombiner   *SimulatedCombiner
	startMine     *SimulatedMine
	endFactory    *SimulatedFactory
	subtype       PathSubtype
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
	mines := make([]Mine, len(solution.mines))
	for i, mine := range solution.mines {
		mines[i] = mine
	}
	factories := make([]Factory, len(solution.factories))
	for i, factory := range solution.factories {
		factories[i] = factory
	}
	combiners := make([]Combiner, len(solution.combiners))
	for i, combiner := range solution.combiners {
		combiners[i] = combiner
	}
	paths := make([]Path, len(solution.paths))
	for i, path := range solution.paths {
		paths[i] = path
	}
	for i, mine := range solution.mines {
		if !s.positionAvailableForMine(factories, mines[:i], combiners, paths, mine) {
			return errors.New("solution includes a mine which position is invalid, can't evaluate this solution")
		}
	}

	for i, factory := range solution.factories {
		if !s.positionAvailableForFactory(factories[:i], mines, combiners, paths, factory.position) {
			return errors.New("solution includes a factory which position is invalid, can't evaluate this solution")
		}
	}
	for i, combiner := range solution.combiners {
		if !s.positionAvailableForCombiner(factories, mines, paths, combiners[:i], combiner) {
			return errors.New("solution includes a combiner which position is invalid, can't evaluate this solution")
		}
	}
	for i, path := range solution.paths {
		for _, conveyor := range path.conveyors {
			if !s.positionAvailableForConveyor(factories, mines, combiners, paths[:i], conveyor) {
				return errors.New("solution includes a factory which position is invalid, can't evaluate this solution")
			}
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
	}

	for i, combiner := range solution.combiners {
		simulation.combiners[i] = SimulatedCombiner{
			combiner:        combiner,
			resources:       []int{0, 0, 0, 0, 0, 0, 0, 0},
			resourcesUpdate: []int{0, 0, 0, 0, 0, 0, 0, 0},
		}
	}

	for _, path := range solution.paths {
		if len(path.conveyors) > 0 {
			startMine, hasAdjacentMine := simulation.adjacentMineToConveyor(path.conveyors[0])
			startCombiner, hasAdjacentStartCombiner := simulation.adjacentCombinerToConveyor(path.conveyors[0], true)
			endFactory, hasAdjacentFactory := simulation.adjacentFactoryToConveyor(path.conveyors[len(path.conveyors)-1])
			endCombiner, hasAdjacentEndCombiner := simulation.adjacentCombinerToConveyor(path.conveyors[0], false)
			var simulatedPath SimulatedPath
			if hasAdjacentMine && hasAdjacentFactory {
				simulatedPath = SimulatedPath{
					startMine:  startMine,
					endFactory: endFactory,
					subtype:    MineToFactory,
				}
			} else if hasAdjacentMine && hasAdjacentEndCombiner {
				simulatedPath = SimulatedPath{
					startMine:   startMine,
					endCombiner: endCombiner,
					subtype:     MineToCombiner,
				}
			} else if hasAdjacentStartCombiner && hasAdjacentEndCombiner {
				simulatedPath = SimulatedPath{
					startCombiner: startCombiner,
					endCombiner:   endCombiner,
					subtype:       CombinerToCombiner,
				}
			} else if hasAdjacentFactory && hasAdjacentStartCombiner {
				simulatedPath = SimulatedPath{
					startCombiner: startCombiner,
					endFactory:    endFactory,
					subtype:       CombinerToFactory,
				}
			} else {
				continue
			}

			simulatedPath.conveyors = make([]SimulatedConveyor, len(path.conveyors))
			for j, conveyor := range path.conveyors {
				simulatedPath.conveyors[j] = SimulatedConveyor{
					conveyor:        conveyor,
					resources:       []int{0, 0, 0, 0, 0, 0, 0, 0},
					resourcesUpdate: []int{0, 0, 0, 0, 0, 0, 0, 0},
				}
			}
			simulation.paths = append(simulation.paths, simulatedPath)
		}
	}
	// TODO: add empty paths
	// Check for paths without conveyors
	// combiner combiner
	for i := range solution.combiners {
		endCombiner, hasEndCombiner := simulation.adjacentCombinerToCombiner(simulation.combiners[i])
		if hasEndCombiner {
			simulation.paths = append(simulation.paths, SimulatedPath{
				startCombiner: &simulation.combiners[i],
				endCombiner:   endCombiner,
				subtype:       CombinerToCombiner,
			})
		}
	}
	// combiner factory
	for i := range solution.combiners {
		endFactory, hasEndFactory := simulation.adjacentFactoryToCombiner(simulation.combiners[i])
		if hasEndFactory {
			simulation.paths = append(simulation.paths, SimulatedPath{
				startCombiner: &simulation.combiners[i],
				endFactory:    endFactory,
				subtype:       CombinerToFactory,
			})
		}
	}
	// mine combiner

	// mine factory

	for i := range scenario.deposits {
		simulation.deposits[i].mines = simulation.adjacentMinesToDeposit(simulation.deposits[i])
	}
	for i := range solution.factories {
		simulation.factories[i].mines = simulation.adjacentMinesToFactory(simulation.factories[i])
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
		for _, egress := range s.factories[i].factory.nextToIngressPositions() {
			if egress == conveyor.Egress() {
				return &s.factories[i], true
			}
		}
	}
	return nil, false
}

func (s *Simulation) simulateOneRound() bool {
	finished := true
	// Transfer resources along paths
	for i := range s.paths {
		// value is copied if used in range
		path := &s.paths[i]
		if len(path.conveyors) > 0 {
			// transfer along conveyors
			for j := range path.conveyors {
				// value is copied if used in range
				conveyor := &path.conveyors[len(path.conveyors)-1-j]
				// Transfer resources from previous conveyor (if present) to current conveyor
				if 0 <= len(path.conveyors)-1-j-1 {
					previousConveyor := &path.conveyors[len(path.conveyors)-1-j-1]
					for k := 0; k < NumResourceTypes; k++ {
						finished = finished && previousConveyor.resources[k] == 0
						conveyor.resourcesUpdate[k] += previousConveyor.resources[k]
						previousConveyor.resourcesUpdate[k] -= previousConveyor.resources[k]
					}
				}
			}
			// last conveyor to end
			lastConveyor := path.conveyors[len(path.conveyors)-1]
			if path.subtype == MineToFactory || path.subtype == CombinerToFactory {
				for j := 0; j < NumResourceTypes; j++ {
					finished = finished && lastConveyor.resources[j] == 0
					path.endFactory.resources[j] += lastConveyor.resources[j]
					lastConveyor.resourcesUpdate[j] -= lastConveyor.resources[j]
				}
			} else if path.subtype == MineToCombiner || path.subtype == CombinerToCombiner {
				for j := 0; j < NumResourceTypes; j++ {
					finished = finished && lastConveyor.resources[j] == 0
					path.endCombiner.resourcesUpdate[j] += lastConveyor.resources[j]
					lastConveyor.resourcesUpdate[j] -= lastConveyor.resources[j]
				}
			}
			// start to first conveyor
			firstConveyor := path.conveyors[0]
			if path.subtype == MineToFactory || path.subtype == MineToCombiner {
				for j := 0; j < NumResourceTypes; j++ {
					finished = finished && path.startMine.resourcesEgress[j] == 0
					firstConveyor.resourcesUpdate[j] += path.startMine.resourcesEgress[j]
					path.startMine.resourcesEgressUpdate[j] -= path.startMine.resourcesEgress[j]
				}
			} else if path.subtype == CombinerToFactory || path.subtype == CombinerToCombiner {
				for j := 0; j < NumResourceTypes; j++ {
					finished = finished && path.startCombiner.resources[j] == 0
					firstConveyor.resourcesUpdate[j] += path.startCombiner.resources[j]
					path.startCombiner.resourcesUpdate[j] -= path.startCombiner.resources[j]
				}
			}
		} else {
			if path.subtype == MineToFactory {
				for j := 0; j < NumResourceTypes; j++ {
					finished = finished && path.startMine.resourcesEgress[j] == 0
					path.endFactory.resources[j] += path.startMine.resourcesEgress[j]
					path.startMine.resourcesEgressUpdate[j] -= path.startMine.resourcesEgress[j]
				}
			} else if path.subtype == MineToCombiner {
				for j := 0; j < NumResourceTypes; j++ {
					finished = finished && path.startMine.resourcesEgress[j] == 0
					path.endCombiner.resources[j] += path.startMine.resourcesEgress[j]
					path.startMine.resourcesEgressUpdate[j] -= path.startMine.resourcesEgress[j]
				}
			} else if path.subtype == CombinerToFactory {
				for j := 0; j < NumResourceTypes; j++ {
					finished = finished && path.startCombiner.resources[j] == 0
					path.endFactory.resources[j] += path.startCombiner.resources[j]
					path.startCombiner.resourcesUpdate[j] -= path.startCombiner.resources[j]
				}
			} else if path.subtype == CombinerToCombiner {
				for j := 0; j < NumResourceTypes; j++ {
					finished = finished && path.startCombiner.resources[j] == 0
					path.endCombiner.resourcesUpdate[j] += path.startCombiner.resources[j]
					path.startCombiner.resourcesUpdate[j] -= path.startCombiner.resources[j]
				}
			}
		}
	}

	// transfer resources from mine ingresses to mine egresses
	for i := range s.mines {
		// value is copied if used in range
		mine := &s.mines[i]
		for j := 0; j < NumResourceTypes; j++ {
			finished = finished && mine.resourcesIngress[j] == 0
			mine.resourcesEgressUpdate[j] += mine.resourcesIngress[j]
			mine.resourcesIngressUpdate[j] -= mine.resourcesIngress[j]
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
			mine.resourcesIngressUpdate[deposit.deposit.subtype] += withdrawAmount
		}
	}

	// apply updates
	for i := range s.mines {
		// value is copied if used in range
		mine := &s.mines[i]
		for j := 0; j < NumResourceTypes; j++ {
			mine.resourcesEgress[j] += mine.resourcesEgressUpdate[j]
			mine.resourcesEgressUpdate[j] = 0
			mine.resourcesIngress[j] += mine.resourcesIngressUpdate[j]
			mine.resourcesIngressUpdate[j] = 0
		}
	}
	for i := range s.combiners {
		// value is copied if used in range
		combiner := &s.combiners[i]
		for j := 0; j < NumResourceTypes; j++ {
			combiner.resources[j] += combiner.resourcesUpdate[j]
			combiner.resourcesUpdate[j] = 0
		}

	}
	for i := range s.paths {
		// value is copied if used in range
		path := &s.paths[i]
		for j := range path.conveyors {
			// value is copied if used in range
			conveyor := &path.conveyors[j]
			for k := 0; k < NumResourceTypes; k++ {
				conveyor.resources[k] += conveyor.resourcesUpdate[k]
				conveyor.resourcesUpdate[k] = 0
			}
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
