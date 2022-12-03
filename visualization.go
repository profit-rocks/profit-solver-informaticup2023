package main

import (
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

func visualizeChromosomes(chromosomes []Chromosome) {
	plotData := plottable{
		grid:   [][]int{{1, 2}, {2, 4}},
		width:  2,
		height: 2,
	}
	p := plot.New()
	pal := moreland.SmoothBlueRed().Palette(255)
	hm := plotter.NewHeatMap(plotData, pal)
	p.Add(hm)
	err := p.Save(4*vg.Inch, 4*vg.Inch, "points.png")
	if err != nil {
		return
	}
}
