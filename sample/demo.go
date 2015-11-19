package main

import (
	"github.com/shibukawa/nanovgo"
	"math"
)

func cosF(a float32) float32 {
	return float32(math.Cos(float64(a)))
}
func sinF(a float32) float32 {
	return float32(math.Sin(float64(a)))
}

func drawLines(ctx *nanovgo.Context, x, y, w, h float32) {
	var t float32 = 0.0
	var pad float32 = 5.0
	s := w/9.0 - pad*2
	joins := []nanovgo.LineCap{nanovgo.MITER, nanovgo.ROUND, nanovgo.BEVEL}
	caps := []nanovgo.LineCap{nanovgo.BUTT, nanovgo.ROUND, nanovgo.SQUARE}

	ctx.Save()
	pts := []float32{
		-s*0.25 + cosF(t*0.3)*s*0.5,
		sinF(t*0.3) * s * 0.5,
		-s * 0.25,
		0,
		s * 0.25,
		0,
		s*0.25 + cosF(-t*0.3)*s*0.5,
		sinF(-t*0.3) * s * 0.5,
	}
	for i, cap := range caps {
		for j, join := range joins {
			fx := x + s*0.5 + float32(i*3+j)/9.0*w + pad
			fy := y - s*0.5 + pad

			ctx.SetLineCap(cap)
			ctx.SetLineJoin(join)

			ctx.SetStrokeWidth(s * 0.3)
			ctx.SetStrokeColor(nanovgo.RGBA(0, 0, 0, 160))
			ctx.BeginPath()
			ctx.MoveTo(fx+pts[0], fy+pts[1])
			ctx.LineTo(fx+pts[2], fy+pts[3])
			ctx.LineTo(fx+pts[4], fy+pts[5])
			ctx.LineTo(fx+pts[6], fy+pts[7])
			ctx.Stroke()

			ctx.SetLineCap(nanovgo.BUTT)
			ctx.SetLineJoin(nanovgo.BEVEL)

			ctx.SetStrokeWidth(1.0)
			ctx.SetStrokeColor(nanovgo.RGBA(0, 192, 255, 255))
			ctx.BeginPath()
			ctx.MoveTo(fx+pts[0], fy+pts[1])
			ctx.LineTo(fx+pts[2], fy+pts[3])
			ctx.LineTo(fx+pts[4], fy+pts[5])
			ctx.LineTo(fx+pts[6], fy+pts[7])
			ctx.Stroke()

		}
	}

	ctx.Restore()
}

func drawWidths(ctx *nanovgo.Context, x, y, width float32) {
	ctx.Save()
	ctx.SetStrokeColor(nanovgo.RGBA(255, 127, 255, 255))
	for i := 0; i < 20; i++ {
		w := (float32(i) + 0.5) * 0.1
		ctx.SetStrokeWidth(w)
		ctx.BeginPath()
		ctx.MoveTo(x, y)
		ctx.LineTo(x+width, y+width*0.3)
		ctx.Stroke()
		y += 10
	}
	ctx.Restore()
}

func drawCaps(ctx *nanovgo.Context, x, y, width float32) {
	caps := []nanovgo.LineCap{nanovgo.BUTT, nanovgo.ROUND, nanovgo.SQUARE}
	var lineWidth float32 = 0.8

	ctx.Save()
	ctx.BeginPath()
	ctx.Rect(x-lineWidth/2, y, width+lineWidth, 40)
	ctx.SetFillColor(nanovgo.RGBA(255, 200, 255, 32))
	ctx.Fill()

	ctx.BeginPath()
	ctx.Rect(x, y, width, 40)
	ctx.SetFillColor(nanovgo.RGBA(255, 200, 255, 32))
	ctx.Fill()

	ctx.SetStrokeWidth(lineWidth)

	for i, cap := range caps {
		ctx.SetLineCap(cap)
		ctx.SetStrokeColor(nanovgo.RGBA(0, 0, 0, 255))
		ctx.BeginPath()
		ctx.MoveTo(x, y+float32(i)*10+5)
		ctx.LineTo(x+width, y+float32(i)*10+5)
		ctx.Stroke()
	}
	ctx.Restore()
}
