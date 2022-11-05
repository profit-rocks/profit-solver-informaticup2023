package main

import (
	"container/heap"
	"errors"
	"log"
	"math"
	"math/rand"
	"sort"
)

type Path = []Conveyor

const NumPathRetries = 10

type Chromosome struct {
	factories []Factory
	mines     []Mine
	paths     []Path
	fitness   int
}

// GeneticAlgorithm contains input data as well as configuration information used by the genetic algorithm.
// Data in this struct is passed around, but never changed. If there is context information that needs to be
// changed, it should probably be stored in a chromosome.
type GeneticAlgorithm struct {
	scenario             Scenario
	iterations           int
	populationSize       int
	mutationProbability  float64
	crossoverProbability float64
	numFactories         int
	numMines             int
	optimum              int
	numPaths             int
}

func (c Chromosome) Solution() Solution {
	solution := Solution{
		factories: make([]Factory, len(c.factories)),
		mines:     make([]Mine, len(c.mines)),
	}
	for i, factory := range c.factories {
		solution.factories[i] = Factory{
			position: factory.position,
			product:  factory.product,
		}
	}
	for i, mine := range c.mines {
		solution.mines[i] = Mine{
			position:  mine.position,
			direction: mine.direction,
		}
	}
	for _, path := range c.paths {
		for _, conveyor := range path {
			solution.conveyors = append(solution.conveyors, conveyor)
		}
	}
	return solution
}

const NumTriesPerChromosome = 10

func (g *GeneticAlgorithm) crossover(chromosome Chromosome, chromosome2 Chromosome) Chromosome {
	newChromosome := Chromosome{}
	for i := 0; i < len(chromosome.mines); i++ {
		if rand.Float64() > g.crossoverProbability {
			newChromosome.mines = append(newChromosome.mines, chromosome.mines[i])
		} else {
			newChromosome.mines = append(newChromosome.mines, chromosome2.mines[i])
		}
	}
	for i := 0; i < len(chromosome.factories); i++ {
		if rand.Float64() > g.crossoverProbability {
			newChromosome.factories = append(newChromosome.factories, chromosome.factories[i])
		} else {
			newChromosome.factories = append(newChromosome.factories, chromosome2.factories[i])
		}
	}
	newChromosome.paths = chromosome.paths
	return newChromosome
}

func (g *GeneticAlgorithm) mutation(chromosome Chromosome) Chromosome {
	newChromosome := Chromosome{}
	for _, mine := range chromosome.mines {
		if rand.Float64() > g.mutationProbability {
			newChromosome.mines = append(newChromosome.mines, mine)
		} else {
			// attach new mine to deposit of old mine.
			// TODO: this does not work correctly when a mine is attached to multiple deposits
			success := false
			for _, deposit := range g.scenario.deposits {
				if deposit.Rectangle().Intersects(Rectangle{Position{mine.Ingress().x - 1, mine.Ingress().y - 1}, 3, 3}) {
					newMine, err := g.getRandomMine(deposit, newChromosome)
					if err == nil {
						newChromosome.mines = append(newChromosome.mines, newMine)
						success = true
						break
					}
				}
			}
			if !success {
				newChromosome.mines = append(newChromosome.mines, mine)
			}
		}
	}
	for _, factory := range chromosome.factories {
		if rand.Float64() > g.mutationProbability {
			newChromosome.factories = append(newChromosome.factories, factory)
		} else {
			newFactory, err := g.getRandomFactory(newChromosome)
			if err != nil {
				newChromosome.factories = append(newChromosome.factories, factory)
			} else {
				newChromosome.factories = append(newChromosome.factories, newFactory)
			}
		}
	}
	newChromosome.paths = chromosome.paths
	return newChromosome
}

func (g *GeneticAlgorithm) evaluateFitness(chromosome Chromosome) int {
	fitness, err := g.scenario.evaluateSolution(chromosome.Solution())
	if err != nil {
		return math.MinInt
	}
	// sum of manhattan distances for each factory to all the mines
	//fitness := 0.0
	//for _, mine := range chromosome.mines {
	//	for _, factory := range chromosome.factories {
	//		fitness += float64(factory.position.ManhattanDist(mine.position))
	//	}
	//}
	return fitness
}

func (g *GeneticAlgorithm) generateChromosomes() ([]Chromosome, error) {
	chromosomes := make([]Chromosome, g.populationSize)
	for i := 0; i < g.populationSize; i++ {
		foundChromosome := false
		for j := 0; j < NumTriesPerChromosome; j++ {
			chromosome, err := g.generateChromosome()
			if err == nil {
				chromosomes[i] = chromosome
				foundChromosome = true
			}
		}
		if !foundChromosome {
			return chromosomes, errors.New("exceeded NumTriesPerChromosome in generateChromosomes, probably trying to place too many factories or mines")
		}
	}
	return chromosomes, nil
}

func (g *GeneticAlgorithm) generateChromosome() (Chromosome, error) {
	chromosome := Chromosome{mines: make([]Mine, 0), factories: make([]Factory, 0)}
	for i := 0; i < g.numMines; i++ {
		deposit := g.scenario.deposits[i%len(g.scenario.deposits)]
		mine, err := g.getRandomMine(deposit, chromosome)
		if err != nil {
			return chromosome, err
		}
		chromosome.mines = append(chromosome.mines, mine)
	}
	for i := 0; i < g.numFactories; i++ {
		factory, err := g.getRandomFactory(chromosome)
		if err != nil {
			return chromosome, err
		}
		chromosome.factories = append(chromosome.factories, factory)
	}
	for i := 0; i < g.numPaths; i++ {
		var path Path
		var err error
		for j := 0; j < NumPathRetries; j++ {
			randomFactory := chromosome.factories[rand.Intn(len(chromosome.factories))]
			randomMine := chromosome.mines[rand.Intn(len(chromosome.mines))]
			path, err = g.getPath(chromosome, randomMine, randomFactory)
			if err == nil {
				break
			}
		}
		if err == nil {
			chromosome.paths = append(chromosome.paths, path)
		} else {
			chromosome.paths = append(chromosome.paths, Path{})
		}
	}
	return chromosome, nil
}

func (g *GeneticAlgorithm) getPath(chromosome Chromosome, mine Mine, factory Factory) (Path, error) {
	var path Path

	startPosition := mine.Egress()
	// Dummy conveyor used for backtracking
	startConveyor := Conveyor{
		position:  Position{startPosition.x - 1, startPosition.y},
		direction: Right,
		length:    Short,
	}
	endPositions := factory.mineEgressPositions()
	queue := PriorityQueue{}
	startItem := Item{
		value:    startConveyor,
		priority: 0,
		index:    0,
	}

	distances := make([][]int, g.scenario.height)
	for i := range distances {
		distances[i] = make([]int, g.scenario.width)
		for j := range distances[i] {
			distances[i][j] = 1000000
		}
	}
	blocked := make([][]bool, g.scenario.height)
	for i := range blocked {
		blocked[i] = make([]bool, g.scenario.width)
		for j := range blocked[i] {
			blocked[i][j] = false
		}
	}
	previousConveyors := make([][]Conveyor, g.scenario.height)
	for i := range previousConveyors {
		previousConveyors[i] = make([]Conveyor, g.scenario.width)
	}

	// keep algorithm from using occupied squares
	for _, deposit := range g.scenario.deposits {
		deposit.Rectangle().ForEach(func(p Position) {
			blocked[p.y][p.x] = true
		})
	}
	for _, obstacle := range g.scenario.obstacles {
		obstacle.ForEach(func(p Position) {
			blocked[p.y][p.x] = true
		})
	}
	for _, m := range chromosome.mines {
		m.RectanglesEach(func(r Rectangle) {
			r.ForEach(func(p Position) {
				blocked[p.y][p.x] = true
			})
		})
	}
	for _, f := range chromosome.factories {
		f.Rectangle().ForEach(func(p Position) {
			blocked[p.y][p.x] = true
		})
	}
	for _, otherPath := range chromosome.paths {
		for _, conveyor := range otherPath {
			conveyor.Rectangle().ForEach(func(p Position) {
				blocked[p.y][p.x] = true
			})
		}
	}

	heap.Init(&queue)
	queue.Push(&startItem)
	heap.Fix(&queue, startItem.index)
	distances[startPosition.y][startPosition.x] = 0
	for queue.Len() > 0 {
		current := queue.Pop().(*Item)
		currentEgress := current.value.Egress()
		finished := false
		for _, p := range endPositions {
			if currentEgress == p {
				path = append(path, current.value)
				finished = true
			}
		}
		if finished {
			break
		}
		if current.priority != distances[currentEgress.y][currentEgress.x] {
			continue
		}
		for _, nextIngress := range currentEgress.NeighborPositions() {
			if !g.inBounds(nextIngress) {
				continue
			}
			for i := 0; i < NumConveyorSubtypes; i++ {
				nextConveyor := ConveyorFromIngressAndSubtype(nextIngress, i)
				nextEgress := nextConveyor.Egress()
				if !g.inBounds(nextEgress) {
					continue
				}
				if current.priority+1 < distances[nextEgress.y][nextEgress.x] {
					isBlocked := false
					nextConveyor.Rectangle().ForEach(func(p Position) {
						if blocked[p.y][p.x] {
							isBlocked = true
						}
					})
					if isBlocked {
						continue
					}
					next := Item{
						value:    nextConveyor,
						priority: current.priority + 1,
						index:    0,
					}
					queue.Push(&next)
					heap.Fix(&queue, next.index)
					previousConveyors[nextEgress.y][nextEgress.x] = current.value
					distances[nextEgress.y][nextEgress.x] = next.priority
				}
			}
		}
	}
	if len(path) == 0 {
		return path, errors.New("no path found")
	}
	currentEgress := path[0].Egress()
	if currentEgress == startPosition {
		return Path{}, nil
	}
	for {
		conveyor := previousConveyors[currentEgress.y][currentEgress.x]
		if conveyor.Egress() == startPosition {
			break
		}
		path = append(path, conveyor)
		currentEgress = conveyor.Egress()
	}
	return path, nil
}

func (g *GeneticAlgorithm) inBounds(p Position) bool {
	if p.x >= g.scenario.width || p.x < 0 || p.y >= g.scenario.height || p.y < 0 {
		return false
	}
	return true
}

func (g *GeneticAlgorithm) getRandomMine(deposit Deposit, chromosome Chromosome) (Mine, error) {
	availableMines := g.scenario.minesAroundDeposit(deposit, chromosome)
	if len(availableMines) != 0 {
		return availableMines[rand.Intn(len(availableMines))], nil
	}
	return Mine{}, errors.New("no mines available")
}

func (g *GeneticAlgorithm) getRandomFactory(chromosome Chromosome) (Factory, error) {
	availablePositions := g.scenario.getAvailableFactoryPositions(chromosome)
	if len(availablePositions) == 0 {
		return Factory{}, errors.New("no factory positions available")
	}
	position := availablePositions[rand.Intn(len(availablePositions))]
	return Factory{position: position, product: 0}, nil
}

func (s *Scenario) getAvailableFactoryPositions(chromosome Chromosome) []Position {
	positions := make([]Position, 0)
	for i := 0; i < s.width; i++ {
		for j := 0; j < s.height; j++ {
			pos := Position{i, j}
			if s.isPositionAvailableForFactory(chromosome.factories, chromosome.mines, pos) {
				positions = append(positions, pos)
			}
		}
	}
	return positions
}

func (s *Scenario) isPositionAvailableForFactory(factories []Factory, mines []Mine, position Position) bool {
	factoryRectangle := Rectangle{
		position: position,
		width:    FactoryWidth,
		height:   FactoryHeight,
	}
	if position.x+FactoryWidth > s.width || position.y+FactoryHeight > s.height {
		return false
	}
	for _, obstacle := range s.obstacles {
		if factoryRectangle.Intersects(obstacle) {
			return false
		}
	}
	for _, factory := range factories {
		if factoryRectangle.Intersects(factory.Rectangle()) {
			return false
		}
	}
	for _, mine := range mines {
		if mine.Intersects(factoryRectangle) {
			return false
		}
	}
	for _, deposit := range s.deposits {
		depositRectangle := deposit.Rectangle()
		extendedDepositRectangle := Rectangle{
			Position{depositRectangle.position.x - 1, depositRectangle.position.y - 1},
			depositRectangle.width + 2,
			depositRectangle.height + 2,
		}
		if factoryRectangle.Intersects(extendedDepositRectangle) {
			// top left
			positionIsCorner := position.y+FactoryHeight == deposit.position.y && position.x+FactoryHeight == deposit.position.x
			// top right
			positionIsCorner = positionIsCorner || (position.y+FactoryHeight == deposit.position.y && position.x == deposit.position.x+deposit.width)
			// bottom left
			positionIsCorner = positionIsCorner || (position.y == deposit.position.y+deposit.height && position.x+FactoryHeight == deposit.position.x)
			// bottom right
			positionIsCorner = positionIsCorner || (position.y == deposit.position.y+deposit.height && position.x == deposit.position.x+deposit.width)
			if !positionIsCorner {
				return false
			}
		}
	}
	return true
}

func (s *Scenario) attachedDepositEgress(mine Mine) (Position, error) {
	ingress := mine.Ingress()
	for _, deposit := range s.deposits {
		depositRectangle := deposit.Rectangle()
		for _, egressPosition := range []Position{{ingress.x + 1, ingress.y}, {ingress.x - 1, ingress.y}, {ingress.x, ingress.y + 1}, {ingress.x, ingress.y - 1}} {
			if depositRectangle.Contains(egressPosition) {
				return egressPosition, nil
			}
		}
	}
	return Position{}, nil
}

func (s *Scenario) isPositionAvailableForMine(factories []Factory, mines []Mine, mine Mine) bool {
	// mine is out of bounds
	boundRectangles := s.boundRectangles()
	if mine.IntersectsAny(boundRectangles) {
		return false
	}
	if mine.IntersectsAny(s.obstacles) {
		return false
	}
	for _, deposit := range s.deposits {
		if mine.Intersects(deposit.Rectangle()) {
			return false
		}
	}
	for _, factory := range factories {
		if mine.Intersects(factory.Rectangle()) {
			return false
		}
	}
	depositEgress, err := s.attachedDepositEgress(mine)
	for _, otherMine := range mines {
		if mine.Egress().NextTo(otherMine.Ingress()) || mine.Ingress().NextTo(otherMine.Egress()) {
			return false
		}
		if err == nil && otherMine.Ingress().NextTo(depositEgress) {
			return false
		}
		if mine.IntersectsMine(otherMine) {
			return false
		}
	}
	return true
}

func (s *Scenario) minesAroundDeposit(deposit Deposit, chromosome Chromosome) []Mine {
	/* For each mine direction, we go counter-clockwise.
	   There is always one case where the mine corner matches the deposit edge.
	   We always use the mine ingress coordinate as our iteration variable */

	positions := make([]Mine, 0)

	// Right
	positions = append(positions, Mine{position: Position{deposit.position.x + deposit.width, deposit.position.y + deposit.height - 1}, direction: Right})
	for i := deposit.position.y + deposit.height - 1; i >= deposit.position.y; i-- {
		positions = append(positions, Mine{position: Position{deposit.position.x + deposit.width + 1, i - 1}, direction: Right})
	}
	for i := deposit.position.x + deposit.width - 1; i >= deposit.position.x; i-- {
		positions = append(positions, Mine{position: Position{i + 1, deposit.position.y - 2}, direction: Right})
	}

	// Bottom
	positions = append(positions, Mine{position: Position{deposit.position.x - 1, deposit.position.y + deposit.height}, direction: Bottom})
	for i := deposit.position.x; i <= deposit.position.x+deposit.width-1; i++ {
		positions = append(positions, Mine{position: Position{i, deposit.position.y + deposit.height + 1}, direction: Bottom})
	}
	for i := deposit.position.y + deposit.height - 1; i >= deposit.position.y; i-- {
		positions = append(positions, Mine{position: Position{deposit.position.x + deposit.width, i + 1}, direction: Bottom})
	}

	// Left
	positions = append(positions, Mine{position: Position{deposit.position.x - 2, deposit.position.y - 1}, direction: Left})
	for i := deposit.position.y; i <= deposit.position.y+deposit.height-1; i++ {
		positions = append(positions, Mine{position: Position{deposit.position.x - 3, i}, direction: Left})
	}
	for i := deposit.position.x; i <= deposit.position.x+deposit.width-1; i++ {
		positions = append(positions, Mine{position: Position{i - 2, deposit.position.y + deposit.height}, direction: Left})
	}

	// Top
	positions = append(positions, Mine{position: Position{deposit.position.x + deposit.width - 1, deposit.position.y - 2}, direction: Top})
	for i := deposit.position.x + deposit.width - 1; i >= deposit.position.x; i-- {
		positions = append(positions, Mine{position: Position{i - 1, deposit.position.y - 3}, direction: Top})
	}
	for i := deposit.position.y; i <= deposit.position.y+deposit.height-1; i++ {
		positions = append(positions, Mine{position: Position{deposit.position.x - 2, i - 2}, direction: Top})
	}

	validPositions := make([]Mine, 0)
	for i := range positions {
		if s.isPositionAvailableForMine(chromosome.factories, chromosome.mines, positions[i]) {
			validPositions = append(validPositions, positions[i])
		}
	}
	return validPositions
}

func (g *GeneticAlgorithm) Run() (Solution, error) {
	chromosomes, err := g.generateChromosomes()
	if err != nil {
		return Solution{}, err
	}
	for i, chromosome := range chromosomes {
		chromosomes[i].fitness = g.evaluateFitness(chromosome)
	}
	for i := 0; i < g.iterations; i++ {
		sort.Slice(chromosomes, func(i, j int) bool {
			return chromosomes[i].fitness > chromosomes[j].fitness
		})
		if g.optimum != NoOptimum && chromosomes[0].fitness == g.optimum {
			log.Println("starting iteration", i+1, "/", g.iterations, "fitness", g.optimum, "(optimal)")
			break
		}
		log.Println("starting iteration", i+1, "/", g.iterations, "fitness", chromosomes[0].fitness)
		chromosomes = chromosomes[:g.populationSize]

		for j := 0; j < g.populationSize; j++ {
			newChromosome := g.crossover(chromosomes[rand.Intn(g.populationSize)], chromosomes[rand.Intn(g.populationSize)])
			newChromosome.fitness = g.evaluateFitness(newChromosome)
			chromosomes = append(chromosomes, newChromosome)
		}
		numChromosomes := len(chromosomes)
		for j := 0; j < numChromosomes; j++ {
			newChromosome := g.mutation(chromosomes[j])
			newChromosome.fitness = g.evaluateFitness(newChromosome)
			chromosomes = append(chromosomes, newChromosome)
		}
	}
	sort.Slice(chromosomes, func(i, j int) bool {
		return chromosomes[i].fitness > chromosomes[j].fitness
	})
	log.Println("final fitness", chromosomes[0].fitness)
	return chromosomes[0].Solution(), nil
}
