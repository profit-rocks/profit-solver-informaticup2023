package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

func main() {
	inputPtr := flag.String("input", "", "Path to input scenario json")
	outputPtr := flag.String("output", "", "Path to output scenario json")
	cpuProfilePtr := flag.String("cpuprofile", "", "Path to output cpu profile")
	flag.Parse()
	if *inputPtr == "" || *outputPtr == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *cpuProfilePtr != "" {
		f, err := os.Create(*cpuProfilePtr)
		if err != nil {
			log.Fatal(err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}
	scenario, err := importScenarioFromJson(*inputPtr)
	if err != nil {
		log.Fatal(err)
	}

	rand.Seed(time.Now().UnixNano())

	optimum, err := TheoreticalOptimum(scenario)
	if err != nil {
		log.Println("no theoretical optimum found")
	} else {
		log.Println("theoretical optimum", optimum)
	}

	geneticAlgorithm := GeneticAlgorithm{
		scenario:             scenario,
		populationSize:       200,
		iterations:           60,
		mutationProbability:  0.18,
		crossoverProbability: 0.7,
		numFactories:         4,
		numMines:             2 * len(scenario.deposits),
		optimum:              optimum,
	}
	solution, err := geneticAlgorithm.Run()

	if err != nil {
		log.Fatal(err)
	}
	err = exportSolution(scenario, solution, *outputPtr)
	if err != nil {
		log.Fatal(err)
	}
}
