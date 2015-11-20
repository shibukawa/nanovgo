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

// State Handling
//
// NanoVG contains state which represents how paths will be rendered.
// The state contains transform, fill and stroke styles, text and font styles,
// and scissor clipping.
//
// Render styles
//
// Fill and stroke render style can be either a solid color or a paint which is a gradient or a pattern.
// Solid color is simply defined as a color value, different kinds of paints can be created
// using nanovgo.LinearGradient(), nanovgo.BoxGradient(), nanovgo.RadialGradient() and nanovgo.ImagePattern().
//
// Current render style can be saved and restored using nvgSave() and nvgRestore().
//
// Transforms
//
// The paths, gradients, patterns and scissor region are transformed by an transformation
// matrix at the time when they are passed to the API.
// The current transformation matrix is a affine matrix:
//   [sx kx tx]
//   [ky sy ty]
//   [ 0  0  1]
// Where: sx,sy define scaling, kx,ky skewing, and tx,ty translation.
// The last row is assumed to be 0,0,1 and is not stored.
//
// Apart from nvgResetTransform(), each transformation function first creates
// specific transformation matrix and pre-multiplies the current transformation by it.
//
// Current coordinate system (transformation) can be saved and restored using nvgSave() and nvgRestore().
//
// Images
//
// NanoVG allows you to load jpg, png, psd, tga, pic and gif files to be used for rendering.
// In addition you can upload your own image. The image loading is provided by stb_image.
// The parameter imageFlags is combination of flags defined in nanovgo.ImageFlags.
//
// Paints
//
// NanoVG supports four types of paints: linear gradient, box gradient, radial gradient and image pattern.
// These can be used as paints for strokes and fills.
//
// Scissoring
//
// Scissoring allows you to clip the rendering into a rectangle. This is useful for various
// user interface cases like rendering a text edit or a timeline.
//
// Paths
//
// Drawing a new shape starts with nvgBeginPath(), it clears all the currently defined paths.
// Then you define one or more paths and sub-paths which describe the shape. The are functions
// to draw common shapes like rectangles and circles, and lower level step-by-step functions,
// which allow to define a path curve by curve.
//
// NanoVG uses even-odd fill rule to draw the shapes. Solid shapes should have counter clockwise
// winding and holes should have counter clockwise order. To specify winding of a path you can
// call nvgPathWinding(). This is useful especially for the common shapes, which are drawn CCW.
//
// Finally you can fill the path using current fill style by calling nvgFill(), and stroke it
// with current stroke style by calling nvgStroke().
//
// The curve segments and sub-paths are transformed by the current transform.
//
// Text
//
// NanoVG allows you to load .ttf files and use the font to render text.
//
// The appearance of the text can be defined by setting the current text style
// and by specifying the fill color. Common text and font settings such as
// font size, letter spacing and text align are supported. Font blur allows you
// to create simple text effects such as drop shadows.
//
// At render time the font face can be set based on the font handles or name.
//
// Font measure functions return values in local space, the calculations are
// carried in the same resolution as the final rendering. This is done because
// the text glyph positions are snapped to the nearest pixels sharp rendering.
//
// The local space means that values are not rotated or scale as per the current
// transformation. For example if you set font size to 12, which would mean that
// line height is 16, then regardless of the current scaling and rotation, the
// returned line height is always 16. Some measures may vary because of the scaling
// since aforementioned pixel snapping.
//
// While this may sound a little odd, the setup allows you to always render the
// same way regardless of scaling. I.e. following works regardless of scaling:
//
//              vg.TextBounds(x, y, "Text me up.", bounds)
//              vg.BeginPath()
//              vg.RoundedRect(bounds[0],bounds[1], bounds[2]-bounds[0], bounds[3]-bounds[1])
//              vg.Fill()
//
// Note: currently only solid color fill is supported for text.
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

// Begin drawing a new frame
// Calls to NanoVGo drawing API should be wrapped in Context.BeginFrame() & Context.EndFrame()
// Context.BeginFrame() defines the size of the window to render to in relation currently
// set viewport (i.e. glViewport on GL backends). Device pixel ration allows to
// control the rendering on Hi-DPI devices.
// For example, GLFW returns two dimension for an opened window: window size and
// frame buffer size. In that case you would set windowWidth/Height to the window size
// devicePixelRatio to: frameBufferWidth / windowWidth.
func (c *Context) BeginFrame(windowWidth, windowHeight int, devicePixelRatio float32) {
	/*log.Printf("Tris: draws:%d  fill:%d  stroke:%d  text:%d  TOT:%d\n",
	c.drawCallCount, c.fillTriCount, c.strokeTriCount, c.textTriCount,
	c.drawCallCount+c.fillTriCount+c.strokeTriCount+c.textTriCount)*/
	c.states = c.states[:]
	c.Save()
	c.Reset()

	c.params.renderViewport(windowWidth, windowHeight)

	c.drawCallCount = 0
	c.fillTriCount = 0
	c.strokeTriCount = 0
	c.textTriCount = 0
}

// Cancels drawing the current frame.
func (c *Context) CancelFrame() {
	c.params.renderCancel()
}

// Ends drawing flushing remaining render state.
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

// Sets the stroke width of the stroke style.
func (c *Context) SetStrokeWidth(width float32) {
	c.getState().strokeWidth = width
}

// Gets the stroke width of the stroke style.
func (c *Context) StrokeWidth() float32 {
	return c.getState().strokeWidth
}

// Sets the miter limit of the stroke style.
// Miter limit controls when a sharp corner is beveled.
func (c *Context) SetMiterLimit(limit float32) {
	c.getState().miterLimit = limit
}

// Gets the miter limit of the stroke style.
func (c *Context) MiterLimit() float32 {
	return c.getState().miterLimit
}

// Sets how the end of the line (cap) is drawn,
// Can be one of: nanovgo.BUTT (default), nanovgo.ROUND, nanovgo.SQUARE.
func (c *Context) SetLineCap(cap LineCap) {
	c.getState().lineCap = cap
}

// Gets how the end of the line (cap) is drawn,
func (c *Context) LineCap() LineCap {
	return c.getState().lineCap
}

// Sets how sharp path corners are drawn.
// Can be one of nanovgo.MITER (default), nanovgo.ROUND, nanovgo.BEVEL.
func (c *Context) SetLineJoin(joint LineCap) {
	c.getState().lineJoin = joint
}

// Gets how sharp path corners are drawn.
func (c *Context) LineJoin() LineCap {
	return c.getState().lineJoin
}

// Sets the transparency applied to all rendered shapes.
// Already transparent paths will get proportionally more transparent as well.
func (c *Context) SetGlobalAlpha(alpha float32) {
	c.getState().alpha = alpha
}

// Gets the transparency applied to all rendered shapes.
func (c *Context) GlobalAlpha() float32 {
	return c.getState().alpha
}

// Premultiplies current coordinate system by specified matrix.
func (ctx *Context) SetTransform(t TransformMatrix) {
	ctx.getState().xform.PreMultiply(t)
}

// Premultiplies current coordinate system by specified matrix.
// The parameters are interpreted as matrix as follows:
//   [a c e]
//   [b d f]
//   [0 0 1]
func (ctx *Context) SetTransformByValue(a, b, c, d, e, f float32) {
	t := TransformMatrix{a, b, c, d, e, f}
	ctx.getState().xform.PreMultiply(t)
}

// Resets current transform to a identity matrix.
func (c *Context) ResetTransform() {
	c.getState().xform.Identity()
}

// Translates current coordinate system.
func (c *Context) Translate(x, y float32) {
	c.getState().xform.PreMultiply(TransformMatrixTranslate(x, y))
}

// Rotates current coordinate system. Angle is specified in radians.
func (c *Context) Rotate(angle float32) {
	c.getState().xform.PreMultiply(TransformMatrixRotate(angle))
}

// Skews the current coordinate system along X axis. Angle is specified in radians.
func (c *Context) SkewX(angle float32) {
	c.getState().xform.PreMultiply(TransformMatrixSkewX(angle))
}

// Skews the current coordinate system along Y axis. Angle is specified in radians.
func (c *Context) SkewY(angle float32) {
	c.getState().xform.PreMultiply(TransformMatrixSkewY(angle))
}

// Scales the current coordinate system.
func (c *Context) Scale(x, y float32) {
	c.getState().xform.PreMultiply(TransformMatrixScale(x, y))
}

// Returns the top part (a-f) of the current transformation matrix.
//   [a c e]
//   [b d f]
//   [0 0 1]
// There should be space for 6 floats in the return buffer for the values a-f.
func (c *Context) CurrentTransform() TransformMatrix {
	return c.getState().xform
}

// Sets current stroke style to a solid color.
func (c *Context) SetStrokeColor(color Color) {
	c.getState().stroke.setPaintColor(color)
}

// Sets current stroke style to a paint, which can be a one of the gradients or a pattern.
func (c *Context) SetStrokePaint(paint Paint) {
	state := c.getState()
	state.stroke = paint
	state.stroke.xform.Multiply(state.xform)
}

// Sets current fill style to a solid color.
func (c *Context) SetFillColor(color Color) {
	c.getState().fill.setPaintColor(color)
}

// Sets current fill style to a paint, which can be a one of the gradients or a pattern.
func (c *Context) SetFillPaint(paint Paint) {
	state := c.getState()
	state.fill = paint
	state.fill.xform.Multiply(state.xform)
}

// Creates image by loading it from the disk from specified file name.
// Returns handle to the image.
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

// Creates image by loading it from the specified chunk of memory.
// Returns handle to the image.
func (c *Context) CreateImageFromMemory(flags ImageFlags, data []byte) int {
	reader := bytes.NewReader(data)
	img, _, err := image.Decode(reader)
	if err != nil {
		return 0
	}
	return c.CreateImageFromGoImage(flags, img)
}

// Creates image by loading it from the specified image.Image object.
// Returns handle to the image.
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

// Creates image from specified image data.
// Returns handle to the image.
func (c *Context) CreateImageRGBA(w, h int, imageFlags ImageFlags, data []byte) int {
	return c.params.renderCreateTexture(nvg_TEXTURE_RGBA, w, h, imageFlags, data)
}

// Updates image data specified by image handle.
func (c *Context) UpdateImage(img int, data []byte) error {
	w, h, err := c.params.renderGetTextureSize(img)
	if err != nil {
		return err
	}
	return c.params.renderUpdateTexture(img, 0, 0, w, h, data)
}

// Returns the dimensions of a created image.
func (c *Context) ImageSize(img int) (int, int, error) {
	return c.params.renderGetTextureSize(img)
}

// Deletes created image.
func (c *Context) DeleteImage(img int) {
	c.params.renderDeleteTexture(img)
}

// Sets the current scissor rectangle.
// The scissor rectangle is transformed by the current transform.
func (c *Context) Scissor(x, y, w, h float32) {
	state := c.getState()

	w = maxF(0.0, w)
	h = maxF(0.0, h)

	state.scissor.xform = TransformMatrixTranslate(x+w*0.5, y+h*0.5)
	state.scissor.xform.Multiply(state.xform)

	state.scissor.extent = [2]float32{w * 0.5, h * 0.5}
}

// Intersects current scissor rectangle with the specified rectangle.
// The scissor rectangle is transformed by the current transform.
// Note: in case the rotation of previous scissor rect differs from
// the current one, the intersection will be done between the specified
// rectangle and the previous scissor rectangle transformed in the current
// transform space. The resulting shape is always rectangle.
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

// Reset and disables scissoring.
func (c *Context) ResetScissor() {
	state := c.getState()

	state.scissor.xform = TransformMatrix{0, 0, 0, 0, 0, 0}
	state.scissor.extent = [2]float32{-1.0, -1.0}
}

// Clears the current path and sub-paths.
func (c *Context) BeginPath() {
	c.commands = c.commands[:0]
	c.cache.clearPathCache()
}

// Starts new sub-path with specified point as first point.
func (c *Context) MoveTo(x, y float32) {
	c.appendCommand([]float32{float32(nvg_MOVETO), x, y})
}

// Adds line segment from the last point in the path to the specified point.
func (c *Context) LineTo(x, y float32) {
	c.appendCommand([]float32{float32(nvg_LINETO), x, y})
}

// Adds cubic bezier segment from last point in the path via two control points to the specified point.
func (c *Context) BezierTo(c1x, c1y, c2x, c2y, x, y float32) {
	c.appendCommand([]float32{float32(nvg_BEZIERTO), c1x, c1y, c2x, c2y, x, y})
}

// Adds quadratic bezier segment from last point in the path via a control point to the specified point.
func (c *Context) QuadTo(cx, cy, x, y float32) {
	x0 := c.commandX
	y0 := c.commandY
	c.appendCommand([]float32{float32(nvg_BEZIERTO),
		x0 + 2.0/3.0*(cx-x0), y0 + 2.0/3.0*(cy-y0),
		x + 2.0/3.0*(cx-x), y + 2.0/3.0*(cy-y),
		x, y,
	})
}

// Creates new circle arc shaped sub-path. The arc center is at cx,cy, the arc radius is r,
// and the arc is drawn from angle a0 to a1, and swept in direction dir (nanovgo.CCW, or nanovgo.CW).
// Angles are specified in radians.
func (c *Context) Arc(cx, cy, r, a0, a1 float32, dir Direction) {
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

// Adds an arc segment at the corner defined by the last path point, and two specified points.
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
	var dir Direction
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

// Creates new rectangle shaped sub-path.
func (c *Context) Rect(x, y, w, h float32) {
	c.appendCommand([]float32{
		float32(nvg_MOVETO), x, y,
		float32(nvg_LINETO), x, y + h,
		float32(nvg_LINETO), x + w, y + h,
		float32(nvg_LINETO), x + w, y,
		float32(nvg_CLOSE),
	})
}

// Creates new rounded rectangle shaped sub-path.
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

// Creates new ellipse shaped sub-path.
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

// Creates new circle shaped sub-path.
func (c *Context) Circle(cx, cy, r float32) {
	c.Ellipse(cx, cy, r, r)
}

// Closes current sub-path with a line segment.
func (c *Context) ClosePath() {
	c.appendCommand([]float32{float32(nvg_CLOSE)})
}

// Sets the current sub-path winding, see nanovgo.Winding.
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

// Fills the current path with current fill style.
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

// Fills the current path with current stroke style.
func (c *Context) Stroke() {
	state := c.getState()
	scale := state.xform.getAverageScale()
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

// Creates font by loading it from the disk from specified file name.
// Returns handle to the font.
func (c *Context) CreateFont(name, filePath string) int {
	return c.fs.AddFont(name, filePath)
}

// Creates image by loading it from the specified memory chunk.
// Returns handle to the font.
func (c *Context) CreateFontFromMemory(name string, data []byte, freeData uint8) int {
	return c.fs.AddFontFromMemory(name, data, freeData)
}

// Finds a loaded font of specified name, and returns handle to it, or -1 if the font is not found.
func (c *Context) FindFont(name string) int {
	return c.fs.GetFontByName(name)
}

// Sets the font size of current text style.
func (c *Context) SetFontSize(size float32) {
	c.getState().fontSize = size
}

// Gets the font size of current text style.
func (c *Context) FontSize() float32 {
	return c.getState().fontSize
}

// Sets the font blur of current text style.
func (c *Context) SetFontBlur(blur float32) {
	c.getState().fontBlur = blur
}

// Gets the font blur of current text style.
func (c *Context) FontBlur() float32 {
	return c.getState().fontBlur
}

// Sets the letter spacing of current text style.
func (c *Context) SetTextLetterSpacing(spacing float32) {
	c.getState().letterSpacing = spacing
}

// Gets the letter spacing of current text style.
func (c *Context) TextLetterSpacing() float32 {
	return c.getState().letterSpacing
}

// Sets the line height of current text style.
func (c *Context) SetTextLineHeight(lineHeight float32) {
	c.getState().lineHeight = lineHeight
}

// Gets the line height of current text style.
func (c *Context) TextLineHeight() float32 {
	return c.getState().lineHeight
}

// Sets the text align of current text style.
func (c *Context) SetTextAlign(align Align) {
	c.getState().textAlign = align
}

// Gets the text align of current text style.
func (c *Context) TextAlign() Align {
	return c.getState().textAlign
}

// Sets the font face based on specified id of current text style.
func (c *Context) SetFontFaceId(font int) {
	c.getState().fontId = font
}

// Gets the font face id of current text style.
func (c *Context) FontFaceId() int {
	return c.getState().fontId
}

// Sets the font face based on specified name of current text style.
func (c *Context) SetFontFace(font string) {
	c.getState().fontId = c.fs.GetFontByName(font)
}

// Gets the font face name of current text style.
func (c *Context) FontFace() string {
	return c.fs.GetFontName()
}

// Draws text string at specified location. If end is specified only the sub-string up to the end is drawn.
func (c *Context) Text(x, y float32, str string) float32 {
	state := c.getState()
	scale := state.getFontScale() * c.devicePxRatio
	invScale := 1.0 / scale
	if state.fontId == fontstashmini.FONS_INVALID {
		return 0
	}

	c.fs.SetSize(state.fontSize * scale)
	c.fs.SetSpacing(state.letterSpacing * scale)
	c.fs.SetBlur(state.fontBlur * scale)
	c.fs.SetAlign(fontstashmini.FONSAlign(state.textAlign))
	c.fs.SetFont(state.fontId)

	runes := []rune(str)

	vertexCount := maxI(2, len(runes)) * 6 // conservative estimate.
	vertexes := c.cache.allocVertexes(vertexCount)

	iter := c.fs.TextIterForRunes(x*scale, y*scale, runes)
	prevIter := iter
	index := 0

	for {
		quad := iter.Next()
		if quad == nil {
			break
		}
		if iter.PrevGlyph.Index == -1 {
			if !c.allocTextAtlas() {
				break // no memory :(
			}
			if index != 0 {
				c.renderText(vertexes[:index])
				index = 0
			}
			iter = prevIter
			quad = iter.Next() // try again
			if iter.PrevGlyph.Index == -1 {
				// still can not find glyph?
				break
			}
		}
		prevIter = iter
		// Transform corners.
		var c [8]float32
		c[0], c[1] = state.xform.Point(quad.X0*invScale, quad.Y0*invScale)
		c[2], c[3] = state.xform.Point(quad.X1*invScale, quad.Y0*invScale)
		c[4], c[5] = state.xform.Point(quad.X1*invScale, quad.Y1*invScale)
		c[6], c[7] = state.xform.Point(quad.X0*invScale, quad.Y1*invScale)
		// Create triangles
		if index+6 <= vertexCount {
			(&vertexes[index]).set(c[0], c[1], quad.S0, quad.T0)
			(&vertexes[index+1]).set(c[4], c[5], quad.S1, quad.T1)
			(&vertexes[index+2]).set(c[2], c[3], quad.S1, quad.T0)
			(&vertexes[index+3]).set(c[0], c[1], quad.S0, quad.T0)
			(&vertexes[index+4]).set(c[6], c[7], quad.S0, quad.T1)
			(&vertexes[index+5]).set(c[4], c[5], quad.S1, quad.T1)
			index += 6
		}
	}
	c.flushTextTexture()
	c.renderText(vertexes[:index])
	return iter.X
}

// Draws multi-line text string at specified location wrapped at the specified width. If end is specified only the sub-string up to the end is drawn.
// White space is stripped at the beginning of the rows, the text is split at word boundaries or when new-line characters are encountered.
// Words longer than the max width are slit at nearest character (i.e. no hyphenation).
// Draws text string at specified location. If end is specified only the sub-string up to the end is drawn.
func (c *Context) TextBox(x, y, breakRowWidth float32, str string) {
	state := c.getState()
	if state.fontId == fontstashmini.FONS_INVALID {
		return
	}
	runes := []rune(str)

	oldAlign := state.textAlign

	var hAlign Align
	if state.textAlign&ALIGN_LEFT != 0 {
		hAlign = ALIGN_LEFT
	} else if state.textAlign&ALIGN_CENTER != 0 {
		hAlign = ALIGN_CENTER
	} else if state.textAlign&ALIGN_RIGHT != 0 {
		hAlign = ALIGN_RIGHT
	}
	vAlign := state.textAlign & (ALIGN_TOP | ALIGN_MIDDLE | ALIGN_BOTTOM | ALIGN_BASELINE)
	state.textAlign = ALIGN_LEFT | vAlign

	_, _, lineH := c.TextMetrics()

	state.textAlign = oldAlign

	for {
		rows := c.textBreakLinesOfRunes(runes, breakRowWidth, 2)
		if rows == nil {
			break
		}
		for i := range rows {
			row := &rows[i]
			text := string(runes[row.StartIndex:row.EndIndex])
			switch hAlign {
			case ALIGN_LEFT:
				c.Text(x, y, text)
			case ALIGN_CENTER:
				c.Text(x+breakRowWidth*0.5-row.Width*0.5, y, text)
			case ALIGN_RIGHT:
				c.Text(x+breakRowWidth-row.Width, y, text)
			}
			y += lineH * state.lineHeight
		}
		runes = runes[rows[len(rows)-1].NextIndex:]
	}
}

// Measures the specified text string. Parameter bounds should be a pointer to float[4],
// if the bounding box of the text should be returned. The bounds value are [xmin,ymin, xmax,ymax]
// Returns the horizontal advance of the measured text (i.e. where the next character should drawn).
// Measured values are returned in local coordinate space.
func (c *Context) TextBounds(x, y float32, str string) (float32, []float32) {
	state := c.getState()
	scale := state.getFontScale() * c.devicePxRatio
	invScale := 1.0 / scale
	if state.fontId == fontstashmini.FONS_INVALID {
		return 0, nil
	}

	c.fs.SetSize(state.fontSize * scale)
	c.fs.SetSpacing(state.letterSpacing * scale)
	c.fs.SetBlur(state.fontBlur * scale)
	c.fs.SetAlign(fontstashmini.FONSAlign(state.textAlign))
	c.fs.SetFont(state.fontId)

	width, bounds := c.fs.TextBounds(x*scale, y*scale, str)
	if bounds != nil {
		bounds[1], bounds[3] = c.fs.LineBounds(y * scale)
		bounds[0] *= invScale
		bounds[1] *= invScale
		bounds[2] *= invScale
		bounds[3] *= invScale
	}
	return width, bounds
}

// Measures the specified multi-text string. Parameter bounds should be a pointer to float[4],
// if the bounding box of the text should be returned. The bounds value are [xmin,ymin, xmax,ymax]
// Measured values are returned in local coordinate space.
func (c *Context) TextBoxBounds(x, y, breakRowWidth float32, str string) [4]float32 {
	return [4]float32{}
}

// Calculates the glyph x positions of the specified text. If end is specified only the sub-string will be used.
// Measured values are returned in local coordinate space.
func (c *Context) TextGlyphPositions(x, y float32, str string) []GlyphPosition {
	state := c.getState()
	scale := state.getFontScale() * c.devicePxRatio
	invScale := 1.0 / scale
	if state.fontId == fontstashmini.FONS_INVALID {
		return nil
	}

	c.fs.SetSize(state.fontSize * scale)
	c.fs.SetSpacing(state.letterSpacing * scale)
	c.fs.SetBlur(state.fontBlur * scale)
	c.fs.SetAlign(fontstashmini.FONSAlign(state.textAlign))
	c.fs.SetFont(state.fontId)

	runes := []rune(str)
	positions := make([]GlyphPosition, 0, len(runes))

	iter := c.fs.TextIterForRunes(x*scale, y*scale, runes)
	prevIter := iter

	for {
		quad := iter.Next()
		if quad == nil {
			break
		}
		if iter.PrevGlyph.Index == -1 && !c.allocTextAtlas() {
			iter = prevIter
			quad = iter.Next() // try again
		}
		prevIter = iter
		positions = append(positions, GlyphPosition{
			Index: iter.CurrentIndex,
			Runes: runes,
			X:     iter.X * invScale,
			MinX:  minF(iter.X, quad.X0) * invScale,
			MaxX:  minF(iter.NextX, quad.X1) * invScale,
		})
	}
	return positions
}

// Returns the vertical metrics based on the current text style.
// Measured values are returned in local coordinate space.
func (c *Context) TextMetrics() (float32, float32, float32) {
	state := c.getState()
	scale := state.getFontScale() * c.devicePxRatio
	invScale := 1.0 / scale
	if state.fontId == fontstashmini.FONS_INVALID {
		return 0, 0, 0
	}

	c.fs.SetSize(state.fontSize * scale)
	c.fs.SetSpacing(state.letterSpacing * scale)
	c.fs.SetBlur(state.fontBlur * scale)
	c.fs.SetAlign(fontstashmini.FONSAlign(state.textAlign))
	c.fs.SetFont(state.fontId)

	ascender, descender, lineH := c.fs.VerticalMetrics()
	return ascender * invScale, descender * invScale, lineH * invScale
}

// Breaks the specified text into lines. If end is specified only the sub-string will be used.
// White space is stripped at the beginning of the rows, the text is split at word boundaries or when new-line characters are encountered.
// Words longer than the max width are slit at nearest character (i.e. no hyphenation).
func (c *Context) TextBreakLines(str string, breakRowWidth float32, maxRows int) []TextRow {
	return c.textBreakLinesOfRunes([]rune(str), breakRowWidth, maxRows)
}

func (c *Context) textBreakLinesOfRunes(runes []rune, breakRowWidth float32, maxRows int) []TextRow {
	state := c.getState()
	scale := state.getFontScale() * c.devicePxRatio
	invScale := 1.0 / scale
	if state.fontId == fontstashmini.FONS_INVALID {
		return nil
	}

	currentType := nvg_SPACE
	prevType := nvg_CHAR

	c.fs.SetSize(state.fontSize * scale)
	c.fs.SetSpacing(state.letterSpacing * scale)
	c.fs.SetBlur(state.fontBlur * scale)
	c.fs.SetAlign(fontstashmini.FONSAlign(state.textAlign))
	c.fs.SetFont(state.fontId)

	breakRowWidth *= scale

	iter := c.fs.TextIterForRunes(0, 0, runes)
	prevIter := iter
	var prevCodePoint rune = 0
	rows := make([]TextRow, 0, maxRows)

	var rowStartX, rowWidth, rowMinX, rowMaxX, wordStartX, wordMinX, breakWidth, breakMaxX float32
	rowStart := -1
	rowEnd := -1
	wordStart := -1
	breakEnd := -1

	for {
		quad := iter.Next()
		if quad == nil {
			break
		}
		if iter.PrevGlyph.Index == -1 && !c.allocTextAtlas() {
			iter = prevIter
			quad = iter.Next() // try again
		}
		prevIter = iter
		switch iter.CodePoint {
		case 9: // \t
			currentType = nvg_SPACE
		case 11: // \v
			currentType = nvg_SPACE
		case 12: // \f
			currentType = nvg_SPACE
		case 0x00a0: // NBSP
			currentType = nvg_SPACE
		case 10: // \n
			if prevCodePoint == 13 {
				currentType = nvg_NEWLINE
			} else {
				currentType = nvg_SPACE
			}
		case 13: // \r
			if prevCodePoint == 13 {
				currentType = nvg_NEWLINE
			} else {
				currentType = nvg_SPACE
			}
		case 0x0085: // NEL
			currentType = nvg_NEWLINE
		default:
			currentType = nvg_CHAR
		}
		if currentType == nvg_NEWLINE {
			// Always handle new lines.
			tmpRowStart := rowStart
			if rowStart == -1 {
				tmpRowStart = iter.CurrentIndex
			}
			if rowEnd == -1 {
				rowEnd = iter.CurrentIndex
			}
			rows = append(rows, TextRow{
				Runes:      runes,
				StartIndex: tmpRowStart,
				EndIndex:   rowEnd,
				Width:      rowWidth * invScale,
				MinX:       rowMinX * invScale,
				MaxX:       rowMaxX * invScale,
				NextIndex:  iter.NextIndex,
			})
			if len(rows) >= maxRows {
				return rows
			}
			// Set null break point
			breakEnd = rowStart
			breakWidth = 0.0
			breakMaxX = 0.0
			rowStart = -1
			rowEnd = -1
			rowMinX = 0
			rowMaxX = 0
			// Indicate to skip the white space at the beginning of the row.

		} else {
			if rowStart == -1 {
				if currentType == nvg_CHAR {
					// The current char is the row so far
					rowStartX = iter.X
					rowStart = iter.CurrentIndex
					rowEnd = iter.NextIndex
					rowWidth = iter.NextX - rowStartX // q.x1 - rowStartX;
					rowMinX = quad.X0 - rowStartX
					rowMaxX = quad.X1 - rowStartX
					wordStart = iter.CurrentIndex
					wordStartX = iter.X
					wordMinX = quad.X0 - rowStartX
					// Set null break point
					breakEnd = rowStart
					breakWidth = 0.0
					breakMaxX = 0.0
				}
			} else {
				nextWidth := iter.NextX - rowStartX
				// track last non-white space character
				if currentType == nvg_CHAR {
					rowEnd = iter.NextIndex
					rowWidth = iter.NextX - rowStartX
					rowMaxX = quad.X1 - rowStartX
				}
				// track last end of a word
				if prevType == nvg_CHAR && currentType == nvg_SPACE {
					breakEnd = iter.CurrentIndex
					breakWidth = rowWidth
					breakMaxX = rowMaxX
				}
				// track last beginning of a word
				if prevType == nvg_SPACE && currentType == nvg_CHAR {
					wordStart = iter.CurrentIndex
					wordStartX = iter.X
					wordMinX = quad.X0 - rowStartX
				}
				// Break to new line when a character is beyond break width.
				if currentType == nvg_CHAR && nextWidth > breakRowWidth {
					// The run length is too long, need to break to new line.
					if breakEnd == rowStart {
						// The current word is longer than the row length, just break it from here.
						rows = append(rows, TextRow{
							Runes:      runes,
							StartIndex: rowStart,
							EndIndex:   iter.CurrentIndex,
							Width:      rowWidth * invScale,
							MinX:       rowMinX * invScale,
							MaxX:       rowMaxX * invScale,
							NextIndex:  iter.CurrentIndex,
						})
						if len(rows) >= maxRows {
							return rows
						}
						rowStartX = iter.X
						rowStart = iter.CurrentIndex
						rowEnd = iter.NextIndex
						rowWidth = iter.NextX - rowStartX
						rowMinX = quad.X0 - rowStartX
						rowMaxX = quad.X1 - rowStartX
						wordStart = iter.CurrentIndex
						wordStartX = iter.X
						wordMinX = quad.X0 - rowStartX
					} else {
						// Break the line from the end of the last word, and start new line from the beginning of the new.
						rows = append(rows, TextRow{
							Runes:      runes,
							StartIndex: rowStart,
							EndIndex:   breakEnd,
							Width:      breakWidth * invScale,
							MinX:       rowMinX * invScale,
							MaxX:       breakMaxX * invScale,
							NextIndex:  wordStart,
						})
						if len(rows) >= maxRows {
							return rows
						}
						rowStartX = wordStartX
						rowStart = wordStart
						rowEnd = iter.NextIndex
						rowWidth = iter.NextX - rowStartX
						rowMinX = wordMinX
						rowMaxX = quad.X1 - rowStartX
						// No change to the word start
					}
					// Set null break point
					breakEnd = rowStart
					breakWidth = 0.0
					breakMaxX = 0.0
				}
			}
		}

		prevCodePoint = iter.CodePoint
		prevType = currentType
	}
	if rowStart != -1 {
		rows = append(rows, TextRow{
			Runes:      runes,
			StartIndex: rowStart,
			EndIndex:   rowEnd,
			Width:      rowWidth * invScale,
			MinX:       rowMinX * invScale,
			MaxX:       rowMaxX * invScale,
			NextIndex:  len(runes),
		})
	}
	return nil
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
			cache.pathWinding(Direction(c.commands[i+1]))
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
		if ptEquals(p0.x, p0.y, p1.x, p1.y, c.distTol) && path.count > 2 {
			path.count--
			p0 = &points[path.count-1]
			path.closed = true
		}

		// Enforce winding.
		if path.count > 2 {
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

func (c *Context) flushTextTexture() {
	dirty := c.fs.ValidateTexture()
	if dirty != nil {
		fontImage := c.fontImages[c.fontImageIdx]
		// Update texture
		if fontImage != 0 {
			data, _, _ := c.fs.GetTextureData()
			x := dirty[0]
			y := dirty[1]
			w := dirty[2] - x
			h := dirty[3] - y
			c.params.renderUpdateTexture(fontImage, x, y, w, h, data)
		}
	}
}

func (c *Context) allocTextAtlas() bool {
	c.flushTextTexture()
	if c.fontImageIdx >= nvg_MAX_FONTIMAGES-1 {
		return false
	}
	var iw, ih int
	// if next fontImage already have a texture
	if c.fontImages[c.fontImageIdx+1] != 0 {
		iw, ih, _ = c.ImageSize(c.fontImageIdx + 1)
	} else { // calculate the new font image size and create it.
		iw, ih, _ = c.ImageSize(c.fontImageIdx)
		if iw > ih {
			ih *= 2
		} else {
			iw *= 2
		}
		if iw > nvg_MAX_FONTIMAGE_SIZE || ih > nvg_MAX_FONTIMAGE_SIZE {
			iw = nvg_MAX_FONTIMAGE_SIZE
			ih = nvg_MAX_FONTIMAGE_SIZE
		}
		c.fontImages[c.fontImageIdx+1] = c.params.renderCreateTexture(nvg_TEXTURE_ALPHA, iw, ih, 0, nil)
	}
	c.fontImageIdx++
	c.fs.ResetAtlas(iw, ih)
	return true
}

func (c *Context) renderText(vertexes []nvgVertex) {
	state := c.getState()
	paint := state.fill

	// Render triangles
	paint.image = c.fontImages[c.fontImageIdx]

	// Apply global alpha
	paint.innerColor.A *= state.alpha
	paint.outerColor.A *= state.alpha

	c.params.renderTriangles(&paint, &state.scissor, vertexes)

	c.drawCallCount++
	c.textTriCount += len(vertexes) / 3
}
