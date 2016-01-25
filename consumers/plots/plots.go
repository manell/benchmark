package statistics

import (
	"time"

	"image/color"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	"github.com/manell/benchmark"
)

func init() {
	sync := make(chan int, 1)
	benchmark.Register("plots", &Plots{
		sync: sync,
		data: make(map[benchmark.Operation][]float64),
	})
}

type point struct{ X, Y float64 }

type Plots struct {
	sync   chan int
	data   map[benchmark.Operation][]float64
	points plotter.XYs
}

func (s *Plots) Run(collector chan *benchmark.Metric, concurrency int) {
	for metric := range collector {
		duration := metric.FinalTime.UnixNano() - metric.StartTime.UnixNano()
		durationMs := float64(duration) / float64(time.Millisecond)

		s.data[*metric.Operation] = append(s.data[*metric.Operation], durationMs)

		p := struct {
			X float64
			Y float64
		}{
			X: float64(metric.FinalTime.UnixNano()) / 1e3,
			Y: durationMs,
		}

		s.points = append(s.points, p)

	}
	s.sync <- 1
}

func (s *Plots) Finalize() {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Plotutil example"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	p.Add(plotter.NewGrid())

	// Make a scatter plotter and set its style.
	scatter, err := plotter.NewScatter(s.points)
	if err != nil {
		panic(err)
	}
	scatter.GlyphStyle.Color = color.RGBA{R: 255, B: 128, A: 255}

	p.Add(scatter)
	p.Legend.Add("scatter", scatter)

	// Save the plot to a PNG file.
	if err := p.Save(128*vg.Inch, 64*vg.Inch, "points.png"); err != nil {
		panic(err)
	}
	<-s.sync
}
