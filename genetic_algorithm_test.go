package main

import "testing"

func geneticAlgorithmFromScenario(scenario Scenario) GeneticAlgorithm {
	return GeneticAlgorithm{
		scenario: scenario,
	}
}

func TestLargeEmptyScenarioIsAvailable(t *testing.T) {
	g := geneticAlgorithmFromScenario(largeEmptyScenario())
	for x := 0; x <= g.scenario.width-FactoryWidth; x++ {
		for y := 0; y <= g.scenario.height-FactoryHeight; y++ {
			if !g.scenario.positionAvailableForFactory([]Factory{}, []Mine{}, []Combiner{}, []Path{}, Position{x, y}) {
				t.Errorf("position %v should be available", Position{x, y})
			}
		}
	}
}

func TestLargeEmptyScenarioBorders(t *testing.T) {
	g := geneticAlgorithmFromScenario(largeEmptyScenario())
	for x := 0; x <= g.scenario.width-FactoryWidth; x++ {
		for y := g.scenario.height - FactoryHeight + 1; y < g.scenario.width; y++ {
			if g.scenario.positionAvailableForFactory([]Factory{}, []Mine{}, []Combiner{}, []Path{}, Position{x, y}) {
				t.Errorf("position %v should not be available", Position{x, y})
			}
		}
	}
	for x := g.scenario.width - FactoryWidth + 1; x < g.scenario.width; x++ {
		for y := 0; y <= g.scenario.height-FactoryHeight; y++ {
			if g.scenario.positionAvailableForFactory([]Factory{}, []Mine{}, []Combiner{}, []Path{}, Position{x, y}) {
				t.Errorf("position %v should not be available", Position{x, y})
			}
		}
	}
}

func TestSmallEmptyScenario(t *testing.T) {
	g := geneticAlgorithmFromScenario(smallEmptyScenario())
	for x := 0; x < g.scenario.width; x++ {
		for y := 0; y < g.scenario.height; y++ {
			pos := Position{x, y}
			if g.scenario.positionAvailableForFactory([]Factory{}, []Mine{}, []Combiner{}, []Path{}, pos) {
				t.Errorf("position %v should not be available", pos)
			}
		}
	}
}

func TestScenarioWithObstacles(t *testing.T) {
	g := geneticAlgorithmFromScenario(scenarioWithObstacle())

	for x := 0; x < g.scenario.width; x++ {
		for y := 0; y < g.scenario.height; y++ {
			pos := Position{x, y}
			if g.scenario.positionAvailableForFactory([]Factory{}, []Mine{}, []Combiner{}, []Path{}, pos) {
				t.Errorf("position %v should not be available", pos)
			}
		}
	}

}

func chromosomeWithSingleFactory() Chromosome {
	return Chromosome{
		factories: []Factory{{position: Position{5, 5}, product: 0}},
	}
}

func TestScenarioWithFactory(t *testing.T) {
	g := geneticAlgorithmFromScenario(largeEmptyScenario())
	chromosome := chromosomeWithSingleFactory()

	for x := 0; x <= g.scenario.width-FactoryWidth; x++ {
		for y := 0; y <= g.scenario.height-FactoryHeight; y++ {
			pos := Position{x, y}
			if x > 5-FactoryWidth && x < 5+FactoryWidth && y > 5-FactoryHeight && y < 5+FactoryWidth {
				if g.scenario.positionAvailableForFactory(chromosome.factories, chromosome.mines, []Combiner{}, chromosome.paths, pos) {
					t.Errorf("position %v should not be available", pos)
				}
			} else {
				if !g.scenario.positionAvailableForFactory(chromosome.factories, chromosome.mines, []Combiner{}, chromosome.paths, pos) {
					t.Errorf("position %v should be available", pos)
				}
			}
		}
	}
}

func TestPositionAvailableForFactory(t *testing.T) {
	scenario := largeEmptyScenario()
	chromosome := chromosomeWithSingleMineForLargeEmptyScenario()
	mines := make([]Mine, len(chromosome.mines))
	for i, mine := range chromosome.mines {
		mines[i] = mine
	}
	for _, mine := range chromosome.mines {
		if scenario.positionAvailableForFactory([]Factory{}, mines, []Combiner{}, []Path{}, mine.position) {
			t.Errorf("position %v should not be available", mine.position)
		}
	}
}

func TestPositionAvailableForCombiner(t *testing.T) {
	scenario, chromosome, err := ImportScenario("fixtures/freePlacesForCombiners.json")
	if err != nil {
		t.Errorf("failed to import fixture: %e", err)
	}
	applicableCombiners := []Combiner{{position: Position{1, 1}, direction: Right}, {position: Position{1, 5}, direction: Bottom}, {position: Position{1, 9}, direction: Left}, {position: Position{1, 13}, direction: Top}}
	for _, combiner := range applicableCombiners {
		if !scenario.positionAvailableForCombiner(chromosome.factories, chromosome.mines, chromosome.paths, chromosome.combiners, combiner) {
			t.Errorf("position %v should be available", combiner)
		}
	}
}

func TestPositionNotAvailableForCombiner(t *testing.T) {
	scenario, chromsome, err := ImportScenario("fixtures/scenarioWithSingleDeposit.json")
	if err != nil {
		t.Errorf("failed to import fixture: %e", err)
	}
	nonApplicableCombiners := []Combiner{{position: Position{1, 4}, direction: Right}, {position: Position{4, 2}, direction: Bottom}, {position: Position{4, 1}, direction: Top}}
	for _, combiner := range nonApplicableCombiners {
		if scenario.positionAvailableForCombiner(chromsome.factories, chromsome.mines, chromsome.paths, chromsome.combiners, combiner) {
			t.Errorf("position %v should not be available", combiner)
		}
	}
}

func TestScenarioWithDeposit(t *testing.T) {
	g := geneticAlgorithmFromScenario(scenarioWithDeposit())

	for x := 0; x <= g.scenario.width-FactoryWidth; x++ {
		for y := 0; y <= g.scenario.height-FactoryHeight; y++ {
			pos := Position{x, y}
			if x == 0 && y == 0 || x == 0 && y == 10 || x == 10 && y == 0 || x == 10 && y == 10 {
				if !g.scenario.positionAvailableForFactory([]Factory{}, []Mine{}, []Combiner{}, []Path{}, pos) {
					t.Errorf("position %v should be available", pos)
				}
			} else {
				if g.scenario.positionAvailableForFactory([]Factory{}, []Mine{}, []Combiner{}, []Path{}, pos) {
					t.Errorf("position %v should not be available", pos)
				}
			}
		}
	}
}

func TestAvailableMinePositions(t *testing.T) {

	validMines := []Mine{
		{position: Position{6, 3}, direction: Right},
		{position: Position{7, 3}, direction: Right},
		{position: Position{8, 3}, direction: Right},
		{position: Position{9, 3}, direction: Right},
		{position: Position{10, 3}, direction: Right},
		{position: Position{11, 4}, direction: Right},
		{position: Position{11, 5}, direction: Right},
		{position: Position{11, 6}, direction: Right},
		{position: Position{11, 7}, direction: Right},
		{position: Position{11, 8}, direction: Right},
		{position: Position{10, 9}, direction: Right},
		{position: Position{4, 10}, direction: Bottom},
		{position: Position{5, 11}, direction: Bottom},
		{position: Position{6, 11}, direction: Bottom},
		{position: Position{7, 11}, direction: Bottom},
		{position: Position{8, 11}, direction: Bottom},
		{position: Position{9, 11}, direction: Bottom},
		{position: Position{10, 10}, direction: Bottom},
		{position: Position{10, 10}, direction: Bottom},
		{position: Position{10, 9}, direction: Bottom},
		{position: Position{10, 8}, direction: Bottom},
		{position: Position{10, 7}, direction: Bottom},
		{position: Position{10, 6}, direction: Bottom},
		{position: Position{3, 4}, direction: Left},
		{position: Position{2, 5}, direction: Left},
		{position: Position{2, 6}, direction: Left},
		{position: Position{2, 7}, direction: Left},
		{position: Position{2, 8}, direction: Left},
		{position: Position{2, 9}, direction: Left},
		{position: Position{3, 10}, direction: Left},
		{position: Position{4, 10}, direction: Left},
		{position: Position{5, 10}, direction: Left},
		{position: Position{6, 10}, direction: Left},
		{position: Position{7, 10}, direction: Left},
		{position: Position{9, 3}, direction: Top},
		{position: Position{8, 2}, direction: Top},
		{position: Position{7, 2}, direction: Top},
		{position: Position{6, 2}, direction: Top},
		{position: Position{5, 2}, direction: Top},
		{position: Position{4, 2}, direction: Top},
		{position: Position{3, 3}, direction: Top},
		{position: Position{3, 4}, direction: Top},
		{position: Position{3, 5}, direction: Top},
		{position: Position{3, 6}, direction: Top},
		{position: Position{3, 7}, direction: Top},
	}

	g := geneticAlgorithmFromScenario(scenarioWithDeposit())
	mines := g.scenario.minePositions(&g.scenario.deposits[0], Chromosome{})

	for _, mine := range mines {
		found := false
		for _, validMine := range validMines {
			if mine.position == validMine.position && mine.direction == validMine.direction {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("mine %v is not valid", mine)
		}
	}

	for _, validMine := range validMines {
		found := false
		for _, mine := range mines {
			if mine.position == validMine.position && mine.direction == validMine.direction {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("mine %v is not found", validMine)
		}
	}
}

func TestTwoMinesSameEgress(t *testing.T) {
	g := geneticAlgorithmFromScenario(scenarioWithDeposit())
	mines := g.scenario.minePositions(&g.scenario.deposits[0], Chromosome{
		factories: []Factory{},
		mines:     []Mine{{position: Position{6, 3}, direction: Right}},
		fitness:   0,
	})
	badMine := Mine{position: Position{3, 3}, direction: Top}
	for _, mine := range mines {
		if mine.position == badMine.position && mine.direction == badMine.direction {
			t.Errorf("mine %v is not valid", mine)
		}
	}
}

func TestPathExistsMineToCombiner(t *testing.T) {
	scenario, chromosome, err := ImportScenario("fixtures/pathExistsMineToCombiner.json")
	if err != nil {
		t.Errorf("failed to import fixture: %e", err)
	}

	g := geneticAlgorithmFromScenario(scenario)

	if len(chromosome.mines) != 1 {
		t.Errorf("Expected number of mines of imported solution to be 1")
	}
	if len(chromosome.combiners) != 1 {
		t.Errorf("Expected number of combiners of imported solution to be 1")
	}
	c := Chromosome{mines: chromosome.mines, combiners: chromosome.combiners}
	var positions []PathEndPosition
	for _, p := range c.combiners[0].NextToIngressPositions() {
		positions = append(positions, PathEndPosition{
			position: p})
	}
	_, _, err = g.path(c.mines[0].Egress(), positions)
	if err != nil {
		t.Errorf("Searching for a path between mine and combiner should not return err %e", err)
	}
}
