package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type Profit struct {
	Height   int      `json:"height"`
	Width    int      `json:"width"`
	Objects  []Object `json:"objects"`
	Products []Object `json:"products"`
	Turns    int      `json:"turns"`
	Time     int      `json:"time"`
}

type Object struct {
	ObjectType string `json:"type"`
	Subtype    int    `json:"subtype"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Resources  []int  `json:"resources"`
	Points     int    `json:"points"`
}

func exportSolution(scenario Scenario, solution Solution, filePath string) error {
	exportableScenario := solutionToProfit(scenario, solution)
	b, err := json.MarshalIndent(exportableScenario, "", " ")
	if err != nil {
		return err
	}
	if filePath == "-" {
		_, err = os.Stdout.Write(b)
		return err
	}
	return os.WriteFile(filePath, b, 0644)
}

func solutionToProfit(scenario Scenario, solution Solution) Profit {
	profit := Profit{
		Height:   scenario.height,
		Width:    scenario.width,
		Objects:  []Object{},
		Products: []Object{},
		Turns:    scenario.turns,
		Time:     100,
	}

	for _, deposit := range scenario.deposits {
		profit.Objects = append(profit.Objects, Object{
			ObjectType: "deposit",
			Subtype:    deposit.subtype,
			X:          deposit.position.x,
			Y:          deposit.position.y,
			Width:      deposit.width,
			Height:     deposit.height,
		})
	}
	for _, obstacle := range scenario.obstacles {
		profit.Objects = append(profit.Objects, Object{
			ObjectType: "obstacle",
			X:          obstacle.position.x,
			Y:          obstacle.position.y,
			Width:      obstacle.width,
			Height:     obstacle.height,
		})
	}

	for _, factory := range solution.factories {
		profit.Objects = append(profit.Objects, Object{
			ObjectType: "factory",
			Subtype:    factory.product,
			X:          factory.position.x,
			Y:          factory.position.y,
		})
	}

	for _, mine := range solution.mines {
		profit.Objects = append(profit.Objects, Object{
			ObjectType: "mine",
			Subtype:    int(mine.direction),
			X:          mine.position.x,
			Y:          mine.position.y,
		})
	}

	for _, path := range solution.paths {
		for _, conveyor := range path.conveyors {
			profit.Objects = append(profit.Objects, Object{
				ObjectType: "conveyor",
				Subtype:    conveyor.Subtype(),
				X:          conveyor.position.x,
				Y:          conveyor.position.y,
			})
		}
	}

	for _, combiner := range solution.combiners {
		profit.Objects = append(profit.Objects, Object{
			ObjectType: "combiner",
			X:          combiner.position.x,
			Y:          combiner.position.y,
			Subtype:    int(combiner.direction),
		})
	}

	for _, product := range scenario.products {
		profit.Products = append(profit.Products, Object{
			ObjectType: "product",
			Subtype:    product.subtype,
			Points:     product.points,
			Resources:  product.resources,
		})
	}
	return profit
}

func importFromProfitJson(path string) (Scenario, Solution, error) {
	var jsonFile *os.File
	var err error
	if path == "-" {
		jsonFile = os.Stdin
	} else {
		jsonFile, err = os.Open(path)
		if err != nil {
			return Scenario{}, Solution{}, err
		}
		defer jsonFile.Close()
	}
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return Scenario{}, Solution{}, err
	}
	var profit Profit
	err = json.Unmarshal(byteValue, &profit)
	if err != nil {
		return Scenario{}, Solution{}, err
	}

	scenario := Scenario{
		width:  profit.Width,
		height: profit.Height,
		turns:  profit.Turns,
		time:   profit.Time,
	}
	if scenario.time <= 0 {
		return Scenario{}, Solution{}, errors.New("time imported from json has to be greater than 0")
	}
	solution := Solution{}
	for _, object := range profit.Objects {
		switch object.ObjectType {
		case "deposit":
			if object.Subtype >= NumResourceTypes || object.Subtype < 0 {
				return Scenario{}, Solution{}, fmt.Errorf("invalid subtype %d for deposit", object.Subtype)
			}
			scenario.deposits = append(scenario.deposits, Deposit{
				position: Position{object.X, object.Y},
				width:    object.Width,
				height:   object.Height,
				subtype:  object.Subtype,
			})
		case "obstacle":
			scenario.obstacles = append(scenario.obstacles, Obstacle{
				position: Position{object.X, object.Y},
				height:   object.Height,
				width:    object.Width,
			})
		case "factory":
			if object.Subtype >= NumProducts || object.Subtype < 0 {
				return Scenario{}, Solution{}, fmt.Errorf("invalid factory subtype %d", object.Subtype)
			}
			solution.factories = append(solution.factories, Factory{
				position: Position{object.X, object.Y},
				product:  object.Subtype,
			})
		case "mine":
			if object.Subtype >= NumDirections || object.Subtype < 0 {
				return Scenario{}, Solution{}, fmt.Errorf("invalid mine subtype: %d", object.Subtype)
			}
			direction := DirectionFromSubtype(object.Subtype)
			solution.mines = append(solution.mines, Mine{
				position:  Position{object.X, object.Y},
				direction: direction,
			})
		case "conveyor":
			if object.Subtype >= NumConveyorSubtypes || object.Subtype < 0 {
				_ = fmt.Errorf("importing a conveyor failed, invalid subtype")
				return Scenario{}, Solution{}, fmt.Errorf("invalid conveyor subtype: %d", object.Subtype)
			}
			direction := DirectionFromSubtype(object.Subtype)
			length := ConveyorLengthFromSubtype(object.Subtype)
			// TODO: Think about building proper paths
			solution.paths = append(solution.paths, Path{
				conveyors: []Conveyor{{
					position:  Position{object.X, object.Y},
					direction: direction,
					length:    length,
				}},
			})
		case "combiner":
			if object.Subtype >= NumDirections || object.Subtype < 0 {
				_ = fmt.Errorf("importing a combiner failed, invalid subtype")
				return Scenario{}, Solution{}, fmt.Errorf("invalid combiner subtype: %d", object.Subtype)
			}
			direction := DirectionFromSubtype(object.Subtype)
			solution.combiners = append(solution.combiners, Combiner{
				position:  Position{object.X, object.Y},
				direction: direction,
			})
		default:
			return Scenario{}, Solution{}, fmt.Errorf("unknown ObjectType: %s", object.ObjectType)
		}
	}

	for _, product := range profit.Products {
		if product.ObjectType != "product" {
			return Scenario{}, Solution{}, fmt.Errorf("expected ObjectType to be 'product', not %s", product.ObjectType)
		}
		scenario.products = append(scenario.products, Product{
			subtype:   product.Subtype,
			points:    product.Points,
			resources: product.Resources,
		})
	}

	// Do a BFS starting at every factory, to determine distance from every mine to it's factory
	combinerMatrix := make([][]*Combiner, scenario.width)
	for i := range combinerMatrix {
		combinerMatrix[i] = make([]*Combiner, scenario.height)
	}
	mineMatrix := make([][]*Mine, scenario.width)
	for i := range mineMatrix {
		mineMatrix[i] = make([]*Mine, scenario.height)
	}
	conveyorMatrix := make([][]*Conveyor, scenario.width)
	for i := range conveyorMatrix {
		conveyorMatrix[i] = make([]*Conveyor, scenario.height)
	}
	for i := range solution.mines {
		mine := &solution.mines[i]
		mine.RectanglesEach(func(rectangle Rectangle) {
			rectangle.ForEach(func(position Position) {
				mineMatrix[position.x][position.y] = mine
			})
		})
	}
	for i := range solution.combiners {
		combiner := &solution.combiners[i]
		combiner.RectanglesEach(func(rectangle Rectangle) {
			rectangle.ForEach(func(position Position) {
				combinerMatrix[position.x][position.y] = combiner
			})
		})
	}
	for i := range solution.paths {
		for j := range solution.paths[i].conveyors {
			conveyor := &solution.paths[i].conveyors[j]
			conveyor.Rectangle().ForEach(func(position Position) {
				conveyorMatrix[position.x][position.y] = conveyor
			})
		}
	}

	for i := range solution.factories {
		distance := 0
		factory := &solution.factories[i]
		positions := factory.NextToIngressPositions()
		visitedPosition := make([][]bool, scenario.width)
		for j := range visitedPosition {
			visitedPosition[j] = make([]bool, scenario.height)
		}
		for {
			distance++
			nextPositions := make([]Position, 0)
			for _, p := range positions {
				if p.x < 0 || p.x >= scenario.width || p.y < 0 || p.y >= scenario.height {
					continue
				}
				if visitedPosition[p.x][p.y] {
					continue
				}
				visitedPosition[p.x][p.y] = true
				if conveyorMatrix[p.x][p.y] != nil && conveyorMatrix[p.x][p.y].Egress() == p {
					conveyorMatrix[p.x][p.y].distance = distance
					nextPositions = append(nextPositions, conveyorMatrix[p.x][p.y].NextToIngressPositions()...)
				}
				if combinerMatrix[p.x][p.y] != nil && combinerMatrix[p.x][p.y].Egress() == p {
					combinerMatrix[p.x][p.y].distance = distance
					nextPositions = append(nextPositions, combinerMatrix[p.x][p.y].NextToIngressPositions()...)
				}
				if mineMatrix[p.x][p.y] != nil && mineMatrix[p.x][p.y].Egress() == p {
					mineMatrix[p.x][p.y].distance = distance
					mineMatrix[p.x][p.y].connectedFactory = factory
					nextPositions = append(nextPositions, mineMatrix[p.x][p.y].NextToIngressPositions()...)
				}
			}
			positions = nextPositions
			if len(positions) == 0 {
				break
			}
		}
	}

	return scenario, solution, nil
}
