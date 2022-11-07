package main

import (
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
	scenario               Scenario
	iterations             int
	populationSize         int
	mutationProbability    float64
	crossoverProbability   float64
	initialMinNumFactories int
	initialMaxNumFactories int
	initialNumMines        int
	optimum                int
	numPaths               int
}

func (c Chromosome) Solution() Solution {
	solution := Solution{
		factories: make([]Factory, len(c.factories)),
		mines:     make([]Mine, len(c.mines)),
		paths:     []Path{},
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
		if len(path) > 0 {
			solution.paths = append(solution.paths, path)
		}
	}
	return solution
}

const NumTriesPerChromosome = 10

func (g *GeneticAlgorithm) crossover(chromosome Chromosome, chromosome2 Chromosome) Chromosome {
	newChromosome := Chromosome{}
	for i := 0; i < int(math.Min(float64(len(chromosome.mines)), float64(len(chromosome2.mines)))); i++ {
		if rand.Float64() > g.crossoverProbability {
			newChromosome.mines = append(newChromosome.mines, chromosome.mines[i])
		} else {
			newChromosome.mines = append(newChromosome.mines, chromosome2.mines[i])
		}
	}
	if rand.Float64() > 0.5 {
		if len(chromosome.mines) > len(chromosome2.mines) {
			for i := len(chromosome2.mines); i < len(chromosome.mines); i++ {
				newChromosome.mines = append(newChromosome.mines, chromosome.mines[i])
			}
		} else {
			for i := len(chromosome.mines); i < len(chromosome2.mines); i++ {
				newChromosome.mines = append(newChromosome.mines, chromosome2.mines[i])
			}
		}
	}
	for i := 0; i < int(math.Min(float64(len(chromosome.factories)), float64(len(chromosome2.factories)))); i++ {
		if rand.Float64() > g.crossoverProbability {
			newChromosome.factories = append(newChromosome.factories, chromosome.factories[i])
		} else {
			newChromosome.factories = append(newChromosome.factories, chromosome2.factories[i])
		}
	}
	if rand.Float64() > 0.5 {
		if len(chromosome.factories) > len(chromosome2.factories) {
			for i := len(chromosome2.factories); i < len(chromosome.factories); i++ {
				newChromosome.factories = append(newChromosome.factories, chromosome.factories[i])
			}
		} else {
			for i := len(chromosome.factories); i < len(chromosome2.factories); i++ {
				newChromosome.factories = append(newChromosome.factories, chromosome2.factories[i])
			}
		}
	}
	for i := 0; i < len(chromosome.paths); i++ {
		if rand.Float64() > g.crossoverProbability {
			newChromosome.paths = append(newChromosome.paths, chromosome.paths[i])
		} else {
			newChromosome.paths = append(newChromosome.paths, chromosome2.paths[i])
		}
	}
	return newChromosome
}

func (c Chromosome) copy() Chromosome {
	newChromosome := Chromosome{
		fitness: 0,
	}
	for _, factory := range c.factories {
		newFactory := Factory{
			position: Position{factory.position.x, factory.position.y},
			product:  0,
		}
		newChromosome.factories = append(newChromosome.factories, newFactory)
	}
	for _, mine := range c.mines {
		newMine := Mine{
			position:         Position{mine.position.x, mine.position.y},
			direction:        mine.direction,
			cachedRectangles: nil,
		}
		newChromosome.mines = append(newChromosome.mines, newMine)
	}
	for _, path := range c.paths {
		newPath := Path{}
		for _, conveyor := range path {
			newConveyor := Conveyor{
				position:  Position{conveyor.position.x, conveyor.position.y},
				direction: conveyor.direction,
				length:    conveyor.length,
			}
			newPath = append(newPath, newConveyor)
		}
		newChromosome.paths = append(newChromosome.paths, newPath)
	}
	return newChromosome
}

func (g *GeneticAlgorithm) addFactoryMutation(chromosome Chromosome) Chromosome {
	newFactory, err := g.randomFactory(chromosome)
	if err == nil {
		chromosome.factories = append(chromosome.factories, newFactory)
	}
	return chromosome
}

func (g *GeneticAlgorithm) removeFactoryMutation(chromosome Chromosome) Chromosome {
	if len(chromosome.factories) > 0 {
		removeIndex := rand.Intn(len(chromosome.factories))
		chromosome.factories[removeIndex] = chromosome.factories[len(chromosome.factories)-1]
		chromosome.factories = chromosome.factories[:len(chromosome.factories)-1]
	}
	return chromosome
}

func (g *GeneticAlgorithm) addMineMutation(chromosome Chromosome) Chromosome {
	newMine, err := g.randomMine(g.scenario.deposits[rand.Intn(len(g.scenario.deposits))], chromosome)
	if err == nil {
		chromosome.mines = append(chromosome.mines, newMine)
	}
	return chromosome
}

func (g *GeneticAlgorithm) removeMineMutation(chromosome Chromosome) Chromosome {
	if len(chromosome.mines) > 0 {
		removeIndex := rand.Intn(len(chromosome.mines))
		chromosome.mines[removeIndex] = chromosome.mines[len(chromosome.mines)-1]
		chromosome.mines = chromosome.mines[:len(chromosome.mines)-1]
	}
	return chromosome
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
					newMine, err := g.randomMine(deposit, newChromosome)
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
			newFactory, err := g.randomFactory(newChromosome)
			if err != nil {
				newChromosome.factories = append(newChromosome.factories, factory)
			} else {
				newChromosome.factories = append(newChromosome.factories, newFactory)
			}
		}
	}
	for _, path := range chromosome.paths {
		if rand.Float64() > g.mutationProbability {
			newChromosome.paths = append(newChromosome.paths, path)
		} else {
			randomFactory := newChromosome.factories[rand.Intn(len(newChromosome.factories))]
			randomMine := newChromosome.mines[rand.Intn(len(newChromosome.mines))]
			newPath, err := g.pathMineToFactory(newChromosome, randomMine, randomFactory)
			if err != nil {
				newChromosome.paths = append(newChromosome.paths, path)
			} else {
				newChromosome.paths = append(newChromosome.paths, newPath)
			}
		}

	}
	return newChromosome
}

func (g *GeneticAlgorithm) evaluateFitness(chromosome Chromosome) int {
	fitness, err := g.scenario.evaluateSolution(chromosome.Solution())
	if err != nil {
		return math.MinInt
	}
	return fitness
}

func (g *GeneticAlgorithm) generateChromosome() (Chromosome, error) {
	chromosome := Chromosome{mines: make([]Mine, 0), factories: make([]Factory, 0)}
	for i := 0; i < g.initialNumMines; i++ {
		deposit := g.scenario.deposits[i%len(g.scenario.deposits)]
		mine, err := g.randomMine(deposit, chromosome)
		if err != nil {
			return chromosome, err
		}
		chromosome.mines = append(chromosome.mines, mine)
	}
	for i := 0; i < g.initialMinNumFactories+rand.Intn(g.initialMaxNumFactories-g.initialMinNumFactories); i++ {
		factory, err := g.randomFactory(chromosome)
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
			path, err = g.pathMineToFactory(chromosome, randomMine, randomFactory)
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
		for j := 0; j < g.populationSize; j++ {
			newChromosome := g.addFactoryMutation(chromosomes[rand.Intn(len(chromosomes))].copy())
			newChromosome.fitness = g.evaluateFitness(newChromosome)
			chromosomes = append(chromosomes, newChromosome)
		}
		for j := 0; j < g.populationSize; j++ {
			newChromosome := g.removeFactoryMutation(chromosomes[rand.Intn(len(chromosomes))].copy())
			newChromosome.fitness = g.evaluateFitness(newChromosome)
			chromosomes = append(chromosomes, newChromosome)
		}
		for j := 0; j < g.populationSize; j++ {
			newChromosome := g.removeMineMutation(chromosomes[rand.Intn(len(chromosomes))].copy())
			newChromosome.fitness = g.evaluateFitness(newChromosome)
			chromosomes = append(chromosomes, newChromosome)
		}
		for j := 0; j < g.populationSize; j++ {
			newChromosome := g.addMineMutation(chromosomes[rand.Intn(len(chromosomes))].copy())
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
