package main

import (
	"errors"
	"log"
	"math/rand"
	"sort"
)

type Path struct {
	conveyors []Conveyor
}

const NumRoundsPerIteration = 1000
const NumMutationsPerRound = 10

const NumPathRetries = 10

type Chromosome struct {
	factories []Factory
	mines     []Mine
	combiners []Combiner
	paths     []Path
	fitness   int
}

// MutationFunction expects a copy of the chromosome which it can modify.
type MutationFunction func(algorithm *GeneticAlgorithm, chromosome Chromosome) (Chromosome, error)

// Mutations contains all mutation functions, performed in multiple layers. Each layer operates on the same set of chromosomes
var Mutations = []MutationFunction{
	(*GeneticAlgorithm).addMineMutation,
	(*GeneticAlgorithm).removeMineMutation,
	(*GeneticAlgorithm).moveMinesMutation,
	(*GeneticAlgorithm).addFactoryMutation,
	(*GeneticAlgorithm).removeFactoryMutation,
	(*GeneticAlgorithm).moveFactoriesMutation,
	(*GeneticAlgorithm).addPathMutation,
	(*GeneticAlgorithm).removePathMutation,
	(*GeneticAlgorithm).movePathMutation,
	(*GeneticAlgorithm).addCombinerMutation,
	(*GeneticAlgorithm).removeCombinerMutation,
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
	chromosomeChannel      chan<- Chromosome
	doneChannel            chan<- bool
}

func removeRandomElement[T any](arr []T) []T {
	removeIndex := rand.Intn(len(arr))
	arr[removeIndex] = arr[len(arr)-1]
	return arr[:len(arr)-1]
}

func removeUniform[T any](arr []T, probability float64) []T {
	for i := 0; i < len(arr); i++ {
		if rand.Float64() < probability {
			arr[i] = arr[len(arr)-1]
			arr = arr[:len(arr)-1]
			i--
		}
	}
	return arr
}

func (c Chromosome) Solution() Solution {
	solution := Solution{
		factories: make([]Factory, len(c.factories)),
		mines:     make([]Mine, len(c.mines)),
		paths:     []Path{},
		combiners: make([]Combiner, len(c.combiners)),
	}
	for i, combiner := range c.combiners {
		solution.combiners[i] = Combiner{
			position:  combiner.position,
			direction: combiner.direction,
		}
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
		if len(path.conveyors) > 0 {
			solution.paths = append(solution.paths, path)
		}
	}
	return solution
}

func (p Path) copy() Path {
	path := Path{}
	for _, c := range p.conveyors {
		path.conveyors = append(path.conveyors, c)
	}
	return path
}

func minInt(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

func (g *GeneticAlgorithm) crossover(chromosome Chromosome, chromosome2 Chromosome) (Chromosome, error) {
	// TODO: only output valid chromosomes
	newChromosome := Chromosome{}
	for i := 0; i < minInt(len(chromosome.mines), len(chromosome2.mines)); i++ {
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
	for i := 0; i < minInt(len(chromosome.factories), len(chromosome2.factories)); i++ {
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
	for i := 0; i < minInt(len(chromosome.paths), len(chromosome2.paths)); i++ {
		if rand.Float64() > g.crossoverProbability {
			newChromosome.paths = append(newChromosome.paths, chromosome.paths[i].copy())
		} else {
			newChromosome.paths = append(newChromosome.paths, chromosome2.paths[i].copy())
		}
	}
	if rand.Float64() > 0.5 {
		if len(chromosome.paths) > len(chromosome2.paths) {
			for i := len(chromosome2.paths); i < len(chromosome.paths); i++ {
				newChromosome.paths = append(newChromosome.paths, chromosome.paths[i].copy())
			}
		} else {
			for i := len(chromosome.paths); i < len(chromosome2.paths); i++ {
				newChromosome.paths = append(newChromosome.paths, chromosome2.paths[i].copy())
			}
		}
	}
	return newChromosome, g.scenario.checkValidity(newChromosome.Solution())
}

func (c Chromosome) copy() Chromosome {
	newChromosome := Chromosome{
		fitness: 0,
	}
	for _, factory := range c.factories {
		newChromosome.factories = append(newChromosome.factories, factory)
	}
	for _, mine := range c.mines {
		newChromosome.mines = append(newChromosome.mines, mine)
	}
	for _, path := range c.paths {
		newChromosome.paths = append(newChromosome.paths, path.copy())
	}
	return newChromosome
}

func (g *GeneticAlgorithm) addFactoryMutation(chromosome Chromosome) (Chromosome, error) {
	newFactory, err := g.scenario.randomFactory(chromosome)
	if err != nil {
		return Chromosome{}, err
	}
	chromosome.factories = append(chromosome.factories, newFactory)
	return chromosome, nil
}

func (g *GeneticAlgorithm) removeFactoryMutation(chromosome Chromosome) (Chromosome, error) {
	if len(chromosome.factories) == 0 {
		return chromosome, errors.New("no factories to remove")
	}
	chromosome.factories = removeRandomElement(chromosome.factories)
	return chromosome, nil
}

func (g *GeneticAlgorithm) addCombinerMutation(chromosome Chromosome) (Chromosome, error) {
	newCombiner, err := g.scenario.randomCombiner(chromosome)
	if err != nil {
		return Chromosome{}, err
	}
	chromosome.combiners = append(chromosome.combiners, newCombiner)
	return chromosome, nil
}

func (g *GeneticAlgorithm) removeCombinerMutation(chromosome Chromosome) (Chromosome, error) {
	if len(chromosome.combiners) == 0 {
		return chromosome, errors.New("no combiners to remove")
	}
	chromosome.combiners = removeRandomElement(chromosome.combiners)
	return chromosome, nil
}

func (g *GeneticAlgorithm) addMineMutation(chromosome Chromosome) (Chromosome, error) {
	newMine, err := g.randomMine(g.scenario.deposits[rand.Intn(len(g.scenario.deposits))], chromosome)
	if err != nil {
		return chromosome, err
	}
	chromosome.mines = append(chromosome.mines, newMine)
	return chromosome, nil
}

func (g *GeneticAlgorithm) removeMineMutation(chromosome Chromosome) (Chromosome, error) {
	if len(chromosome.mines) == 0 {
		return chromosome, errors.New("no mines to remove")
	}
	chromosome.mines = removeRandomElement(chromosome.mines)
	return chromosome, nil
}

func (g *GeneticAlgorithm) addPathMutation(chromosome Chromosome) (Chromosome, error) {
	if len(chromosome.mines) == 0 || len(chromosome.factories) == 0 {
		return chromosome, errors.New("no mines or factories to add path")
	}
	// TODO: take product subtypes into account
	for j := 0; j < NumPathRetries; j++ {
		randomFactory := chromosome.factories[rand.Intn(len(chromosome.factories))]
		randomMine := chromosome.mines[rand.Intn(len(chromosome.mines))]
		newPath, err := g.pathMineToFactory(chromosome, randomMine, randomFactory)
		if err == nil {
			chromosome.paths = append(chromosome.paths, newPath)
			return chromosome, nil
		}
	}
	return chromosome, errors.New("could not find a path")
}

func (g *GeneticAlgorithm) removePathMutation(chromosome Chromosome) (Chromosome, error) {
	if len(chromosome.paths) == 0 {
		return chromosome, errors.New("no paths to remove")
	}
	chromosome.paths = removeRandomElement(chromosome.paths)
	return chromosome, nil
}

func (g *GeneticAlgorithm) moveMinesMutation(chromosome Chromosome) (Chromosome, error) {
	newChromosome := Chromosome{
		factories: chromosome.factories,
		paths:     chromosome.paths,
	}
	newChromosome.mines = removeUniform(chromosome.mines, g.mutationProbability)
	for i := len(newChromosome.mines); i < len(chromosome.mines); i++ {
		mine := chromosome.mines[i]
		// TODO: this might move the mine to a different deposit
		for _, deposit := range g.scenario.deposits {
			if deposit.Rectangle().Intersects(Rectangle{Position{mine.Ingress().x - 1, mine.Ingress().y - 1}, 3, 3}) {
				newMine, err := g.randomMine(deposit, newChromosome)
				if err == nil {
					newChromosome.mines = append(newChromosome.mines, newMine)
					break
				}
			}
		}
	}
	return newChromosome, nil
}

func (g *GeneticAlgorithm) moveFactoriesMutation(chromosome Chromosome) (Chromosome, error) {
	newChromosome := Chromosome{
		mines: chromosome.mines,
		paths: chromosome.paths,
	}
	newChromosome.factories = removeUniform(chromosome.factories, g.mutationProbability)
	for i := len(newChromosome.factories); i < len(chromosome.factories); i++ {
		factory, err := g.scenario.randomFactory(newChromosome)
		if err != nil {
			factory.product = chromosome.factories[i].product
			newChromosome.factories = append(newChromosome.factories, factory)
		}

	}
	return newChromosome, nil
}

func (g *GeneticAlgorithm) movePathMutation(chromosome Chromosome) (Chromosome, error) {
	newChromosome := Chromosome{
		mines:     chromosome.mines,
		factories: chromosome.factories,
	}
	if len(chromosome.factories) == 0 || len(chromosome.mines) == 0 {
		return newChromosome, errors.New("no factories or mines")
	}

	newChromosome.paths = removeUniform(chromosome.paths, g.mutationProbability)
	for i := len(newChromosome.paths); i < len(chromosome.paths); i++ {
		// TODO: maybe create path from previous factory to previous mine?
		for j := 0; j < NumPathRetries; j++ {
			randomFactory := chromosome.factories[rand.Intn(len(chromosome.factories))]
			randomMine := chromosome.mines[rand.Intn(len(chromosome.mines))]
			newPath, err := g.pathMineToFactory(newChromosome, randomMine, randomFactory)
			if err == nil {
				newChromosome.paths = append(newChromosome.paths, newPath)
				break
			}
		}
	}
	return newChromosome, nil
}

func (g *GeneticAlgorithm) evaluateFitness(chromosome Chromosome) int {
	fitness, err := g.scenario.evaluateSolution(chromosome.Solution())
	if err != nil {
		return -1
	}
	return fitness
}

func (g *GeneticAlgorithm) generateChromosomes() []Chromosome {
	chromosomes := make([]Chromosome, g.populationSize)
	for i := 0; i < g.populationSize; i++ {
		chromosomes[i] = Chromosome{mines: make([]Mine, 0), factories: make([]Factory, 0), paths: make([]Path, 0)}
	}
	return chromosomes
}

func (g *GeneticAlgorithm) Run() {
	chromosomes := g.generateChromosomes()
	for i, chromosome := range chromosomes {
		chromosomes[i].fitness = g.evaluateFitness(chromosome)
	}
	for i := 0; i < g.iterations; i++ {
		sort.Slice(chromosomes, func(i, j int) bool {
			return chromosomes[i].fitness > chromosomes[j].fitness
		})
		g.chromosomeChannel <- chromosomes[0]
		if g.optimum != NoOptimum && chromosomes[0].fitness == g.optimum {
			log.Println("starting iteration", i+1, "/", g.iterations, "fitness", g.optimum, "(optimal)")
			break
		}
		chromosomes = chromosomes[:g.populationSize]
		log.Println("starting iteration", i+1, "/", g.iterations, "max fitness", chromosomes[0].fitness, "min fitness", chromosomes[len(chromosomes)-1].fitness)

		for j := 0; j < g.populationSize; j++ {
			newChromosome, err := g.crossover(chromosomes[rand.Intn(g.populationSize)], chromosomes[rand.Intn(g.populationSize)])
			if err != nil {
				continue
			}
			newChromosome.fitness = g.evaluateFitness(newChromosome)
			chromosomes = append(chromosomes, newChromosome)
		}

		numChromosomesBeforeMutation := len(chromosomes)
		for j := 0; j < NumRoundsPerIteration; j++ {
			chromosome := chromosomes[rand.Intn(numChromosomesBeforeMutation)]
			for k := 0; k < NumMutationsPerRound; k++ {
				mutation := Mutations[rand.Intn(len(Mutations))]
				newChromosome, err := mutation(g, chromosome.copy())
				if err == nil {
					chromosome = newChromosome
					chromosome.fitness = g.evaluateFitness(chromosome)
					chromosomes = append(chromosomes, chromosome)
				}
			}
		}
	}
	sort.Slice(chromosomes, func(i, j int) bool {
		return chromosomes[i].fitness > chromosomes[j].fitness
	})
	g.doneChannel <- true
}
