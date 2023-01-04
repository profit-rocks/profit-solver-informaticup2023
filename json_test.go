package main

import (
	"fmt"
	"testing"
)

func TestImportMineWithInvalidSubtype(t *testing.T) {
	for i := 1; i < 3; i++ {
		_, _, err := ImportScenario(fmt.Sprintf("fixtures/mineWithInvalidSubtype%d.json", i))
		if err == nil {
			t.Errorf("importing invalid subtype should return an error")
		}
	}
}

func TestImportConveyerWithInvalidSubtype(t *testing.T) {
	for i := 1; i < 3; i++ {
		_, _, err := ImportScenario(fmt.Sprintf("fixtures/conveyorWithInvalidSubtype%d.json", i))
		if err == nil {
			t.Errorf("importing invalid subtype should return an error")
		}
	}
}

func TestImportCombiner(t *testing.T) {
	_, chromosome, err := ImportScenario("fixtures/singleCombiner.json")
	if err != nil {
		t.Errorf("importing combiner should not throw an error")
	}
	if len(chromosome.combiners) == 0 {
		t.Errorf("a combiner should be added to the chromsome")
	}
}
