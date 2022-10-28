package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type ExportableScenario struct {
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
	exportableScenario := solutionToExportableScenario(scenario, solution)
	b, err := json.MarshalIndent(exportableScenario, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, b, 0644)
}

func solutionToExportableScenario(scenario Scenario, solution Solution) ExportableScenario {
	exportableScenario := ExportableScenario{
		Height:   scenario.height,
		Width:    scenario.width,
		Objects:  []Object{},
		Products: []Object{},
		Turns:    100,
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

	for _, conveyor := range solution.conveyors {
		exportableScenario.Objects = append(exportableScenario.Objects, Object{
			ObjectType: "conveyor",
			Subtype:    conveyor.Subtype(),
			X:          conveyor.position.x,
			Y:          conveyor.position.y,
		})
	}
	return exportableScenario
}

func importScenarioFromJson(path string) Scenario {
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	byteValue, _ := io.ReadAll(jsonFile)
	var importedScenario ExportableScenario
	err = json.Unmarshal(byteValue, &importedScenario)
	if err != nil {
		fmt.Println(err)
		return Scenario{}
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
			fmt.Println("Unknown ObjectType: ", object.ObjectType)
		}
	}

	return scenario
}
