package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type ProfitStruct struct {
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
	exportableScenario := solutionToExportableProfitStruct(scenario, solution)
	b, err := json.MarshalIndent(exportableScenario, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, b, 0644)
}

func solutionToExportableProfitStruct(scenario Scenario, solution Solution) ProfitStruct {
	exportableScenario := ProfitStruct{
		Height:   scenario.height,
		Width:    scenario.width,
		Objects:  []Object{},
		Products: []Object{},
		Turns:    scenario.turns,
		Time:     100,
	}

	for _, deposit := range scenario.deposits {
		exportableScenario.Objects = append(exportableScenario.Objects, Object{
			ObjectType: "deposit",
			Subtype:    deposit.subtype,
			X:          deposit.position.x,
			Y:          deposit.position.y,
			Width:      deposit.width,
			Height:     deposit.height,
		})
	}
	for _, obstacle := range scenario.obstacles {
		exportableScenario.Objects = append(exportableScenario.Objects, Object{
			ObjectType: "obstacle",
			X:          obstacle.position.x,
			Y:          obstacle.position.y,
			Width:      obstacle.width,
			Height:     obstacle.height,
		})
	}

	for _, factory := range solution.factories {
		exportableScenario.Objects = append(exportableScenario.Objects, Object{
			ObjectType: "factory",
			Subtype:    factory.product,
			X:          factory.position.x,
			Y:          factory.position.y,
		})
	}

	for _, mine := range solution.mines {
		exportableScenario.Objects = append(exportableScenario.Objects, Object{
			ObjectType: "mine",
			Subtype:    int(mine.direction),
			X:          mine.position.x,
			Y:          mine.position.y,
		})
	}

	for _, path := range solution.paths {
		for _, conveyor := range path.conveyors {
			exportableScenario.Objects = append(exportableScenario.Objects, Object{
				ObjectType: "conveyor",
				Subtype:    conveyor.Subtype(),
				X:          conveyor.position.x,
				Y:          conveyor.position.y,
			})
		}
	}

	for _, product := range scenario.products {
		exportableScenario.Products = append(exportableScenario.Products, Object{
			ObjectType: "product",
			Subtype:    product.subtype,
			Points:     product.points,
			Resources:  product.resources,
		})
	}
	return exportableScenario
}

func importScenarioFromJson(path string) (Scenario, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return Scenario{}, err
	}
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return Scenario{}, err
	}
	var importedScenario ProfitStruct
	err = json.Unmarshal(byteValue, &importedScenario)
	if err != nil {
		return Scenario{}, err
	}

	scenario := Scenario{
		width:  importedScenario.Width,
		height: importedScenario.Height,
		turns:  importedScenario.Turns,
	}
	for _, object := range importedScenario.Objects {
		switch object.ObjectType {
		case "deposit":
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
		default:
			return Scenario{}, fmt.Errorf("unknown ObjectType: %s", object.ObjectType)
		}
	}

	for _, product := range importedScenario.Products {
		if product.ObjectType != "product" {
			return Scenario{}, fmt.Errorf("expected ObjectType to be 'product', not %s", product.ObjectType)
		}
		scenario.products = append(scenario.products, Product{
			subtype:   product.Subtype,
			points:    product.Points,
			resources: product.Resources,
		})
	}
	return scenario, nil
}

func importSolutionFromJson(path string) (Solution, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return Solution{}, err
	}
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return Solution{}, err
	}
	var importedSolution ProfitStruct
	err = json.Unmarshal(byteValue, &importedSolution)
	if err != nil {
		return Solution{}, err
	}

	solution := Solution{}
	for _, object := range importedSolution.Objects {
		switch object.ObjectType {
		case "factory":
			solution.factories = append(solution.factories, Factory{
				position: Position{object.X, object.Y},
				product:  object.Subtype,
			})
		case "mine":
			if object.Subtype > 3 || object.Subtype < 0 {
				_ = fmt.Errorf("importing a mine failed, invalid subtype")
				continue
			}
			direction := DirectionFromSubtype(object.Subtype)
			solution.mines = append(solution.mines, Mine{
				position:  Position{object.X, object.Y},
				direction: direction,
			})
		case "conveyor":
			if object.Subtype > 7 || object.Subtype < 0 {
				_ = fmt.Errorf("importing a conveyor failed, invalid subtype")
				continue
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
		case "deposit":
			continue
		case "obstacle":
			continue
		default:
			return Solution{}, fmt.Errorf("unknown ObjectType: %s", object.ObjectType)
		}
	}
	return solution, nil
}
