package utils

import "testing"

// test code for func Plot(timeline []int64, plotData [][]float64, plotLabels []string,)
func ExamplePlot() {
	timeline := []int64{1, 2, 3, 4, 5}
	plotData := [][]float64{
		{1, 2, 3, 4, 5},
		{5, 4, 3, 2, 1},
	}
	plotLabels := []string{"label1", "label2"}
	title := "title"
	xlabel := "xlabel"
	ylabel := "ylabel"
	output := "output.png"
	timeFormat := "2006-01-02 15:04:05"

	Plot(timeline, plotData, plotLabels, title, xlabel, ylabel, output, timeFormat)
}

// test code for func Plot(timeline []int64, plotData [][]float64, plotLabels []string,)
func TestPlot(t *testing.T) {
	timeline := []int64{1, 2, 3, 4, 5}
	plotData := [][]float64{
		{1, 2, 3, 4, 5},
		{5, 4, 3, 2, 1},
	}
	plotLabels := []string{"label1", "label2"}
	title := "title"
	xlabel := "xlabel"
	ylabel := "ylabel"
	output := "output.png"
	timeFormat := "2006-01-02 15:04:05"

	testDir, err := InitTestDir("TestPlot")
	if err != nil {
		t.Fatalf("Error initializing test directory: %v", err)
	}
	output = testDir + "/" + output

	Plot(timeline, plotData, plotLabels, title, xlabel, ylabel, output, timeFormat)
	t.Logf("Graph saved to %s", output)
}
