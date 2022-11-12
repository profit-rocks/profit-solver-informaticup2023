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

	optimum, err := TheoreticalOptimumNoProducts(scenario)
	if err != nil {
		log.Println("no theoretical optimum found")
	} else {
		log.Println("theoretical optimum", optimum)
	}

	geneticAlgorithm := GeneticAlgorithm{
		scenario:             scenario,
		populationSize:       200,
		iterations:           120,
		mutationProbability:  0.18,
		crossoverProbability: 0.7,
		optimum:              optimum,
	}
	solution := geneticAlgorithm.Run()

	err = exportSolution(scenario, solution, *outputPtr)
	if err != nil {
		log.Fatal(err)
	}
}
