package main

type Position struct {
	x int
	y int
}

type Rectangle struct {
	position Position
	width    int
	height   int
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
	return abs(p.x-other.x) <= 1 && abs(p.y-other.y) <= 1
}

func (p Position) ManhattanDist(other Position) int {
	return abs(p.x-other.x) + abs(p.y-other.y)
}

func (r Rectangle) Contains(p Position) bool {
	return p.x >= r.position.x && p.x < r.position.x+r.width && p.y >= r.position.y && p.y < r.position.y+r.height
}

func (r Rectangle) Intersects(other Rectangle) bool {
	return r.position.x < other.position.x+other.width && r.position.x+r.width > other.position.x && r.position.y < other.position.y+other.height && r.position.y+r.height > other.position.y
}
