package chartings

import (
	"bytes"
	"fmt"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func Plot(insidePts, outsidePts plotter.XYs) (*bytes.Buffer, error) {
	if len(insidePts) == 0 || len(outsidePts) == 0 {
		return nil, fmt.Errorf("no data points to plot")
	}

	// Create a new plot.
	p := plot.New()
	p.Title.Text = "Inside vs Outside Temperature"
	p.X.Label.Text = "Time (minutes)"
	p.Y.Label.Text = "Temperature (°C)"
	p.Legend.TextStyle.Font.Typeface = "Ubuntu-Regular"
	p.Title.TextStyle.Font.Typeface = "Ubuntu-Bold"

	// Add the inside and outside temperature lines.
	err := plotutil.AddLines(p,
		"Inside", insidePts,
		"Outside", outsidePts)
	if err != nil {
		panic(err)
	}

	// Max and min temperature values
	maxInsideTemp := -1000.0
	for _, pt := range insidePts {
		if pt.Y > maxInsideTemp {
			maxInsideTemp = pt.Y
		}
	}

	// Save the plot to a PNG file.
	if err := p.Save(10*vg.Inch, 4*vg.Inch, "temperature_plot.png"); err != nil {
		panic(err)
	}

	// Render to *bytes.Buffer
	buf := new(bytes.Buffer)
	wt, err := p.WriterTo(6*vg.Inch, 3*vg.Inch, "png")
	if err != nil {
		panic(err)
	}
	wt.WriteTo(buf)

	return buf, nil
}
