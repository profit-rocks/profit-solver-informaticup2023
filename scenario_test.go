package main

import "testing"

func TestDirectionFromSubtype(t *testing.T) {
	subtypes := [NumConveyorSubtypes]int{0, 1, 2, 3, 4, 5, 6, 7}
	directions := [NumConveyorSubtypes]Direction{Right, Bottom, Left, Top, Right, Bottom, Left, Top}
	for i := 0; i < NumConveyorSubtypes; i++ {
		if DirectionFromSubtype(subtypes[i]) != directions[i] {
			t.Errorf("subtype %d should result in Direction %d", subtypes[i], directions[i])
		}
	}
}
