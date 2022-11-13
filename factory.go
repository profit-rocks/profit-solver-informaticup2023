package main

import (
	"errors"
	"math/rand"
)

const FactoryWidth = 5
const FactoryHeight = 5

type Factory struct {
	position Position
	product  int
}

func (f Factory) Rectangle() Rectangle {
	return Rectangle{
		position: f.position,
		width:    FactoryWidth,
		height:   FactoryHeight,
	}
}

func (f Factory) nextToIngressPositions() []Position {
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

func (s *Scenario) positionAvailableForFactory(factories []Factory, mines []Mine, paths []Path, position Position) bool {
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
	for _, path := range paths {
		for _, conveyor := range path.conveyors {
			if factoryRectangle.Intersects(conveyor.Rectangle()) {
				return false
			}
		}
	}
	return true
}

func (s *Scenario) factoryPositions(chromosome Chromosome) []Position {
	positions := make([]Position, 0)
	for i := 0; i < s.width; i++ {
		for j := 0; j < s.height; j++ {
			pos := Position{i, j}
			if s.positionAvailableForFactory(chromosome.factories, chromosome.mines, chromosome.paths, pos) {
				positions = append(positions, pos)
			}
		}
	}
	return positions
}

func (s *Scenario) randomFactory(chromosome Chromosome) (Factory, error) {
	availablePositions := s.factoryPositions(chromosome)
	if len(availablePositions) == 0 {
		return Factory{}, errors.New("no factory positions available")
	}
	position := availablePositions[rand.Intn(len(availablePositions))]
	subtype := s.products[rand.Intn(len(s.products))].subtype
	return Factory{position: position, product: subtype}, nil
}
