package main

import (
	"errors"
	"math"
	"math/rand"
	"sort"
	"time"
)

type Chromosome struct {
	factories []Factory
	mines     []Mine
	fitness   float64
}

const NumTriesPerChromosome = 10

func crossover(chromosome Chromosome, chromosome2 Chromosome, probability float64, scenario Scenario) Chromosome {
	newChromosome := Chromosome{}
	for i := 0; i < scenario.numFactories; i++ {
		rand.Seed(time.Now().UnixNano())
		if rand.Float64() > probability {
			newChromosome.factories = append(newChromosome.factories, chromosome.factories[i])
		} else {
			newChromosome.factories = append(newChromosome.factories, chromosome2.factories[i])
		}
		//TODO: Check if still valid
	}
	return newChromosome
}

func evaluateFitness(chromosome Chromosome, scenario Scenario) float64 {
	// TODO: use A* or other metric

	for i, factory := range chromosome.factories {
		if !isPositionAvailableForFactory(scenario, chromosome.factories[:i], factory.position) {
			return math.Inf(1)
		}
	}

	// sum of manhattan distances for each factory to all the deposits
	fitness := 0.0
	for _, deposit := range scenario.deposits {
		for _, factory := range chromosome.factories {
			fitness += math.Abs(float64(factory.position.x-deposit.position.x)) + math.Abs(float64(factory.position.y-deposit.position.y))
		}
	}
	return fitness
}

func generateChromosomes(n int, scenario Scenario) ([]Chromosome, error) {
	chromosomes := make([]Chromosome, n)
	for i := 0; i < n; i++ {
		foundChromosome := false
		for j := 0; j < NumTriesPerChromosome; j++ {
			chromosome, err := generateChromosome(scenario)
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

func generateChromosome(scenario Scenario) (Chromosome, error) {
	chromosome := Chromosome{mines: make([]Mine, 0)}
	factories := make([]Factory, scenario.numFactories)
	for i := 0; i < scenario.numFactories; i++ {
		var err error
		factories[i], err = getRandomFactory(scenario, factories[0:i])
		if err != nil {
			return chromosome, err
		}
	}
	chromosome.factories = factories
	return chromosome, nil
}

func getRandomFactory(scenario Scenario, additionalFactories []Factory) (Factory, error) {
	availablePositions := getAvailableFactoryPositions(scenario, additionalFactories)
	rand.Seed(time.Now().UnixNano())
	//fmt.Printf("Found %d available positions for a factory.\n", len(availablePositions))
	if len(availablePositions) == 0 {
		return Factory{}, errors.New("no factory positions available")
	}
	position := availablePositions[rand.Intn(len(availablePositions))]
	return Factory{position: position, product: 0}, nil
}

func getAvailableFactoryPositions(scenario Scenario, additionalFactories []Factory) []Position {
	positions := make([]Position, 0)
	for i := 0; i < scenario.width; i++ {
		for j := 0; j < scenario.height; j++ {
			if isPositionAvailableForFactory(scenario, additionalFactories, Position{
				x: i,
				y: j,
			}) {
				positions = append(positions, Position{
					x: i,
					y: j,
				})
			}
		}
	}
	return positions
}

func isPositionAvailableForFactory(scenario Scenario, additionalFactories []Factory, position Position) bool {
	scenario.factories = additionalFactories
	for i := position.x; i < position.x+FactoryWidth; i++ {
		for j := position.y; j < position.y+FactoryHeight; j++ {
			if !isPositionAvailableForFactoryCell(scenario, Position{
				x: i,
				y: j,
			}) {
				return false
			}
		}
	}
	return true
}

func isPositionAvailableForFactoryCell(scenario Scenario, position Position) bool {
	if position.x >= scenario.width || position.y >= scenario.height {
		return false
	}
	for _, deposit := range scenario.deposits {
		xOverlap := position.x >= deposit.position.x-1 && position.x < deposit.position.x+deposit.width+1
		yOverlap := position.y >= deposit.position.y-1 && position.y < deposit.position.y+deposit.height+1
		if xOverlap && yOverlap {
			positionIsCorner := position.y == deposit.position.y-1 && position.x == deposit.position.x-1
			positionIsCorner = positionIsCorner || (position.y == deposit.position.y+deposit.height && position.x == deposit.position.x-1)
			positionIsCorner = positionIsCorner || (position.y == deposit.position.y-1 && position.x == deposit.position.x+deposit.width)
			positionIsCorner = positionIsCorner || (position.y == deposit.position.y+deposit.height && position.x == deposit.position.x+deposit.width)
			if !positionIsCorner {
				return false
			}
		}
	}
	for _, factory := range scenario.factories {
		xOverlap := position.x >= factory.position.x && position.x < factory.position.x+FactoryWidth
		yOverlap := position.y >= factory.position.y && position.y < factory.position.y+FactoryHeight
		if xOverlap && yOverlap {
			return false
		}
	}
	for _, obstacles := range scenario.obstacles {
		xOverlap := position.x >= obstacles.position.x && position.x < obstacles.position.x+obstacles.width
		yOverlap := position.y >= obstacles.position.y && position.y < obstacles.position.y+obstacles.height
		if xOverlap && yOverlap {
			return false
		}
	}
	return true
}

func runGeneticAlgorithm(maxIterations int, scenario Scenario, populationSize int, crossoverProbability float64) (Scenario, error) {
	chromosomes, err := generateChromosomes(populationSize, scenario)
	if err != nil {
		return scenario, err
	}
	for i, chromosome := range chromosomes {
		chromosomes[i].fitness = evaluateFitness(chromosome, scenario)
	}
	for i := 0; i < maxIterations; i++ {
		sort.Slice(chromosomes, func(i, j int) bool {
			return chromosomes[i].fitness < chromosomes[j].fitness
		})
		chromosomes = chromosomes[:populationSize]

		for j := 0; j < populationSize; j++ {
			rand.Seed(time.Now().UnixNano())
			newChromosome := crossover(chromosomes[rand.Intn(populationSize)], chromosomes[rand.Intn(populationSize)], crossoverProbability, scenario)
			newChromosome.fitness = evaluateFitness(newChromosome, scenario)
			chromosomes = append(chromosomes, newChromosome)
		}
	}
	sort.Slice(chromosomes, func(i, j int) bool {
		return chromosomes[i].fitness < chromosomes[j].fitness
	})
	scenario.factories = chromosomes[0].factories
	return scenario, nil
}
