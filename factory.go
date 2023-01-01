package main

import (
	"errors"
	"math/rand"
)

const NumProducts = 8

const FactoryWidth = 5
const FactoryHeight = 5

type Factory struct {
	position Position
	product  int
	distance int
}

func (f Factory) Rectangle() Rectangle {
	return Rectangle{
		position: f.position,
		width:    FactoryWidth,
		height:   FactoryHeight,
	}
}

func (f Factory) NextToIngressPositions() []Position {
	positions := make([]Position, 0)
	for i := 0; i < FactoryWidth; i++ {
		positions = append(positions, Position{
			x: f.position.x + i,
			y: f.position.y - 1,
		})
		positions = append(positions, Position{
			x: f.position.x + i,
			y: f.position.y + FactoryHeight,
		})
	}
	for i := 0; i < FactoryHeight; i++ {
		positions = append(positions, Position{
			x: f.position.x - 1,
			y: f.position.y + i,
		})
		positions = append(positions, Position{
			x: f.position.x + FactoryWidth,
			y: f.position.y + i,
		})
	}
	return positions
}

func (f Factory) ingressPositions() []Position {
	positions := make([]Position, 0)
	for i := 0; i < FactoryWidth; i++ {
		positions = append(positions, Position{
			x: f.position.x + i,
			y: f.position.y,
		})
		positions = append(positions, Position{
			x: f.position.x + i,
			y: f.position.y + FactoryHeight - 1,
		})
	}
	for i := 0; i < FactoryHeight; i++ {
		positions = append(positions, Position{
			x: f.position.x,
			y: f.position.y + i,
		})
		positions = append(positions, Position{
			x: f.position.x + FactoryWidth - 1,
			y: f.position.y + i,
		})
	}
	return positions
}

func (s *Scenario) positionAvailableForFactory(factories []Factory, mines []Mine, combiners []Combiner, paths []Path, position Position) bool {
	factoryRectangle := Rectangle{
		position: position,
		width:    FactoryWidth,
		height:   FactoryHeight,
	}
	if position.x+FactoryWidth > s.width || position.y+FactoryHeight > s.height {
		return false
	}
	for _, obstacle := range s.obstacles {
		if factoryRectangle.Intersects(obstacle) {
			return false
		}
	}
	for _, factory := range factories {
		if factoryRectangle.Intersects(factory.Rectangle()) {
			return false
		}
	}
	for _, mine := range mines {
		if mine.Intersects(factoryRectangle) {
			return false
		}
	}
	for _, deposit := range s.deposits {
		depositRectangle := deposit.Rectangle()
		extendedDepositRectangle := Rectangle{
			Position{depositRectangle.position.x - 1, depositRectangle.position.y - 1},
			depositRectangle.width + 2,
			depositRectangle.height + 2,
			nil,
		}
		if factoryRectangle.Intersects(extendedDepositRectangle) {
			// top left
			positionIsCorner := position.y+FactoryHeight == deposit.position.y && position.x+FactoryHeight == deposit.position.x
			// top right
			positionIsCorner = positionIsCorner || (position.y+FactoryHeight == deposit.position.y && position.x == deposit.position.x+deposit.width)
			// bottom left
			positionIsCorner = positionIsCorner || (position.y == deposit.position.y+deposit.height && position.x+FactoryHeight == deposit.position.x)
			// bottom right
			positionIsCorner = positionIsCorner || (position.y == deposit.position.y+deposit.height && position.x == deposit.position.x+deposit.width)
			if !positionIsCorner {
				return false
			}
		}
	}
	for _, combiner := range combiners {
		if combiner.Intersects(factoryRectangle) {
			return false
		}
	}
	for _, path := range paths {
		for _, conveyor := range path.conveyors {
			if factoryRectangle.Intersects(*conveyor.Rectangle()) {
				return false
			}
		}
	}
	return true
}

func (s *Scenario) randomFactory(chromosome Chromosome) (Factory, error) {
	rng := NewUniqueRNG(s.width * s.height)
	var n int
	done := false
	for !done {
		n, done = rng.Next()
		x := n % s.width
		y := n / s.width
		pos := Position{x, y}
		if s.positionAvailableForFactory(chromosome.factories, chromosome.mines, chromosome.combiners, chromosome.paths, pos) {
			subtype := s.products[rand.Intn(len(s.products))].subtype
			return Factory{position: pos, product: subtype}, nil
		}
	}
	return Factory{}, errors.New("no position available for factory")
}
