package main

import (
	"fmt"
	"github.com/shibukawa/nanovgo"
	"math"

	"log"
)

type DemoData struct {
	fontNormal, fontBold, fontIcons int
	images                          [12]int
}

func (d *DemoData) loadData(vg *nanovgo.Context) {
	for i := 0; i < 12; i++ {
		path := fmt.Sprintf("images/image%d.jpg", i+1)
		d.images[i] = vg.CreateImage(path, 0)
		if d.images[i] == 0 {
			log.Fatalf("Could not load %s", path)
		}
	}

	d.fontIcons = vg.CreateFont("icons", "entypo.ttf")
	if d.fontIcons == -1 {
		log.Fatalln("Could not add font icons.")
	}
	d.fontNormal = vg.CreateFont("sans", "Roboto-Regular.ttf")
	if d.fontNormal == -1 {
		log.Fatalln("Could not add font italic.")
	}
	d.fontBold = vg.CreateFont("sans-bold", "Roboto-Bold.ttf")
	if d.fontBold == -1 {
		log.Fatalln("Could not add font bold.")
	}
}

func (d *DemoData) freeData(vg *nanovgo.Context) {
	for _, img := range d.images {
		vg.DeleteImage(img)
	}
}

func cosF(a float32) float32 {
	return float32(math.Cos(float64(a)))
}
func sinF(a float32) float32 {
	return float32(math.Sin(float64(a)))
}
func sqrtF(a float32) float32 {
	return float32(math.Sqrt(float64(a)))
}

func drawWindow(ctx *nanovgo.Context, title string, x, y, w, h float32) {
	var cornerRadius float32 = 3.0

	ctx.Save()
	//      ctx.ClearState(vg);

	// Window
	ctx.BeginPath()
	ctx.RoundedRect(x, y, w, h, cornerRadius)
	ctx.SetFillColor(nanovgo.RGBA(28, 30, 34, 192))
	//      ctx.FillColor(vg, nanovgo.RGBA(0,0,0,128));
	ctx.Fill()

	// Drop shadow
	shadowPaint := nanovgo.BoxGradient(x, y+2, w, h, cornerRadius*2, 10, nanovgo.RGBA(0, 0, 0, 128), nanovgo.RGBA(0, 0, 0, 0))
	ctx.BeginPath()
	ctx.Rect(x-10, y-10, w+20, h+30)
	ctx.RoundedRect(x, y, w, h, cornerRadius)
	ctx.PathWinding(nanovgo.HOLE)
	ctx.SetFillPaint(shadowPaint)
	ctx.Fill()

	// Header
	headerPaint := nanovgo.LinearGradient(x, y, x, y+15, nanovgo.RGBA(255, 255, 255, 8), nanovgo.RGBA(0, 0, 0, 16))
	ctx.BeginPath()
	ctx.RoundedRect(x+1, y+1, w-2, 30, cornerRadius-1)
	ctx.SetFillPaint(headerPaint)
	ctx.Fill()
	ctx.BeginPath()
	ctx.MoveTo(x+0.5, y+0.5+30)
	ctx.LineTo(x+0.5+w-1, y+0.5+30)
	ctx.SetStrokeColor(nanovgo.RGBA(0, 0, 0, 32))
	ctx.Stroke()

	ctx.SetFontSize(18.0)
	ctx.SetFontFace("sans-bold")
	ctx.SetTextAlign(nanovgo.ALIGN_CENTER | nanovgo.ALIGN_MIDDLE)

	ctx.SetFontBlur(2)
	ctx.SetFillColor(nanovgo.RGBA(0, 0, 0, 128))
	ctx.Text(x+w/2, y+16+1, title)

	ctx.SetFontBlur(0)
	ctx.SetFillColor(nanovgo.RGBA(220, 220, 220, 160))
	ctx.Text(x+w/2, y+16, title)

	ctx.Restore()
}

func isBlack(col nanovgo.Color) bool {
	return col.R == 0.0 && col.G == 0.0 && col.B == 0.0 && col.A == 0.0
}

func cpToUTF8(cp int) string {
	return string([]rune{rune(cp)})
}

func drawButton(ctx *nanovgo.Context, preicon int, text string, x, y, w, h float32, col nanovgo.Color) {
	var cornerRadius float32 = 4.0
	var iw, tw float32

	var alpha uint8
	if isBlack(col) {
		alpha = 16
	} else {
		alpha = 32
	}
	bg := nanovgo.LinearGradient(x, y, x, y+h, nanovgo.RGBA(255, 255, 255, alpha), nanovgo.RGBA(0, 0, 0, alpha))
	ctx.BeginPath()
	ctx.RoundedRect(x+1, y+1, w-2, h-2, cornerRadius-1)
	if !isBlack(col) {
		ctx.SetFillColor(col)
		ctx.Fill()
	}
	ctx.SetFillPaint(bg)
	ctx.Fill()

	ctx.BeginPath()
	ctx.RoundedRect(x+0.5, y+0.5, w-1, h-1, cornerRadius-0.5)
	ctx.SetStrokeColor(nanovgo.RGBA(0, 0, 0, 48))
	ctx.Stroke()

	ctx.SetFontSize(20.0)
	ctx.SetFontFace("sans-bold")
	tw, _ = ctx.TextBounds(0, 0, text)
	if preicon != 0 {
		ctx.SetFontSize(h * 1.3)
		ctx.SetFontFace("icons")
		iw, _ = ctx.TextBounds(0, 0, cpToUTF8(preicon))
		iw += h * 0.15
	}

	if preicon != 0 {
		ctx.SetFontSize(h * 1.3)
		ctx.SetFontFace("icons")
		ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 96))
		ctx.SetTextAlign(nanovgo.ALIGN_LEFT | nanovgo.ALIGN_MIDDLE)
		ctx.Text(x+w*0.5-tw*0.5-iw*0.75, y+h*0.5, cpToUTF8(preicon))
	}

	ctx.SetFontSize(20.0)
	ctx.SetFontFace("sans-bold")
	ctx.SetTextAlign(nanovgo.ALIGN_LEFT | nanovgo.ALIGN_MIDDLE)
	ctx.SetFillColor(nanovgo.RGBA(0, 0, 0, 160))
	ctx.Text(x+w*0.5-tw*0.5+iw*0.25, y+h*0.5-1, text)
	ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 160))
	ctx.Text(x+w*0.5-tw*0.5+iw*0.25, y+h*0.5, text)
}

func drawEyes(ctx *nanovgo.Context, x, y, w, h, mx, my, t float32) {
	ex := w * 0.23
	ey := h * 0.5
	lx := x + ex
	ly := y + ey
	rx := x + w - ex
	ry := y + ey
	var dx, dy, d, br float32
	if ex < ey {
		br = ex * 0.5
	} else {
		br = ey * 0.5
	}
	blink := float32(1.0 - math.Pow(float64(sinF(t*0.5)), 200)*0.8)

	bg1 := nanovgo.LinearGradient(x, y+h*0.5, x+w*0.1, y+h, nanovgo.RGBA(0, 0, 0, 32), nanovgo.RGBA(0, 0, 0, 16))
	ctx.BeginPath()
	ctx.Ellipse(lx+3.0, ly+16.0, ex, ey)
	ctx.Ellipse(rx+3.0, ry+16.0, ex, ey)
	ctx.SetFillPaint(bg1)
	ctx.Fill()

	bg2 := nanovgo.LinearGradient(x, y+h*0.25, x+w*0.1, y+h, nanovgo.RGBA(220, 220, 220, 255), nanovgo.RGBA(128, 128, 128, 255))
	ctx.BeginPath()
	ctx.Ellipse(lx, ly, ex, ey)
	ctx.Ellipse(rx, ry, ex, ey)
	ctx.SetFillPaint(bg2)
	ctx.Fill()

	dx = (mx - rx) / (ex * 10)
	dy = (my - ry) / (ey * 10)
	d = sqrtF(dx*dx + dy*dy)
	if d > 1.0 {
		dx /= d
		dy /= d
	}
	dx *= ex * 0.4
	dy *= ey * 0.5
	ctx.BeginPath()
	ctx.Ellipse(lx+dx, ly+dy+ey*0.25*(1.0-blink), br, br*blink)
	ctx.SetFillColor(nanovgo.RGBA(32, 32, 32, 255))
	ctx.Fill()

	dx = (mx - rx) / (ex * 10)
	dy = (my - ry) / (ey * 10)
	d = sqrtF(dx*dx + dy*dy)
	if d > 1.0 {
		dx /= d
		dy /= d
	}
	dx *= ex * 0.4
	dy *= ey * 0.5
	ctx.BeginPath()
	ctx.Ellipse(rx+dx, ry+dy+ey*0.25*(1.0-blink), br, br*blink)
	ctx.SetFillColor(nanovgo.RGBA(32, 32, 32, 255))
	ctx.Fill()
	dx = (mx - rx) / (ex * 10)
	dy = (my - ry) / (ey * 10)
	d = sqrtF(dx*dx + dy*dy)
	if d > 1.0 {
		dx /= d
		dy /= d
	}
	dx *= ex * 0.4
	dy *= ey * 0.5
	ctx.BeginPath()
	ctx.Ellipse(rx+dx, ry+dy+ey*0.25*(1.0-blink), br, br*blink)
	ctx.SetFillColor(nanovgo.RGBA(32, 32, 32, 255))
	ctx.Fill()

	gloss1 := nanovgo.RadialGradient(lx-ex*0.25, ly-ey*0.5, ex*0.1, ex*0.75, nanovgo.RGBA(255, 255, 255, 128), nanovgo.RGBA(255, 255, 255, 0))
	ctx.BeginPath()
	ctx.Ellipse(lx, ly, ex, ey)
	ctx.SetFillPaint(gloss1)
	ctx.Fill()

	gloss2 := nanovgo.RadialGradient(rx-ex*0.25, ry-ey*0.5, ex*0.1, ex*0.75, nanovgo.RGBA(255, 255, 255, 128), nanovgo.RGBA(255, 255, 255, 0))
	ctx.BeginPath()
	ctx.Ellipse(rx, ry, ex, ey)
	ctx.SetFillPaint(gloss2)
	ctx.Fill()
}

func drawGraph(ctx *nanovgo.Context, x, y, w, h, t float32) {
	var sx, sy [6]float32
	dx := w / 5.0

	samples := []float32{
		(1 + sinF(t*1.2345+cosF(t*0.33457)*0.44)) * 0.5,
		(1 + sinF(t*0.68363+cosF(t*1.3)*1.55)) * 0.5,
		(1 + sinF(t*1.1642+cosF(t*0.33457)*1.24)) * 0.5,
		(1 + sinF(t*0.56345+cosF(t*1.63)*0.14)) * 0.5,
		(1 + sinF(t*1.6245+cosF(t*0.254)*0.3)) * 0.5,
		(1 + sinF(t*0.345+cosF(t*0.03)*0.6)) * 0.5,
	}

	for i := 0; i < 6; i++ {
		sx[i] = x + float32(i)*dx
		sy[i] = y + h*samples[i]*0.8
	}

	// Graph background
	bg := nanovgo.LinearGradient(x, y, x, y+h, nanovgo.RGBA(0, 160, 192, 0), nanovgo.RGBA(0, 160, 192, 64))
	ctx.BeginPath()
	ctx.MoveTo(sx[0], sy[0])
	for i := 1; i < 6; i++ {
		ctx.BezierTo(sx[i-1]+dx*0.5, sy[i-1], sx[i]-dx*0.5, sy[i], sx[i], sy[i])
	}
	ctx.LineTo(x+w, y+h)
	ctx.LineTo(x, y+h)
	ctx.SetFillPaint(bg)
	ctx.Fill()

	// Graph line
	ctx.BeginPath()
	ctx.MoveTo(sx[0], sy[0]+2)
	for i := 1; i < 6; i++ {
		ctx.BezierTo(sx[i-1]+dx*0.5, sy[i-1]+2, sx[i]-dx*0.5, sy[i]+2, sx[i], sy[i]+2)
	}
	ctx.SetStrokeColor(nanovgo.RGBA(0, 0, 0, 32))
	ctx.SetStrokeWidth(3.0)
	ctx.Stroke()

	ctx.BeginPath()
	ctx.MoveTo(sx[0], sy[0])
	for i := 1; i < 6; i++ {
		ctx.BezierTo(sx[i-1]+dx*0.5, sy[i-1], sx[i]-dx*0.5, sy[i], sx[i], sy[i])
	}
	ctx.SetStrokeColor(nanovgo.RGBA(0, 160, 192, 255))
	ctx.SetStrokeWidth(3.0)
	ctx.Stroke()

	// Graph sample pos
	for i := 0; i < 6; i++ {
		bg = nanovgo.RadialGradient(sx[i], sy[i]+2, 3.0, 8.0, nanovgo.RGBA(0, 0, 0, 32), nanovgo.RGBA(0, 0, 0, 0))
		ctx.BeginPath()
		ctx.Rect(sx[i]-10, sy[i]-10+2, 20, 20)
		ctx.SetFillPaint(bg)
		ctx.Fill()
	}

	ctx.BeginPath()
	for i := 0; i < 6; i++ {
		ctx.Circle(sx[i], sy[i], 4.0)
	}
	ctx.SetFillColor(nanovgo.RGBA(0, 160, 192, 255))
	ctx.Fill()
	ctx.BeginPath()
	for i := 0; i < 6; i++ {
		ctx.Circle(sx[i], sy[i], 2.0)
	}
	ctx.SetFillColor(nanovgo.RGBA(220, 220, 220, 255))
	ctx.Fill()

	ctx.SetStrokeWidth(1.0)
}

func drawColorWheel(ctx *nanovgo.Context, x, y, w, h, t float32) {
	var r0, r1, ax, ay, bx, by, aeps, r float32
	hue := sinF(t * 0.12)

	ctx.Save()
	/*      ctx.BeginPath()
	ctx.Rect(x,y,w,h)
	ctx.FillColor(nanovgo.RGBA(255,0,0,128))
	ctx.Fill()*/

	cx := x + w*0.5
	cy := y + h*0.5
	if w < h {
		r1 = w*0.5 - 5.0
	} else {
		r1 = h*0.5 - 5.0
	}
	r0 = r1 - 20.0
	aeps = 0.5 / r1 // half a pixel arc length in radians (2pi cancels out).

	for i := 0; i < 6; i++ {
		a0 := float32(i)/6.0*nanovgo.PI*2.0 - aeps
		a1 := float32(i+1.0)/6.0*nanovgo.PI*2.0 + aeps
		ctx.BeginPath()
		ctx.Arc(cx, cy, r0, a0, a1, nanovgo.CW)
		ctx.Arc(cx, cy, r1, a1, a0, nanovgo.CCW)
		ctx.ClosePath()
		ax = cx + cosF(a0)*(r0+r1)*0.5
		ay = cy + sinF(a0)*(r0+r1)*0.5
		bx = cx + cosF(a1)*(r0+r1)*0.5
		by = cy + sinF(a1)*(r0+r1)*0.5
		paint := nanovgo.LinearGradient(ax, ay, bx, by, nanovgo.HSLA(a0/(nanovgo.PI*2), 1.0, 0.55, 255), nanovgo.HSLA(a1/(nanovgo.PI*2), 1.0, 0.55, 255))
		ctx.SetFillPaint(paint)
		ctx.Fill()
	}

	ctx.BeginPath()
	ctx.Circle(cx, cy, r0-0.5)
	ctx.Circle(cx, cy, r1+0.5)
	ctx.SetStrokeColor(nanovgo.RGBA(0, 0, 0, 64))
	ctx.SetStrokeWidth(1.0)
	ctx.Stroke()

	// Selector
	ctx.Translate(cx, cy)
	ctx.Rotate(hue * nanovgo.PI * 2)

	// Marker on
	ctx.SetStrokeWidth(2.0)
	ctx.BeginPath()
	ctx.Rect(r0-1, -3, r1-r0+2, 6)
	ctx.SetStrokeColor(nanovgo.RGBA(255, 255, 255, 192))
	ctx.Stroke()

	paint := nanovgo.BoxGradient(r0-3, -5, r1-r0+6, 10, 2, 4, nanovgo.RGBA(0, 0, 0, 128), nanovgo.RGBA(0, 0, 0, 0))
	ctx.BeginPath()
	ctx.Rect(r0-2-10, -4-10, r1-r0+4+20, 8+20)
	ctx.Rect(r0-2, -4, r1-r0+4, 8)
	ctx.PathWinding(nanovgo.HOLE)
	ctx.SetFillPaint(paint)
	ctx.Fill()

	// Center triangle
	r = r0 - 6
	ax = cosF(120.0/180.0*nanovgo.PI) * r
	ay = sinF(120.0/180.0*nanovgo.PI) * r
	bx = cosF(-120.0/180.0*nanovgo.PI) * r
	by = sinF(-120.0/180.0*nanovgo.PI) * r
	ctx.BeginPath()
	ctx.MoveTo(r, 0)
	ctx.LineTo(ax, ay)
	ctx.LineTo(bx, by)
	ctx.ClosePath()
	paint = nanovgo.LinearGradient(r, 0, ax, ay, nanovgo.HSLA(hue, 1.0, 0.5, 255), nanovgo.RGBA(255, 255, 255, 255))
	ctx.SetFillPaint(paint)
	ctx.Fill()
	paint = nanovgo.LinearGradient((r+ax)*0.5, (0+ay)*0.5, bx, by, nanovgo.RGBA(0, 0, 0, 0), nanovgo.RGBA(0, 0, 0, 255))
	ctx.SetFillPaint(paint)
	ctx.Fill()
	ctx.SetStrokeColor(nanovgo.RGBA(0, 0, 0, 64))
	ctx.Stroke()

	// Select circle on triangle
	ax = cosF(120.0/180.0*nanovgo.PI) * r * 0.3
	ay = sinF(120.0/180.0*nanovgo.PI) * r * 0.4
	ctx.SetStrokeWidth(2.0)
	ctx.BeginPath()
	ctx.Circle(ax, ay, 5)
	ctx.SetStrokeColor(nanovgo.RGBA(255, 255, 255, 192))
	ctx.Stroke()

	paint = nanovgo.RadialGradient(ax, ay, 7, 9, nanovgo.RGBA(0, 0, 0, 64), nanovgo.RGBA(0, 0, 0, 0))
	ctx.BeginPath()
	ctx.Rect(ax-20, ay-20, 40, 40)
	ctx.Circle(ax, ay, 7)
	ctx.PathWinding(nanovgo.HOLE)
	ctx.SetFillPaint(paint)
	ctx.Fill()

	ctx.Restore()
}

func drawLines(ctx *nanovgo.Context, x, y, w, h, t float32) {
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
	var lineWidth float32 = 8.0

	ctx.Save()
	ctx.BeginPath()
	ctx.Rect(x-lineWidth/2, y, width+lineWidth, 40)
	ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 32))
	ctx.Fill()

	ctx.BeginPath()
	ctx.Rect(x, y, width, 40)
	ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 32))
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
