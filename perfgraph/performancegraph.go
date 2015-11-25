package perfgraph

import (
	"fmt"
	"github.com/shibukawa/nanovgo"
	"time"
)

const (
	nvgGraphHistoryCount = 100
)

var backgroundColor = nanovgo.RGBA(0, 0, 0, 128)
var graphColor  = nanovgo.RGBA(255, 192, 0, 128)
var titleTextColor  = nanovgo.RGBA(255, 192, 0, 128)
var fpsTextColor = nanovgo.RGBA(240, 240, 240, 255)
var averageTextColor  = nanovgo.RGBA(240, 240, 240, 160)
var msTextColor = nanovgo.RGBA(240, 240, 240, 255)

// PerfGraph shows FPS counter on NanoVGo application
type PerfGraph struct {
	name     string
	fontFace string
	values   [nvgGraphHistoryCount]float32
	head     int

	startTime      time.Time
	lastUpdateTime time.Time
}

// NewPerfGraph creates PerfGraph instance
func NewPerfGraph(name, fontFace string) *PerfGraph {
	return &PerfGraph{
		name:           name,
		fontFace:       fontFace,
		startTime:      time.Now(),
		lastUpdateTime: time.Now(),
	}
}

// UpdateGraph updates timer it is needed to show graph
func (pg *PerfGraph) UpdateGraph() (timeFromStart, frameTime float32) {
	timeNow := time.Now()
	timeFromStart = float32(timeNow.Sub(pg.startTime)/time.Millisecond) * 0.001
	frameTime = float32(timeNow.Sub(pg.lastUpdateTime)/time.Millisecond) * 0.001
	pg.lastUpdateTime = timeNow

	pg.head = (pg.head + 1) % nvgGraphHistoryCount
	pg.values[pg.head] = frameTime
	return
}

// RenderGraph shows graph
func (pg *PerfGraph) RenderGraph(ctx *nanovgo.Context, x, y float32) {
	avg := pg.GetGraphAverage()
	var w float32 = 200
	var h float32 = 35

	ctx.BeginPath()
	ctx.Rect(x, y, w, h)
	ctx.SetFillColor(backgroundColor)
	ctx.Fill()

	ctx.BeginPath()
	ctx.MoveTo(x, y+h)
	for i := 0; i < nvgGraphHistoryCount; i++ {
		v := float32(1.0) / float32(0.00001+pg.values[(pg.head+i)%nvgGraphHistoryCount])
		if v > 80.0 {
			v = 80.0
		}
		vx := x + float32(i)/float32(nvgGraphHistoryCount-1)*w
		vy := y + h - ((v / 80.0) * h)
		ctx.LineTo(vx, vy)
	}
	ctx.LineTo(x+w, y+h)
	ctx.SetFillColor(graphColor)
	ctx.Fill()

	ctx.SetFontFace(pg.fontFace)

	if len(pg.name) > 0 {
		ctx.SetFontSize(14.0)
		ctx.SetTextAlign(nanovgo.AlignLeft | nanovgo.AlignTop)
		ctx.SetFillColor(titleTextColor)
		ctx.Text(x+3, y+1, pg.name)
	}

	ctx.SetFontSize(18.0)
	ctx.SetTextAlign(nanovgo.AlignRight | nanovgo.AlignTop)
	ctx.SetFillColor(fpsTextColor)
	ctx.Text(x+w-3, y+1, fmt.Sprintf("%.2f FPS", 1.0/avg))

	ctx.SetFontSize(15.0)
	ctx.SetTextAlign(nanovgo.AlignRight | nanovgo.AlignBottom)
	ctx.SetFillColor(averageTextColor)
	ctx.Text(x+w-3, y+h+1, fmt.Sprintf("%.2f ms", avg*1000.0))
}

// GetGraphAverage returns average value of graph.
func (pg *PerfGraph) GetGraphAverage() float32 {
	var average float32
	for _, value := range pg.values {
		average += value
	}
	return average / float32(nvgGraphHistoryCount)
}
