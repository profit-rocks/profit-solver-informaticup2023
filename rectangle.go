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
