package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

type Chromosome struct {
	factories []Factory
	mines     []Mine
	fitness   float64
}

const NumTriesPerChromosome = 10

func chromosomeToExportableScenario(scenario Scenario, chromosome Chromosome) ExportableScenario {
	exportableScenario := ExportableScenario{Height: scenario.height, Width: scenario.width, Turns: 100, Time: 100}

	for _, deposit := range scenario.deposits {
		exportableScenario.Objects = append(exportableScenario.Objects, Object{
			ObjectType: "deposit",
			Subtype:    deposit.subtype,
			X:          deposit.position.x,
			Y:          deposit.position.y,
			Width:      deposit.width,
			Height:     deposit.height,
		})
	}
	for _, obstacle := range scenario.obstacles {
		exportableScenario.Objects = append(exportableScenario.Objects, Object{
			ObjectType: "obstacle",
			X:          obstacle.position.x,
			Y:          obstacle.position.y,
			Width:      obstacle.width,
			Height:     obstacle.height,
		})

	}

	for _, factory := range chromosome.factories {
		exportableScenario.Objects = append(exportableScenario.Objects, Object{
			ObjectType: "factory",
			Subtype:    factory.product,
			X:          factory.position.x,
			Y:          factory.position.y,
		})
	}
	return exportableScenario
}

func crossover(chromosome Chromosome, chromosome2 Chromosome, probability float64, scenario Scenario) Chromosome {
	newChromosome := Chromosome{}
	for i := 0; i < scenario.numFactories; i++ {
		if rand.Float64() > probability {
			newChromosome.factories = append(newChromosome.factories, chromosome.factories[i])
		} else {
			newChromosome.factories = append(newChromosome.factories, chromosome2.factories[i])
		}
	}
	return newChromosome
}

func mutation(chromosome Chromosome, probability float64, scenario Scenario) Chromosome {
	newChromosome := Chromosome{}
	for _, factory := range chromosome.factories {
		fl := rand.Float64()
		if fl > probability {
			newChromosome.factories = append(newChromosome.factories, factory)
		} else {
			newFactory, err := getRandomFactory(scenario, newChromosome)
			if err != nil {
				newChromosome.factories = append(newChromosome.factories, factory)
			} else {
				newChromosome.factories = append(newChromosome.factories, newFactory)
			}
		}
	}
	return newChromosome
}

func evaluateFitness(chromosome Chromosome, scenario Scenario) float64 {
	// TODO: use A* or other metric

	allFactories := chromosome.factories
	for i, factory := range chromosome.factories {
		chromosome.factories = allFactories[:i]
		if !isPositionAvailableForFactory(scenario, chromosome, factory.position) {
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
	chromosome := Chromosome{mines: make([]Mine, 0), factories: make([]Factory, 0)}
	for i := 0; i < scenario.numFactories; i++ {
		factory, err := getRandomFactory(scenario, chromosome)
		if err != nil {
			return chromosome, err
		}
		chromosome.factories = append(chromosome.factories, factory)
	}
	return chromosome, nil
}

func getRandomFactory(scenario Scenario, chromosome Chromosome) (Factory, error) {
	availablePositions := getAvailableFactoryPositions(scenario, chromosome)
	//fmt.Printf("Found %d available positions for a factory.\n", len(availablePositions))
	if len(availablePositions) == 0 {
		return Factory{}, errors.New("no factory positions available")
	}
	position := availablePositions[rand.Intn(len(availablePositions))]
	return Factory{position: position, product: 0}, nil
}

func getAvailableFactoryPositions(scenario Scenario, chromosome Chromosome) []Position {
	positions := make([]Position, 0)
	for i := 0; i < scenario.width; i++ {
		for j := 0; j < scenario.height; j++ {
			if isPositionAvailableForFactory(scenario, chromosome, Position{
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

func isPositionAvailableForFactory(scenario Scenario, chromosome Chromosome, position Position) bool {
	for i := position.x; i < position.x+FactoryWidth; i++ {
		for j := position.y; j < position.y+FactoryHeight; j++ {
			if !isPositionAvailableForFactoryCell(scenario, chromosome, Position{
				x: i,
				y: j,
			}) {
				return false
			}
		}
	}
	return true
}

func isPositionAvailableForFactoryCell(scenario Scenario, chromosome Chromosome, position Position) bool {
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
	for _, factory := range chromosome.factories {
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

func runGeneticAlgorithm(maxIterations int, scenario Scenario, populationSize int, mutationProbability float64, crossoverProbability float64) (ExportableScenario, error) {
	chromosomes, err := generateChromosomes(populationSize, scenario)
	if err != nil {
		return ExportableScenario{}, err
	}
	for i, chromosome := range chromosomes {
		chromosomes[i].fitness = evaluateFitness(chromosome, scenario)
	}
	for i := 0; i < maxIterations; i++ {
		sort.Slice(chromosomes, func(i, j int) bool {
			return chromosomes[i].fitness < chromosomes[j].fitness
		})
		fmt.Println("starting iteration", i+1, "/", maxIterations, "fitness", chromosomes[0].fitness)
		chromosomes = chromosomes[:populationSize]

		for j := 0; j < populationSize; j++ {
			newChromosome := crossover(chromosomes[rand.Intn(populationSize)], chromosomes[rand.Intn(populationSize)], crossoverProbability, scenario)
			newChromosome.fitness = evaluateFitness(newChromosome, scenario)
			chromosomes = append(chromosomes, newChromosome)
		}
		numChromosomes := len(chromosomes)
		for j := 0; j < numChromosomes; j++ {
			newChromosome := mutation(chromosomes[j], mutationProbability, scenario)
			newChromosome.fitness = evaluateFitness(newChromosome, scenario)
			chromosomes = append(chromosomes, newChromosome)
		}
	}
	sort.Slice(chromosomes, func(i, j int) bool {
		return chromosomes[i].fitness < chromosomes[j].fitness
	})
	fmt.Println("final fitness", chromosomes[0].fitness)
	return chromosomeToExportableScenario(scenario, chromosomes[0]), nil
}
