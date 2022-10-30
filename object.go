package main

const FactoryWidth = 5
const FactoryHeight = 5

type Deposit struct {
	position Position
	width    int
	height   int
	subtype  int
}

type Obstacle = Rectangle

type Factory struct {
	position Position
	product  int
}

// Direction is the relative position of the egress
type Direction int

const (
	Right  Direction = iota
	Bottom Direction = iota
	Left   Direction = iota
	Top    Direction = iota
)

type Mine struct {
	position  Position
	direction Direction
}

type Product struct {
	subtype   int
	points    int
	resources []int
}

type ConveyorLength int

const (
	Short ConveyorLength = iota
	Long  ConveyorLength = iota
)

type Conveyor struct {
	position  Position
	direction Direction
	length    ConveyorLength
}

func (c Conveyor) Subtype() int {
	return (int(c.length) << 2) | int(c.direction)
}

// Scenario is the input to any algorithm that solves Profit!
type Scenario struct {
	width     int
	height    int
	deposits  []Deposit
	obstacles []Obstacle
	products  []Product
	turns     int
}

// Solution is the output of any algorithm that solves Profit!
type Solution struct {
	factories []Factory
	mines     []Mine
	conveyors []Conveyor
}

func (d Deposit) Rectangle() Rectangle {
	return Rectangle{
		position: d.position,
		width:    d.width,
		height:   d.height,
	}
}

func (f Factory) Rectangle() Rectangle {
	return Rectangle{
		position: f.position,
		width:    FactoryWidth,
		height:   FactoryHeight,
	}
}

func (m Mine) Egress() Position {
	if m.direction == Right {
		return Position{m.position.x + 2, m.position.y + 1}
	} else if m.direction == Bottom {
		return Position{m.position.x, m.position.y + 2}
	} else if m.direction == Left {
		return Position{m.position.x - 1, m.position.y}
	}
	// Top
	return Position{m.position.x + 1, m.position.y - 1}
}

func (m Mine) Ingress() Position {
	if m.direction == Right {
		return Position{m.position.x - 1, m.position.y + 1}
	} else if m.direction == Bottom {
		return Position{m.position.x, m.position.y - 1}
	} else if m.direction == Left {
		return Position{m.position.x + 2, m.position.y}
	}
	// Top
	return Position{m.position.x + 1, m.position.y + 2}
}

func (m Mine) Rectangles() []Rectangle {
	switch m.direction {
	case Right:
		return []Rectangle{{m.position, 2, 1}, {Position{m.position.x - 1, m.position.y + 1}, 4, 1}}
	case Bottom:
		return []Rectangle{{Position{m.position.x, m.position.y - 1}, 1, 4}, {Position{m.position.x + 1, m.position.y}, 1, 2}}
	case Left:
		return []Rectangle{{Position{m.position.x - 1, m.position.y}, 4, 1}, {Position{m.position.x, m.position.y + 1}, 2, 1}}
	case Top:
		return []Rectangle{{Position{m.position.x, m.position.y}, 1, 2}, {Position{m.position.x + 1, m.position.y - 1}, 1, 4}}
	}
	return []Rectangle{}
}

func (m Mine) Intersects(other Rectangle) bool {
	for _, r := range m.Rectangles() {
		if r.Intersects(other) {
			return true
		}
	}
	return false
}

func (f Factory) mineEgressPositions() []Position {
	positions := make([]Position, 0)
	for i := 0; i < FactoryWidth; i++ {
		positions = append(positions, Position{
			x: f.position.x + i,
			y: f.position.y - 1,
		})
		positions = append(positions, Position{
			x: f.position.x + i,
			y: f.position.y + FactoryHeight,
		})
	}
	for i := 0; i < FactoryHeight; i++ {
		positions = append(positions, Position{
			x: f.position.x - 1,
			y: f.position.y + i,
		})
		positions = append(positions, Position{
			x: f.position.x + FactoryWidth,
			y: f.position.y + i,
		})
	}
	return positions
}

func (d Deposit) mineIngressPositions() []Position {
	positions := make([]Position, 0)
	for i := 0; i < d.width; i++ {
		positions = append(positions, Position{
			x: d.position.x + i,
			y: d.position.y - 1,
		})
		positions = append(positions, Position{
			x: d.position.x + i,
			y: d.position.y + d.height,
		})
	}
	for i := 0; i < d.height; i++ {
		positions = append(positions, Position{
			x: d.position.x - 1,
			y: d.position.y + i,
		})
		positions = append(positions, Position{
			x: d.position.x + d.width,
			y: d.position.y + i,
		})
	}
	return positions
}

func (s Scenario) boundRectangles() []Rectangle {
	return []Rectangle{{Position{0, -1}, s.width, 1},
		{Position{-1, 0}, 1, s.height},
		{Position{s.width, 0}, 1, s.height},
		{Position{0, s.height}, s.width, 1}}
}
