package main

import (
	"errors"
)

type Combiner struct {
	position         Position
	direction        Direction
	connectedFactory *Factory
	distance         int
}

func (c *Combiner) Ingresses() []Position {
	if c.direction == Right {
		return []Position{{c.position.x - 1, c.position.y}, {c.position.x - 1, c.position.y - 1}, {c.position.x - 1, c.position.y + 1}}
	} else if c.direction == Bottom {
		return []Position{{c.position.x, c.position.y - 1}, {c.position.x - 1, c.position.y - 1}, {c.position.x + 1, c.position.y - 1}}
	} else if c.direction == Left {
		return []Position{{c.position.x + 1, c.position.y}, {c.position.x + 1, c.position.y - 1}, {c.position.x + 1, c.position.y + 1}}
	}
	// Top
	return []Position{{c.position.x, c.position.y + 1}, {c.position.x - 1, c.position.y + 1}, {c.position.x + 1, c.position.y + 1}}
}

func (c *Combiner) NextToIngressPositions() []Position {
	if c.direction == Right {
		return []Position{{c.position.x - 2, c.position.y}, {c.position.x - 2, c.position.y - 1}, {c.position.x - 2, c.position.y + 1}, {c.position.x - 1, c.position.y - 2}, {c.position.x - 1, c.position.y + 2}}
	} else if c.direction == Bottom {
		return []Position{{c.position.x, c.position.y - 2}, {c.position.x - 1, c.position.y - 2}, {c.position.x + 1, c.position.y - 2}, {c.position.x - 2, c.position.y - 1}, {c.position.x + 2, c.position.y - 1}}
	} else if c.direction == Left {
		return []Position{{c.position.x + 2, c.position.y}, {c.position.x + 2, c.position.y - 1}, {c.position.x + 2, c.position.y + 1}, {c.position.x + 1, c.position.y - 2}, {c.position.x + 1, c.position.y + 2}}
	}
	// Top
	return []Position{{c.position.x, c.position.y + 2}, {c.position.x - 1, c.position.y + 2}, {c.position.x + 1, c.position.y + 2}, {c.position.x - 2, c.position.y + 1}, {c.position.x + 2, c.position.y + 1}}
}

func (c *Combiner) Egress() Position {
	if c.direction == Right {
		return Position{c.position.x + 1, c.position.y}
	} else if c.direction == Bottom {
		return Position{c.position.x, c.position.y + 1}
	} else if c.direction == Left {
		return Position{c.position.x - 1, c.position.y}
	}
	// Top
	return Position{c.position.x, c.position.y - 1}
}

func (c *Combiner) NextToIngressRectangles() []Rectangle {
	if c.direction == Right {
		return []Rectangle{
			{Position{c.position.x - 2, c.position.y - 1}, 1, 3, nil},
			{Position{c.position.x - 1, c.position.y - 2}, 1, 1, nil},
			{Position{c.position.x - 1, c.position.y + 2}, 1, 1, nil},
		}
	} else if c.direction == Bottom {
		return []Rectangle{
			{Position{c.position.x - 1, c.position.y - 2}, 3, 1, nil},
			{Position{c.position.x - 2, c.position.y - 1}, 1, 1, nil},
			{Position{c.position.x + 2, c.position.y - 1}, 1, 1, nil},
		}
	} else if c.direction == Left {
		return []Rectangle{
			{Position{c.position.x + 2, c.position.y - 1}, 1, 3, nil},
			{Position{c.position.x + 1, c.position.y + 2}, 1, 1, nil},
			{Position{c.position.x + 1, c.position.y - 2}, 1, 1, nil},
		}
	}
	// Top
	return []Rectangle{
		{Position{c.position.x - 1, c.position.y + 2}, 3, 1, nil},
		{Position{c.position.x - 2, c.position.y + 1}, 1, 1, nil},
		{Position{c.position.x + 2, c.position.y + 1}, 1, 1, nil},
	}
}

func (s *Scenario) PositionAvailableForCombiner(factories []Factory, mines []Mine, paths []Path, combiners []Combiner, combiner Combiner) bool {
	// combiner is out of bounds
	boundRectangles := s.BoundRectangles()
	if combiner.IntersectsAny(boundRectangles) {
		return false
	}
	if combiner.IntersectsAny(s.obstacles) {
		return false
	}
	for _, deposit := range s.deposits {
		if combiner.Intersects(deposit.Rectangle()) {
			return false
		}
		for _, rectangle := range combiner.NextToIngressRectangles() {
			if deposit.Rectangle().Intersects(rectangle) {
				return false
			}
		}
	}
	for _, factory := range factories {
		if combiner.Intersects(factory.Rectangle()) {
			return false
		}
	}
	for _, mine := range mines {
		foundIntersection := false
		combiner.RectanglesEach(func(r Rectangle) {
			if mine.Intersects(r) {
				foundIntersection = true
			}
		})
		if foundIntersection {
			return false
		}
	}
	for _, c := range combiners {
		foundIntersection := false
		combiner.RectanglesEach(func(r Rectangle) {
			if c.Intersects(r) {
				foundIntersection = true
			}
		})
		if foundIntersection {
			return false
		}
	}
	for _, path := range paths {
		for _, conveyor := range path.conveyors {
			if combiner.Intersects(*conveyor.Rectangle()) {
				return false
			}
		}
	}
	return true
}

func (c *Combiner) Intersects(other Rectangle) bool {
	res := false
	c.RectanglesEach(func(r Rectangle) {
		if r.Intersects(other) {
			res = true
		}
	})
	return res
}

func (c *Combiner) IntersectsAny(rectangles []Rectangle) bool {
	res := false
	c.RectanglesEach(func(r1 Rectangle) {
		for _, r2 := range rectangles {
			if r1.Intersects(r2) {
				res = true
			}
		}
	})
	return res
}

func (c *Combiner) RectanglesEach(f func(Rectangle)) {
	switch c.direction {
	case Right:
		f(Rectangle{Position{c.position.x - 1, c.position.y - 1}, 2, 3, nil})
		f(Rectangle{Position{c.position.x + 1, c.position.y}, 1, 1, nil})
	case Bottom:
		f(Rectangle{Position{c.position.x - 1, c.position.y - 1}, 3, 2, nil})
		f(Rectangle{Position{c.position.x, c.position.y + 1}, 1, 1, nil})
	case Left:
		f(Rectangle{Position{c.position.x, c.position.y - 1}, 2, 3, nil})
		f(Rectangle{Position{c.position.x - 1, c.position.y}, 1, 1, nil})
	case Top:
		f(Rectangle{Position{c.position.x - 1, c.position.y}, 3, 2, nil})
		f(Rectangle{Position{c.position.x, c.position.y - 1}, 1, 1, nil})
	}
}

func (s *Scenario) RandomCombiner(chromosome Chromosome) (Combiner, error) {
	rng := NewUniqueRNG(s.width * s.height)
	var n int
	done := false
	for !done {
		n, done = rng.Next()
		x := n % s.width
		y := n / s.width
		pos := Position{x, y}
		directionRng := NewUniqueRNG(4)
		for i := 0; i < 4; i++ {
			direction, _ := directionRng.Next()
			combiner := Combiner{position: pos, direction: Direction(direction)}
			if s.PositionAvailableForCombiner(chromosome.factories, chromosome.mines, chromosome.paths, chromosome.combiners, combiner) {
				return combiner, nil
			}
		}
	}
	return Combiner{}, errors.New("no position available for factory")
}
