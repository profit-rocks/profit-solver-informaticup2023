package main

type Position struct {
	x int
	y int
}

type Rectangle struct {
	position  Position
	width     int
	height    int
	positions []Position
}

func (r Rectangle) Equals(rectangle Rectangle) bool {
	return r.position == rectangle.position && r.width == rectangle.width && r.height == rectangle.height
}

func (p Position) NeighborPositions() []Position {
	return []Position{{p.x + 1, p.y}, {p.x - 1, p.y}, {p.x, p.y + 1}, {p.x, p.y - 1}}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (p Position) NextTo(other Position) bool {
	if p == other {
		return false
	}
	xClose := abs(p.x-other.x) <= 1
	yClose := abs(p.y-other.y) <= 1
	return xClose && p.y == other.y || yClose && p.x == other.x
}

func (r Rectangle) Contains(p Position) bool {
	return p.x >= r.position.x && p.x < r.position.x+r.width && p.y >= r.position.y && p.y < r.position.y+r.height
}

func (r Rectangle) Intersects(other Rectangle) bool {
	return r.position.x < other.position.x+other.width && r.position.x+r.width > other.position.x && r.position.y < other.position.y+other.height && r.position.y+r.height > other.position.y
}

func (r Rectangle) ForEach(f func(Position)) {
	for x := r.position.x; x < r.position.x+r.width; x++ {
		for y := r.position.y; y < r.position.y+r.height; y++ {
			f(Position{x, y})
		}
	}
}

func (r *Rectangle) Positions() []Position {
	if r.positions != nil {
		return r.positions
	}
	res := make([]Position, r.width*r.height)
	for x := 0; x < r.width; x++ {
		for y := 0; y < r.height; y++ {
			res[x*r.height+y] = Position{r.position.x + x, r.position.y + y}
		}
	}
	r.positions = res
	return res
}
