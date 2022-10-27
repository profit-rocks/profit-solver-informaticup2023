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

// Orientation is the relative position of the egress
type Orientation int

const (
	Right  Orientation = iota
	Bottom Orientation = iota
	Left   Orientation = iota
	Top    Orientation = iota
)

type Mine struct {
	position    Position
	orientation Orientation
}

type ConveyorLength int

const (
	Short ConveyorLength = iota
	Long  ConveyorLength = iota
)

type Conveyor struct {
	position    Position
	orientation Orientation
	length      ConveyorLength
}

func (c Conveyor) Subtype() int {
	return (int(c.length) << 2) | int(c.orientation)
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
