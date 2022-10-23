package main

import (
	"flag"
	"fmt"
	"os"
)

const FactoryWidth = 5
const FactoryHeight = 5
const NumFactories = 4

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

func main() {
	inputPtr := flag.String("input", "", "Path to input scenario json")
	outputPtr := flag.String("output", "", "Path to output scenario json")
	flag.Parse()
	if *inputPtr == "" || *outputPtr == "" {
		flag.Usage()
		os.Exit(1)
	}
	scenario := importScenarioFromJson(*inputPtr)
	scenario, err := runGeneticAlgorithm(40, scenario, 200, 0.7)

	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
	exportScenario(scenario, *outputPtr)
}
