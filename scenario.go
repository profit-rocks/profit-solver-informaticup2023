package main

type Obstacle = Rectangle

// Direction is the relative position of the egress
type Direction int

const (
	Right  Direction = iota
	Bottom Direction = iota
	Left   Direction = iota
	Top    Direction = iota
)

const NumDirections = 4

type Product struct {
	subtype   int
	points    int
	resources []int
}

// Scenario is the input to any algorithm that solves Profit!
type Scenario struct {
	width     int
	height    int
	deposits  []Deposit
	obstacles []Obstacle
	products  []Product
	turns     int
	time      int
}

func (s *Scenario) BoundRectangles() []Rectangle {
	return []Rectangle{{Position{0, -1}, s.width, 1, nil},
		{Position{-1, 0}, 1, s.height, nil},
		{Position{s.width, 0}, 1, s.height, nil},
		{Position{0, s.height}, s.width, 1, nil}}
}

func (s *Scenario) InBounds(position Position) bool {
	return !(position.y < 0 || position.y >= s.height || position.x < 0 || position.x >= s.width)
}

func DirectionFromSubtype(subtype int) Direction {
	return Direction(subtype & 3)
}
