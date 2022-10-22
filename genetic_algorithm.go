package main

import (
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

	// sum of manhattan distances for each factory to all the deposits
	fitness := 0.0
	for _, deposit := range scenario.deposits {
		for _, factory := range chromosome.factories {
			fitness += math.Abs(float64(factory.position.x-deposit.position.x)) + math.Abs(float64(factory.position.y-deposit.position.y))
		}
	}
	return fitness
}

func generateChromosomes(n int, scenario Scenario) []Chromosome {
	chromosomes := make([]Chromosome, n)
	for i := 0; i < n; i++ {
		chromosomes[i] = generateChromosome(scenario)
	}
	return chromosomes
}

func generateChromosome(scenario Scenario) Chromosome {
	chromosome := Chromosome{mines: make([]Mine, 0)}
	factories := make([]Factory, scenario.numFactories)
	for i := 0; i < scenario.numFactories; i++ {
		factories[i] = getRandomFactory(scenario, factories[0:i])
	}
	chromosome.factories = factories
	return chromosome
}

func getRandomFactory(scenario Scenario, additionalFactories []Factory) Factory {
	scenario.factories = additionalFactories
	availablePositions := getAvailableFactoryPositions(scenario)
	rand.Seed(time.Now().UnixNano())
	//fmt.Printf("Found %d available positions for a factory.\n", len(availablePositions))
	position := availablePositions[rand.Intn(len(availablePositions))]
	return Factory{position: position, product: 0}
}

func getAvailableFactoryPositions(scenario Scenario) []Position {
	positions := make([]Position, 0)
	for i := 0; i < scenario.width; i++ {
		for j := 0; j < scenario.height; j++ {
			if isPositionAvailableForFactory(scenario, Position{
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

func isPositionAvailableForFactory(scenario Scenario, position Position) bool {
	for i := position.x; i < position.x+FACTORY_WIDTH; i++ {
		for j := position.y; j < position.y+FACTORY_HEIGHT; j++ {
			if !isPositionFree(scenario, Position{
				x: i,
				y: j,
			}) || i+FACTORY_WIDTH > scenario.width || j+FACTORY_HEIGHT > scenario.height {
				return false
			}
		}
	}
	return true
}

func isPositionFree(scenario Scenario, position Position) bool {
	for _, deposit := range scenario.deposits {
		xOverlap := position.x >= deposit.position.x && position.x < deposit.position.x+deposit.width
		yOverlap := position.y >= deposit.position.y && position.y < deposit.position.y+deposit.height
		if xOverlap && yOverlap {
			return false
		}
	}
	for _, factory := range scenario.factories {
		xOverlap := position.x >= factory.position.x && position.x < factory.position.x+FACTORY_WIDTH
		yOverlap := position.y >= factory.position.y && position.y < factory.position.y+FACTORY_HEIGHT
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

func runGeneticAlgorithm(maxIterations int, scenario Scenario, populationSize int, crossoverProbability float64) Scenario {
	chromosomes := generateChromosomes(populationSize, scenario)
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
	return scenario
}
