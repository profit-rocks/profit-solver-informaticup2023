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

func exportScenario(scenario ExportableScenario, path string) error {
	b, err := json.MarshalIndent(scenario, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
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
		width:        importedScenario.Width,
		height:       importedScenario.Height,
		turns:        importedScenario.Turns,
		numFactories: NumFactories,
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
