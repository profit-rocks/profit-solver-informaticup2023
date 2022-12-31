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
	factories []Factory
	mines     []Mine
	combiners []Combiner
	paths     []Path
	fitness   int
}

// MutationFunction expects a copy of the chromosome which it can modify.
type MutationFunction func(algorithm *GeneticAlgorithm, chromosome Chromosome) (Chromosome, error)

// Mutations contains all mutation functions, performed in multiple layers. Each layer operates on the same set of chromosomes
var MutationsWithoutPaths = []MutationFunction{
	(*GeneticAlgorithm).addMineMutation,
	(*GeneticAlgorithm).removeMineMutation,
	(*GeneticAlgorithm).moveMinesMutation,
	(*GeneticAlgorithm).addFactoryMutation,
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
	crossoverProbability    float64
	initialMinNumFactories  int
	initialMaxNumFactories  int
	initialNumMines         int
	optimum                 int
	numPaths                int
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

func (c Chromosome) Solution() Solution {
	solution := Solution{
		factories: make([]Factory, len(c.factories)),
		mines:     make([]Mine, 0, len(c.mines)),
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
	for _, mine := range c.mines {
		if mine.connectedFactory != nil {
			solution.mines = append(solution.mines, Mine{
				position:         mine.position,
				direction:        mine.direction,
				connectedFactory: mine.connectedFactory,
				distance:         mine.distance,
			})
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
	path.connectedFactory = p.connectedFactory
	return path
}

func minInt(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

func (c Chromosome) copy() Chromosome {
	newChromosome := Chromosome{
		fitness:   0,
		mines:     make([]Mine, len(c.mines)),
		factories: make([]Factory, len(c.factories)),
		combiners: make([]Combiner, len(c.combiners)),
		paths:     make([]Path, len(c.paths)),
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

func (chromosome *Chromosome) getPathEndPositionsForProduct(product int) []PathEndPosition {
	endPositions := make([]PathEndPosition, 0)
	for i, factory := range chromosome.factories {
		if factory.product == product {
			for _, pos := range factory.NextToIngressPositions() {
				endPositions = append(endPositions, PathEndPosition{pos, &chromosome.factories[i], factory.distance})
			}
		}
	}
	for _, combiner := range chromosome.combiners {
		if combiner.connectedFactory != nil && combiner.connectedFactory.product == product {
			for _, pos := range combiner.NextToIngressPositions() {
				endPositions = append(endPositions, PathEndPosition{pos, combiner.connectedFactory, combiner.distance})
			}
		}
	}
	for _, mine := range chromosome.mines {
		if mine.connectedFactory != nil && mine.connectedDeposit.subtype == product {
			for _, pos := range mine.NextToIngressPositions() {
				endPositions = append(endPositions, PathEndPosition{pos, mine.connectedFactory, mine.distance})
			}
		}
	}
	for _, path := range chromosome.paths {
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

func (g *GeneticAlgorithm) addPathMineToFactoryMutation(chromosome Chromosome) (Chromosome, error) {
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
			newPath, distance, err := g.path(chromosome, startPosition, endPositions)
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
		newPath, distance, err := g.path(chromosome, startPosition, endPositions)
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
		chromosomes[i] = Chromosome{mines: make([]Mine, 0), factories: make([]Factory, 0), paths: make([]Path, 0), combiners: make([]Combiner, 0)}
	}
	return chromosomes
}

func (g *GeneticAlgorithm) Run() {
	chromosomes := g.generateChromosomes()
	g.initializeCellInfoWithScenario()
	for i, chromosome := range chromosomes {
		chromosomes[i].fitness = g.evaluateFitness(chromosome)
	}
	for i := 0; g.iterations == 0 || i < g.iterations; i++ {
		sort.Slice(chromosomes, func(i, j int) bool {
			return chromosomes[i].fitness > chromosomes[j].fitness
		})
		if g.logChromosomesDir != "" {
			err := exportChromosomes(g.scenario, i, chromosomes, g.logChromosomesDir)
			if err != nil {
				log.Fatal("could not export chromosomes: ", err)
			}
		}
		chromosomes = chromosomes[:g.populationSize]
		log.Println("starting iteration", i+1, "/", g.iterations, "max fitness", chromosomes[0].fitness, "min fitness", chromosomes[len(chromosomes)-1].fitness)

		for j := 0; j < NumRoundsPerIteration; j++ {
			chromosome := chromosomes[rand.Intn(g.populationSize)]
			chromosomeWithoutPath := chromosome.copy()
			// reset paths and mines
			chromosomeWithoutPath.paths = make([]Path, 0)
			for x := range chromosomeWithoutPath.mines {
				chromosomeWithoutPath.mines[x].connectedFactory = nil
			}

			for k := 0; k < NumMutationsPerRound; k++ {
				rng := NewUniqueRNG(len(MutationsWithoutPaths))
				done := false
				var mutationIndex int
				for !done {
					mutationIndex, done = rng.Next()
					mutation := MutationsWithoutPaths[mutationIndex]
					newChromosome, err := mutation(g, chromosomeWithoutPath.copy())
					if err == nil {
						chromosomeWithoutPath = newChromosome
						chromosomeWithPaths := newChromosome.copy()
						for _, comb := range chromosomeWithPaths.combiners {
							chromosomeWithPaths, _ = g.addPathCombinerToFactory(chromosomeWithPaths, comb)
						}
						for _, comb := range chromosomeWithPaths.combiners {
							chromosomeWithPaths, _ = g.addPathCombinerToFactory(chromosomeWithPaths, comb)
						}
						// Before building paths, we have to update the cell Info
						g.populateCellInfoWithNewChromosome(chromosomeWithPaths)
						for m := 0; m < len(chromosomeWithPaths.mines); m++ {
							newChromosomeWithPaths, err2 := g.addPathMineToFactoryMutation(chromosomeWithPaths)
							if err2 == nil {
								newChromosomeWithPaths.fitness = g.evaluateFitness(newChromosomeWithPaths)
								// if the new chromosome is invalid, it won't get valid by building more paths
								if newChromosomeWithPaths.fitness == -1 {
									break
								}
								chromosomeWithPaths = newChromosomeWithPaths.copy()
								chromosomes = append(chromosomes, newChromosomeWithPaths)
								g.chromosomeChannel <- newChromosomeWithPaths
							}
						}
						break
					}
				}
				if done {
					log.Println("all mutations without paths failed, trying different chromosome")
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
