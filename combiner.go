package main

type Combiner struct {
	position  Position
	direction Direction
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

func (s *Scenario) positionAvailableForCombiner(factories []Factory, mines []Mine, paths []Path, combiners []Combiner, combiner Combiner) bool {
	// combiner is out of bounds
	boundRectangles := s.boundRectangles()
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
		for _, egress := range combiner.Ingresses() {
			for _, p := range deposit.nextToEgressPositions() {
				if p == egress {
					return false
				}
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
			if combiner.Intersects(conveyor.Rectangle()) {
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
		f(Rectangle{Position{c.position.x - 1, c.position.y - 1}, 2, 3})
		f(Rectangle{Position{c.position.x + 1, c.position.y}, 1, 1})
	case Bottom:
		f(Rectangle{Position{c.position.x - 1, c.position.y - 1}, 3, 2})
		f(Rectangle{Position{c.position.x, c.position.y + 1}, 1, 1})
	case Left:
		f(Rectangle{Position{c.position.x, c.position.y - 1}, 2, 3})
		f(Rectangle{Position{c.position.x - 1, c.position.y}, 1, 1})
	case Top:
		f(Rectangle{Position{c.position.x - 1, c.position.y}, 3, 2})
		f(Rectangle{Position{c.position.x, c.position.y - 1}, 1, 1})
	}
}
