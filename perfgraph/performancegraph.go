package perfgraph

import (
	"fmt"
	"github.com/shibukawa/nanovgo"
	"time"
)

const (
	nvg_GRAPH_HISTORY_COUNT = 100
)

type GraphRenderStyle int

const (
	RENDER_FPS GraphRenderStyle = iota
	RENDER_MS
	RENDER_PERCENT
)

var backgroundColor nanovgo.Color = nanovgo.RGBA(0, 0, 0, 128)
var graphColor nanovgo.Color = nanovgo.RGBA(255, 192, 0, 128)
var titleTextColor nanovgo.Color = nanovgo.RGBA(255, 192, 0, 128)
var fpsTextColor nanovgo.Color = nanovgo.RGBA(240, 240, 240, 255)
var averageTextColor nanovgo.Color = nanovgo.RGBA(240, 240, 240, 160)
var msTextColor nanovgo.Color = nanovgo.RGBA(240, 240, 240, 255)

type PerfGraph struct {
	style    GraphRenderStyle
	name     string
	fontFace string
	values   [nvg_GRAPH_HISTORY_COUNT]float32
	head     int

	startTime      time.Time
	lastUpdateTime time.Time
}

func NewPerfGraph(style GraphRenderStyle, name, fontFace string) *PerfGraph {
	return &PerfGraph{
		style:          style,
		name:           name,
		fontFace:       fontFace,
		startTime:      time.Now(),
		lastUpdateTime: time.Now(),
	}
}

func (pg *PerfGraph) UpdateGraph() (timeFromStart, frameTime float32) {
	timeNow := time.Now()
	timeFromStart = float32(timeNow.Sub(pg.startTime)/time.Millisecond) * 0.001
	frameTime = float32(timeNow.Sub(pg.lastUpdateTime)/time.Millisecond) * 0.001
	pg.lastUpdateTime = timeNow

	pg.head = (pg.head + 1) % nvg_GRAPH_HISTORY_COUNT
	pg.values[pg.head] = frameTime
	return
}

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
	switch pg.style {
	case RENDER_FPS:
		for i := 0; i < nvg_GRAPH_HISTORY_COUNT; i++ {
			v := float32(1.0) / float32(0.00001+pg.values[(pg.head+i)%nvg_GRAPH_HISTORY_COUNT])
			if v > 80.0 {
				v = 80.0
			}
			vx := float32(i) / float32(nvg_GRAPH_HISTORY_COUNT-1) * w
			vy := y + h - ((v / 80.0) * h)
			ctx.LineTo(vx, vy)
		}
	case RENDER_PERCENT:
		for i := 0; i < nvg_GRAPH_HISTORY_COUNT; i++ {
			v := pg.values[(pg.head+i)%nvg_GRAPH_HISTORY_COUNT]
			if v > 100.0 {
				v = 100.0
			}
			vx := float32(i) / float32(nvg_GRAPH_HISTORY_COUNT-1) * w
			vy := y + h - ((v / 100.0) * h)
			ctx.LineTo(vx, vy)
		}
	case RENDER_MS:
		for i := 0; i < nvg_GRAPH_HISTORY_COUNT; i++ {
			v := pg.values[(pg.head+i)%nvg_GRAPH_HISTORY_COUNT] * 1000.0
			if v > 20.0 {
				v = 20.0
			}
			vx := float32(i) / float32(nvg_GRAPH_HISTORY_COUNT-1) * w
			vy := y + h - ((v / 20.0) * h)
			ctx.LineTo(vx, vy)
		}
	}
	ctx.LineTo(x+w, y+h)
	ctx.SetFillColor(graphColor)
	ctx.Fill()

	ctx.SetFontFace(pg.fontFace)

	if len(pg.name) > 0 {
		ctx.SetFontSize(14.0)
		ctx.SetTextAlign(nanovgo.ALIGN_LEFT | nanovgo.ALIGN_TOP)
		ctx.SetFillColor(titleTextColor)
		ctx.Text(x+3, y+1, pg.name)
	}

	switch pg.style {
	case RENDER_FPS:
		ctx.SetFontSize(18.0)
		ctx.SetTextAlign(nanovgo.ALIGN_RIGHT | nanovgo.ALIGN_TOP)
		ctx.SetFillColor(fpsTextColor)
		ctx.Text(x+w-3, y+1, fmt.Sprintf("%.2f FPS", 1.0/avg))

		ctx.SetFontSize(15.0)
		ctx.SetTextAlign(nanovgo.ALIGN_RIGHT | nanovgo.ALIGN_BOTTOM)
		ctx.SetFillColor(averageTextColor)
		ctx.Text(x+w-3, y+1, fmt.Sprintf("%.2f ms", avg*1000.0))
	case RENDER_PERCENT:
		ctx.SetFontSize(18.0)
		ctx.SetTextAlign(nanovgo.ALIGN_RIGHT | nanovgo.ALIGN_TOP)
		ctx.SetFillColor(averageTextColor)
		ctx.Text(x+w-3, y+1, fmt.Sprintf("%.1f %%", avg))
	case RENDER_MS:
		ctx.SetFontSize(18.0)
		ctx.SetTextAlign(nanovgo.ALIGN_RIGHT | nanovgo.ALIGN_TOP)
		ctx.SetFillColor(msTextColor)
		ctx.Text(x+w-3, y+1, fmt.Sprintf("%.2f ms", avg*1000.0))
	}
}

func (pg *PerfGraph) GetGraphAverage() float32 {
	var average float32
	for _, value := range pg.values {
		average += value
	}
	return average / float32(nvg_GRAPH_HISTORY_COUNT)
}
