package main

import (
	"fmt"
	"math/rand"
	"time"
)

const FACTORY_WIDTH = 5
const FACTORY_HEIGHT = 5

type Position struct {
	x int
	y int
}

type Deposit struct {
	position Position
	width    int
	height   int
	subtype  int
}

type Obstacle struct {
	position Position
	height   int
	width    int
}

type Scenario struct {
	width        int
	height       int
	deposits     []Deposit
	obstacles    []Obstacle
	factories    []Factory
	turns        int
	numFactories int
}

type Factory struct {
	position Position
	product  int
}

type Mine struct {
	position    Position
	orientation int
}

type Chromosome struct {
	factories []Factory
	mines     []Mine
}

func crossover(chromosome Chromosome, chromosome2 Chromosome, probability float32) Chromosome {
	var newChromosome Chromosome = chromosome
	return newChromosome
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
	rand.Seed(time.Now().UnixNano() + int64(len(additionalFactories)))
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

func getDefaultScenario() Scenario {
	deposits := make([]Deposit, 2)
	deposits[0] = Deposit{width: 5, height: 5, subtype: 1, position: Position{0, 0}}
	deposits[1] = Deposit{width: 5, height: 5, subtype: 0, position: Position{10, 3}}

	obstacles := make([]Obstacle, 1)
	obstacles[0] = Obstacle{
		position: Position{8, 10},
		height:   4,
		width:    4,
	}
	return Scenario{
		width:        40,
		height:       40,
		deposits:     deposits,
		obstacles:    obstacles,
		numFactories: 10,
	}
}

func main() {
	var scenario Scenario = getDefaultScenario()
	fmt.Println(scenario)
	var populationSize int = 30
	//var crossoverProbability float32 = 0.7
	//var mutationProbability float32 = 0.5

	chromosomes := generateChromosomes(populationSize, scenario)
	fmt.Println(chromosomes)
}
