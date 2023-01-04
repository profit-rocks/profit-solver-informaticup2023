package main

import (
	"errors"
	"log"
	"math/rand"
	"sort"
)

type Path struct {
	conveyors        []Conveyor
	connectedFactory *Factory
}

const NumRoundsPerIteration = 50
const NumMutationsPerRound = 20

type Chromosome struct {
	factories   []Factory
	mines       []Mine
	combiners   []Combiner
	paths       []Path
	fitness     int
	neededTurns int
}

// MutationFunction expects a copy of the chromosome which it can modify.
type MutationFunction func(algorithm *GeneticAlgorithm, chromosome Chromosome) (Chromosome, error)

// Mutations contains all mutation functions, performed in multiple layers. Each layer operates on the same set of chromosomes
var Mutations = []MutationFunction{
	(*GeneticAlgorithm).addMineMutation,
	(*GeneticAlgorithm).removeMineMutation,
	(*GeneticAlgorithm).moveMinesMutation,
	(*GeneticAlgorithm).addFactoryMutation,
	(*GeneticAlgorithm).changeProduct,
	(*GeneticAlgorithm).removeFactoryMutation,
	(*GeneticAlgorithm).moveFactoriesMutation,
	(*GeneticAlgorithm).addCombinerMutation,
	(*GeneticAlgorithm).removeCombinerMutation,
	(*GeneticAlgorithm).moveCombinersMutation,
}

// GeneticAlgorithm contains input data as well as configuration information used by the genetic algorithm.
// Data in this struct is passed around, but never changed. If there is context information that needs to be
// changed, it should probably be stored in a chromosome.
type GeneticAlgorithm struct {
	scenario                Scenario
	iterations              int
	populationSize          int
	mutationProbability     float64
	optimum                 int
	chromosomeChannel       chan<- Chromosome
	doneChannel             chan<- bool
	logChromosomesDir       string
	visualizeChromosomesDir string
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

func (p Path) copy() Path {
	path := Path{}
	for _, c := range p.conveyors {
		path.conveyors = append(path.conveyors, c)
	}
	path.connectedFactory = p.connectedFactory
	return path
}

func minInt(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

func (c Chromosome) Copy() Chromosome {
	newChromosome := Chromosome{
		fitness:     c.fitness,
		neededTurns: c.neededTurns,
		mines:       make([]Mine, len(c.mines)),
		factories:   make([]Factory, len(c.factories)),
		combiners:   make([]Combiner, len(c.combiners)),
		paths:       make([]Path, len(c.paths)),
	}
	for k, factory := range c.factories {
		newChromosome.factories[k] = factory
	}
	for k, mine := range c.mines {
		newChromosome.mines[k] = mine
	}
	for k, path := range c.paths {
		newChromosome.paths[k] = path.copy()
	}
	for k, combiner := range c.combiners {
		newChromosome.combiners[k] = combiner
	}
	return newChromosome
}

func (c Chromosome) CleanCopy() Chromosome {
	newChromosome := Chromosome{
		fitness:     c.fitness,
		neededTurns: c.neededTurns,
		mines:       make([]Mine, 0, len(c.mines)),
		factories:   make([]Factory, len(c.factories)),
		combiners:   make([]Combiner, len(c.combiners)),
		paths:       make([]Path, len(c.paths)),
	}
	for k, factory := range c.factories {
		newChromosome.factories[k] = factory
	}
	for _, mine := range c.mines {
		if mine.connectedFactory != nil {
			newChromosome.mines = append(newChromosome.mines, mine)
		}
	}
	for k, path := range c.paths {
		newChromosome.paths[k] = path.copy()
	}
	for k, combiner := range c.combiners {
		newChromosome.combiners[k] = combiner
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

func (g *GeneticAlgorithm) changeProduct(chromosome Chromosome) (Chromosome, error) {
	if len(chromosome.factories) == 0 {
		return chromosome, errors.New("no factories to change product")
	}
	factory := &chromosome.factories[rand.Intn(len(chromosome.factories))]
	factory.product = rand.Intn(len(g.scenario.products))
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
	if err != nil {
		return Chromosome{}, err
	}
	return chromosome, nil
}

func (g *GeneticAlgorithm) removeCombinerMutation(chromosome Chromosome) (Chromosome, error) {
	if len(chromosome.combiners) == 0 {
		return chromosome, errors.New("no combiners to remove")
	}
	chromosome.combiners = removeRandomElement(chromosome.combiners)
	return chromosome, nil
}

func (g *GeneticAlgorithm) moveCombinersMutation(chromosome Chromosome) (Chromosome, error) {
	newChromosome := Chromosome{
		factories: chromosome.factories,
		paths:     chromosome.paths,
		mines:     chromosome.mines,
	}
	newChromosome.combiners = removeUniform(chromosome.combiners, g.mutationProbability)
	for i := len(newChromosome.combiners); i < len(chromosome.combiners); i++ {
		combiner, err := g.scenario.randomCombiner(newChromosome)
		if err == nil {
			newChromosome.combiners = append(newChromosome.combiners, combiner)
		}
	}
	return newChromosome, nil
}

func (g *GeneticAlgorithm) addMineMutation(chromosome Chromosome) (Chromosome, error) {
	randomDeposit := &g.scenario.deposits[rand.Intn(len(g.scenario.deposits))]
	newMine, err := g.randomMine(randomDeposit, chromosome)
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

func (c *Chromosome) getPathEndPositionsForProduct(product int) []PathEndPosition {
	endPositions := make([]PathEndPosition, 0)
	for i, factory := range c.factories {
		if factory.product == product {
			for _, pos := range factory.NextToIngressPositions() {
				endPositions = append(endPositions, PathEndPosition{pos, &c.factories[i], factory.distance})
			}
		}
	}
	for _, combiner := range c.combiners {
		if combiner.connectedFactory != nil && combiner.connectedFactory.product == product {
			for _, pos := range combiner.NextToIngressPositions() {
				endPositions = append(endPositions, PathEndPosition{pos, combiner.connectedFactory, combiner.distance})
			}
		}
	}
	for _, mine := range c.mines {
		if mine.connectedFactory != nil && mine.connectedDeposit.subtype == product {
			for _, pos := range mine.NextToIngressPositions() {
				endPositions = append(endPositions, PathEndPosition{pos, mine.connectedFactory, mine.distance})
			}
		}
	}
	for _, path := range c.paths {
		if path.connectedFactory.product == product {
			for _, conveyor := range path.conveyors {
				for _, pos := range conveyor.NextToIngressPositions() {
					endPositions = append(endPositions, PathEndPosition{pos, path.connectedFactory, conveyor.distance})
				}
			}
		}
	}
	return endPositions
}

func (g *GeneticAlgorithm) addPathMineToFactory(chromosome Chromosome) (Chromosome, error) {
	if len(chromosome.mines) == 0 || len(chromosome.factories) == 0 {
		return chromosome, errors.New("no mines or factories to add path")
	}
	mineRng := NewUniqueRNG(len(chromosome.mines))
	mineDone := false
	var mineIndex int
	for !mineDone {
		mineIndex, mineDone = mineRng.Next()
		randomMine := chromosome.mines[mineIndex]
		if randomMine.connectedFactory != nil {
			continue
		}
		viableProducts := make([]Product, 0)
		for _, product := range g.scenario.products {
			if product.resources[randomMine.connectedDeposit.subtype] > 0 {
				for _, factory := range chromosome.factories {
					if factory.product == product.subtype {
						viableProducts = append(viableProducts, product)
						break
					}
				}
			}
		}
		if len(viableProducts) == 0 {
			continue
		}
		rng := NewUniqueRNG(len(viableProducts))
		done := false
		var index int
		for !done {
			index, done = rng.Next()
			startPosition := randomMine.Egress()
			randomProduct := viableProducts[index].subtype
			endPositions := chromosome.getPathEndPositionsForProduct(randomProduct)
			newPath, distance, err := findPath(startPosition, endPositions, g.scenario)
			if err == nil {
				chromosome.mines[mineIndex].connectedFactory = newPath.connectedFactory
				chromosome.mines[mineIndex].distance = distance + 1
				chromosome.paths = append(chromosome.paths, newPath)
				return chromosome, nil
			}
		}
	}
	return chromosome, errors.New("could not find a path")
}

func (g *GeneticAlgorithm) addPathCombinerToFactory(chromosome Chromosome, combiner Combiner) (Chromosome, error) {
	if len(chromosome.factories) == 0 {
		return chromosome, errors.New("no factories to add path")
	}
	rng := NewUniqueRNG(len(g.scenario.products))
	done := false
	var index int
	for !done {
		index, done = rng.Next()
		randomProduct := g.scenario.products[index].subtype
		endPositions := chromosome.getPathEndPositionsForProduct(randomProduct)
		startPosition := combiner.Egress()
		newPath, distance, err := findPath(startPosition, endPositions, g.scenario)
		if err == nil {
			combiner.connectedFactory = newPath.connectedFactory
			combiner.distance = distance + 1
			chromosome.paths = append(chromosome.paths, newPath)
			return chromosome, nil
		}
	}
	return chromosome, errors.New("could not find a path")
}

func (g *GeneticAlgorithm) moveMinesMutation(chromosome Chromosome) (Chromosome, error) {
	newChromosome := Chromosome{
		factories: chromosome.factories,
		paths:     chromosome.paths,
		combiners: chromosome.combiners,
	}
	newChromosome.mines = removeUniform(chromosome.mines, g.mutationProbability)
	for i := len(newChromosome.mines); i < len(chromosome.mines); i++ {
		mine := chromosome.mines[i]
		// TODO: this might move the mine to a different deposit
		for i, deposit := range g.scenario.deposits {
			if deposit.Rectangle().Intersects(Rectangle{Position{mine.Ingress().x - 1, mine.Ingress().y - 1}, 3, 3, nil}) {
				newMine, err := g.randomMine(&g.scenario.deposits[i], newChromosome)
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
		mines:     chromosome.mines,
		paths:     chromosome.paths,
		combiners: chromosome.combiners,
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

func (g *GeneticAlgorithm) generateChromosomes() []Chromosome {
	chromosomes := make([]Chromosome, g.populationSize)
	for i := 0; i < g.populationSize; i++ {
		chromosomes[i] = Chromosome{mines: make([]Mine, 0), factories: make([]Factory, 0), paths: make([]Path, 0), combiners: make([]Combiner, 0), fitness: 0, neededTurns: g.scenario.turns}
	}
	return chromosomes
}

func (g *GeneticAlgorithm) Run() {
	chromosomes := g.generateChromosomes()
	initializeCellInfo(g.scenario)
	for i := 0; g.iterations == 0 || i < g.iterations; i++ {
		sort.Slice(chromosomes, func(i, j int) bool {
			if chromosomes[i].fitness == chromosomes[j].fitness {
				return chromosomes[i].neededTurns < chromosomes[j].neededTurns
			}
			return chromosomes[i].fitness > chromosomes[j].fitness
		})
		if g.logChromosomesDir != "" {
			err := exportChromosomes(g.scenario, i, chromosomes, g.logChromosomesDir)
			if err != nil {
				log.Fatal("could not export chromosomes: ", err)
			}
		}
		chromosomes = chromosomes[:g.populationSize]
		log.Println("starting iteration", i+1, "/", g.iterations, "max fitness", chromosomes[0].fitness, "turns", chromosomes[0].neededTurns, "min fitness", chromosomes[len(chromosomes)-1].fitness, "turns", chromosomes[len(chromosomes)-1].neededTurns)

		for j := 0; j < NumRoundsPerIteration; j++ {
			chromosome := chromosomes[rand.Intn(g.populationSize)].Copy()
			chromosome.resetPaths()

			for k := 0; k < NumMutationsPerRound; k++ {
				rng := NewUniqueRNG(len(Mutations))
				done := false
				var mutationIndex int
				for !done {
					mutationIndex, done = rng.Next()
					mutation := Mutations[mutationIndex]
					newChromosome, err := mutation(g, chromosome.Copy())
					if err == nil {
						chromosome = newChromosome
						for _, c := range g.chromosomesWithPaths(newChromosome.Copy()) {
							chromosomes = append(chromosomes, c)
							g.chromosomeChannel <- c
						}
						break
					}
				}
				if done {
					log.Println("all mutations failed, trying different chromosome")
				}
			}
		}
		if g.visualizeChromosomesDir != "" {
			err := g.visualizeChromosomes(chromosomes, i, g.visualizeChromosomesDir)
			if err != nil {
				log.Fatal("could not visualize chromosomes: ", err)
			}
		}
	}
	g.doneChannel <- true
}

func (c *Chromosome) resetPaths() {
	c.paths = make([]Path, 0)
	for x := range c.mines {
		c.mines[x].connectedFactory = nil
	}
}

func (g *GeneticAlgorithm) chromosomesWithPaths(chromosome Chromosome) []Chromosome {
	var chromosomes []Chromosome
	// fitness is always 0 for chromosomes without mines or factories
	if len(chromosome.factories) == 0 || len(chromosome.mines) == 0 {
		return chromosomes
	}
	// Before building paths, we have to update the cell Info
	populateCellInfoWithNewChromosome(chromosome, g.scenario)
	for _, comb := range chromosome.combiners {
		chromosome, _ = g.addPathCombinerToFactory(chromosome, comb)
	}
	for m := 0; m < len(chromosome.mines); m++ {
		newChromosome, err := g.addPathMineToFactory(chromosome)
		if err == nil {
			newChromosome.fitness, newChromosome.neededTurns, err = g.scenario.evaluateChromosome(newChromosome)
			if err != nil {
				break
			}
			chromosome = newChromosome.Copy()
			chromosomes = append(chromosomes, newChromosome)
		}
	}
	return chromosomes
}
