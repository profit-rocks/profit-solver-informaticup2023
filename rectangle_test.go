package main

import "testing"

func TestRightMineRectangles(t *testing.T) {
	mine := Mine{Position{1, 0}, Right}

	rectangles := mine.Rectangles()
	validRectangles := []Rectangle{
		{Position{1, 0}, 2, 1},
		{Position{0, 1}, 4, 1},
	}

	for i := range validRectangles {
		if rectangles[i] != validRectangles[i] {
			t.Errorf("Rectangle %d is not valid", i)
		}
	}
}

func TestBottomMineRectangles(t *testing.T) {
	mine := Mine{Position{0, 1}, Bottom}

	rectangles := mine.Rectangles()
	validRectangles := []Rectangle{
		{Position{0, 0}, 1, 4},
		{Position{1, 1}, 1, 2},
	}

	for i := range validRectangles {
		if rectangles[i] != validRectangles[i] {
			t.Errorf("Rectangle %d is not valid", i)
		}
	}
}

func TestLeftMineRectangles(t *testing.T) {
	mine := Mine{Position{1, 0}, Left}

	rectangles := mine.Rectangles()
	validRectangles := []Rectangle{
		{Position{0, 0}, 4, 1},
		{Position{1, 1}, 2, 1},
	}

	for i := range validRectangles {
		if rectangles[i] != validRectangles[i] {
			t.Errorf("Rectangle %d is not valid", i)
		}
	}
}

func TestTopMineRectangles(t *testing.T) {
	mine := Mine{Position{0, 1}, Top}

	rectangles := mine.Rectangles()
	validRectangles := []Rectangle{
		{Position{0, 1}, 1, 2},
		{Position{1, 0}, 1, 4},
	}

	for i := range validRectangles {
		if rectangles[i] != validRectangles[i] {
			t.Errorf("Rectangle %d is not valid", i)
		}
	}
}
