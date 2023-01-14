package main

import (
	"reflect"
	"testing"
)

func mineRectanglesSlice(mine *Mine) []Rectangle {
	rectangles := make([]Rectangle, 0)
	mine.RectanglesEach(func(r Rectangle) {
		rectangles = append(rectangles, r)
	})
	return rectangles
}

func TestRightMineRectangles(t *testing.T) {
	mine := Mine{position: Position{1, 0}, direction: Right}

	rectangles := mineRectanglesSlice(&mine)
	validRectangles := []Rectangle{
		{Position{1, 0}, 2, 1, nil},
		{Position{0, 1}, 4, 1, nil},
	}

	for i := range validRectangles {
		if !rectangles[i].Equals(validRectangles[i]) {
			t.Errorf("Rectangle %d is not valid", i)
		}
	}
}

func TestBottomMineRectangles(t *testing.T) {
	mine := Mine{position: Position{0, 1}, direction: Bottom}

	rectangles := mineRectanglesSlice(&mine)
	validRectangles := []Rectangle{
		{Position{0, 0}, 1, 4, nil},
		{Position{1, 1}, 1, 2, nil},
	}

	for i := range validRectangles {
		if !rectangles[i].Equals(validRectangles[i]) {
			t.Errorf("Rectangle %d is not valid", i)
		}
	}
}

func TestLeftMineRectangles(t *testing.T) {
	mine := Mine{position: Position{1, 0}, direction: Left}

	rectangles := mineRectanglesSlice(&mine)
	validRectangles := []Rectangle{
		{Position{0, 0}, 4, 1, nil},
		{Position{1, 1}, 2, 1, nil},
	}

	for i := range validRectangles {
		if !rectangles[i].Equals(validRectangles[i]) {
			t.Errorf("Rectangle %d is not valid", i)
		}
	}
}

func TestTopMineRectangles(t *testing.T) {
	mine := Mine{position: Position{0, 1}, direction: Top}

	rectangles := mineRectanglesSlice(&mine)
	validRectangles := []Rectangle{
		{Position{0, 1}, 1, 2, nil},
		{Position{1, 0}, 1, 4, nil},
	}

	for i := range validRectangles {
		if !rectangles[i].Equals(validRectangles[i]) {
			t.Errorf("Rectangle %d is not valid", i)
		}
	}
}

func TestRectangle_Positions(t *testing.T) {
	rect := Rectangle{position: Position{0, 0}, width: 3, height: 2}
	positions := []Position{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}, {1, 2}}
	if !reflect.DeepEqual(rect.Positions(), positions) {
		t.Error("Rectangle should have positions", positions)
	}
}

func TestRectangle_PositionsCaching(t *testing.T) {
	rect := Rectangle{position: Position{1, 1}, width: 1, height: 1}
	positions := []Position{{1, 1}}
	rect.Positions()
	if !reflect.DeepEqual(rect.Positions(), positions) {
		t.Error("Rectangle should have cached positions", positions)

	}
}
