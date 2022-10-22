package main

import "fmt"

type Deposit struct {
	xPosition int
	yPosition int
	width     int
	height    int
	resource  int
}

type Obstacle struct {
	xPosition int
	yPosition int
	height    int
	width     int
}

type Scenario struct {
	width     int
	height    int
	deposits  []Deposit
	obstacles []Obstacle
}

func getDefaultScenario() Scenario {
	deposits := make([]Deposit, 2)
	deposits[0] = Deposit{width: 5, height: 5, resource: 1, xPosition: 0, yPosition: 0}
	deposits[1] = Deposit{width: 5, height: 5, resource: 0, xPosition: 10, yPosition: 3}

	obstacles := make([]Obstacle, 1)
	obstacles[0] = Obstacle{
		xPosition: 8,
		yPosition: 10,
		height:    4,
		width:     4,
	}
	return Scenario{
		width:     40,
		height:    40,
		deposits:  deposits,
		obstacles: obstacles,
	}
}

func main() {

	var scenario Scenario = getDefaultScenario()
	fmt.Println(scenario)

}
