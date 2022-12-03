package main

import (
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette/moreland"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// Adapted from: https://medium.com/@balazs.dianiska/generating-heatmaps-with-go-83988b22c000
type plottable struct {
	grid   [][]int
	width  int
	height int
}

func (p plottable) Dims() (c, r int) {
	return p.width, p.height
}
func (p plottable) X(c int) float64 {
	return float64(c)
}
func (p plottable) Y(r int) float64 {
	return float64(r)
}
func (p plottable) Z(c, r int) float64 {
	return float64(p.grid[c][r])
}

func (g GeneticAlgorithm) visualizeChromosomes(chromosomes []Chromosome, iteration int) {
	factoryFrequencies := make([][]int, g.scenario.width)
	for i := range factoryFrequencies {
		factoryFrequencies[i] = make([]int, g.scenario.height)
	}

	for _, c := range chromosomes {
		for _, f := range c.factories {
			f.Rectangle().ForEach(func(p Position) {
				factoryFrequencies[p.x][p.y] += 1
			})
		}
	}

	plotData := plottable{
		grid:   factoryFrequencies,
		width:  g.scenario.width,
		height: g.scenario.height,
	}
	p := plot.New()
	pal := moreland.SmoothBlueRed().Palette(255)
	hm := plotter.NewHeatMap(plotData, pal)
	p.Add(hm)
	err := p.Save(10*vg.Inch, 10*vg.Inch, fmt.Sprintf("visuals/iteration_%d_f.png", iteration))
	if err != nil {
		return
	}
}
