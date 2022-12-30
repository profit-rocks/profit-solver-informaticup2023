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

type PositionFactory struct {
	position Position
	factory  *Factory
	distance int
}

type Conveyor struct {
	position  Position
	direction Direction
	length    ConveyorLength
	distance  int
}

func ConveyorLengthFromSubtype(subtype int) ConveyorLength {
	return ConveyorLength(subtype >> 2)
}

func ConveyorFromIngressAndSubtype(ingress Position, subtype int) Conveyor {
	length := ConveyorLengthFromSubtype(subtype)
	direction := DirectionFromSubtype(subtype)
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

func (c Conveyor) NextToIngressPositions() []Position {
	return c.Ingress().NeighborPositions()
}

func (s *Scenario) positionAvailableForConveyor(factories []Factory, mines []Mine, combiners []Combiner, paths []Path, conveyor Conveyor) bool {
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
	for _, combiner := range combiners {
		if combiner.Intersects(conveyor.Rectangle()) {
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
		for _, pathConveyor := range path.conveyors {
			isValidOverlap := checkOverlapIsValid(conveyor, pathConveyor)
			if !isValidOverlap {
				return false
			}
		}
	}
	return true
}

func checkOverlapIsValid(conveyor1 Conveyor, conveyor2 Conveyor) bool {
	isValidOverlap := true
	if conveyor1.Rectangle().Intersects(conveyor2.Rectangle()) {
		conveyor2.Rectangle().ForEach(func(p Position) {
			if p == conveyor1.Egress() || p == conveyor1.Ingress() {
				isValidOverlap = false
			}
		})
		conveyor1.Rectangle().ForEach(func(p Position) {
			if p == conveyor2.Egress() || p == conveyor2.Ingress() {
				isValidOverlap = false
			}
		})
	}
	return isValidOverlap
}

// CellInfo contains information on each cell that is used while pathfinding.
type CellInfo struct {
	distance            int
	blocked             bool
	isConveyorMiddle    bool
	numEgressNeighbors  int8
	numIngressNeighbors int8
	previousConveyor    Conveyor
}

// We allocate a 2D array of CellInfo structs in static storage to decrease the number of dynamic allocations.
var cellInfo [100][100]CellInfo

func (g *GeneticAlgorithm) populateCellInfoWithIngress(ingress Position) {
	for _, p := range ingress.NeighborPositions() {
		if !g.scenario.inBounds(p) {
			continue
		}
		cellInfo[p.y][p.x].numIngressNeighbors += 1
	}
}

func (g *GeneticAlgorithm) populateCellInfoWithEgress(egress Position) {
	for _, p := range egress.NeighborPositions() {
		if !g.scenario.inBounds(p) {
			continue
		}
		cellInfo[p.y][p.x].numEgressNeighbors += 1
	}
}

func (g *GeneticAlgorithm) blockCellInfoWithRectangle(rectangle Rectangle) {
	rectangle.ForEach(func(p Position) {
		cellInfo[p.y][p.x].blocked = true
	})
}

func (g *GeneticAlgorithm) populateCellInfoWithConveyor(conveyor Conveyor) {
	conveyor.Rectangle().ForEach(func(p Position) {
		if p != conveyor.Egress() && p != conveyor.Ingress() {
			cellInfo[p.y][p.x].isConveyorMiddle = true
		}
	})
}

func (g *GeneticAlgorithm) populateCellInfo(chromosome Chromosome) {
	for y := 0; y < g.scenario.height; y++ {
		for x := 0; x < g.scenario.width; x++ {
			cellInfo[y][x] = CellInfo{
				distance:            1000000,
				blocked:             false,
				numIngressNeighbors: 0,
				numEgressNeighbors:  0,
			}
		}
	}

	// keep algorithm from using occupied squares
	for _, deposit := range g.scenario.deposits {
		for _, p := range deposit.nextToEgressPositions() {
			if !g.scenario.inBounds(p) {
				continue
			}
			cellInfo[p.y][p.x].numEgressNeighbors += 1
		}
		g.blockCellInfoWithRectangle(deposit.Rectangle())
	}
	for _, obstacle := range g.scenario.obstacles {
		g.blockCellInfoWithRectangle(obstacle)
	}
	for _, m := range chromosome.mines {
		g.populateCellInfoWithIngress(m.Ingress())
		g.populateCellInfoWithEgress(m.Egress())

		m.RectanglesEach(func(r Rectangle) {
			g.blockCellInfoWithRectangle(r)
		})
	}
	for _, c := range chromosome.combiners {
		for _, p := range c.Ingresses() {
			g.populateCellInfoWithIngress(p)
		}
		g.populateCellInfoWithEgress(c.Egress())

		c.RectanglesEach(func(r Rectangle) {
			g.blockCellInfoWithRectangle(r)
		})
	}
	for _, f := range chromosome.factories {
		for _, p := range f.nextToIngressPositions() {
			if !g.scenario.inBounds(p) {
				continue
			}
			cellInfo[p.y][p.x].numIngressNeighbors += 1
		}
		g.blockCellInfoWithRectangle(f.Rectangle())
	}
	for _, otherPath := range chromosome.paths {
		for _, conveyor := range otherPath.conveyors {
			g.populateCellInfoWithIngress(conveyor.Ingress())
			g.populateCellInfoWithEgress(conveyor.Egress())
			g.populateCellInfoWithConveyor(conveyor)
			g.blockCellInfoWithRectangle(conveyor.Rectangle())
		}
	}
}

func (g *GeneticAlgorithm) path(chromosome Chromosome, startPosition Position, endPositions []PositionFactory) (Path, *Factory, int, error) {
	var path Path

	g.populateCellInfo(chromosome)

	// Dummy conveyor used for backtracking
	startConveyor := Conveyor{
		position:  Position{startPosition.x - 1, startPosition.y},
		direction: Right,
		length:    Short,
	}
	queue := PriorityQueue{}
	startItem := Item{
		value:    startConveyor,
		priority: 0,
	}

	// TODO: Conveyors of same path may overlap
	// TODO: Conveyors of same path may violate ingress-egress-rules
	heap.Init(&queue)
	queue.Push(&startItem)

	cellInfo[startPosition.y][startPosition.x].distance = 0
	var factory *Factory
	var initialDistance int
	for queue.Len() > 0 {
		current := queue.Pop().(*Item)
		currentConveyor := current.value
		currentEgress := current.value.Egress()
		finished := false
		for _, p := range endPositions {
			if currentEgress == p.position && cellInfo[currentEgress.y][currentEgress.x].numIngressNeighbors <= 1 {
				path.conveyors = append(path.conveyors, current.value)
				factory = p.factory
				initialDistance = p.distance
				finished = true
			}
		}
		if finished {
			break
		}
		if cellInfo[currentEgress.y][currentEgress.x].numIngressNeighbors >= 1 || current.priority != cellInfo[currentEgress.y][currentEgress.x].distance {
			continue
		}
		for _, nextIngress := range currentConveyor.EgressNeighborPositions() {
			if !g.scenario.inBounds(nextIngress) {
				continue
			}
			if cellInfo[nextIngress.y][nextIngress.x].numEgressNeighbors >= 1 && currentConveyor.Egress() != startPosition || cellInfo[nextIngress.y][nextIngress.x].numEgressNeighbors >= 2 {
				continue
			}
			for i := 0; i < NumConveyorSubtypes; i++ {
				nextConveyor := ConveyorFromIngressAndSubtype(nextIngress, i)
				nextEgress := nextConveyor.Egress()
				if !g.scenario.inBounds(nextEgress) || nextConveyor.Rectangle().Intersects(currentConveyor.Rectangle()) {
					continue
				}
				if current.priority+1 < cellInfo[nextEgress.y][nextEgress.x].distance {
					isBlocked := false
					nextConveyor.Rectangle().ForEach(func(p Position) {
						if cellInfo[p.y][p.x].blocked {
							if !(cellInfo[p.y][p.x].isConveyorMiddle && p != nextConveyor.Egress() && p != nextConveyor.Ingress()) {
								isBlocked = true
							}
						}
					})
					if isBlocked {
						continue
					}
					next := Item{
						value:    nextConveyor,
						priority: current.priority + 1,
					}
					queue.Push(&next)
					cellInfo[nextEgress.y][nextEgress.x].previousConveyor = current.value
					cellInfo[nextEgress.y][nextEgress.x].distance = next.priority
				}
			}
		}
	}
	if len(path.conveyors) == 0 {
		return path, nil, 0, errors.New("no path found")
	}
	maxDistance := 0
	currentEgress := path.conveyors[0].Egress()
	if currentEgress == startPosition {
		return Path{}, factory, maxDistance, nil
	}
	for {
		conveyor := cellInfo[currentEgress.y][currentEgress.x].previousConveyor
		if conveyor.Egress() == startPosition {
			break
		}
		path.conveyors = append(path.conveyors, conveyor)
		currentEgress = conveyor.Egress()
	}
	for i := range path.conveyors {
		path.conveyors[i].distance = initialDistance + i + 1
		maxDistance = initialDistance + i + 1
	}
	// Reverse the path
	var pathMineToFactory Path
	for i := range path.conveyors {
		pathMineToFactory.conveyors = append(pathMineToFactory.conveyors, path.conveyors[len(path.conveyors)-i-1])
	}
	return pathMineToFactory, factory, maxDistance, nil
}
