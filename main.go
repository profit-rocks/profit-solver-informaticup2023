package main

const FACTORY_WIDTH = 5
const FACTORY_HEIGHT = 5
const NUM_FACTORIES = 10

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
	scenario := importScenarioFromJson("exampleScenario.json")

	scenario = runGeneticAlgorithm(40, scenario, 200, 0.7)

	exportScenario(scenario, "exampleScenarioWithFactories.json")
}
