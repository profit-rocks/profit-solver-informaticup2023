package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

const FactoryWidth = 5
const FactoryHeight = 5

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

type Factory struct {
	position Position
	product  int
}

type Mine struct {
	position    Position
	orientation int
}

// Scenario is the input to any algorithm that solves Profit!
type Scenario struct {
	width     int
	height    int
	deposits  []Deposit
	obstacles []Obstacle
	turns     int
}

// Solution is the output of any algorithm that solves Profit!
type Solution struct {
	factories []Factory
	mines     []Mine
}

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
	scenario := importScenarioFromJson(*inputPtr)

	rand.Seed(time.Now().UnixNano())
	solution, err := runGeneticAlgorithm(200, scenario, 120, 0.18, 0.7, 4)

	if err != nil {
		log.Fatal(err)
	}
	err = exportSolution(scenario, solution, *outputPtr)
	if err != nil {
		log.Fatal(err)
	}
}
