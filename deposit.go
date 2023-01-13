package main

type Deposit struct {
	position Position
	width    int
	height   int
	subtype  int
}

func (d Deposit) Rectangle() Rectangle {
	return Rectangle{
		position: d.position,
		width:    d.width,
		height:   d.height,
	}
}

func (d Deposit) NextToEgressPositions() []Position {
	positions := make([]Position, 2*d.width+2*d.height)
	for i := 0; i < d.width; i++ {
		positions[i] = Position{
			x: d.position.x + i,
			y: d.position.y - 1,
		}
		positions[d.width+i] = Position{
			x: d.position.x + i,
			y: d.position.y + d.height,
		}
	}
	for i := 0; i < d.height; i++ {
		positions[2*d.width+i] = Position{
			x: d.position.x - 1,
			y: d.position.y + i,
		}
		positions[2*d.width+d.height+i] = Position{
			x: d.position.x + d.width,
			y: d.position.y + i,
		}
	}
	return positions
}

func (d Deposit) EgressPositions() []Position {
	positions := make([]Position, 0)
	for i := 0; i < d.width; i++ {
		positions = append(positions, Position{
			x: d.position.x + i,
			y: d.position.y,
		})
		positions = append(positions, Position{
			x: d.position.x + i,
			y: d.position.y + d.height - 1,
		})
	}
	for i := 0; i < d.height; i++ {
		positions = append(positions, Position{
			x: d.position.x,
			y: d.position.y + i,
		})
		positions = append(positions, Position{
			x: d.position.x + d.width - 1,
			y: d.position.y + i,
		})
	}
	return positions
}
