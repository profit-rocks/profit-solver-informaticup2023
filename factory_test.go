package main

import "testing"

func TestIsPositionAvailableForFactory(t *testing.T) {
	deposit := Deposit{position: Position{2, 2}, subtype: 0, height: 4, width: 8}
	scenario := Scenario{width: 20, height: 20, deposits: []Deposit{deposit}}
	if scenario.positionAvailableForFactory([]Factory{}, []Mine{}, []Combiner{}, []Path{}, Position{0, 0}) {
		t.Error("Factory position should not be available")
	}
}

func TestRandomFactoryNoSpace(t *testing.T) {
	scenario := Scenario{width: 4, height: 4, products: []Product{{subtype: 0, points: 1, resources: []int{1, 0, 0, 0, 0, 0, 0, 0}}}}
	_, err := scenario.randomFactory(Chromosome{})
	if err == nil {
		t.Error("No space for factory, but no error was returned")
	}
}
