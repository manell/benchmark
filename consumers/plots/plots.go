package statistics

import (
	"flag"
	"image/color"

	"time"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	"github.com/manell/benchmark"
)

var (
	loaded = flag.Bool("plots", false, "Use plots module.")
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
	sync          chan int
	data          map[benchmark.Operation][]float64
	points        plotter.XYs
	start         float64
	intervalLat   float64
	intervalCount float64
	intervalLats  plotter.XYs
	intervalsTps  plotter.XYs
	concurrency   int
}

func (s *Plots) Loaded() bool { return *loaded }

func (s *Plots) Run(collector chan *benchmark.Metric, iterations, concurrency int) {
	s.concurrency = concurrency
	for metric := range collector {
		if s.start == 0 {
			s.start = float64(metric.StartTime.UnixNano())
		}
		p := struct {
			X float64
			Y float64
		}{
			X: (float64(metric.FinalTime.UnixNano()) - s.start) / 1e9,
			Y: float64(metric.Duration.Nanoseconds()) / 1e6,
		}

		s.points = append(s.points, p)

		s.intervalLat += float64(metric.Duration.Nanoseconds()) / 1e6
		s.intervalCount++

		if s.intervalCount >= float64(iterations/400) {
			timeOp := (float64(metric.FinalTime.UnixNano()) - s.start) / 1e9
			avgLat := s.intervalLat / s.intervalCount

			if len(s.intervalsTps) == 0 {
				pTps := struct {
					X float64
					Y float64
				}{
					X: timeOp,
					Y: 0,
				}

				s.intervalsTps = append(s.intervalsTps, pTps)
			}

			p := struct {
				X float64
				Y float64
			}{
				X: timeOp,
				Y: avgLat,
			}

			s.intervalLats = append(s.intervalLats, p)

			pTps := struct {
				X float64
				Y float64
			}{
				X: timeOp,
				Y: (1 / (avgLat / 1e3)) * float64(s.concurrency),
			}

			s.intervalsTps = append(s.intervalsTps, pTps)

			s.intervalLat = 0
			s.intervalCount = 0
		}

	}

	s.sync <- 1
}

func (s *Plots) DrawTps() {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Response Time"
	p.X.Label.Text = "Time (seconds)"
	p.Y.Label.Text = "Response time (ms)"

	p.Add(plotter.NewGrid())

	// Make a line plotter and set its style.
	l, err := plotter.NewLine(s.intervalsTps)
	if err != nil {
		panic(err)
	}

	l.LineStyle.Color = color.RGBA{B: 255, A: 255}

	p.Add(l)
	p.Legend.Add("line", l)

	// Save the plot to a PNG file.
	if err := p.Save(128*vg.Inch, 32*vg.Inch, "tps.png"); err != nil {
		panic(err)
	}
}

func (s *Plots) Finalize(d time.Duration) {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Response Time"
	p.X.Label.Text = "Time (seconds)"
	p.Y.Label.Text = "Response time (ms)"

	p.Add(plotter.NewGrid())

	// Make a scatter plotter and set its style.
	scatter, err := plotter.NewScatter(s.points)
	if err != nil {
		panic(err)
	}
	scatter.GlyphStyle.Color = color.RGBA{R: 255, B: 128, A: 255}

	// Make a line plotter and set its style.
	l, err := plotter.NewLine(s.intervalLats)
	if err != nil {
		panic(err)
	}
	l.LineStyle.Width = vg.Points(1)
	l.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}
	l.LineStyle.Color = color.RGBA{B: 255, A: 255}

	p.Add(scatter, l)
	p.Legend.Add("scatter", scatter)
	p.Legend.Add("line", l)

	// Save the plot to a PNG file.
	if err := p.Save(128*vg.Inch, 24*vg.Inch, "points.png"); err != nil {
		panic(err)
	}

	s.DrawTps()

	<-s.sync
}
