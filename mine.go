package main

import (
	"errors"
	"math/rand"
)

type Mine struct {
	position         Position
	direction        Direction
	connectedDeposit *Deposit
	connectedFactory *Factory
	distance         int
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
		f(Rectangle{m.position, 2, 1, nil})
		f(Rectangle{Position{m.position.x - 1, m.position.y + 1}, 4, 1, nil})
	case Bottom:
		f(Rectangle{Position{m.position.x, m.position.y - 1}, 1, 4, nil})
		f(Rectangle{Position{m.position.x + 1, m.position.y}, 1, 2, nil})
	case Left:
		f(Rectangle{Position{m.position.x - 1, m.position.y}, 4, 1, nil})
		f(Rectangle{Position{m.position.x, m.position.y + 1}, 2, 1, nil})
	case Top:
		f(Rectangle{Position{m.position.x, m.position.y}, 1, 2, nil})
		f(Rectangle{Position{m.position.x + 1, m.position.y - 1}, 1, 4, nil})
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

func (m *Mine) NextToIngressPositions() []Position {
	ingress := m.Ingress()
	if m.direction == Right {
		return []Position{{ingress.x - 1, ingress.y}, {ingress.x, ingress.y - 1}, {ingress.x, ingress.y + 1}}
	} else if m.direction == Bottom {
		return []Position{{ingress.x, ingress.y - 1}, {ingress.x - 1, ingress.y}, {ingress.x + 1, ingress.y}}
	} else if m.direction == Left {
		return []Position{{ingress.x + 1, ingress.y}, {ingress.x, ingress.y - 1}, {ingress.x, ingress.y + 1}}
	}
	//Top
	return []Position{{ingress.x, ingress.y + 1}, {ingress.x - 1, ingress.y}, {ingress.x + 1, ingress.y}}
}

func (s *Scenario) positionAvailableForMine(factories []Factory, mines []Mine, combiners []Combiner, paths []Path, mine Mine) bool {
	// mine is out of bounds
	boundRectangles := s.boundRectangles()
	if mine.IntersectsAny(boundRectangles) {
		return false
	}
	if mine.IntersectsAny(s.obstacles) {
		return false
	}
	for _, deposit := range s.deposits {
		if mine.Intersects(deposit.Rectangle()) {
			return false
		}
	}
	for _, factory := range factories {
		if mine.Intersects(factory.Rectangle()) {
			return false
		}
	}
	depositEgress, err := s.attachedDepositEgress(mine)
	for _, otherMine := range mines {
		if mine.Egress().NextTo(otherMine.Ingress()) || mine.Ingress().NextTo(otherMine.Egress()) {
			return false
		}
		if err == nil && otherMine.Ingress().NextTo(depositEgress) {
			return false
		}
		if mine.IntersectsMine(otherMine) {
			return false
		}
	}
	for _, combiner := range combiners {
		foundIntersection := false
		mine.RectanglesEach(func(r Rectangle) {
			if combiner.Intersects(r) {
				foundIntersection = true
			}
		})
		if foundIntersection {
			return false
		}
	}
	for _, path := range paths {
		for _, conveyor := range path.conveyors {
			if mine.Intersects(*conveyor.Rectangle()) {
				return false
			}
		}
	}
	return true
}

func (s *Scenario) minePositions(deposit *Deposit, chromosome Chromosome) []Mine {
	/* For each mine direction, we go counter-clockwise.
	   There is always one case where the mine corner matches the deposit edge.
	   We always use the mine ingress coordinate as our iteration variable */

	positions := make([]Mine, 0)

	// Right
	positions = append(positions, Mine{position: Position{deposit.position.x + deposit.width, deposit.position.y + deposit.height - 1}, direction: Right})
	for i := deposit.position.y + deposit.height - 1; i >= deposit.position.y; i-- {
		positions = append(positions, Mine{position: Position{deposit.position.x + deposit.width + 1, i - 1}, direction: Right})
	}
	for i := deposit.position.x + deposit.width - 1; i >= deposit.position.x; i-- {
		positions = append(positions, Mine{position: Position{i + 1, deposit.position.y - 2}, direction: Right})
	}

	// Bottom
	positions = append(positions, Mine{position: Position{deposit.position.x - 1, deposit.position.y + deposit.height}, direction: Bottom})
	for i := deposit.position.x; i <= deposit.position.x+deposit.width-1; i++ {
		positions = append(positions, Mine{position: Position{i, deposit.position.y + deposit.height + 1}, direction: Bottom})
	}
	for i := deposit.position.y + deposit.height - 1; i >= deposit.position.y; i-- {
		positions = append(positions, Mine{position: Position{deposit.position.x + deposit.width, i + 1}, direction: Bottom})
	}

	// Left
	positions = append(positions, Mine{position: Position{deposit.position.x - 2, deposit.position.y - 1}, direction: Left})
	for i := deposit.position.y; i <= deposit.position.y+deposit.height-1; i++ {
		positions = append(positions, Mine{position: Position{deposit.position.x - 3, i}, direction: Left})
	}
	for i := deposit.position.x; i <= deposit.position.x+deposit.width-1; i++ {
		positions = append(positions, Mine{position: Position{i - 2, deposit.position.y + deposit.height}, direction: Left})
	}

	// Top
	positions = append(positions, Mine{position: Position{deposit.position.x + deposit.width - 1, deposit.position.y - 2}, direction: Top})
	for i := deposit.position.x + deposit.width - 1; i >= deposit.position.x; i-- {
		positions = append(positions, Mine{position: Position{i - 1, deposit.position.y - 3}, direction: Top})
	}
	for i := deposit.position.y; i <= deposit.position.y+deposit.height-1; i++ {
		positions = append(positions, Mine{position: Position{deposit.position.x - 2, i - 2}, direction: Top})
	}

	validPositions := make([]Mine, 0)
	for i := range positions {
		if s.positionAvailableForMine(chromosome.factories, chromosome.mines, chromosome.combiners, chromosome.paths, positions[i]) {
			validPositions = append(validPositions, positions[i])
		}
	}
	return validPositions
}

func (g *GeneticAlgorithm) randomMine(deposit *Deposit, chromosome Chromosome) (Mine, error) {
	availableMines := g.scenario.minePositions(deposit, chromosome)
	if len(availableMines) != 0 {
		randomMine := availableMines[rand.Intn(len(availableMines))]
		randomMine.connectedDeposit = deposit
		return randomMine, nil
	}
	return Mine{}, errors.New("no mines available")
}

func (s *Scenario) attachedDepositEgress(mine Mine) (Position, error) {
	ingress := mine.Ingress()
	for _, deposit := range s.deposits {
		depositRectangle := deposit.Rectangle()
		for _, egressPosition := range []Position{{ingress.x + 1, ingress.y}, {ingress.x - 1, ingress.y}, {ingress.x, ingress.y + 1}, {ingress.x, ingress.y - 1}} {
			if depositRectangle.Contains(egressPosition) {
				return egressPosition, nil
			}
		}
	}
	return Position{}, errors.New("no attached deposit")
}
