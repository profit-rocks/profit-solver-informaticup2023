package main

import (
	"errors"
	"log"
	"math"
	"math/rand"
	"sort"
)

type Chromosome struct {
	factories []Factory
	mines     []Mine
	fitness   float64
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
}

func (c Chromosome) Solution() Solution {
	solution := Solution{
		factories: make([]Factory, len(c.factories)),
		mines:     make([]Mine, len(c.mines)),
	}
	copy(solution.factories, c.factories)
	copy(solution.mines, c.mines)
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
	return newChromosome
}

func (g *GeneticAlgorithm) mutation(chromosome Chromosome) Chromosome {
	newChromosome := Chromosome{}
	for _, mine := range chromosome.mines {
		if rand.Float64() > g.mutationProbability {
			newChromosome.mines = append(newChromosome.mines, mine)
		} else {
			newMine, err := g.getRandomMine(newChromosome)
			if err != nil {
				newChromosome.mines = append(newChromosome.mines, mine)
			} else {
				newChromosome.mines = append(newChromosome.mines, newMine)
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
	return newChromosome
}

func (g *GeneticAlgorithm) evaluateFitness(chromosome Chromosome) float64 {
	// TODO: use A* or other metric
	for i, mine := range chromosome.mines {
		copiedChromosome := chromosome
		copiedChromosome.mines = chromosome.mines[:i]
		if !g.isPositionAvailableForMine(copiedChromosome, mine) {
			return math.Inf(1)
		}
	}

	for i, factory := range chromosome.factories {
		copiedChromosome := chromosome
		copiedChromosome.factories = chromosome.factories[:i]
		if !g.isPositionAvailableForFactory(copiedChromosome, factory.position) {
			return math.Inf(1)
		}
	}

	// sum of manhattan distances for each factory to all the deposits
	fitness := 0.0
	for _, mine := range chromosome.mines {
		for _, factory := range chromosome.factories {
			fitness += factory.position.ManhattanDist(mine.position)
		}
	}
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
			return chromosomes, errors.New("exceeded NumTriesPerChromosome in generateChromosomes, probably trying to place too many factories")
		}
	}
	return chromosomes, nil
}

func (g *GeneticAlgorithm) generateChromosome() (Chromosome, error) {
	chromosome := Chromosome{mines: make([]Mine, 0), factories: make([]Factory, 0)}
	for i := 0; i < g.numMines; i++ {
		mine, err := g.getRandomMine(chromosome)
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
	return chromosome, nil
}

func (g *GeneticAlgorithm) getRandomMine(chromosome Chromosome) (Mine, error) {
	for _, deposit := range g.scenario.deposits {
		availableMines := g.minesAroundDeposit(deposit, chromosome)
		if len(availableMines) != 0 {
			return availableMines[rand.Intn(len(availableMines))], nil
		}
	}
	return Mine{}, errors.New("no mines available")
}

func (g *GeneticAlgorithm) getRandomFactory(chromosome Chromosome) (Factory, error) {
	availablePositions := g.getAvailableFactoryPositions(chromosome)
	if len(availablePositions) == 0 {
		return Factory{}, errors.New("no factory positions available")
	}
	position := availablePositions[rand.Intn(len(availablePositions))]
	return Factory{position: position, product: 0}, nil
}

func (g *GeneticAlgorithm) getAvailableFactoryPositions(chromosome Chromosome) []Position {
	positions := make([]Position, 0)
	for i := 0; i < g.scenario.width; i++ {
		for j := 0; j < g.scenario.height; j++ {
			pos := Position{i, j}
			if g.isPositionAvailableForFactory(chromosome, pos) {
				positions = append(positions, pos)
			}
		}
	}
	return positions
}

func (g *GeneticAlgorithm) isPositionAvailableForFactory(chromosome Chromosome, position Position) bool {
	factoryRectangle := Rectangle{
		position: position,
		width:    FactoryWidth,
		height:   FactoryHeight,
	}
	if position.x+FactoryWidth > g.scenario.width || position.y+FactoryHeight > g.scenario.height {
		return false
	}
	for _, obstacle := range g.scenario.obstacles {
		if factoryRectangle.Intersects(obstacle) {
			return false
		}
	}
	for _, factory := range chromosome.factories {
		if factoryRectangle.Intersects(factory.Rectangle()) {
			return false
		}
	}
	for _, deposit := range g.scenario.deposits {
		depositRectangle := deposit.Rectangle()
		extendedDepositRectangle := Rectangle{
			position: Position{
				x: depositRectangle.position.x - 1,
				y: depositRectangle.position.y - 1,
			},
			width:  depositRectangle.width + 2,
			height: depositRectangle.height + 2,
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

func (g *GeneticAlgorithm) isPositionAvailableForMine(chromosome Chromosome, mine Mine) bool {
	// mine is out of bounds
	boundRectangles := g.scenario.boundRectangles()
	for _, borderRectangle := range boundRectangles {
		if mine.Intersects(borderRectangle) {
			return false
		}
	}
	for _, obstacle := range g.scenario.obstacles {
		if mine.Intersects(obstacle) {
			return false
		}
	}
	for _, deposit := range g.scenario.deposits {
		if mine.Intersects(deposit.Rectangle()) {
			return false
		}
	}
	for _, factory := range chromosome.factories {
		if mine.Intersects(factory.Rectangle()) {
			return false
		}
	}
	// TODO: check if one mine's egress is adjacent to another mine's ingress
	for _, otherMine := range chromosome.mines {
		for _, rectangle := range otherMine.Rectangles() {
			if mine.Intersects(rectangle) {
				return false
			}
		}
	}
	return true
}

func (g *GeneticAlgorithm) minesAroundDeposit(deposit Deposit, chromosome Chromosome) []Mine {
	/* For each mine direction, we go counter-clockwise.
	   There is always one case where the mine corner matches the deposit edge.
	   We always use the mine ingress coordinate as our iteration variable */

	positions := make([]Mine, 0)

	// Right
	positions = append(positions, Mine{Position{deposit.position.x + deposit.width, deposit.position.y + deposit.height - 1}, Right})
	for i := deposit.position.y + deposit.height - 1; i >= deposit.position.y; i-- {
		positions = append(positions, Mine{Position{deposit.position.x + deposit.width + 1, i - 1}, Right})
	}
	for i := deposit.position.x + deposit.width - 1; i >= deposit.position.x; i-- {
		positions = append(positions, Mine{Position{i + 1, deposit.position.y - 2}, Right})
	}

	// Bottom
	positions = append(positions, Mine{Position{deposit.position.x - 1, deposit.position.y + deposit.height}, Bottom})
	for i := deposit.position.x; i <= deposit.position.x+deposit.width-1; i++ {
		positions = append(positions, Mine{Position{i, deposit.position.y + deposit.height + 1}, Bottom})
	}
	for i := deposit.position.y + deposit.height - 1; i >= deposit.position.y; i-- {
		positions = append(positions, Mine{Position{deposit.position.x + deposit.width, i + 1}, Bottom})
	}

	// Left
	positions = append(positions, Mine{Position{deposit.position.x - 2, deposit.position.y - 1}, Left})
	for i := deposit.position.y; i <= deposit.position.y+deposit.height-1; i++ {
		positions = append(positions, Mine{Position{deposit.position.x - 3, i}, Left})
	}
	for i := deposit.position.x; i <= deposit.position.x+deposit.width-1; i++ {
		positions = append(positions, Mine{Position{i - 2, deposit.position.y + deposit.height}, Left})
	}

	// Top
	positions = append(positions, Mine{Position{deposit.position.x + deposit.width - 1, deposit.position.y - 2}, Top})
	for i := deposit.position.x + deposit.width - 1; i >= deposit.position.x; i-- {
		positions = append(positions, Mine{Position{i - 1, deposit.position.y - 3}, Top})
	}
	for i := deposit.position.y; i <= deposit.position.y+deposit.height-1; i++ {
		positions = append(positions, Mine{Position{deposit.position.x - 2, i - 2}, Top})
	}

	validPositions := make([]Mine, 0)
	for _, position := range positions {
		if g.isPositionAvailableForMine(chromosome, position) {
			validPositions = append(validPositions, position)
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
			return chromosomes[i].fitness < chromosomes[j].fitness
		})
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
		return chromosomes[i].fitness < chromosomes[j].fitness
	})
	log.Println("final fitness", chromosomes[0].fitness)
	return chromosomes[0].Solution(), nil
}
