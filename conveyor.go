package main

import (
	"errors"

	"gopkg.in/eapache/queue.v1"
)

const NumConveyorSubtypes = 8
const InfiniteDistance = 1000000

type ConveyorLength int

const (
	Short ConveyorLength = 3
	Long  ConveyorLength = 4
)

type PathEndPosition struct {
	position         Position
	connectedFactory *Factory
	distance         int
}

type Conveyor struct {
	position  Position
	direction Direction
	length    ConveyorLength
	distance  int
	rectangle Rectangle
}

type QueueItem struct {
	conveyor Conveyor
	distance int
}

func ConveyorLengthFromSubtype(subtype int) ConveyorLength {
	return ConveyorLength(3 + subtype>>2)
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

func (c Conveyor) NextToEgressPositions() []Position {
	// return positions clockwise
	p := c.Egress()
	if c.direction == Right {
		return []Position{{p.x, p.y - 1}, {p.x + 1, p.y}, {p.x, p.y + 1}}
	} else if c.direction == Bottom {
		return []Position{{p.x + 1, p.y}, {p.x, p.y + 1}, {p.x - 1, p.y}}
	} else if c.direction == Left {
		return []Position{{p.x, p.y + 1}, {p.x - 1, p.y}, {p.x, p.y - 1}}
	}
	// Top
	return []Position{{p.x - 1, p.y}, {p.x, p.y - 1}, {p.x + 1, p.y}}
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
	return ((int(c.length) - 3) << 2) | int(c.direction)
}

func (c *Conveyor) Rectangle() *Rectangle {
	if c.rectangle.width != 0 {
		return &c.rectangle
	}
	if c.direction == Right || c.direction == Bottom {
		c.rectangle.position = c.Ingress()
	} else {
		c.rectangle.position = c.Egress()
	}

	if c.length == Short {
		if c.direction == Right || c.direction == Left {
			c.rectangle.width = 3
			c.rectangle.height = 1
		} else {
			c.rectangle.width = 1
			c.rectangle.height = 3
		}
	} else if c.length == Long {
		if c.direction == Right || c.direction == Left {
			c.rectangle.width = 4
			c.rectangle.height = 1
		} else {
			c.rectangle.width = 1
			c.rectangle.height = 4
		}
	}
	return &c.rectangle
}

func (c Conveyor) NextToIngressPositions() []Position {
	ingress := c.Ingress()
	if c.direction == Right {
		return []Position{{ingress.x - 1, ingress.y}, {ingress.x, ingress.y - 1}, {ingress.x, ingress.y + 1}}
	} else if c.direction == Bottom {
		return []Position{{ingress.x, ingress.y - 1}, {ingress.x - 1, ingress.y}, {ingress.x + 1, ingress.y}}
	} else if c.direction == Left {
		return []Position{{ingress.x + 1, ingress.y}, {ingress.x, ingress.y - 1}, {ingress.x, ingress.y + 1}}
	}
	//Top
	return []Position{{ingress.x, ingress.y + 1}, {ingress.x - 1, ingress.y}, {ingress.x + 1, ingress.y}}
}

func (c Conveyor) Positions(i int) Position {
	if c.direction == Right || c.direction == Left {
		return Position{c.position.x + i - 1, c.position.y}
	}
	return Position{c.position.x, c.position.y + i - 1}
}

func (s *Scenario) PositionAvailableForConveyor(factories []Factory, mines []Mine, combiners []Combiner, paths []Path, conveyor Conveyor) bool {
	boundRectangles := s.BoundRectangles()
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
		if mine.Intersects(*conveyor.Rectangle()) {
			return false
		}
	}
	for _, combiner := range combiners {
		if combiner.Intersects(*conveyor.Rectangle()) {
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
	if conveyor1.Rectangle().Intersects(*conveyor2.Rectangle()) {
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

// CellInfo contains information on each cell that is used during pathfinding.
type CellInfo struct {
	distance                   int
	blocked                    bool
	isConveyorMiddle           bool
	numEgressNeighbors         int8
	numIngressNeighbors        int8
	currentConveyor            Conveyor
	previousEgress             Position
	blockedByObstacleOrDeposit bool
}

// We allocate a 2D array of CellInfo structs in static storage to decrease the number of dynamic allocations.
var cellInfo [100][100]CellInfo

func populateCellInfoWithIngress(ingress Position, scenario Scenario) {
	for _, p := range ingress.NeighborPositions() {
		if !scenario.InBounds(p) {
			continue
		}
		cellInfo[p.y][p.x].numIngressNeighbors += 1
	}
}

func populateCellInfoWithEgress(egress Position, scenario Scenario) {
	for _, p := range egress.NeighborPositions() {
		if !scenario.InBounds(p) {
			continue
		}
		cellInfo[p.y][p.x].numEgressNeighbors += 1
	}
}

func blockCellInfoWithRectangle(rectangle Rectangle) {
	rectangle.ForEach(func(p Position) {
		cellInfo[p.y][p.x].blocked = true
	})
}

func blockCellInfoWithRectangleOfDepositOrObstacle(rectangle Rectangle) {
	rectangle.ForEach(func(p Position) {
		cellInfo[p.y][p.x].blocked = true
		cellInfo[p.y][p.x].blockedByObstacleOrDeposit = true
	})
}

func populateCellInfoWithConveyor(conveyor Conveyor) {
	conveyor.Rectangle().ForEach(func(p Position) {
		if p != conveyor.Egress() && p != conveyor.Ingress() {
			cellInfo[p.y][p.x].isConveyorMiddle = true
		}
	})
}

func resetCellInfo(scenario Scenario) {
	for y := 0; y < scenario.height; y++ {
		for x := 0; x < scenario.width; x++ {
			cellInfo[y][x].distance = InfiniteDistance
		}
	}
}

func addNewPathToCellInfo(path Path, scenario Scenario) {
	for _, conveyor := range path.conveyors {
		populateCellInfoWithIngress(conveyor.Ingress(), scenario)
		populateCellInfoWithEgress(conveyor.Egress(), scenario)
		populateCellInfoWithConveyor(conveyor)
		blockCellInfoWithRectangle(*conveyor.Rectangle())
	}

}

func initializeCellInfo(scenario Scenario) {
	for _, deposit := range scenario.deposits {
		blockCellInfoWithRectangleOfDepositOrObstacle(deposit.Rectangle())
	}
	for _, obstacle := range scenario.obstacles {
		blockCellInfoWithRectangleOfDepositOrObstacle(obstacle)
	}
}

func populateCellInfoWithNewChromosome(chromosome Chromosome, scenario Scenario) {
	// Cell info has to be initialized for the scenario before we can populate it with the chromosome!
	for y := 0; y < scenario.height; y++ {
		for x := 0; x < scenario.width; x++ {
			if !cellInfo[y][x].blockedByObstacleOrDeposit {
				cellInfo[y][x].blocked = false
				cellInfo[y][x].distance = InfiniteDistance
				cellInfo[y][x].isConveyorMiddle = false
				cellInfo[y][x].numEgressNeighbors = 0
				cellInfo[y][x].numIngressNeighbors = 0
			}
		}
	}

	// keep algorithm from using occupied squares
	for _, deposit := range scenario.deposits {
		for _, p := range deposit.NextToEgressPositions() {
			if !scenario.InBounds(p) {
				continue
			}
			cellInfo[p.y][p.x].numEgressNeighbors += 1
		}
	}
	for _, m := range chromosome.mines {
		populateCellInfoWithIngress(m.Ingress(), scenario)
		populateCellInfoWithEgress(m.Egress(), scenario)

		m.RectanglesEach(func(r Rectangle) {
			blockCellInfoWithRectangle(r)
		})
	}
	for _, c := range chromosome.combiners {
		for _, p := range c.Ingresses() {
			populateCellInfoWithIngress(p, scenario)
		}
		populateCellInfoWithEgress(c.Egress(), scenario)

		c.RectanglesEach(func(r Rectangle) {
			blockCellInfoWithRectangle(r)
		})
	}
	for _, f := range chromosome.factories {
		for _, p := range f.NextToIngressPositions() {
			if !scenario.InBounds(p) {
				continue
			}
			cellInfo[p.y][p.x].numIngressNeighbors += 1
		}
		blockCellInfoWithRectangle(f.Rectangle())
	}
	for _, otherPath := range chromosome.paths {
		for _, conveyor := range otherPath.conveyors {
			populateCellInfoWithIngress(conveyor.Ingress(), scenario)
			populateCellInfoWithEgress(conveyor.Egress(), scenario)
			populateCellInfoWithConveyor(conveyor)
			blockCellInfoWithRectangle(*conveyor.Rectangle())
		}
	}
}

// findPath uses the cellInfo to check if placing a conveyor is legal and to perform backtracking
// make sure to initialize and populate the cellInfo before calling findPath
func findPath(startPosition Position, endPositions []PathEndPosition, scenario Scenario) (Path, int, error) {
	var path Path

	resetCellInfo(scenario)
	// Dummy conveyor used for backtracking
	startConveyor := Conveyor{
		position:  Position{startPosition.x - 1, startPosition.y},
		direction: Right,
		length:    Short,
	}
	q := queue.New()
	startItem := QueueItem{
		conveyor: startConveyor,
		distance: 0,
	}

	// TODO: Conveyors of same path may overlap
	// TODO: Conveyors of same path may violate ingress-egress-rules
	q.Add(&startItem)

	cellInfo[startPosition.y][startPosition.x].distance = 0
	var factory *Factory
	var initialDistance int
	for q.Length() > 0 {
		current := q.Remove().(*QueueItem)
		currentConveyor := current.conveyor
		currentEgress := currentConveyor.Egress()
		finished := false
		for _, p := range endPositions {
			if currentEgress == p.position && cellInfo[currentEgress.y][currentEgress.x].numIngressNeighbors <= 1 {
				if p.position != startPosition {
					path.conveyors = append(path.conveyors, cellInfo[currentEgress.y][currentEgress.x].currentConveyor)
				} else {
					path.conveyors = append(path.conveyors, currentConveyor)
				}
				factory = p.connectedFactory
				initialDistance = p.distance
				finished = true
			}
		}
		if finished {
			break
		}
		if cellInfo[currentEgress.y][currentEgress.x].numIngressNeighbors >= 1 || current.distance != cellInfo[currentEgress.y][currentEgress.x].distance {
			continue
		}
		var nextIngresses []Position
		if currentEgress == startPosition {
			nextIngresses = startPosition.NeighborPositions()
		} else {
			nextIngresses = currentConveyor.NextToEgressPositions()
		}
		for z, nextIngress := range nextIngresses {
			if !scenario.InBounds(nextIngress) {
				continue
			}
			if cellInfo[nextIngress.y][nextIngress.x].numEgressNeighbors >= 1 && currentEgress != startPosition || cellInfo[nextIngress.y][nextIngress.x].numEgressNeighbors >= 2 {
				continue
			}
			for i := 0; i < NumConveyorSubtypes; i++ {
				nextConveyor := ConveyorFromIngressAndSubtype(nextIngress, i)
				if (currentConveyor.overlapsWith(nextConveyor, z) || currentConveyor.buildsLoopWith(nextConveyor, z)) && currentEgress != startPosition {
					continue
				}
				nextEgress := nextConveyor.Egress()
				if !scenario.InBounds(nextEgress) {
					continue
				}
				if current.distance+1 < cellInfo[nextEgress.y][nextEgress.x].distance {
					isBlocked := false
					for m := 0; m < int(nextConveyor.length); m++ {
						p := nextConveyor.Positions(m)
						if !scenario.InBounds(p) {
							isBlocked = true
							break
						}
						if cellInfo[p.y][p.x].blocked {
							if !(cellInfo[p.y][p.x].isConveyorMiddle && p != nextConveyor.Egress() && p != nextConveyor.Ingress()) {
								isBlocked = true
							}
						}
					}
					if isBlocked {
						continue
					}
					next := QueueItem{
						conveyor: nextConveyor,
						distance: current.distance + 1,
					}
					q.Add(&next)
					cellInfo[nextEgress.y][nextEgress.x].previousEgress = currentEgress
					cellInfo[nextEgress.y][nextEgress.x].currentConveyor = nextConveyor
					cellInfo[nextEgress.y][nextEgress.x].distance = next.distance
				}
			}
		}
	}
	if len(path.conveyors) == 0 {
		return path, 0, errors.New("no path found")
	}
	maxDistance := 0
	currentEgress := path.conveyors[0].Egress()
	if currentEgress == startPosition {
		return Path{
			connectedFactory: factory,
		}, maxDistance, nil
	}
	for {
		previousEgress := cellInfo[currentEgress.y][currentEgress.x].previousEgress
		if previousEgress == startPosition {
			break
		}
		path.conveyors = append(path.conveyors, cellInfo[previousEgress.y][previousEgress.x].currentConveyor)
		currentEgress = previousEgress
	}
	for i := range path.conveyors {
		path.conveyors[i].distance = initialDistance + i + 1
		maxDistance = initialDistance + i + 1
	}

	addNewPathToCellInfo(path, scenario)
	path.connectedFactory = factory
	return path, maxDistance, nil
}

func (c Conveyor) overlapsWith(conveyor Conveyor, nextIngressIndex int) bool {
	// The formula calculates the forbidden direction based on our direction and the new ingress position we are on
	return (nextIngressIndex+1+int(c.direction))%4 == int(conveyor.direction)

}

func (c Conveyor) buildsLoopWith(conveyor Conveyor, nextIngressIndex int) bool {
	// Adjacent conveyors with the same length and opposite directions always result in a loop of two conveyors
	return (c.length == conveyor.length || nextIngressIndex == 1) && (c.direction+2)%4 == conveyor.direction
}
