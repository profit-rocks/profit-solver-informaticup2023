package main

import "testing"

func TestConveyor_Egress(t *testing.T) {
	conveyors := []Conveyor{
		{direction: Right, position: Position{5, 5}, length: Short},
		{direction: Right, position: Position{5, 5}, length: Long},
		{direction: Bottom, position: Position{5, 5}, length: Short},
		{direction: Bottom, position: Position{5, 5}, length: Long},
		{direction: Left, position: Position{5, 5}, length: Short},
		{direction: Left, position: Position{5, 5}, length: Long},
		{direction: Top, position: Position{5, 5}, length: Short},
		{direction: Top, position: Position{5, 5}, length: Long},
	}
	expectedEgresses := []Position{
		{6, 5},
		{7, 5},
		{5, 6},
		{5, 7},
		{4, 5},
		{4, 5},
		{5, 4},
		{5, 4},
	}
	for i, conveyor := range conveyors {
		if conveyor.Egress() != expectedEgresses[i] {
			t.Errorf("Egress() = %v, want %v", conveyor.Egress(), expectedEgresses[i])
		}
	}
}

func TestConveyor_Ingress(t *testing.T) {
	conveyors := []Conveyor{
		{direction: Right, position: Position{5, 5}, length: Short},
		{direction: Right, position: Position{5, 5}, length: Long},
		{direction: Bottom, position: Position{5, 5}, length: Short},
		{direction: Bottom, position: Position{5, 5}, length: Long},
		{direction: Left, position: Position{5, 5}, length: Short},
		{direction: Left, position: Position{5, 5}, length: Long},
		{direction: Top, position: Position{5, 5}, length: Short},
		{direction: Top, position: Position{5, 5}, length: Long},
	}
	expectedIngresses := []Position{
		{4, 5},
		{4, 5},
		{5, 4},
		{5, 4},
		{6, 5},
		{7, 5},
		{5, 6},
		{5, 7},
	}
	for i, conveyor := range conveyors {
		if conveyor.Ingress() != expectedIngresses[i] {
			t.Errorf("Ingress() = %v, want %v", conveyor.Ingress(), expectedIngresses[i])
		}
	}
}

func TestConveyor_NextToEgressPositions(t *testing.T) {
	conveyors := []Conveyor{
		{direction: Right, position: Position{5, 5}, length: Short},
		{direction: Right, position: Position{5, 5}, length: Long},
		{direction: Bottom, position: Position{5, 5}, length: Short},
		{direction: Bottom, position: Position{5, 5}, length: Long},
		{direction: Left, position: Position{5, 5}, length: Short},
		{direction: Left, position: Position{5, 5}, length: Long},
		{direction: Top, position: Position{5, 5}, length: Short},
		{direction: Top, position: Position{5, 5}, length: Long},
	}
	expectedPositions := [][]Position{
		{{6, 4}, {7, 5}, {6, 6}},
		{{7, 4}, {8, 5}, {7, 6}},
		{{6, 6}, {5, 7}, {4, 6}},
		{{6, 7}, {5, 8}, {4, 7}},
		{{4, 6}, {3, 5}, {4, 4}},
		{{4, 6}, {3, 5}, {4, 4}},
		{{4, 4}, {5, 3}, {6, 4}},
		{{4, 4}, {5, 3}, {6, 4}},
	}
	for i, conveyor := range conveyors {
		for j, pos := range conveyor.NextToEgressPositions() {
			if pos != expectedPositions[i][j] {
				t.Errorf("NextToEgressPositions() = %v, want %v", pos, expectedPositions[i][j])
			}
		}
	}
}
