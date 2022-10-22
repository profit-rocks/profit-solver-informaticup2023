package main

import (
	"encoding/json"
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

func exportScenario(scenario Scenario, path string) {
	exportableScenario := ExportableScenario{Height: scenario.height, Width: scenario.width, Turns: 100, Time: 100}

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

	for _, factory := range scenario.factories {
		exportableScenario.Objects = append(exportableScenario.Objects, Object{
			ObjectType: "factory",
			Subtype:    factory.product,
			X:          factory.position.x,
			Y:          factory.position.y,
		})

	}

	b, _ := json.MarshalIndent(exportableScenario, "", " ")
	_ = os.WriteFile("test.json", b, 0644)
}
