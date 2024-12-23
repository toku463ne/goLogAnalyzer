package utils

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// make plots. x-axis: time, y-axis: count
func Plot(timeline []int64, plotData [][]float64, plotLabels []string,
	title, xlabel, ylabel, output, timeFormat string) error {
	// Create a new plot
	p := plot.New()
	// Set title
	p.Title.Text = title
	// Set x-axis label
	p.X.Label.Text = xlabel
	// Set y-axis label
	p.Y.Label.Text = ylabel
	// Set x-axis time format
	p.X.Tick.Marker = plot.TimeTicks{Format: timeFormat}
	// Add data
	for i, data := range plotData {
		// Create a new line plot
		pts := make(plotter.XYs, len(data))
		for j := range data {
			pts[j].X = float64(timeline[j])
			pts[j].Y = data[j]
		}
		l, err := plotter.NewLine(pts)
		if err != nil {
			return err
		}
		// Set line plot label
		// Add line plot to the plot
		p.Add(l)
		// Add line plot to the legend
		p.Legend.Add(plotLabels[i], l)
	}
	// Set legend
	p.Legend.Top = true
	p.Legend.Left = true
	// Save the plot to a file
	if err := p.Save(8*vg.Inch, 4*vg.Inch, output); err != nil {
		return err
	}
	return nil
}
