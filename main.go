package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

// TODO: is this enough when running in docker?
const PercentTimeUsed = 90
const MaxFinishingTime = 2

func main() {
	inputPtr := flag.String("input", "-", "Path to input scenario json")
	outputPtr := flag.String("output", "-", "Path to output scenario json")
	seedPtr := flag.Int64("seed", 0, "Seed for random number generator")
	cpuProfilePtr := flag.String("cpuprofile", "", "Path to output cpu profile")
	itersPtr := flag.Int("iters", 50, "Number of iterations to run. Use 0 for unlimited")
	logChromosomesDirPtr := flag.String("logdir", "", "Directory to log top chromosomes in each iteration")
	visualizeChromosomesDirPtr := flag.String("visualizedir", "", "Directory to visualize chromosomes in each iteration")
	endOnOptimalPtr := flag.Bool("endonoptimal", false, "End when optimal solution is found")
	exportPtr := flag.String("exporter", "scenario", "Export type, either \"scenario\" or \"solution\"")
	flag.Parse()
	if *inputPtr == "" || *outputPtr == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *cpuProfilePtr != "" {
		f, err := os.Create(*cpuProfilePtr)
		if err != nil {
			log.Fatal("could not create cpu profile file: ", err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Fatal("could not start cpu profiling: ", err)
		}
		defer pprof.StopCPUProfile()
	}
	var exporter Exporter
	if *exportPtr == "scenario" {
		exporter = ScenarioExporter{}
	} else if *exportPtr == "solution" {
		exporter = SolutionExporter{}
	} else {
		log.Fatal("unknown exporter ", *exportPtr)
	}
	scenario, _, err := ImportScenario(*inputPtr)
	if err != nil {
		log.Fatal("could not import scenario: ", err)
	}
	var seed int64
	if *seedPtr != 0 {
		seed = *seedPtr
	} else {
		seed = time.Now().UnixNano()
	}
	log.Println("using seed", seed)
	rand.Seed(seed)

	optimum, err := TheoreticalOptimum(scenario)
	if err != nil {
		log.Println("no theoretical optimum found")
	} else {
		log.Println("theoretical optimum", optimum)
	}

	chromosomeChannel := make(chan Chromosome)
	doneChannel := make(chan bool)

	geneticAlgorithm := GeneticAlgorithm{
		scenario:                  scenario,
		populationSize:            100,
		iterations:                *itersPtr,
		moveObjectProbability:     0.18,
		numMutatedChromosomes:     50,
		numMutationsPerChromosome: 20,
		numCrossovers:             10,
		optimum:                   optimum,
		chromosomeChannel:         chromosomeChannel,
		doneChannel:               doneChannel,
		logChromosomesDir:         *logChromosomesDirPtr,
		visualizeChromosomesDir:   *visualizeChromosomesDirPtr,
	}
	go geneticAlgorithm.Run()

	var timeChannel <-chan time.Time
	if scenario.time != 0 {
		buffer := time.Duration(scenario.time) * time.Second * (100 - PercentTimeUsed) / 100
		if buffer > MaxFinishingTime*time.Second {
			buffer = MaxFinishingTime * time.Second
		}
		deadline := time.Duration(scenario.time)*time.Second - buffer
		timeChannel = time.After(deadline)
	} else {
		timeChannel = make(<-chan time.Time)
	}
	var chromosome Chromosome

	done := false
	for !done {
		var newChromosome Chromosome
		select {
		case <-timeChannel:
			log.Println("terminating: time is up")
			done = true
		case <-doneChannel:
			log.Println("terminating: max iters reached")
			done = true
		case newChromosome = <-chromosomeChannel:
			if newChromosome.fitness > chromosome.fitness || (newChromosome.fitness == chromosome.fitness && newChromosome.neededTurns < chromosome.neededTurns) {
				chromosome = newChromosome
				if optimum != NoOptimum && chromosome.fitness == optimum && endOnOptimalPtr != nil && *endOnOptimalPtr {
					log.Println("terminating: optimal solution found")
					done = true
				}
			}
		}
	}
	log.Println("final fitness", chromosome.fitness, "turns", chromosome.neededTurns)

	err = chromosome.CopyWithoutDisconnectedMines().Export(scenario, exporter, *outputPtr)
	if err != nil {
		log.Fatal("could not export solution: ", err)
	}
}
