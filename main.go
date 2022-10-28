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

type Deposit struct {
	position Position
	width    int
	height   int
	subtype  int
}

type Obstacle = Rectangle

type Factory struct {
	position Position
	product  int
}

// Direction is the relative position of the egress
type Direction int

const (
	Right  Direction = iota
	Bottom Direction = iota
	Left   Direction = iota
	Top    Direction = iota
)

type Mine struct {
	position    Position
	orientation Direction
}

type ConveyorLength int

const (
	Short ConveyorLength = iota
	Long  ConveyorLength = iota
)

type Conveyor struct {
	position  Position
	direction Direction
	length    ConveyorLength
}

func (c Conveyor) Subtype() int {
	return (int(c.length) << 2) | int(c.direction)
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
	conveyors []Conveyor
}

func (m Mine) Egress() Position {
	if m.orientation == Right {
		return Position{m.position.x + 2, m.position.y + 1}
	} else if m.orientation == Bottom {
		return Position{m.position.x, m.position.y + 2}
	} else if m.orientation == Left {
		return Position{m.position.x - 1, m.position.y}
	}
	// Top
	return Position{m.position.x + 1, m.position.y - 1}
}

func (m Mine) Ingress() Position {
	if m.orientation == Right {
		return Position{m.position.x - 1, m.position.y + 1}
	} else if m.orientation == Bottom {
		return Position{m.position.x, m.position.y - 1}
	} else if m.orientation == Left {
		return Position{m.position.x + 2, m.position.y}
	}
	// Top
	return Position{m.position.x + 1, m.position.y + 2}
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
