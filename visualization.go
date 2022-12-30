package main

import (
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette/moreland"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"os"
)

const NumLoggedChromosomesPerIteration = 5

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

func (g *GeneticAlgorithm) visualizeChromosomes(chromosomes []Chromosome, iteration int, dir string) error {
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return err
	}
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
	err = saveGrid(factoryFrequencies, fmt.Sprintf("%s/f_iteration_%d.png", dir, iteration))
	if err != nil {
		return err
	}
	err = saveGrid(combinerFrequencies, fmt.Sprintf("%s/com_iteration_%d.png", dir, iteration))
	if err != nil {
		return err
	}
	err = saveGrid(conveyorFrequencies, fmt.Sprintf("%s/con_iteration_%d.png", dir, iteration))
	if err != nil {
		return err
	}
	return saveGrid(mineFrequencies, fmt.Sprintf("%s/m_iteration_%d.png", dir, iteration))
}

func saveGrid(grid [][]int, path string) error {
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
	return p.Save(vg.Length(width)*vg.Inch, vg.Length(height)*vg.Inch, path)
}

func exportChromosomes(scenario Scenario, i int, chromosomes []Chromosome, dir string) error {
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return err
	}
	for j := 0; j < NumLoggedChromosomesPerIteration; j++ {
		if j < len(chromosomes) {
			err = exportSolution(scenario, chromosomes[j].Solution(), fmt.Sprintf("%s/iteration_%d_ch_%d.json", dir, i, j))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
