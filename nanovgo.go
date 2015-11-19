package nanovgo

import (
	"bytes"
	"github.com/shibukawa/nanovgo/fontstashmini"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
)

type Context struct {
	params         nvgParams
	commands       []float32
	commandX       float32
	commandY       float32
	states         []nvgState
	cache          nvgPathCache
	tessTol        float32
	distTol        float32
	fringeWidth    float32
	devicePxRatio  float32
	fs             *fontstashmini.FontStash
	fontImages     []int
	fontImageIdx   int
	drawCallCount  int
	fillTriCount   int
	strokeTriCount int
	textTriCount   int
}

func (c *Context) Delete() {

	for i, fontImage := range c.fontImages {
		if fontImage != 0 {
			c.DeleteImage(fontImage)
			c.fontImages[i] = 0
		}
	}
	c.params.renderDelete()
}

func (c *Context) BeginFrame(windowWidth, windowHeight int, devicePixelRatio float32) {
	log.Printf("Tris: draws:%d  fill:%d  stroke:%d  text:%d  TOT:%d\n",
		c.drawCallCount, c.fillTriCount, c.strokeTriCount, c.textTriCount,
		c.drawCallCount+c.fillTriCount+c.strokeTriCount+c.textTriCount)
	c.states = c.states[:]
	c.Save()
	c.Reset()

	c.params.renderViewport(windowWidth, windowHeight)

	c.drawCallCount = 0
	c.fillTriCount = 0
	c.strokeTriCount = 0
	c.textTriCount = 0
}

func (c *Context) CancelFrame() {
	c.params.renderCancel()
}

func (c *Context) EndFrame() {
	c.params.renderFlush()
	if c.fontImageIdx != 0 {
		fontImage := c.fontImages[c.fontImageIdx]
		if fontImage == 0 {
			return
		}
		iw, ih, _ := c.ImageSize(fontImage)
		j := 0
		for i := 0; i < c.fontImageIdx; i++ {
			nw, nh, _ := c.ImageSize(c.fontImages[i])
			if nw < iw || nh < ih {
				c.DeleteImage(c.fontImages[i])
			} else {
				c.fontImages[j] = c.fontImages[i]
				j++
			}
		}
		// make current font image to first
		c.fontImages[j] = c.fontImages[0]
		j++
		c.fontImages[0] = fontImage
		c.fontImageIdx = 0
		// clear all image after j
		for i := j; i < nvg_MAX_FONTIMAGES; i++ {
			c.fontImages[i] = 0
		}
	}
}

// Pushes and saves the current render state into a state stack.
// A matching Restore() must be used to restore the state.
func (c *Context) Save() {
	if len(c.states) >= nvg_MAX_STATES {
		return
	}
	if len(c.states) > 0 {
		c.states = append(c.states, c.states[len(c.states)-1])
	} else {
		c.states = append(c.states, nvgState{})
	}
}

// Pops and restores current render state.
func (c *Context) Restore() {
	nStates := len(c.states)
	if nStates > 1 {
		c.states = c.states[:nStates-1]
	}
}

// Resets current render state to default values. Does not affect the render state stack.
func (c *Context) Reset() {
	c.getState().reset()
}

func (c *Context) SetStrokeWidth(width float32) {
	c.getState().strokeWidth = width
}

func (c *Context) StrokeWidth() float32 {
	return c.getState().strokeWidth
}

func (c *Context) SetMiterLimit(limit float32) {
	c.getState().miterLimit = limit
}

func (c *Context) MiterLimit() float32 {
	return c.getState().miterLimit
}

func (c *Context) SetLineCap(cap LineCap) {
	c.getState().lineCap = cap
}

func (c *Context) LineCap() LineCap {
	return c.getState().lineCap
}

func (c *Context) SetLineJoin(joint LineCap) {
	c.getState().lineJoin = joint
}

func (c *Context) LineJoin() LineCap {
	return c.getState().lineJoin
}

func (c *Context) SetGlobalAlpha(alpha float32) {
	c.getState().alpha = alpha
}

func (c *Context) GlobalAlpha() float32 {
	return c.getState().alpha
}

func (ctx *Context) Transform(a, b, c, d, e, f float32) {
	t := TransformMatrix{a, b, c, d, e, f}
	ctx.getState().xform.PreMultiply(t)
}

func (c *Context) ResetTransform() {
	c.getState().xform.Identity()
}

func (c *Context) Translate(x, y float32) {
	c.getState().xform.PreMultiply(TransformMatrixTranslate(x, y))
}

func (c *Context) Rotate(angle float32) {
	c.getState().xform.PreMultiply(TransformMatrixRotate(angle))
}

func (c *Context) SkewX(angle float32) {
	c.getState().xform.PreMultiply(TransformMatrixSkewX(angle))
}

func (c *Context) SkewY(angle float32) {
	c.getState().xform.PreMultiply(TransformMatrixSkewY(angle))
}

func (c *Context) Scale(x, y float32) {
	c.getState().xform.PreMultiply(TransformMatrixScale(x, y))
}

func (c *Context) CurrentTransform() TransformMatrix {
	return c.getState().xform
}

func (c *Context) SetStrokeColor(color Color) {
	c.getState().stroke.setPaintColor(color)
}

func (c *Context) SetStrokePaint(paint Paint) {
	state := c.getState()
	state.stroke = paint
	state.stroke.xform.Multiply(state.xform)
}

func (c *Context) SetFillColor(color Color) {
	c.getState().fill.setPaintColor(color)
}

func (c *Context) SetFillPaint(paint Paint) {
	state := c.getState()
	state.fill = paint
	state.fill.xform.Multiply(state.xform)
}

func (c *Context) CreateImage(filePath string, flags ImageFlags) int {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return 0
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return 0
	}
	return c.CreateImageFromGoImage(flags, img)
}

func (c *Context) CreateImageFromMemory(flags ImageFlags, data []byte) int {
	reader := bytes.NewReader(data)
	img, _, err := image.Decode(reader)
	if err != nil {
		return 0
	}
	return c.CreateImageFromGoImage(flags, img)
}

func (c *Context) CreateImageFromGoImage(imageFlag ImageFlags, img image.Image) int {
	bounds := img.Bounds()
	size := bounds.Size()
	rgba, ok := img.(*image.RGBA)
	if ok {
		return c.CreateImageRGBA(size.X, size.Y, imageFlag, rgba.Pix)
	}
	rgba = image.NewRGBA(bounds)
	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	return c.CreateImageRGBA(size.X, size.Y, imageFlag, rgba.Pix)
}

func (c *Context) CreateImageRGBA(w, h int, imageFlags ImageFlags, data []byte) int {
	return c.params.renderCreateTexture(nvg_TEXTURE_RGBA, w, h, imageFlags, data)
}

func (c *Context) ImageSize(img int) (int, int, error) {
	return c.params.renderGetTextureSize(img)
}

func (c *Context) DeleteImage(img int) {
	c.params.renderDeleteTexture(img)
}

func (c *Context) Scissor(x, y, w, h float32) {
	state := c.getState()

	w = maxF(0.0, w)
	h = maxF(0.0, h)

	state.scissor.xform = TransformMatrixTranslate(x+w*0.5, y+h*0.5)
	state.scissor.xform.Multiply(state.xform)

	state.scissor.extent = [2]float32{w * 0.5, h * 0.5}
}

func (c *Context) IntersectScissor(x, y, w, h float32) {
	state := c.getState()

	if state.scissor.extent[0] < 0 {
		c.Scissor(x, y, w, h)
		return
	}

	pXform := state.scissor.xform
	ex := state.scissor.extent[0]
	ey := state.scissor.extent[1]

	invXform := state.xform.Inverse()
	pXform.Multiply(invXform)

	teX := ex * absF(pXform[0]) * ey * absF(pXform[2])
	teY := ex * absF(pXform[1]) * ey * absF(pXform[3])
	rect := intersectRects(pXform[4]-teX, pXform[5]-teY, teX*2, teY*2, x, y, w, h)
	c.Scissor(rect[0], rect[1], rect[2], rect[3])
}

func (c *Context) ResetScissor() {
	state := c.getState()

	state.scissor.xform = TransformMatrix{0, 0, 0, 0, 0, 0}
	state.scissor.extent = [2]float32{-1.0, -1.0}
}

func (c *Context) BeginPath() {
	c.commands = c.commands[:0]
	c.cache.clearPathCache()
}

func (c *Context) MoveTo(x, y float32) {
	c.appendCommand([]float32{float32(nvg_MOVETO), x, y})
}

func (c *Context) LineTo(x, y float32) {
	c.appendCommand([]float32{float32(nvg_LINETO), x, y})
}

func (c *Context) BezierTo(c1x, c1y, c2x, c2y, x, y float32) {
	c.appendCommand([]float32{float32(nvg_BEZIERTO), c1x, c1y, c2x, c2y, x, y})
}

func (c *Context) QuadTo(cx, cy, x, y float32) {
	x0 := c.commandX
	y0 := c.commandY
	c.appendCommand([]float32{float32(nvg_BEZIERTO),
		x0 + 2.0/3.0*(cx-x0), y0 + 2.0/3.0*(cy-y0),
		x + 2.0/3.0*(cx-x), y + 2.0/3.0*(cy-y),
		x, y,
	})
}

func (c *Context) Arc(cx, cy, r, a0, a1 float32, dir Winding) {
	var move nvgCommands
	if len(c.commands) > 0 {
		move = nvg_LINETO
	} else {
		move = nvg_MOVETO
	}

	// Clamp angles
	da := a1 - a0
	if dir == CW {
		if absF(da) >= PI*2 {
			da = PI * 2
		} else {
			for da < 0.0 {
				da += PI * 2
			}
		}
	} else {
		if absF(da) >= PI*2 {
			da = -PI * 2
		} else {
			for da > 0.0 {
				da -= PI * 2
			}
		}
	}
	// Split arc into max 90 degree segments.
	nDivs := clampI(int(absF(da)/(PI*0.5)+0.5), 1, 5)
	hda := da / float32(nDivs) / 2.0
	sin, cos := sinCosF(hda)
	kappa := absF(4.0 / 3.0 * (1.0 - cos) / sin)

	if dir == CCW {
		kappa = -kappa
	}
	values := make([]float32, 0, 3+5*7+100)
	var px, py, pTanX, pTanY float32

	for i := 0; i <= nDivs; i++ {
		a := a0 + da*float32(i)/float32(nDivs)
		dy, dx := sinCosF(a)
		x := cx + dx*r
		y := cy + dy*r
		tanX := -dy * r * kappa
		tanY := dx * r * kappa
		if i == 0 {
			values = append(values, float32(move), x, y)
		} else {
			values = append(values, float32(nvg_BEZIERTO), px+pTanX, py+pTanY, x-tanX, y-tanY, x, y)
		}
		px = x
		py = y
		pTanX = tanX
		pTanY = tanY
	}
	c.appendCommand(values)
}

func (c *Context) ArcTo(x1, y1, x2, y2, radius float32) {
	if len(c.commands) == 0 {
		return
	}
	x0 := c.commandX
	y0 := c.commandY

	// Handle degenerate cases.
	if ptEquals(x0, y0, x1, y1, c.distTol) ||
		ptEquals(x1, y1, x2, y2, c.distTol) ||
		distPtSeg(x1, y1, x0, y0, x2, y2) < c.distTol*c.distTol ||
		radius < c.distTol {
		c.LineTo(x1, y1)
		return
	}

	// Calculate tangential circle to lines (x0,y0)-(x1,y1) and (x1,y1)-(x2,y2).
	dx0 := x0 - x1
	dy0 := y0 - y1
	dx1 := x2 - x1
	dy1 := y2 - y1
	_, dx0, dy0 = normalize(dx0, dy0)
	_, dx1, dy1 = normalize(dx1, dy1)
	a := acosF(dx0*dx1 + dy0*dy1)
	d := radius / tanF(a/2.0)

	if d > 10000.0 {
		c.LineTo(x1, y1)
		return
	}
	var cx, cy, a0, a1 float32
	var dir Winding
	if cross(dx0, dy0, dx1, dy1) > 0.0 {
		cx = x1 + dx0*d + dy0*radius
		cy = y1 + dy0*d + -dx0*radius
		a0 = atan2F(dx0, -dy0)
		a1 = atan2F(-dx1, dy1)
		dir = CW
	} else {
		cx = x1 + dx0*d + -dy0*radius
		cy = y1 + dy0*d + dx0*radius
		a0 = atan2F(-dx0, dy0)
		a1 = atan2F(dx1, -dy1)
		dir = CCW
	}
	c.Arc(cx, cy, radius, a0, a1, dir)
}

func (c *Context) Rect(x, y, w, h float32) {
	c.appendCommand([]float32{
		float32(nvg_MOVETO), x, y,
		float32(nvg_LINETO), x, y + h,
		float32(nvg_LINETO), x + w, y + h,
		float32(nvg_LINETO), x + w, y,
		float32(nvg_CLOSE),
	})
}

func (c *Context) RoundRect(x, y, w, h, r float32) {
	if r < 0.1 {
		c.Rect(x, y, w, h)
	} else {
		rx := minF(r, absF(w)*0.5) * signF(w)
		ry := minF(r, absF(h)*0.5) * signF(h)
		c.appendCommand([]float32{
			float32(nvg_MOVETO), x, y + ry,
			float32(nvg_LINETO), x, y + h - ry,
			float32(nvg_BEZIERTO), x, y + h - ry*(1-KAPPA90), x + rx*(1-KAPPA90), y + h, x + rx, y + h,
			float32(nvg_LINETO), x + w - rx, y + h,
			float32(nvg_BEZIERTO), x + w - rx*(1-KAPPA90), y + h, x + w, y + h - ry*(1-KAPPA90), x + w, y + h - ry,
			float32(nvg_LINETO), x + w, y + ry,
			float32(nvg_BEZIERTO), x + w, y + ry*(1-KAPPA90), x + w - rx*(1-KAPPA90), y, x + w - rx, y,
			float32(nvg_LINETO), x + rx, y,
			float32(nvg_BEZIERTO), x + rx*(1-KAPPA90), y, x, y + ry*(1-KAPPA90), x, y + ry,
			float32(nvg_CLOSE),
		})
	}
}

func (c *Context) Ellipse(cx, cy, rx, ry float32) {
	c.appendCommand([]float32{
		float32(nvg_MOVETO), cx - rx, cy,
		float32(nvg_BEZIERTO), cx - rx, cy + ry*KAPPA90, cx - rx*KAPPA90, cy + ry, cx, cy + ry,
		float32(nvg_BEZIERTO), cx + rx*KAPPA90, cy + ry, cx + rx, cy + ry*KAPPA90, cx + rx, cy,
		float32(nvg_BEZIERTO), cx + rx, cy - ry*KAPPA90, cx + rx*KAPPA90, cy - ry, cx, cy - ry,
		float32(nvg_BEZIERTO), cx - rx*KAPPA90, cy - ry, cx - rx, cy - ry*KAPPA90, cx - rx, cy,
		float32(nvg_CLOSE),
	})
}

func (c *Context) Circle(cx, cy, r float32) {
	c.Ellipse(cx, cy, r, r)
}

func (c *Context) ClosePath() {
	c.appendCommand([]float32{float32(nvg_CLOSE)})
}

func (c *Context) PathWinding(dir Winding) {
	c.appendCommand([]float32{float32(nvg_WINDING), float32(dir)})
}

func (c *Context) DebugDumpPathCache() {
	log.Printf("Dumping %d cached paths\n", len(c.cache.paths))
	for i := 0; i < len(c.cache.paths); i++ {
		path := &c.cache.paths[i]
		log.Printf(" - Path %d\n", i)
		if len(path.fills) > 0 {
			log.Printf("   - fill: %d\n", len(path.fills))
			for _, fill := range path.fills {
				log.Printf("%f\t%f\n", fill.x, fill.y)
			}
		}
		if len(path.strokes) > 0 {
			log.Printf("   - strokes: %d\n", len(path.strokes))
			for _, stroke := range path.strokes {
				log.Printf("%f\t%f\n", stroke.x, stroke.y)
			}
		}
	}
}

func (c *Context) Fill() {
	state := c.getState()
	fillPaint := state.fill
	c.flattenPaths()
	if c.params.edgeAntiAlias() {
		c.cache.expandFill(c.fringeWidth, MITER, 2.4, c.fringeWidth)
	} else {
		c.cache.expandFill(0.0, MITER, 2.4, c.fringeWidth)
	}

	// Apply global alpha
	fillPaint.innerColor.A *= state.alpha
	fillPaint.outerColor.A *= state.alpha

	c.params.renderFill(&fillPaint, &state.scissor, c.fringeWidth, c.cache.bounds, c.cache.paths)

	// Count triangles
	for i := 0; i < len(c.cache.paths); i++ {
		path := &c.cache.paths[i]
		c.fillTriCount += len(path.fills) - 2
		c.strokeTriCount += len(path.strokes) - 2
		c.drawCallCount += 2
	}
}

func (c *Context) Stroke() {
	state := c.getState()
	scale := getAverageScale(state.xform)
	strokeWidth := clampF(state.strokeWidth*scale, 0.0, 200.0)
	strokePaint := state.stroke

	if strokeWidth < c.fringeWidth {
		// If the stroke width is less than pixel size, use alpha to emulate coverage.
		// Since coverage is area, scale by alpha*alpha.
		alpha := clampF(strokeWidth/c.fringeWidth, 0.0, 1.0)
		strokePaint.innerColor.A *= alpha * alpha
		strokePaint.outerColor.A *= alpha * alpha
		strokeWidth = c.fringeWidth
	}

	// Apply global alpha
	strokePaint.innerColor.A *= state.alpha
	strokePaint.outerColor.A *= state.alpha

	c.flattenPaths()
	if c.params.edgeAntiAlias() {
		c.cache.expandStroke(strokeWidth*0.5+c.fringeWidth*0.5, state.lineCap, state.lineJoin, state.miterLimit, c.fringeWidth, c.tessTol)
	} else {
		c.cache.expandStroke(strokeWidth*0.5, state.lineCap, state.lineJoin, state.miterLimit, c.fringeWidth, c.tessTol)
	}
	c.params.renderStroke(&strokePaint, &state.scissor, c.fringeWidth, strokeWidth, c.cache.paths)

	// Count triangles
	for i := 0; i < len(c.cache.paths); i++ {
		path := &c.cache.paths[i]
		c.strokeTriCount += len(path.strokes) - 2
		c.drawCallCount += 2
	}
}

func createInternal(params nvgParams) (*Context, error) {
	context := &Context{
		params:     params,
		states:     make([]nvgState, 0, nvg_MAX_STATES),
		fontImages: make([]int, nvg_MAX_FONTIMAGES),
		commands:   make([]float32, 0, nvg_INIT_COMMANDS_SIZE),
		cache: nvgPathCache{
			points:   make([]nvgPoint, 0, nvg_INIT_POINTS_SIZE),
			paths:    make([]nvgPath, 0, nvg_INIT_PATHS_SIZE),
			vertexes: make([]nvgVertex, 0, nvg_INIT_VERTS_SIZE),
		},
	}
	context.Save()
	context.Reset()
	context.setDevicePixelRatio(1.0)
	context.params.renderCreate()

	context.fs = fontstashmini.New(nvg_INIT_FONTIMAGE_SIZE, nvg_INIT_FONTIMAGE_SIZE)

	context.fontImages[0] = context.params.renderCreateTexture(nvg_TEXTURE_ALPHA, nvg_INIT_FONTIMAGE_SIZE, nvg_INIT_FONTIMAGE_SIZE, 0, nil)
	context.fontImageIdx = 0

	return context, nil
}

func (c *Context) setDevicePixelRatio(ratio float32) {
	c.tessTol = 0.25 / ratio
	c.distTol = 0.01 / ratio
	c.fringeWidth = 1.0 / ratio
	c.devicePxRatio = ratio
}

func (c *Context) getState() *nvgState {
	return &c.states[len(c.states)-1]
}

func (c *Context) appendCommand(vals []float32) {
	xForm := c.getState().xform

	if nvgCommands(vals[0]) != nvg_CLOSE && nvgCommands(vals[0]) != nvg_WINDING {
		c.commandX = vals[len(vals)-2]
		c.commandY = vals[len(vals)-1]
	}

	i := 0
	for i < len(vals) {
		switch nvgCommands(vals[i]) {
		case nvg_MOVETO:
			vals[i+1], vals[i+2] = xForm.Point(vals[i+1], vals[i+2])
			i += 3
		case nvg_LINETO:
			vals[i+1], vals[i+2] = xForm.Point(vals[i+1], vals[i+2])
			i += 3
		case nvg_BEZIERTO:
			vals[i+1], vals[i+2] = xForm.Point(vals[i+1], vals[i+2])
			vals[i+3], vals[i+4] = xForm.Point(vals[i+3], vals[i+4])
			vals[i+5], vals[i+6] = xForm.Point(vals[i+5], vals[i+6])
			i += 7
		case nvg_CLOSE:
			i++
		case nvg_WINDING:
			i += 2
		default:
			i++
		}

	}
	c.commands = append(c.commands, vals...)
}

func (c *Context) flattenPaths() {
	cache := &c.cache
	if len(cache.paths) > 0 {
		return
	}
	// Flatten
	i := 0
	for i < len(c.commands) {
		switch nvgCommands(c.commands[i]) {
		case nvg_MOVETO:
			cache.addPath()
			cache.addPoint(c.commands[i+1], c.commands[i+2], nvg_PT_CORNER, c.distTol)
			i += 3
		case nvg_LINETO:
			cache.addPoint(c.commands[i+1], c.commands[i+2], nvg_PT_CORNER, c.distTol)
			i += 3
		case nvg_BEZIERTO:
			last := cache.lastPoint()
			if last != nil {
				cache.tesselateBezier(
					last.x, last.y,
					c.commands[i+1], c.commands[i+2],
					c.commands[i+3], c.commands[i+4],
					c.commands[i+5], c.commands[i+6], 0, nvg_PT_CORNER, c.tessTol, c.distTol)
			}
			i += 7
		case nvg_CLOSE:
			cache.closePath()
			i++
		case nvg_WINDING:
			cache.pathWinding(Winding(c.commands[i+1]))
			i += 2
		default:
			i++
		}
	}

	cache.bounds = [4]float32{1e6, 1e6, -1e6, -1e6}

	// Calculate the direction and length of line segments.
	for j := 0; j < len(cache.paths); j++ {
		path := &cache.paths[j]
		points := cache.points[path.first:]
		p0 := &points[path.count-1]
		p1Index := 0
		p1 := &points[p1Index]
		if ptEquals(p0.x, p0.y, p1.x, p1.y, c.distTol) {
			path.count--
			p0 = &points[path.count-1]
			path.closed = true
		}
		points = points[:path.count]
		// Enforce winding.
		if path.count > 0 {
			area := polyArea(points)
			if path.winding == CCW && area < 0.0 {
				polyReverse(points)
			} else if path.winding == CW && area > 0.0 {
				polyReverse(points)
			}
		}

		for i := 0; i < path.count; i++ {
			// Calculate segment direction and length
			p0.len, p0.dx, p0.dy = normalize(p1.x-p0.x, p1.y-p0.y)
			// Update bounds
			cache.bounds = [4]float32{
				minF(cache.bounds[0], p0.x),
				minF(cache.bounds[1], p0.y),
				maxF(cache.bounds[2], p0.x),
				maxF(cache.bounds[3], p0.y),
			}
			// Advance
			p1Index++
			p0 = p1
			if len(points) != p1Index {
				p1 = &points[p1Index]
			}
		}
	}
}
