package main

import "testing"

func combinerRectanglesSlice(combiner *Combiner) []Rectangle {
	rectangles := make([]Rectangle, 0)
	combiner.RectanglesEach(func(r Rectangle) {
		rectangles = append(rectangles, r)
	})
	return rectangles
}

func TestRightCombinerRectangles(t *testing.T) {
	combiner := Combiner{position: Position{1, 1}, direction: Right}

	rectangles := combinerRectanglesSlice(&combiner)
	validRectangles := []Rectangle{
		{Position{0, 0}, 2, 3},
		{Position{2, 1}, 1, 1},
	}

	for i := range validRectangles {
		if rectangles[i] != validRectangles[i] {
			t.Errorf("Rectangle %d is not valid", i)
		}
	}
}

func TestBottomCombinerRectangles(t *testing.T) {
	combiner := Combiner{position: Position{1, 1}, direction: Bottom}

	rectangles := combinerRectanglesSlice(&combiner)
	validRectangles := []Rectangle{
		{Position{0, 0}, 3, 2},
		{Position{1, 2}, 1, 1},
	}

	for i := range validRectangles {
		if rectangles[i] != validRectangles[i] {
			t.Errorf("Rectangle %d is not valid", i)
		}
	}
}

func TestLeftCombinerRectangles(t *testing.T) {
	combiner := Combiner{position: Position{1, 1}, direction: Left}

	rectangles := combinerRectanglesSlice(&combiner)
	validRectangles := []Rectangle{
		{Position{1, 0}, 2, 3},
		{Position{0, 1}, 1, 1},
	}

	for i := range validRectangles {
		if rectangles[i] != validRectangles[i] {
			t.Errorf("Rectangle %d is not valid", i)
		}
	}
}

func TestTopCombinerRectangles(t *testing.T) {
	combiner := Combiner{position: Position{1, 1}, direction: Top}

	rectangles := combinerRectanglesSlice(&combiner)
	validRectangles := []Rectangle{
		{Position{0, 1}, 3, 2},
		{Position{1, 0}, 1, 1},
	}

	for i := range validRectangles {
		if rectangles[i] != validRectangles[i] {
			t.Errorf("Rectangle %d is not valid", i)
		}
	}
}

func TestCombinerNextToIngressPositions(t *testing.T) {
	combiner := Combiner{
		position:  Position{2, 2},
		direction: Right,
	}
	validPositionsRight := []Position{{0, 2}, {0, 1}, {0, 3}, {1, 0}, {1, 4}}
	validPositionsBottom := []Position{{2, 0}, {1, 0}, {3, 0}, {0, 1}, {4, 1}}
	validPositionsLeft := []Position{{4, 2}, {4, 1}, {4, 3}, {3, 0}, {3, 4}}
	validPositionsTop := []Position{{2, 4}, {1, 4}, {3, 4}, {0, 3}, {4, 3}}

	validPositions := [][]Position{validPositionsRight, validPositionsBottom, validPositionsLeft, validPositionsTop}
	for i, direction := range []Direction{Right, Bottom, Left, Top} {
		combiner.direction = direction
		result := combiner.NextToIngressPositions()
		if len(result) != len(validPositions[i]) {
			t.Errorf("number of positions should be 5")
		}
		for j, p := range result {
			if p != validPositions[i][j] {
				t.Errorf("position %v of combiner with subtype %d does not match %v", p, combiner.direction, validPositions[i][j])
			}
		}
	}
}
