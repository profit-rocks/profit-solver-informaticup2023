package main

import (
	"container/heap"
	"errors"
)

const NumConveyorSubtypes = 8

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

func (s *Scenario) positionAvailableForConveyor(factories []Factory, mines []Mine, paths []Path, conveyor Conveyor) bool {
	boundRectangles := s.boundRectangles()
	for _, rectangle := range boundRectangles {
		if conveyor.Rectangle().Intersects(rectangle) {
			return false
		}
	}
	for _, obstacle := range s.obstacles {
		if conveyor.Rectangle().Intersects(obstacle) {
			return false
		}
	}
	for _, factory := range factories {
		if conveyor.Rectangle().Intersects(factory.Rectangle()) {
			return false
		}
	}
	for _, mine := range mines {
		if mine.Intersects(conveyor.Rectangle()) {
			return false
		}
	}
	for _, deposit := range s.deposits {
		depositRectangle := deposit.Rectangle()
		if conveyor.Rectangle().Intersects(depositRectangle) {
			return false
		}
	}
	for _, path := range paths {
		for _, pathConveyor := range path {
			if conveyor.Rectangle().Intersects(pathConveyor.Rectangle()) {
				return false
			}
		}
	}
	return true
}

func (g *GeneticAlgorithm) getPathMineToFactory(chromosome Chromosome, mine Mine, factory Factory) (Path, error) {
	var path Path

	startPosition := mine.Egress()
	// Dummy conveyor used for backtracking
	startConveyor := Conveyor{
		position:  Position{startPosition.x - 1, startPosition.y},
		direction: Right,
		length:    Short,
	}
	endPositions := factory.nextToIngressPositions()
	queue := PriorityQueue{}
	startItem := Item{
		value:    startConveyor,
		priority: 0,
		index:    0,
	}

	// TODO: Conveyors of same path may overlap
	// TODO: Conveyors of same path may violate ingress-egress-rules
	distances := make([][]int, g.scenario.height)
	for i := range distances {
		distances[i] = make([]int, g.scenario.width)
		for j := range distances[i] {
			distances[i][j] = 1000000
		}
	}
	blocked := make([][]bool, g.scenario.height)
	for i := range blocked {
		blocked[i] = make([]bool, g.scenario.width)
		for j := range blocked[i] {
			blocked[i][j] = false
		}
	}
	blockedForConveyorIngress := make([][]int, g.scenario.height)
	for i := range blockedForConveyorIngress {
		blockedForConveyorIngress[i] = make([]int, g.scenario.width)
		for j := range blockedForConveyorIngress[i] {
			blockedForConveyorIngress[i][j] = 0
		}
	}
	blockedForConveyorEgress := make([][]int, g.scenario.height)
	for i := range blockedForConveyorEgress {
		blockedForConveyorEgress[i] = make([]int, g.scenario.width)
		for j := range blockedForConveyorEgress[i] {
			blockedForConveyorEgress[i][j] = 0
		}
	}
	previousConveyors := make([][]Conveyor, g.scenario.height)
	for i := range previousConveyors {
		previousConveyors[i] = make([]Conveyor, g.scenario.width)
	}

	// keep algorithm from using occupied squares
	for _, deposit := range g.scenario.deposits {
		for _, position := range deposit.mineIngressPositions() {
			if !g.scenario.inBounds(position) {
				continue
			}
			blockedForConveyorIngress[position.y][position.x] += 1
		}
		deposit.Rectangle().ForEach(func(p Position) {
			blocked[p.y][p.x] = true
		})
	}
	for _, obstacle := range g.scenario.obstacles {
		obstacle.ForEach(func(p Position) {
			blocked[p.y][p.x] = true
		})
	}
	for _, m := range chromosome.mines {
		for _, position := range m.Ingress().NeighborPositions() {
			if !g.scenario.inBounds(position) {
				continue
			}
			blockedForConveyorEgress[position.y][position.x] += 1
		}
		for _, position := range m.Egress().NeighborPositions() {
			if !g.scenario.inBounds(position) {
				continue
			}
			blockedForConveyorIngress[position.y][position.x] += 1
		}
		m.RectanglesEach(func(r Rectangle) {
			r.ForEach(func(p Position) {
				blocked[p.y][p.x] = true
			})
		})
	}
	for _, f := range chromosome.factories {
		for _, position := range f.nextToIngressPositions() {
			if !g.scenario.inBounds(position) {
				continue
			}
			blockedForConveyorEgress[position.y][position.x] += 1
		}
		f.Rectangle().ForEach(func(p Position) {
			blocked[p.y][p.x] = true
		})
	}
	for _, otherPath := range chromosome.paths {
		for _, conveyor := range otherPath {
			for _, position := range conveyor.Ingress().NeighborPositions() {
				if !g.scenario.inBounds(position) {
					continue
				}
				blockedForConveyorEgress[position.y][position.x] += 1
			}
			for _, position := range conveyor.Egress().NeighborPositions() {
				if !g.scenario.inBounds(position) {
					continue
				}
				blockedForConveyorIngress[position.y][position.x] += 1
			}
			conveyor.Rectangle().ForEach(func(p Position) {
				blocked[p.y][p.x] = true
			})
		}
	}

	heap.Init(&queue)
	queue.Push(&startItem)
	heap.Fix(&queue, startItem.index)
	distances[startPosition.y][startPosition.x] = 0
	for queue.Len() > 0 {
		current := queue.Pop().(*Item)
		currentConveyor := current.value
		currentEgress := current.value.Egress()
		finished := false
		for _, p := range endPositions {
			if currentEgress == p && blockedForConveyorEgress[currentEgress.y][currentEgress.x] <= 1 {
				path = append(path, current.value)
				finished = true
			}
		}
		if finished {
			break
		}
		if blockedForConveyorEgress[currentEgress.y][currentEgress.x] >= 1 {
			continue
		}
		if current.priority != distances[currentEgress.y][currentEgress.x] {
			continue
		}
		for _, nextIngress := range currentConveyor.EgressNeighborPositions() {
			if !g.scenario.inBounds(nextIngress) {
				continue
			}
			if blockedForConveyorIngress[nextIngress.y][nextIngress.x] >= 1 && currentConveyor.Egress() != startPosition || blockedForConveyorIngress[nextIngress.y][nextIngress.x] >= 2 {
				continue
			}
			for i := 0; i < NumConveyorSubtypes; i++ {
				nextConveyor := ConveyorFromIngressAndSubtype(nextIngress, i)
				nextEgress := nextConveyor.Egress()
				if !g.scenario.inBounds(nextEgress) || nextConveyor.Rectangle().Intersects(currentConveyor.Rectangle()) {
					continue
				}
				if current.priority+1 < distances[nextEgress.y][nextEgress.x] {
					isBlocked := false
					nextConveyor.Rectangle().ForEach(func(p Position) {
						if blocked[p.y][p.x] {
							isBlocked = true
						}
					})
					if isBlocked {
						continue
					}
					next := Item{
						value:    nextConveyor,
						priority: current.priority + 1,
						index:    0,
					}
					queue.Push(&next)
					heap.Fix(&queue, next.index)
					previousConveyors[nextEgress.y][nextEgress.x] = current.value
					distances[nextEgress.y][nextEgress.x] = next.priority
				}
			}
		}
	}
	if len(path) == 0 {
		return path, errors.New("no path found")
	}
	currentEgress := path[0].Egress()
	if currentEgress == startPosition {
		return Path{}, nil
	}
	for {
		conveyor := previousConveyors[currentEgress.y][currentEgress.x]
		if conveyor.Egress() == startPosition {
			break
		}
		path = append(path, conveyor)
		currentEgress = conveyor.Egress()
	}
	// Reverse the path
	var pathMineToFactory Path
	for i := range path {
		pathMineToFactory = append(pathMineToFactory, path[len(path)-i-1])
	}
	return pathMineToFactory, nil
}
