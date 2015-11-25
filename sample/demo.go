package main

import (
	"fmt"
	"github.com/shibukawa/nanovgo"
	"math"

	"log"
	"strconv"
)

// DemoData keeps font and image handlers
type DemoData struct {
	fontNormal, fontBold, fontIcons int
	images                          []int
}

func (d *DemoData) loadData(ctx *nanovgo.Context) {
	for i := 0; i < 12; i++ {
		path := fmt.Sprintf("images/image%d.jpg", i+1)
		d.images = append(d.images, ctx.CreateImage(path, 0))
		if d.images[i] == 0 {
			log.Fatalf("Could not load %s", path)
		}
	}

	d.fontIcons = ctx.CreateFont("icons", "entypo.ttf")
	if d.fontIcons == -1 {
		log.Fatalln("Could not add font icons.")
	}
	d.fontNormal = ctx.CreateFont("sans", "Roboto-Regular.ttf")
	if d.fontNormal == -1 {
		log.Fatalln("Could not add font italic.")
	}
	d.fontBold = ctx.CreateFont("sans-bold", "Roboto-Bold.ttf")
	if d.fontBold == -1 {
		log.Fatalln("Could not add font bold.")
	}
}

func (d *DemoData) freeData(ctx *nanovgo.Context) {
	for _, img := range d.images {
		ctx.DeleteImage(img)
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
func clampF(a, min, max float32) float32 {
	if a < min {
		return min
	}
	if a > max {
		return max
	}
	return a
}
func absF(a float32) float32 {
	if a > 0.0 {
		return a
	}
	return -a
}
func maxF(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func isBlack(col nanovgo.Color) bool {
	return col.R == 0.0 && col.G == 0.0 && col.B == 0.0 && col.A == 0.0
}

func cpToUTF8(cp int) string {
	return string([]rune{rune(cp)})
}

func drawWindow(ctx *nanovgo.Context, title string, x, y, w, h float32) {
	var cornerRadius float32 = 3.0

	ctx.Save()
	defer ctx.Restore()
	//      ctx.Reset();

	// Window
	ctx.BeginPath()
	ctx.RoundedRect(x, y, w, h, cornerRadius)
	ctx.SetFillColor(nanovgo.RGBA(28, 30, 34, 192))
	//      ctx.FillColor(nanovgo.RGBA(0,0,0,128));
	ctx.Fill()

	// Drop shadow
	shadowPaint := nanovgo.BoxGradient(x, y+2, w, h, cornerRadius*2, 10, nanovgo.RGBA(0, 0, 0, 128), nanovgo.RGBA(0, 0, 0, 0))
	ctx.BeginPath()
	ctx.Rect(x-10, y-10, w+20, h+30)
	ctx.RoundedRect(x, y, w, h, cornerRadius)
	ctx.PathWinding(nanovgo.Hole)
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
	ctx.SetTextAlign(nanovgo.AlignCenter | nanovgo.AlignMiddle)

	ctx.SetFontBlur(2)
	ctx.SetFillColor(nanovgo.RGBA(0, 0, 0, 128))
	ctx.Text(x+w/2, y+16+1, title)

	ctx.SetFontBlur(0)
	ctx.SetFillColor(nanovgo.RGBA(220, 220, 220, 160))
	ctx.Text(x+w/2, y+16, title)
}

func drawSearchBox(ctx *nanovgo.Context, text string, x, y, w, h float32) {
	cornerRadius := h/2 - 1

	// Edit
	bg := nanovgo.BoxGradient(x, y+1.5, w, h, h/2, 5, nanovgo.RGBA(0, 0, 0, 16), nanovgo.RGBA(0, 0, 0, 92))
	ctx.BeginPath()
	ctx.RoundedRect(x, y, w, h, cornerRadius)
	ctx.SetFillPaint(bg)
	ctx.Fill()

	/*      ctx.BeginPath();
	        ctx.RoundedRect(x+0.5f,y+0.5f, w-1,h-1, cornerRadius-0.5f);
	        ctx.StrokeColor(ctx.RGBA(0,0,0,48));
	        ctx.Stroke();*/

	ctx.SetFontSize(h * 1.3)
	ctx.SetFontFace("icons")
	ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 64))
	ctx.SetTextAlign(nanovgo.AlignCenter | nanovgo.AlignMiddle)
	ctx.Text(x+h*0.55, y+h*0.55, cpToUTF8(iconSEARCH))

	ctx.SetFontSize(20.0)
	ctx.SetFontFace("sans")
	ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 32))

	ctx.SetTextAlign(nanovgo.AlignLeft | nanovgo.AlignMiddle)
	ctx.Text(x+h*1.05, y+h*0.5, text)

	ctx.SetFontSize(h * 1.3)
	ctx.SetFontFace("icons")
	ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 32))
	ctx.SetTextAlign(nanovgo.AlignCenter | nanovgo.AlignMiddle)
	ctx.Text(x+w-h*0.55, y+h*0.55, cpToUTF8(iconCIRCLEDCROSS))
}

func drawDropDown(ctx *nanovgo.Context, text string, x, y, w, h float32) {

	var cornerRadius float32 = 4.0

	bg := nanovgo.LinearGradient(x, y, x, y+h, nanovgo.RGBA(255, 255, 255, 16), nanovgo.RGBA(0, 0, 0, 16))
	ctx.BeginPath()
	ctx.RoundedRect(x+1, y+1, w-2, h-2, cornerRadius-1)
	ctx.SetFillPaint(bg)
	ctx.Fill()

	ctx.BeginPath()
	ctx.RoundedRect(x+0.5, y+0.5, w-1, h-1, cornerRadius-0.5)
	ctx.SetStrokeColor(nanovgo.RGBA(0, 0, 0, 48))
	ctx.Stroke()

	ctx.SetFontSize(20.0)
	ctx.SetFontFace("sans")
	ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 160))
	ctx.SetTextAlign(nanovgo.AlignLeft | nanovgo.AlignMiddle)
	ctx.Text(x+h*0.3, y+h*0.5, text)

	ctx.SetFontSize(h * 1.3)
	ctx.SetFontFace("icons")
	ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 64))
	ctx.SetTextAlign(nanovgo.AlignCenter | nanovgo.AlignMiddle)
	ctx.Text(x+w-h*0.5, y+h*0.5, cpToUTF8(iconCHEVRONRIGHT))
}

func drawEditBoxBase(ctx *nanovgo.Context, x, y, w, h float32) {
	// Edit
	bg := nanovgo.BoxGradient(x+1, y+1+1.5, w-2, h-2, 3, 4, nanovgo.RGBA(255, 255, 255, 32), nanovgo.RGBA(32, 32, 32, 32))
	ctx.BeginPath()
	ctx.RoundedRect(x+1, y+1, w-2, h-2, 4-1)
	ctx.SetFillPaint(bg)
	ctx.Fill()

	ctx.BeginPath()
	ctx.RoundedRect(x+0.5, y+0.5, w-1, h-1, 4-0.5)
	ctx.SetStrokeColor(nanovgo.RGBA(0, 0, 0, 48))
	ctx.Stroke()
}

func drawEditBox(ctx *nanovgo.Context, text string, x, y, w, h float32) {

	drawEditBoxBase(ctx, x, y, w, h)

	ctx.SetFontSize(20.0)
	ctx.SetFontFace("sans")
	ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 64))
	ctx.SetTextAlign(nanovgo.AlignLeft | nanovgo.AlignMiddle)
	ctx.Text(x+h*0.3, y+h*0.5, text)
}

func drawLabel(ctx *nanovgo.Context, text string, x, y, w, h float32) {

	ctx.SetFontSize(18.0)
	ctx.SetFontFace("sans")
	ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 128))

	ctx.SetTextAlign(nanovgo.AlignLeft | nanovgo.AlignMiddle)
	ctx.Text(x, y+h*0.5, text)
}

func drawEditBoxNum(ctx *nanovgo.Context, text, units string, x, y, w, h float32) {
	drawEditBoxBase(ctx, x, y, w, h)

	uw, _ := ctx.TextBounds(0, 0, units)

	ctx.SetFontSize(18.0)
	ctx.SetFontFace("sans")
	ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 64))
	ctx.SetTextAlign(nanovgo.AlignRight | nanovgo.AlignMiddle)
	ctx.Text(x+w-h*0.3, y+h*0.5, units)

	ctx.SetFontSize(20.0)
	ctx.SetFontFace("sans")
	ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 128))
	ctx.SetTextAlign(nanovgo.AlignRight | nanovgo.AlignMiddle)
	ctx.Text(x+w-uw-h*0.5, y+h*0.5, text)
}

func drawCheckBox(ctx *nanovgo.Context, text string, x, y, w, h float32) {

	ctx.SetFontSize(18.0)
	ctx.SetFontFace("sans")
	ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 160))

	ctx.SetTextAlign(nanovgo.AlignLeft | nanovgo.AlignMiddle)
	ctx.Text(x+28, y+h*0.5, text)

	bg := nanovgo.BoxGradient(x+1, y+float32(int(h*0.5))-9+1, 18, 18, 3, 3, nanovgo.RGBA(0, 0, 0, 32), nanovgo.RGBA(0, 0, 0, 92))
	ctx.BeginPath()
	ctx.RoundedRect(x+1, y+float32(int(h*0.5))-9, 18, 18, 3)
	ctx.SetFillPaint(bg)
	ctx.Fill()

	ctx.SetFontSize(40)
	ctx.SetFontFace("icons")
	ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 128))
	ctx.SetTextAlign(nanovgo.AlignCenter | nanovgo.AlignMiddle)
	ctx.Text(x+9+2, y+h*0.5, cpToUTF8(iconCHECK))
}

func drawButton(ctx *nanovgo.Context, preicon int, text string, x, y, w, h float32, col nanovgo.Color) {
	var cornerRadius float32 = 4.0
	var iw float32

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
	tw, _ := ctx.TextBounds(0, 0, text)
	if preicon != 0 {
		ctx.SetFontSize(h * 1.3)
		ctx.SetFontFace("icons")
		iw, _ = ctx.TextBounds(0, 0, cpToUTF8(preicon))
		iw += h * 0.15

		ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 96))
		ctx.SetTextAlign(nanovgo.AlignLeft | nanovgo.AlignMiddle)
		ctx.Text(x+w*0.5-tw*0.5-iw*0.75, y+h*0.5, cpToUTF8(preicon))
	}

	ctx.SetFontSize(20.0)
	ctx.SetFontFace("sans-bold")
	ctx.SetTextAlign(nanovgo.AlignLeft | nanovgo.AlignMiddle)
	ctx.SetFillColor(nanovgo.RGBA(0, 0, 0, 160))
	ctx.Text(x+w*0.5-tw*0.5+iw*0.25, y+h*0.5-1, text)
	ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 160))
	ctx.Text(x+w*0.5-tw*0.5+iw*0.25, y+h*0.5, text)
}

func drawSlider(ctx *nanovgo.Context, pos, x, y, w, h float32) {
	cy := y + float32(int(h*0.5))
	kr := float32(int(h * 0.25))

	ctx.Save()
	defer ctx.Restore()
	//      ctx.ClearState(vg);

	// Slot
	bg := nanovgo.BoxGradient(x, cy-2+1, w, 4, 2, 2, nanovgo.RGBA(0, 0, 0, 32), nanovgo.RGBA(0, 0, 0, 128))
	ctx.BeginPath()
	ctx.RoundedRect(x, cy-2, w, 4, 2)
	ctx.SetFillPaint(bg)
	ctx.Fill()

	// Knob Shadow
	bg = nanovgo.RadialGradient(x+float32(int(pos*w)), cy+1, kr-3, kr+3, nanovgo.RGBA(0, 0, 0, 64), nanovgo.RGBA(0, 0, 0, 0))
	ctx.BeginPath()
	ctx.Rect(x+float32(int(pos*w))-kr-5, cy-kr-5, kr*2+5+5, kr*2+5+5+3)
	ctx.Circle(x+float32(int(pos*w)), cy, kr)
	ctx.PathWinding(nanovgo.Hole)
	ctx.SetFillPaint(bg)
	ctx.Fill()

	// Knob
	knob := nanovgo.LinearGradient(x, cy-kr, x, cy+kr, nanovgo.RGBA(255, 255, 255, 16), nanovgo.RGBA(0, 0, 0, 16))
	ctx.BeginPath()
	ctx.Circle(x+float32(int(pos*w)), cy, kr-1)
	ctx.SetFillColor(nanovgo.RGBA(40, 43, 48, 255))
	ctx.Fill()
	ctx.SetFillPaint(knob)
	ctx.Fill()

	ctx.BeginPath()
	ctx.Circle(x+float32(int(pos*w)), cy, kr-0.5)
	ctx.SetStrokeColor(nanovgo.RGBA(0, 0, 0, 92))
	ctx.Stroke()
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

func drawSpinner(ctx *nanovgo.Context, cx, cy, r, t float32) {
	a0 := 0.0 + t*6
	a1 := nanovgo.PI + t*6
	r0 := r
	r1 := r * 0.75

	ctx.Save()
	defer ctx.Restore()

	ctx.BeginPath()
	ctx.Arc(cx, cy, r0, a0, a1, nanovgo.Clockwise)
	ctx.Arc(cx, cy, r1, a1, a0, nanovgo.CounterClockwise)
	ctx.ClosePath()
	ax := cx + cosF(a0)*(r0+r1)*0.5
	ay := cy + sinF(a0)*(r0+r1)*0.5
	bx := cx + cosF(a1)*(r0+r1)*0.5
	by := cy + sinF(a1)*(r0+r1)*0.5
	paint := nanovgo.LinearGradient(ax, ay, bx, by, nanovgo.RGBA(0, 0, 0, 0), nanovgo.RGBA(0, 0, 0, 128))
	ctx.SetFillPaint(paint)
	ctx.Fill()
}

func drawThumbnails(ctx *nanovgo.Context, x, y, w, h float32, images []int, t float32) {
	var cornerRadius float32 = 3.0

	var thumb float32 = 60.0
	var arry float32 = 30.5
	stackh := float32(len(images)/2)*(thumb+10) + 10
	u := (1 + cosF(t*0.5)) * 0.5
	u2 := (1 - cosF(t*0.2)) * 0.5

	ctx.Save()
	defer ctx.Restore()

	// Drop shadow
	shadowPaint := nanovgo.BoxGradient(x, y+4, w, h, cornerRadius*2, 20, nanovgo.RGBA(0, 0, 0, 128), nanovgo.RGBA(0, 0, 0, 0))
	ctx.BeginPath()
	ctx.Rect(x-10, y-10, w+20, h+30)
	ctx.RoundedRect(x, y, w, h, cornerRadius)
	ctx.PathWinding(nanovgo.Hole)
	ctx.SetFillPaint(shadowPaint)
	ctx.Fill()

	// Window
	ctx.BeginPath()
	ctx.RoundedRect(x, y, w, h, cornerRadius)
	ctx.MoveTo(x-10, y+arry)
	ctx.LineTo(x+1, y+arry-11)
	ctx.LineTo(x+1, y+arry+11)
	ctx.SetFillColor(nanovgo.RGBA(200, 200, 200, 255))
	ctx.Fill()

	ctx.Block(func() {
		ctx.Scissor(x, y, w, h)
		ctx.Translate(0, -(stackh-h)*u)

		dv := 1.0 / float32(len(images)-1)

		for i, imageID := range images {
			tx := x + 10.0
			ty := y + 10.0
			tx += float32(i%2) * (thumb + 10.0)
			ty += float32(i/2) * (thumb + 10.0)
			imgW, imgH, _ := ctx.ImageSize(imageID)
			var iw, ih, ix, iy float32
			if imgW < imgH {
				iw = thumb
				ih = iw * float32(imgH) / float32(imgW)
				ix = 0
				iy = -(ih - thumb) * 0.5
			} else {
				ih = thumb
				iw = ih * float32(imgW) / float32(imgH)
				ix = -(iw - thumb) * 0.5
				iy = 0
			}

			v := float32(i) * dv
			a := clampF((u2-v)/dv, 0, 1)

			if a < 1.0 {
				drawSpinner(ctx, tx+thumb/2, ty+thumb/2, thumb*0.25, t)
			}

			imgPaint := nanovgo.ImagePattern(tx+ix, ty+iy, iw, ih, 0.0/180.0*nanovgo.PI, imageID, a)
			ctx.BeginPath()
			ctx.RoundedRect(tx, ty, thumb, thumb, 5)
			ctx.SetFillPaint(imgPaint)
			ctx.Fill()

			shadowPaint := nanovgo.BoxGradient(tx-1, ty, thumb+2, thumb+2, 5, 3, nanovgo.RGBA(0, 0, 0, 128), nanovgo.RGBA(0, 0, 0, 0))
			ctx.BeginPath()
			ctx.Rect(tx-5, ty-5, thumb+10, thumb+10)
			ctx.RoundedRect(tx, ty, thumb, thumb, 6)
			ctx.PathWinding(nanovgo.Hole)
			ctx.SetFillPaint(shadowPaint)
			ctx.Fill()

			ctx.BeginPath()
			ctx.RoundedRect(tx+0.5, ty+0.5, thumb-1, thumb-1, 4-0.5)
			ctx.SetStrokeWidth(1.0)
			ctx.SetStrokeColor(nanovgo.RGBA(255, 255, 255, 192))
			ctx.Stroke()
		}
	})

	// Hide fades
	fadePaint := nanovgo.LinearGradient(x, y, x, y+6, nanovgo.RGBA(200, 200, 200, 255), nanovgo.RGBA(200, 200, 200, 0))
	ctx.BeginPath()
	ctx.Rect(x+4, y, w-8, 6)
	ctx.SetFillPaint(fadePaint)
	ctx.Fill()

	fadePaint = nanovgo.LinearGradient(x, y+h, x, y+h-6, nanovgo.RGBA(200, 200, 200, 255), nanovgo.RGBA(200, 200, 200, 0))
	ctx.BeginPath()
	ctx.Rect(x+4, y+h-6, w-8, 6)
	ctx.SetFillPaint(fadePaint)
	ctx.Fill()

	// Scroll bar
	shadowPaint = nanovgo.BoxGradient(x+w-12+1, y+4+1, 8, h-8, 3, 4, nanovgo.RGBA(0, 0, 0, 32), nanovgo.RGBA(0, 0, 0, 92))
	ctx.BeginPath()
	ctx.RoundedRect(x+w-12, y+4, 8, h-8, 3)
	ctx.SetFillPaint(shadowPaint)
	//      ctx.FillColor(ctx.RGBA(255,0,0,128));
	ctx.Fill()

	scrollH := (h / stackh) * (h - 8)
	shadowPaint = nanovgo.BoxGradient(x+w-12-1, y+4+(h-8-scrollH)*u-1, 8, scrollH, 3, 4, nanovgo.RGBA(220, 220, 220, 255), nanovgo.RGBA(128, 128, 128, 255))
	ctx.BeginPath()
	ctx.RoundedRect(x+w-12+1, y+4+1+(h-8-scrollH)*u, 8-2, scrollH-2, 2)
	ctx.SetFillPaint(shadowPaint)
	//      ctx.FillColor(ctx.RGBA(0,0,0,128));
	ctx.Fill()
}

func drawColorWheel(ctx *nanovgo.Context, x, y, w, h, t float32) {
	var r0, r1, ax, ay, bx, by, aeps, r float32
	hue := sinF(t * 0.12)

	ctx.Save()
	defer ctx.Restore()
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
		ctx.Arc(cx, cy, r0, a0, a1, nanovgo.Clockwise)
		ctx.Arc(cx, cy, r1, a1, a0, nanovgo.CounterClockwise)
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
	ctx.PathWinding(nanovgo.Hole)
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
	ctx.PathWinding(nanovgo.Hole)
	ctx.SetFillPaint(paint)
	ctx.Fill()
}

func drawLines(ctx *nanovgo.Context, x, y, w, h, t float32) {
	var pad float32 = 5.0
	s := w/9.0 - pad*2
	joins := []nanovgo.LineCap{nanovgo.Miter, nanovgo.Round, nanovgo.Bevel}
	caps := []nanovgo.LineCap{nanovgo.Butt, nanovgo.Round, nanovgo.Square}

	ctx.Save()
	defer ctx.Restore()

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

			ctx.SetLineCap(nanovgo.Butt)
			ctx.SetLineJoin(nanovgo.Bevel)

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
}

func drawWidths(ctx *nanovgo.Context, x, y, width float32) {
	ctx.Save()
	defer ctx.Restore()

	ctx.SetStrokeColor(nanovgo.RGBA(0, 0, 0, 255))
	for i := 0; i < 20; i++ {
		w := (float32(i) + 0.5) * 0.1
		ctx.SetStrokeWidth(w)
		ctx.BeginPath()
		ctx.MoveTo(x, y)
		ctx.LineTo(x+width, y+width*0.3)
		ctx.Stroke()
		y += 10
	}
}

func drawCaps(ctx *nanovgo.Context, x, y, width float32) {
	caps := []nanovgo.LineCap{nanovgo.Butt, nanovgo.Round, nanovgo.Square}
	var lineWidth float32 = 8.0

	ctx.Save()
	defer ctx.Restore()

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
}

func drawParagraph(ctx *nanovgo.Context, x, y, width, height, mx, my float32) {
	text := "This is longer chunk of text.\n  \n  Would have used lorem ipsum but she    was busy jumping over the lazy dog with the fox and all the men who came to the aid of the party."

	ctx.Save()
	defer ctx.Restore()

	ctx.SetFontSize(18.0)
	ctx.SetFontFace("sans")
	ctx.SetTextAlign(nanovgo.AlignLeft | nanovgo.AlignTop)
	_, _, lineh := ctx.TextMetrics()
	// The text break API can be used to fill a large buffer of rows,
	// or to iterate over the text just few lines (or just one) at a time.
	// The "next" variable of the last returned item tells where to continue.
	runes := []rune(text)

	var gx, gy float32
	var gutter int
	lnum := 0

	for _, row := range ctx.TextBreakLinesRune(runes, width) {
		hit := mx > x && mx < (x+width) && my >= y && my < (y+lineh)

		ctx.BeginPath()
		var alpha uint8
		if hit {
			alpha = 64
		} else {
			alpha = 16
		}
		ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, alpha))
		ctx.Rect(x, y, row.Width, lineh)
		ctx.Fill()

		ctx.SetFillColor(nanovgo.RGBA(255, 255, 255, 255))
		ctx.TextRune(x, y, runes[row.StartIndex:row.EndIndex])

		if hit {
			var caretX float32
			if mx < x+row.Width/2 {
				caretX = x
			} else {
				caretX = x + row.Width
			}
			px := x
			lineRune := runes[row.StartIndex:row.EndIndex]
			glyphs := ctx.TextGlyphPositionsRune(x, y, lineRune)
			for j, glyph := range glyphs {
				x0 := glyph.X
				var x1 float32
				if j+1 < len(glyphs) {
					x1 = glyphs[j+1].X
				} else {
					x1 = x + row.Width
				}
				gx = x0*0.3 + x1*0.7
				if mx >= px && mx < gx {
					caretX = glyph.X
				}
				px = gx
			}
			ctx.BeginPath()
			ctx.SetFillColor(nanovgo.RGBA(255, 192, 0, 255))
			ctx.Rect(caretX, y, 1, lineh)
			ctx.Fill()

			gutter = lnum + 1
			gx = x - 10
			gy = y + lineh/2
		}
		lnum++
		y += lineh
	}

	if gutter > 0 {
		txt := strconv.Itoa(gutter)

		ctx.SetFontSize(13.0)
		ctx.SetTextAlign(nanovgo.AlignRight | nanovgo.AlignMiddle)

		_, bounds := ctx.TextBounds(gx, gy, txt)

		ctx.BeginPath()
		ctx.SetFillColor(nanovgo.RGBA(255, 192, 0, 255))
		ctx.RoundedRect(
			float32(int(bounds[0]-4)),
			float32(int(bounds[1]-2)),
			float32(int(bounds[2]-bounds[0])+8),
			float32(int(bounds[3]-bounds[1])+4),
			float32(int(bounds[3]-bounds[1])+4)/2.0-1.0)
		ctx.Fill()

		ctx.SetFillColor(nanovgo.RGBA(32, 32, 32, 255))
		ctx.Text(gx, gy, txt)
	}

	y += 20.0

	ctx.SetFontSize(13.0)
	ctx.SetTextAlign(nanovgo.AlignLeft | nanovgo.AlignTop)
	ctx.SetTextLineHeight(1.2)

	bounds := ctx.TextBoxBounds(x, y, 150, "Hover your mouse over the text to see calculated caret position.")

	// Fade the tooltip out when close to it.
	gx = absF((mx - (bounds[0]+bounds[2])*0.5) / (bounds[0] - bounds[2]))
	gy = absF((my - (bounds[1]+bounds[3])*0.5) / (bounds[1] - bounds[3]))
	a := maxF(gx, gy) - 0.5
	a = clampF(a, 0, 1)
	ctx.SetGlobalAlpha(a)

	ctx.BeginPath()
	ctx.SetFillColor(nanovgo.RGBA(220, 220, 220, 255))
	ctx.RoundedRect(bounds[0]-2, bounds[1]-2, float32(int(bounds[2]-bounds[0])+4), float32(int(bounds[3]-bounds[1])+4), 3)
	px := float32(int((bounds[2] + bounds[0]) / 2))
	ctx.MoveTo(px, bounds[1]-10)
	ctx.LineTo(px+7, bounds[1]+1)
	ctx.LineTo(px-7, bounds[1]+1)
	ctx.Fill()

	ctx.SetFillColor(nanovgo.RGBA(0, 0, 0, 220))
	ctx.TextBox(x, y, 150, "Hover your mouse over the text to see calculated caret position.")
}

func drawScissor(ctx *nanovgo.Context, x, y, t float32) {
	ctx.Save()

	// Draw first rect and set scissor to it's area.
	ctx.Translate(x, y)
	ctx.Rotate(nanovgo.DegToRad(5))
	ctx.BeginPath()
	ctx.Rect(-20, -20, 60, 40)
	ctx.SetFillColor(nanovgo.RGBA(255, 0, 0, 255))
	ctx.Fill()
	ctx.Scissor(-20, -20, 60, 40)

	// Draw second rectangle with offset and rotation.
	ctx.Translate(40, 0)
	ctx.Rotate(t)

	// Draw the intended second rectangle without any scissoring.
	ctx.Save()
	ctx.ResetScissor()
	ctx.BeginPath()
	ctx.Rect(-20, -10, 60, 30)
	ctx.SetFillColor(nanovgo.RGBA(255, 128, 0, 64))
	ctx.Fill()
	ctx.Restore()

	// Draw second rectangle with combined scissoring.
	ctx.IntersectScissor(-20, -10, 60, 30)
	ctx.BeginPath()
	ctx.Rect(-20, -10, 60, 30)
	ctx.SetFillColor(nanovgo.RGBA(255, 128, 0, 255))
	ctx.Fill()

	ctx.Restore()
}
