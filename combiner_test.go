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
