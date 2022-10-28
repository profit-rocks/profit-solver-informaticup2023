package main

import "math"

type Position struct {
	x int
	y int
}

type Rectangle struct {
	position Position
	width    int
	height   int
}

func (p Position) ManhattanDist(other Position) float64 {
	return math.Abs(float64(p.x-other.x)) + math.Abs(float64(p.y-other.y))
}

func (r Rectangle) Contains(p Position) bool {
	return p.x >= r.position.x && p.x < r.position.x+r.width && p.y >= r.position.y && p.y < r.position.y+r.height
}

func (r Rectangle) Intersects(other Rectangle) bool {
	return r.position.x < other.position.x+other.width && r.position.x+r.width > other.position.x && r.position.y < other.position.y+other.height && r.position.y+r.height > other.position.y
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

func (scenario Scenario) boundRectangles() []Rectangle {
	return []Rectangle{{Position{0, -1}, scenario.width, 1},
		{Position{-1, 0}, 1, scenario.height},
		{Position{scenario.width, 0}, 1, scenario.height},
		{Position{0, scenario.height}, scenario.width, 1}}
}
