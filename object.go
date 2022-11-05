package main

const FactoryWidth = 5
const FactoryHeight = 5
const NumConveyorSubtypes = 8

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

	cachedRectangles []Rectangle
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

func ConveyorLengthFromSubtype(subtype int) ConveyorLength {
	return ConveyorLength(subtype >> 2)
}

func ConveyorDirectionFromSubtype(subtype int) Direction {
	return Direction(subtype & 3)
}

func ConveyorFromIngressAndSubtype(ingress Position, subtype int) Conveyor {
	length := ConveyorLengthFromSubtype(subtype)
	direction := ConveyorDirectionFromSubtype(subtype)
	var position Position
	if direction == Right {
		position = Position{ingress.x + 1, ingress.y}
	} else if direction == Bottom {
		position = Position{ingress.x, ingress.y + 1}
	} else if direction == Left {
		if length == Short {
			position = Position{ingress.x - 1, ingress.y}
		} else {
			position = Position{ingress.x - 2, ingress.y}
		}
	} else if direction == Top {
		if length == Short {
			position = Position{ingress.x, ingress.y - 1}
		} else {
			position = Position{ingress.x, ingress.y - 2}
		}
	}
	return Conveyor{
		position:  position,
		direction: direction,
		length:    length,
	}
}

func (c Conveyor) Egress() Position {
	if c.direction == Right {
		if c.length == Short {
			return Position{c.position.x + 1, c.position.y}
		} else {
			return Position{c.position.x + 2, c.position.y}
		}
	} else if c.direction == Bottom {
		if c.length == Short {
			return Position{c.position.x, c.position.y + 1}
		} else {
			return Position{c.position.x, c.position.y + 2}
		}
	} else if c.direction == Left {
		return Position{c.position.x - 1, c.position.y}
	}
	// Top
	return Position{c.position.x, c.position.y - 1}
}

func (c Conveyor) EgressNeighborPositions() []Position {
	p := c.Egress()
	if c.direction == Right {
		return []Position{{p.x + 1, p.y}, {p.x, p.y + 1}, {p.x, p.y - 1}}
	} else if c.direction == Bottom {
		return []Position{{p.x + 1, p.y}, {p.x - 1, p.y}, {p.x, p.y + 1}}
	} else if c.direction == Left {
		return []Position{{p.x - 1, p.y}, {p.x, p.y + 1}, {p.x, p.y - 1}}
	}
	// Top
	return []Position{{p.x + 1, p.y}, {p.x - 1, p.y}, {p.x, p.y - 1}}
}

func (c Conveyor) Ingress() Position {
	if c.direction == Right {
		return Position{c.position.x - 1, c.position.y}
	} else if c.direction == Bottom {
		return Position{c.position.x, c.position.y - 1}
	} else if c.direction == Left {
		if c.length == Short {
			return Position{c.position.x + 1, c.position.y}
		} else {
			return Position{c.position.x + 2, c.position.y}
		}
	}
	// Top
	if c.length == Short {
		return Position{c.position.x, c.position.y + 1}
	}
	return Position{c.position.x, c.position.y + 2}
}

func (c Conveyor) Subtype() int {
	return (int(c.length) << 2) | int(c.direction)
}

func (c Conveyor) Rectangle() Rectangle {
	r := Rectangle{}
	if c.direction == Right || c.direction == Bottom {
		r.position = c.Ingress()
	} else {
		r.position = c.Egress()
	}

	if c.length == Short {
		if c.direction == Right || c.direction == Left {
			r.width = 3
			r.height = 1
		} else {
			r.width = 1
			r.height = 3
		}
	} else if c.length == Long {
		if c.direction == Right || c.direction == Left {
			r.width = 4
			r.height = 1
		} else {
			r.width = 1
			r.height = 4
		}
	}
	return r
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
	paths     []Path
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

func (m *Mine) Egress() Position {
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

func (m *Mine) Ingress() Position {
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

func (m *Mine) RectanglesEach(f func(Rectangle)) {
	switch m.direction {
	case Right:
		f(Rectangle{m.position, 2, 1})
		f(Rectangle{Position{m.position.x - 1, m.position.y + 1}, 4, 1})
	case Bottom:
		f(Rectangle{Position{m.position.x, m.position.y - 1}, 1, 4})
		f(Rectangle{Position{m.position.x + 1, m.position.y}, 1, 2})
	case Left:
		f(Rectangle{Position{m.position.x - 1, m.position.y}, 4, 1})
		f(Rectangle{Position{m.position.x, m.position.y + 1}, 2, 1})
	case Top:
		f(Rectangle{Position{m.position.x, m.position.y}, 1, 2})
		f(Rectangle{Position{m.position.x + 1, m.position.y - 1}, 1, 4})
	}
}

func (m *Mine) Intersects(other Rectangle) bool {
	res := false
	m.RectanglesEach(func(r Rectangle) {
		if r.Intersects(other) {
			res = true
		}
	})
	return res
}

func (m *Mine) IntersectsAny(rectangles []Rectangle) bool {
	res := false
	m.RectanglesEach(func(r1 Rectangle) {
		for _, r2 := range rectangles {
			if r1.Intersects(r2) {
				res = true
			}
		}
	})
	return res
}

func (m *Mine) IntersectsMine(m2 Mine) bool {
	res := false
	m.RectanglesEach(func(r Rectangle) {
		if m2.Intersects(r) {
			res = true
		}
	})
	return res
}

func (f Factory) nextToIngressPositions() []Position {
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

func (s *Scenario) boundRectangles() []Rectangle {
	return []Rectangle{{Position{0, -1}, s.width, 1},
		{Position{-1, 0}, 1, s.height},
		{Position{s.width, 0}, 1, s.height},
		{Position{0, s.height}, s.width, 1}}
}

func (r Rectangle) ForEach(f func(Position)) {
	for x := r.position.x; x < r.position.x+r.width; x++ {
		for y := r.position.y; y < r.position.y+r.height; y++ {
			f(Position{x, y})
		}
	}
}
