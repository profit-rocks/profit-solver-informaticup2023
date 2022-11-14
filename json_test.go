package main

import (
	"fmt"
	"testing"
)

func TestImportMineWithInvalidSubtype(t *testing.T) {
	for i := 1; i < 3; i++ {
		_, _, err := importFromProfitJson(fmt.Sprintf("fixtures/mineWithInvalidSubtype%d.json", i))
		if err == nil {
			t.Errorf("importing invalid subtype should return an error")
		}
	}
}

func TestImportConveyerWithInvalidSubtype(t *testing.T) {
	for i := 1; i < 3; i++ {
		_, _, err := importFromProfitJson(fmt.Sprintf("fixtures/conveyorWithInvalidSubtype%d.json", i))
		if err == nil {
			t.Errorf("importing invalid subtype should return an error")
		}
	}
}
