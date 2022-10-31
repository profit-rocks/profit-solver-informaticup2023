package main

import (
	"errors"
	"github.com/draffensperger/golp"
)

const NoOptimum = -1

func TheoreticalOptimum(scenario Scenario) (int, error) {
	// variables: number of used resources (8), number of units for each product (8)
	lp := golp.NewLP(32, 16)

	// all variables must take integer arguments
	for i := 0; i < 16; i++ {
		lp.SetInt(i, true)
	}

	// all variables must be non-negative
	for i := 0; i < 16; i++ {
		row := [16]float64{}
		row[i] = 1
		err := lp.AddConstraint(row[:], golp.GE, 0)
		if err != nil {
			return NoOptimum, err
		}
	}

	// constraint: each resource needs to be used as often as the number of product units requires
	for i := 0; i < 8; i++ {
		row := [16]float64{}
		row[i] = -1
		for j := range scenario.products {
			row[8+j] = float64(scenario.products[j].resources[i])
		}
		err := lp.AddConstraint(row[:], golp.EQ, 0)
		if err != nil {
			return NoOptimum, err
		}
	}

	// constraint: total resource requirements must not exceed max resources
	for i := 0; i < 8; i++ {
		row := [16]float64{}
		row[i] = 1
		available := 0
		for _, deposit := range scenario.deposits {
			if deposit.subtype == i {
				available += DepositResourceFactor * deposit.width * deposit.height
			}
		}
		err := lp.AddConstraint(row[:], golp.LE, float64(available))
		if err != nil {
			return NoOptimum, err
		}
	}

	// maximize total points
	objective := [16]float64{}
	for i := range scenario.products {
		objective[8+i] = float64(scenario.products[i].points)
	}
	lp.SetObjFn(objective[:])
	lp.SetMaximize()

	sol := lp.Solve()
	if sol == golp.OPTIMAL {
		return int(lp.Objective()), nil
	}
	return NoOptimum, errors.New("no optimal solution found")
}
