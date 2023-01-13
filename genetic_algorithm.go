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
	(*GeneticAlgorithm).AddMine,
	(*GeneticAlgorithm).RemoveMine,
	(*GeneticAlgorithm).MoveMines,
	(*GeneticAlgorithm).AddFactory,
	(*GeneticAlgorithm).ChangeProduct,
	(*GeneticAlgorithm).RemoveFactory,
	(*GeneticAlgorithm).MoveFactories,
	(*GeneticAlgorithm).AddCombiner,
	(*GeneticAlgorithm).RemoveCombiner,
	(*GeneticAlgorithm).MoveCombiners,
}

// GeneticAlgorithm contains input data as well as configuration information used by the genetic algorithm.
// Data in this struct is passed around, but never changed. If there is context information that needs to be
// changed, it should probably be stored in a chromosome.
type GeneticAlgorithm struct {
	scenario                  Scenario
	iterations                int
	populationSize            int
	numMutatedChromosomes     int
	numMutationsPerChromosome int
	numCrossovers             int
	moveObjectProbability     float64
	optimum                   int
	chromosomeChannel         chan<- Chromosome
	doneChannel               chan<- bool
	logChromosomesDir         string
	visualizeChromosomesDir   string
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

func (p Path) Copy() Path {
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

func maxInt(x int, y int) int {
	if x > y {
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
		newChromosome.paths[k] = path.Copy()
	}
	for k, combiner := range c.combiners {
		newChromosome.combiners[k] = combiner
	}
	return newChromosome
}

func (c Chromosome) CopyWithoutDisconnectedMines() Chromosome {
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
		newChromosome.paths[k] = path.Copy()
	}
	for k, combiner := range c.combiners {
		newChromosome.combiners[k] = combiner
	}
	return newChromosome
}

func (g *GeneticAlgorithm) AddFactory(chromosome Chromosome) (Chromosome, error) {
	newFactory, err := g.scenario.RandomFactory(chromosome)
	if err != nil {
		return Chromosome{}, err
	}
	chromosome.factories = append(chromosome.factories, newFactory)
	return chromosome, nil
}

func (g *GeneticAlgorithm) ChangeProduct(chromosome Chromosome) (Chromosome, error) {
	if len(chromosome.factories) == 0 {
		return chromosome, errors.New("no factories to change product")
	}
	factory := &chromosome.factories[rand.Intn(len(chromosome.factories))]
	factory.product = rand.Intn(len(g.scenario.products))
	return chromosome, nil
}

func (g *GeneticAlgorithm) RemoveFactory(chromosome Chromosome) (Chromosome, error) {
	if len(chromosome.factories) == 0 {
		return chromosome, errors.New("no factories to remove")
	}
	chromosome.factories = removeRandomElement(chromosome.factories)
	return chromosome, nil
}

func (g *GeneticAlgorithm) AddCombiner(chromosome Chromosome) (Chromosome, error) {
	newCombiner, err := g.scenario.RandomCombiner(chromosome)
	if err != nil {
		return Chromosome{}, err
	}
	chromosome.combiners = append(chromosome.combiners, newCombiner)
	if err != nil {
		return Chromosome{}, err
	}
	return chromosome, nil
}

func (g *GeneticAlgorithm) RemoveCombiner(chromosome Chromosome) (Chromosome, error) {
	if len(chromosome.combiners) == 0 {
		return chromosome, errors.New("no combiners to remove")
	}
	chromosome.combiners = removeRandomElement(chromosome.combiners)
	return chromosome, nil
}

func (g *GeneticAlgorithm) MoveCombiners(chromosome Chromosome) (Chromosome, error) {
	newChromosome := Chromosome{
		factories: chromosome.factories,
		paths:     chromosome.paths,
		mines:     chromosome.mines,
	}
	newChromosome.combiners = removeUniform(chromosome.combiners, g.moveObjectProbability)
	for i := len(newChromosome.combiners); i < len(chromosome.combiners); i++ {
		combiner, err := g.scenario.RandomCombiner(newChromosome)
		if err == nil {
			newChromosome.combiners = append(newChromosome.combiners, combiner)
		}
	}
	return newChromosome, nil
}

func (g *GeneticAlgorithm) AddMine(chromosome Chromosome) (Chromosome, error) {
	randomDeposit := &g.scenario.deposits[rand.Intn(len(g.scenario.deposits))]
	newMine, err := g.scenario.RandomMine(randomDeposit, chromosome)
	if err != nil {
		return chromosome, err
	}
	chromosome.mines = append(chromosome.mines, newMine)
	return chromosome, nil
}

func (g *GeneticAlgorithm) RemoveMine(chromosome Chromosome) (Chromosome, error) {
	if len(chromosome.mines) == 0 {
		return chromosome, errors.New("no mines to remove")
	}
	chromosome.mines = removeRandomElement(chromosome.mines)
	return chromosome, nil
}

func (c *Chromosome) getPathEndPositionsForProduct(product int) []PathEndPosition {
	length := 0
	for _, factory := range c.factories {
		if factory.product == product {
			length += 20
		}
	}
	for _, combiner := range c.combiners {
		if combiner.connectedFactory != nil && combiner.connectedFactory.product == product {
			length += 5
		}
	}
	for _, mine := range c.mines {
		if mine.connectedFactory != nil && mine.connectedDeposit.subtype == product {
			length += 3
		}
	}
	for _, path := range c.paths {
		if path.connectedFactory.product == product {
			length += len(path.conveyors) * 3
		}
	}
	endPositions := make([]PathEndPosition, 0, length)
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

func (g *GeneticAlgorithm) AddPathMineToFactory(chromosome Chromosome) (Chromosome, error) {
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

func (g *GeneticAlgorithm) AddPathCombinerToFactory(chromosome Chromosome, combiner Combiner) (Chromosome, error) {
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

func (g *GeneticAlgorithm) MoveMines(chromosome Chromosome) (Chromosome, error) {
	newChromosome := Chromosome{
		factories: chromosome.factories,
		paths:     chromosome.paths,
		combiners: chromosome.combiners,
	}
	newChromosome.mines = removeUniform(chromosome.mines, g.moveObjectProbability)
	for i := len(newChromosome.mines); i < len(chromosome.mines); i++ {
		mine := chromosome.mines[i]
		// TODO: this might move the mine to a different deposit
		for i, deposit := range g.scenario.deposits {
			if deposit.Rectangle().Intersects(Rectangle{Position{mine.Ingress().x - 1, mine.Ingress().y - 1}, 3, 3, nil}) {
				newMine, err := g.scenario.RandomMine(&g.scenario.deposits[i], newChromosome)
				if err == nil {
					newChromosome.mines = append(newChromosome.mines, newMine)
					break
				}
			}
		}
	}
	return newChromosome, nil
}

func (g *GeneticAlgorithm) MoveFactories(chromosome Chromosome) (Chromosome, error) {
	newChromosome := Chromosome{
		mines:     chromosome.mines,
		paths:     chromosome.paths,
		combiners: chromosome.combiners,
	}
	newChromosome.factories = removeUniform(chromosome.factories, g.moveObjectProbability)
	for i := len(newChromosome.factories); i < len(chromosome.factories); i++ {
		factory, err := g.scenario.RandomFactory(newChromosome)
		if err != nil {
			factory.product = chromosome.factories[i].product
			newChromosome.factories = append(newChromosome.factories, factory)
		}

	}
	return newChromosome, nil
}

func (g *GeneticAlgorithm) GenerateChromosomes() []Chromosome {
	chromosomes := make([]Chromosome, g.populationSize)
	for i := 0; i < g.populationSize; i++ {
		chromosomes[i] = Chromosome{mines: make([]Mine, 0), factories: make([]Factory, 0), paths: make([]Path, 0), combiners: make([]Combiner, 0), fitness: 0, neededTurns: g.scenario.turns}
	}
	return chromosomes
}

func (g *GeneticAlgorithm) Run() {
	chromosomes := g.GenerateChromosomes()
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
		if i > 0 {
			log.Println("iteration", i, "/", g.iterations, "max fitness", chromosomes[0].fitness, "turns", chromosomes[0].neededTurns, "min fitness", chromosomes[len(chromosomes)-1].fitness, "turns", chromosomes[len(chromosomes)-1].neededTurns)
		}
		bestCrossover := Chromosome{}
		for j := 0; j < g.numCrossovers; j++ {
			index1 := rand.Intn(g.populationSize)
			index2 := rand.Intn(g.populationSize)
			chromosome1 := chromosomes[minInt(index1, index2)]
			chromosome2 := chromosomes[maxInt(index1, index2)]
			newChromosome, err := g.Crossover(chromosome1, chromosome2)
			if err == nil {
				newChromosome.resetPaths()
				for _, c := range g.ChromosomesWithPaths(newChromosome.Copy()) {
					chromosomes = append(chromosomes, c)
					if c.fitness > bestCrossover.fitness || c.fitness == bestCrossover.fitness && c.neededTurns < bestCrossover.neededTurns {
						bestCrossover = c
					}
					g.chromosomeChannel <- c
				}
			}
		}
		log.Println("iteration", i+1, "/", g.iterations, "best crossover fitness", bestCrossover.fitness, "turns", bestCrossover.neededTurns)

		for j := 0; j < g.numMutatedChromosomes; j++ {
			chromosome := chromosomes[rand.Intn(len(chromosomes)-j)].Copy()
			chromosome.resetPaths()

			for k := 0; k < g.numMutationsPerChromosome; k++ {
				rng := NewUniqueRNG(len(Mutations))
				done := false
				var mutationIndex int
				for !done {
					mutationIndex, done = rng.Next()
					mutation := Mutations[mutationIndex]
					newChromosome, err := mutation(g, chromosome.Copy())
					if err == nil {
						chromosome = newChromosome
						for _, c := range g.ChromosomesWithPaths(newChromosome.Copy()) {
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
			err := g.VisualizeChromosomes(chromosomes, i, g.visualizeChromosomesDir)
			if err != nil {
				log.Fatal("could not visualize chromosomes: ", err)
			}
		}
	}
	g.doneChannel <- true
}

func (g *GeneticAlgorithm) Crossover(chromosome1 Chromosome, chromosome2 Chromosome) (Chromosome, error) {
	newChromosome := chromosome1.Copy()
	newChromosome.paths = make([]Path, 0)
	newChromosome.fitness = 0
	for _, factory := range chromosome2.factories {
		same := false
		for _, oldFactory := range newChromosome.factories {
			if oldFactory.Rectangle().Intersects(factory.Rectangle()) && oldFactory.product == factory.product {
				same = true
				break
			}
		}
		if same {
			continue
		}
		if g.scenario.PositionAvailableForFactory(newChromosome.factories, newChromosome.mines, newChromosome.combiners, newChromosome.paths, factory.position) {
			newChromosome.factories = append(newChromosome.factories, factory)
		} else {
			newFactory, err := g.scenario.RandomFactory(newChromosome)
			if err == nil {
				newFactory.product = factory.product
				newChromosome.factories = append(newChromosome.factories, newFactory)
			}
		}
	}
	for _, mine := range chromosome2.mines {
		if g.scenario.PositionAvailableForMine(newChromosome.factories, newChromosome.mines, newChromosome.combiners, newChromosome.paths, mine) {
			newChromosome.mines = append(newChromosome.mines, mine)
		} else {
			newMine, err := g.scenario.RandomMine(mine.connectedDeposit, newChromosome)
			if err == nil {
				newChromosome.mines = append(newChromosome.mines, newMine)
			}
		}
	}
	for _, combiner := range chromosome2.combiners {
		if g.scenario.PositionAvailableForCombiner(newChromosome.factories, newChromosome.mines, newChromosome.paths, newChromosome.combiners, combiner) {
			newChromosome.combiners = append(newChromosome.combiners, combiner)
		}
	}
	return newChromosome, nil
}

func (c *Chromosome) resetPaths() {
	c.paths = make([]Path, 0)
	for x := range c.mines {
		c.mines[x].connectedFactory = nil
	}
}

func (g *GeneticAlgorithm) ChromosomesWithPaths(chromosome Chromosome) []Chromosome {
	chromosomes := make([]Chromosome, 0, len(chromosome.mines))
	// fitness is always 0 for chromosomes without mines or factories
	if len(chromosome.factories) == 0 || len(chromosome.mines) == 0 {
		return chromosomes
	}
	// Before building paths, we have to update the cell Info
	populateCellInfoWithNewChromosome(chromosome, g.scenario)
	for _, comb := range chromosome.combiners {
		chromosome, _ = g.AddPathCombinerToFactory(chromosome, comb)
	}
	for m := 0; m < len(chromosome.mines); m++ {
		newChromosome, err := g.AddPathMineToFactory(chromosome)
		if err == nil {
			newChromosome.fitness, newChromosome.neededTurns, err = g.scenario.EvaluateChromosome(newChromosome.CopyWithoutDisconnectedMines())
			if err != nil {
				break
			}
			chromosome = newChromosome.Copy()
			chromosomes = append(chromosomes, newChromosome)
		}
	}
	return chromosomes
}
