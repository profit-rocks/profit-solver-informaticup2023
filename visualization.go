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
	combinerFrequencies := make([][]int, g.scenario.width)
	for i := range combinerFrequencies {
		combinerFrequencies[i] = make([]int, g.scenario.height)
	}
	conveyorFrequencies := make([][]int, g.scenario.width)
	for i := range conveyorFrequencies {
		conveyorFrequencies[i] = make([]int, g.scenario.height)
	}
	mineFrequencies := make([][]int, g.scenario.width)
	for i := range mineFrequencies {
		mineFrequencies[i] = make([]int, g.scenario.height)
	}
	for _, c := range chromosomes {
		for _, f := range c.factories {
			f.Rectangle().ForEach(func(p Position) {
				factoryFrequencies[p.x][g.scenario.height-1-p.y] += 1
			})
		}
		for _, combiner := range c.combiners {
			combiner.RectanglesEach(func(r Rectangle) {
				r.ForEach(func(p Position) {
					combinerFrequencies[p.x][g.scenario.height-1-p.y] += 1
				})
			})
		}
		for _, path := range c.paths {
			for _, conveyor := range path.conveyors {
				conveyor.Rectangle().ForEach(func(p Position) {
					conveyorFrequencies[p.x][g.scenario.height-1-p.y] += 1
				})
			}
		}
		for _, mine := range c.mines {
			mine.RectanglesEach(func(r Rectangle) {
				r.ForEach(func(p Position) {
					mineFrequencies[p.x][g.scenario.height-1-p.y] += 1
				})
			})
		}
	}
	saveGrid(factoryFrequencies, fmt.Sprintf("visuals/f_iteration_%d.png", iteration))
	saveGrid(combinerFrequencies, fmt.Sprintf("visuals/com_iteration_%d.png", iteration))
	saveGrid(conveyorFrequencies, fmt.Sprintf("visuals/con_iteration_%d.png", iteration))
	saveGrid(mineFrequencies, fmt.Sprintf("visuals/m_iteration_%d.png", iteration))
}

func saveGrid(grid [][]int, path string) {
	width := len(grid)
	height := len(grid[0])
	plotData := plottable{
		grid:   grid,
		width:  width,
		height: height,
	}
	p := plot.New()
	pal := moreland.SmoothBlueRed().Palette(255)
	hm := plotter.NewHeatMap(plotData, pal)
	p.Add(hm)
	err := p.Save(vg.Length(width)*vg.Inch, vg.Length(height)*vg.Inch, path)
	if err != nil {
		return
	}
}
