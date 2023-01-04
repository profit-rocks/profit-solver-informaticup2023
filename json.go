package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type Exporter interface {
	Export(scenario Scenario, chromosome Chromosome) any
}

// ScenarioExporter exports the scenario and the chromosome, useful for debugging on profit.phinau.de
type ScenarioExporter struct{}

// SolutionExporter exports the chromosome as specified by the problem statement
type SolutionExporter struct{}

type SerializedScenario struct {
	Height   int                        `json:"height"`
	Width    int                        `json:"width"`
	Objects  []SerializedScenarioObject `json:"objects"`
	Products []SerializedScenarioObject `json:"products"`
	Turns    int                        `json:"turns"`
	Time     int                        `json:"time"`
}

type SerializedScenarioObject struct {
	ObjectType string `json:"type"`
	Subtype    int    `json:"subtype"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Resources  []int  `json:"resources"`
	Points     int    `json:"points"`
}

type SolutionExport []SolutionExportObject

type SolutionExportObject struct {
	ObjectType string `json:"type"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	Subtype    int    `json:"subtype"`
}

func (c Chromosome) Export(scenario Scenario, exporter Exporter, filePath string) error {
	b, err := json.MarshalIndent(exporter.Export(scenario, c), "", " ")
	if err != nil {
		return err
	}
	if filePath == "-" {
		_, err = os.Stdout.Write(b)
		return err
	}
	return os.WriteFile(filePath, b, 0644)
}

func (_ SolutionExporter) Export(s Scenario, c Chromosome) any {
	export := SolutionExport{}
	for _, factory := range c.factories {
		export = append(export, SolutionExportObject{
			ObjectType: "factory",
			Subtype:    factory.product,
			X:          factory.position.x,
			Y:          factory.position.y,
		})
	}

	for _, mine := range c.mines {
		export = append(export, SolutionExportObject{
			ObjectType: "mine",
			Subtype:    int(mine.direction),
			X:          mine.position.x,
			Y:          mine.position.y,
		})
	}

	for _, path := range c.paths {
		for _, conveyor := range path.conveyors {
			export = append(export, SolutionExportObject{
				ObjectType: "conveyor",
				Subtype:    conveyor.Subtype(),
				X:          conveyor.position.x,
				Y:          conveyor.position.y,
			})
		}
	}

	for _, combiner := range c.combiners {
		export = append(export, SolutionExportObject{
			ObjectType: "combiner",
			X:          combiner.position.x,
			Y:          combiner.position.y,
			Subtype:    int(combiner.direction),
		})
	}
	return export
}

func (_ ScenarioExporter) Export(s Scenario, c Chromosome) any {
	export := SerializedScenario{
		Height:   s.height,
		Width:    s.width,
		Objects:  []SerializedScenarioObject{},
		Products: []SerializedScenarioObject{},
		Turns:    s.turns,
		Time:     100,
	}

	for _, deposit := range s.deposits {
		export.Objects = append(export.Objects, SerializedScenarioObject{
			ObjectType: "deposit",
			Subtype:    deposit.subtype,
			X:          deposit.position.x,
			Y:          deposit.position.y,
			Width:      deposit.width,
			Height:     deposit.height,
		})
	}
	for _, obstacle := range s.obstacles {
		export.Objects = append(export.Objects, SerializedScenarioObject{
			ObjectType: "obstacle",
			X:          obstacle.position.x,
			Y:          obstacle.position.y,
			Width:      obstacle.width,
			Height:     obstacle.height,
		})
	}

	for _, factory := range c.factories {
		export.Objects = append(export.Objects, SerializedScenarioObject{
			ObjectType: "factory",
			Subtype:    factory.product,
			X:          factory.position.x,
			Y:          factory.position.y,
		})
	}

	for _, mine := range c.mines {
		export.Objects = append(export.Objects, SerializedScenarioObject{
			ObjectType: "mine",
			Subtype:    int(mine.direction),
			X:          mine.position.x,
			Y:          mine.position.y,
		})
	}

	for _, path := range c.paths {
		for _, conveyor := range path.conveyors {
			export.Objects = append(export.Objects, SerializedScenarioObject{
				ObjectType: "conveyor",
				Subtype:    conveyor.Subtype(),
				X:          conveyor.position.x,
				Y:          conveyor.position.y,
			})
		}
	}

	for _, combiner := range c.combiners {
		export.Objects = append(export.Objects, SerializedScenarioObject{
			ObjectType: "combiner",
			X:          combiner.position.x,
			Y:          combiner.position.y,
			Subtype:    int(combiner.direction),
		})
	}

	for _, product := range s.products {
		export.Products = append(export.Products, SerializedScenarioObject{
			ObjectType: "product",
			Subtype:    product.subtype,
			Points:     product.points,
			Resources:  product.resources,
		})
	}
	return export
}

func ImportScenario(path string) (Scenario, Chromosome, error) {
	var jsonFile *os.File
	var err error
	if path == "-" {
		jsonFile = os.Stdin
	} else {
		jsonFile, err = os.Open(path)
		if err != nil {
			return Scenario{}, Chromosome{}, err
		}
		defer jsonFile.Close()
	}
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return Scenario{}, Chromosome{}, err
	}
	var s SerializedScenario
	err = json.Unmarshal(byteValue, &s)
	if err != nil {
		return Scenario{}, Chromosome{}, err
	}

	scenario := Scenario{
		width:  s.Width,
		height: s.Height,
		turns:  s.Turns,
		time:   s.Time,
	}
	if scenario.time <= 0 {
		return Scenario{}, Chromosome{}, errors.New("time imported from json has to be greater than 0")
	}
	chromosome := Chromosome{}
	for _, object := range s.Objects {
		switch object.ObjectType {
		case "deposit":
			if object.Subtype >= NumResourceTypes || object.Subtype < 0 {
				return Scenario{}, Chromosome{}, fmt.Errorf("invalid subtype %d for deposit", object.Subtype)
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
				return Scenario{}, Chromosome{}, fmt.Errorf("invalid factory subtype %d", object.Subtype)
			}
			chromosome.factories = append(chromosome.factories, Factory{
				position: Position{object.X, object.Y},
				product:  object.Subtype,
			})
		case "mine":
			if object.Subtype >= NumDirections || object.Subtype < 0 {
				return Scenario{}, Chromosome{}, fmt.Errorf("invalid mine subtype: %d", object.Subtype)
			}
			direction := DirectionFromSubtype(object.Subtype)
			chromosome.mines = append(chromosome.mines, Mine{
				position:  Position{object.X, object.Y},
				direction: direction,
			})
		case "conveyor":
			if object.Subtype >= NumConveyorSubtypes || object.Subtype < 0 {
				_ = fmt.Errorf("importing a conveyor failed, invalid subtype")
				return Scenario{}, Chromosome{}, fmt.Errorf("invalid conveyor subtype: %d", object.Subtype)
			}
			direction := DirectionFromSubtype(object.Subtype)
			length := ConveyorLengthFromSubtype(object.Subtype)
			// TODO: Think about building proper paths
			chromosome.paths = append(chromosome.paths, Path{
				conveyors: []Conveyor{{
					position:  Position{object.X, object.Y},
					direction: direction,
					length:    length,
				}},
			})
		case "combiner":
			if object.Subtype >= NumDirections || object.Subtype < 0 {
				_ = fmt.Errorf("importing a combiner failed, invalid subtype")
				return Scenario{}, Chromosome{}, fmt.Errorf("invalid combiner subtype: %d", object.Subtype)
			}
			direction := DirectionFromSubtype(object.Subtype)
			chromosome.combiners = append(chromosome.combiners, Combiner{
				position:  Position{object.X, object.Y},
				direction: direction,
			})
		default:
			return Scenario{}, Chromosome{}, fmt.Errorf("unknown ObjectType: %s", object.ObjectType)
		}
	}

	for _, product := range s.Products {
		if product.ObjectType != "product" {
			return Scenario{}, Chromosome{}, fmt.Errorf("expected ObjectType to be 'product', not %s", product.ObjectType)
		}
		scenario.products = append(scenario.products, Product{
			subtype:   product.Subtype,
			points:    product.Points,
			resources: product.Resources,
		})
	}

	chromosome.determineDistancesFromMinesToFactories(scenario)

	return scenario, chromosome, nil
}

// we perform a BFS from all factories to the mines
func (c *Chromosome) determineDistancesFromMinesToFactories(scenario Scenario) {
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
	for i := range c.mines {
		mine := &c.mines[i]
		mine.RectanglesEach(func(rectangle Rectangle) {
			rectangle.ForEach(func(position Position) {
				mineMatrix[position.x][position.y] = mine
			})
		})
	}
	for i := range c.combiners {
		combiner := &c.combiners[i]
		combiner.RectanglesEach(func(rectangle Rectangle) {
			rectangle.ForEach(func(position Position) {
				combinerMatrix[position.x][position.y] = combiner
			})
		})
	}
	for i := range c.paths {
		for j := range c.paths[i].conveyors {
			conveyor := &c.paths[i].conveyors[j]
			conveyor.Rectangle().ForEach(func(position Position) {
				conveyorMatrix[position.x][position.y] = conveyor
			})
		}
	}

	for i := range c.factories {
		distance := 0
		factory := &c.factories[i]
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
}
