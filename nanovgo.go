package nanovgo

import (
	"bytes"
	"github.com/shibukawa/nanovgo/fontstashmini"
	"image"
	_ "image/jpeg" // to read jpeg
	_ "image/png"  // to read png
	"log"
	"os"
)

// Context is an entry point object to use NanoVGo API and created by NewContext() function.
//
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
// using LinearGradient(), BoxGradient(), RadialGradient() and ImagePattern().
//
// Current render style can be saved and restored using Save() and Restore().
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
// Apart from ResetTransform(), each transformation function first creates
// specific transformation matrix and pre-multiplies the current transformation by it.
//
// Current coordinate system (transformation) can be saved and restored using Save() and Restore().
//
// Images
//
// NanoVG allows you to load jpg, png, psd, tga, pic and gif files to be used for rendering.
// In addition you can upload your own image. The image loading is provided by stb_image.
// The parameter imageFlags is combination of flags defined in ImageFlags.
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
// Drawing a new shape starts with BeginPath(), it clears all the currently defined paths.
// Then you define one or more paths and sub-paths which describe the shape. The are functions
// to draw common shapes like rectangles and circles, and lower level step-by-step functions,
// which allow to define a path curve by curve.
//
// NanoVG uses even-odd fill rule to draw the shapes. Solid shapes should have counter clockwise
// winding and holes should have counter clockwise order. To specify winding of a path you can
// call PathWinding(). This is useful especially for the common shapes, which are drawn CCW.
//
// Finally you can fill the path using current fill style by calling Fill(), and stroke it
// with current stroke style by calling Stroke().
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

// Delete is called when tearing down NanoVGo context
func (c *Context) Delete() {

	for i, fontImage := range c.fontImages {
		if fontImage != 0 {
			c.DeleteImage(fontImage)
			c.fontImages[i] = 0
		}
	}
	c.params.renderDelete()
}

// BeginFrame begins drawing a new frame
// Calls to NanoVGo drawing API should be wrapped in Context.BeginFrame() & Context.EndFrame()
// Context.BeginFrame() defines the size of the window to render to in relation currently
// set viewport (i.e. glViewport on GL backends). Device pixel ration allows to
// control the rendering on Hi-DPI devices.
// For example, GLFW returns two dimension for an opened window: window size and
// frame buffer size. In that case you would set windowWidth/Height to the window size
// devicePixelRatio to: frameBufferWidth / windowWidth.
func (c *Context) BeginFrame(windowWidth, windowHeight int, devicePixelRatio float32) {
	c.states = c.states[:0]
	c.Save()
	c.Reset()

	c.setDevicePixelRatio(devicePixelRatio)
	c.params.renderViewport(windowWidth, windowHeight)

	c.drawCallCount = 0
	c.fillTriCount = 0
	c.strokeTriCount = 0
	c.textTriCount = 0
}

// CancelFrame cancels drawing the current frame.
func (c *Context) CancelFrame() {
	c.params.renderCancel()
}

// EndFrame ends drawing flushing remaining render state.
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
		for i := j; i < nvgMaxFontImages; i++ {
			c.fontImages[i] = 0
		}
	}
}

// Save pushes and saves the current render state into a state stack.
// A matching Restore() must be used to restore the state.
func (c *Context) Save() {
	if len(c.states) >= nvgMaxStates {
		return
	}
	if len(c.states) > 0 {
		c.states = append(c.states, c.states[len(c.states)-1])
	} else {
		c.states = append(c.states, nvgState{})
	}
}

// Restore pops and restores current render state.
func (c *Context) Restore() {
	nStates := len(c.states)
	if nStates > 1 {
		c.states = c.states[:nStates-1]
	}
}

// Block makes Save/Restore block.
func (c *Context) Block(block func()) {
	c.Save()
	defer c.Restore()
	block()
}

// Reset resets current render state to default values. Does not affect the render state stack.
func (c *Context) Reset() {
	c.getState().reset()
}

// SetStrokeWidth sets the stroke width of the stroke style.
func (c *Context) SetStrokeWidth(width float32) {
	c.getState().strokeWidth = width
}

// StrokeWidth gets the stroke width of the stroke style.
func (c *Context) StrokeWidth() float32 {
	return c.getState().strokeWidth
}

// SetMiterLimit sets the miter limit of the stroke style.
// Miter limit controls when a sharp corner is beveled.
func (c *Context) SetMiterLimit(limit float32) {
	c.getState().miterLimit = limit
}

// MiterLimit gets the miter limit of the stroke style.
func (c *Context) MiterLimit() float32 {
	return c.getState().miterLimit
}

// SetLineCap sets how the end of the line (cap) is drawn,
// Can be one of: Butt (default), Round, Squre.
func (c *Context) SetLineCap(cap LineCap) {
	c.getState().lineCap = cap
}

// LineCap gets how the end of the line (cap) is drawn,
func (c *Context) LineCap() LineCap {
	return c.getState().lineCap
}

// SetLineJoin sets how sharp path corners are drawn.
// Can be one of Miter (default), Round, Bevel.
func (c *Context) SetLineJoin(joint LineCap) {
	c.getState().lineJoin = joint
}

// LineJoin gets how sharp path corners are drawn.
func (c *Context) LineJoin() LineCap {
	return c.getState().lineJoin
}

// SetGlobalAlpha sets the transparency applied to all rendered shapes.
// Already transparent paths will get proportionally more transparent as well.
func (c *Context) SetGlobalAlpha(alpha float32) {
	c.getState().alpha = alpha
}

// GlobalAlpha gets the transparency applied to all rendered shapes.
func (c *Context) GlobalAlpha() float32 {
	return c.getState().alpha
}

// SetTransform premultiplies current coordinate system by specified matrix.
func (c *Context) SetTransform(t TransformMatrix) {
	state := c.getState()
	state.xform = state.xform.PreMultiply(t)
}

// SetTransformByValue premultiplies current coordinate system by specified matrix.
// The parameters are interpreted as matrix as follows:
//   [a c e]
//   [b d f]
//   [0 0 1]
func (cx *Context) SetTransformByValue(a, b, c, d, e, f float32) {
	t := TransformMatrix{a, b, c, d, e, f}
	state := cx.getState()
	state.xform = state.xform.PreMultiply(t)
}

// ResetTransform resets current transform to a identity matrix.
func (c *Context) ResetTransform() {
	state := c.getState()
	state.xform = IdentityMatrix()
}

// Translate translates current coordinate system.
func (c *Context) Translate(x, y float32) {
	state := c.getState()
	state.xform = state.xform.PreMultiply(TranslateMatrix(x, y))
}

// Rotate rotates current coordinate system. Angle is specified in radians.
func (c *Context) Rotate(angle float32) {
	state := c.getState()
	state.xform = state.xform.PreMultiply(RotateMatrix(angle))
}

// SkewX skews the current coordinate system along X axis. Angle is specified in radians.
func (c *Context) SkewX(angle float32) {
	state := c.getState()
	state.xform = state.xform.PreMultiply(SkewXMatrix(angle))
}

// SkewY skews the current coordinate system along Y axis. Angle is specified in radians.
func (c *Context) SkewY(angle float32) {
	state := c.getState()
	state.xform = state.xform.PreMultiply(SkewYMatrix(angle))
}

// Scale scales the current coordinate system.
func (c *Context) Scale(x, y float32) {
	state := c.getState()
	state.xform = state.xform.PreMultiply(ScaleMatrix(x, y))
}

// CurrentTransform returns the top part (a-f) of the current transformation matrix.
//   [a c e]
//   [b d f]
//   [0 0 1]
// There should be space for 6 floats in the return buffer for the values a-f.
func (c *Context) CurrentTransform() TransformMatrix {
	return c.getState().xform
}

// SetStrokeColor sets current stroke style to a solid color.
func (c *Context) SetStrokeColor(color Color) {
	c.getState().stroke.setPaintColor(color)
}

// SetStrokePaint sets current stroke style to a paint, which can be a one of the gradients or a pattern.
func (c *Context) SetStrokePaint(paint Paint) {
	state := c.getState()
	state.stroke = paint
	state.stroke.xform = state.stroke.xform.Multiply(state.xform)
}

// SetFillColor sets current fill style to a solid color.
func (c *Context) SetFillColor(color Color) {
	c.getState().fill.setPaintColor(color)
}

// SetFillPaint sets current fill style to a paint, which can be a one of the gradients or a pattern.
func (c *Context) SetFillPaint(paint Paint) {
	state := c.getState()
	state.fill = paint
	state.fill.xform = state.fill.xform.Multiply(state.xform)
}

// CreateImage creates image by loading it from the disk from specified file name.
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

// CreateImageFromMemory creates image by loading it from the specified chunk of memory.
// Returns handle to the image.
func (c *Context) CreateImageFromMemory(flags ImageFlags, data []byte) int {
	reader := bytes.NewReader(data)
	img, _, err := image.Decode(reader)
	if err != nil {
		return 0
	}
	return c.CreateImageFromGoImage(flags, img)
}

// CreateImageFromGoImage creates image by loading it from the specified image.Image object.
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

// CreateImageRGBA creates image from specified image data.
// Returns handle to the image.
func (c *Context) CreateImageRGBA(w, h int, imageFlags ImageFlags, data []byte) int {
	return c.params.renderCreateTexture(nvgTextureRGBA, w, h, imageFlags, data)
}

// UpdateImage updates image data specified by image handle.
func (c *Context) UpdateImage(img int, data []byte) error {
	w, h, err := c.params.renderGetTextureSize(img)
	if err != nil {
		return err
	}
	return c.params.renderUpdateTexture(img, 0, 0, w, h, data)
}

// ImageSize returns the dimensions of a created image.
func (c *Context) ImageSize(img int) (int, int, error) {
	return c.params.renderGetTextureSize(img)
}

// DeleteImage deletes created image.
func (c *Context) DeleteImage(img int) {
	c.params.renderDeleteTexture(img)
}

// Scissor sets the current scissor rectangle.
// The scissor rectangle is transformed by the current transform.
func (c *Context) Scissor(x, y, w, h float32) {
	state := c.getState()

	w = maxF(0.0, w)
	h = maxF(0.0, h)

	state.scissor.xform = TranslateMatrix(x+w*0.5, y+h*0.5).Multiply(state.xform)
	state.scissor.extent = [2]float32{w * 0.5, h * 0.5}
}

// IntersectScissor calculates intersects current scissor rectangle with the specified rectangle.
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

	pXform := state.scissor.xform.Multiply(state.xform.Inverse())
	ex := state.scissor.extent[0]
	ey := state.scissor.extent[1]

	teX := ex * absF(pXform[0]) * ey * absF(pXform[2])
	teY := ex * absF(pXform[1]) * ey * absF(pXform[3])
	rect := intersectRects(pXform[4]-teX, pXform[5]-teY, teX*2, teY*2, x, y, w, h)
	c.Scissor(rect[0], rect[1], rect[2], rect[3])
}

// ResetScissor resets and disables scissoring.
func (c *Context) ResetScissor() {
	state := c.getState()

	state.scissor.xform = TransformMatrix{0, 0, 0, 0, 0, 0}
	state.scissor.extent = [2]float32{-1.0, -1.0}
}

// BeginPath clears the current path and sub-paths.
func (c *Context) BeginPath() {
	c.commands = c.commands[:0]
	c.cache.clearPathCache()
}

// MoveTo starts new sub-path with specified point as first point.
func (c *Context) MoveTo(x, y float32) {
	c.appendCommand([]float32{float32(nvgMOVETO), x, y})
}

// LineTo adds line segment from the last point in the path to the specified point.
func (c *Context) LineTo(x, y float32) {
	c.appendCommand([]float32{float32(nvgLINETO), x, y})
}

// BezierTo adds cubic bezier segment from last point in the path via two control points to the specified point.
func (c *Context) BezierTo(c1x, c1y, c2x, c2y, x, y float32) {
	c.appendCommand([]float32{float32(nvgBEZIERTO), c1x, c1y, c2x, c2y, x, y})
}

// QuadTo adds quadratic bezier segment from last point in the path via a control point to the specified point.
func (c *Context) QuadTo(cx, cy, x, y float32) {
	x0 := c.commandX
	y0 := c.commandY
	c.appendCommand([]float32{float32(nvgBEZIERTO),
		x0 + 2.0/3.0*(cx-x0), y0 + 2.0/3.0*(cy-y0),
		x + 2.0/3.0*(cx-x), y + 2.0/3.0*(cy-y),
		x, y,
	})
}

// Arc creates new circle arc shaped sub-path. The arc center is at cx,cy, the arc radius is r,
// and the arc is drawn from angle a0 to a1, and swept in direction dir (CounterClockwise, or Clockwise).
// Angles are specified in radians.
func (c *Context) Arc(cx, cy, r, a0, a1 float32, dir Direction) {
	var move nvgCommands
	if len(c.commands) > 0 {
		move = nvgLINETO
	} else {
		move = nvgMOVETO
	}

	// Clamp angles
	da := a1 - a0
	if dir == Clockwise {
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

	if dir == CounterClockwise {
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
			values = append(values, float32(nvgBEZIERTO), px+pTanX, py+pTanY, x-tanX, y-tanY, x, y)
		}
		px = x
		py = y
		pTanX = tanX
		pTanY = tanY
	}
	c.appendCommand(values)
}

// ArcTo adds an arc segment at the corner defined by the last path point, and two specified points.
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
		dir = Clockwise
	} else {
		cx = x1 + dx0*d + -dy0*radius
		cy = y1 + dy0*d + dx0*radius
		a0 = atan2F(-dx0, dy0)
		a1 = atan2F(dx1, -dy1)
		dir = CounterClockwise
	}
	c.Arc(cx, cy, radius, a0, a1, dir)
}

// Rect creates new rectangle shaped sub-path.
func (c *Context) Rect(x, y, w, h float32) {
	c.appendCommand([]float32{
		float32(nvgMOVETO), x, y,
		float32(nvgLINETO), x, y + h,
		float32(nvgLINETO), x + w, y + h,
		float32(nvgLINETO), x + w, y,
		float32(nvgCLOSE),
	})
}

// RoundedRect creates new rounded rectangle shaped sub-path.
func (c *Context) RoundedRect(x, y, w, h, r float32) {
	if r < 0.1 {
		c.Rect(x, y, w, h)
	} else {
		rx := minF(r, absF(w)*0.5) * signF(w)
		ry := minF(r, absF(h)*0.5) * signF(h)
		c.appendCommand([]float32{
			float32(nvgMOVETO), x, y + ry,
			float32(nvgLINETO), x, y + h - ry,
			float32(nvgBEZIERTO), x, y + h - ry*(1-Kappa90), x + rx*(1-Kappa90), y + h, x + rx, y + h,
			float32(nvgLINETO), x + w - rx, y + h,
			float32(nvgBEZIERTO), x + w - rx*(1-Kappa90), y + h, x + w, y + h - ry*(1-Kappa90), x + w, y + h - ry,
			float32(nvgLINETO), x + w, y + ry,
			float32(nvgBEZIERTO), x + w, y + ry*(1-Kappa90), x + w - rx*(1-Kappa90), y, x + w - rx, y,
			float32(nvgLINETO), x + rx, y,
			float32(nvgBEZIERTO), x + rx*(1-Kappa90), y, x, y + ry*(1-Kappa90), x, y + ry,
			float32(nvgCLOSE),
		})
	}
}

// Ellipse creates new ellipse shaped sub-path.
func (c *Context) Ellipse(cx, cy, rx, ry float32) {
	c.appendCommand([]float32{
		float32(nvgMOVETO), cx - rx, cy,
		float32(nvgBEZIERTO), cx - rx, cy + ry*Kappa90, cx - rx*Kappa90, cy + ry, cx, cy + ry,
		float32(nvgBEZIERTO), cx + rx*Kappa90, cy + ry, cx + rx, cy + ry*Kappa90, cx + rx, cy,
		float32(nvgBEZIERTO), cx + rx, cy - ry*Kappa90, cx + rx*Kappa90, cy - ry, cx, cy - ry,
		float32(nvgBEZIERTO), cx - rx*Kappa90, cy - ry, cx - rx, cy - ry*Kappa90, cx - rx, cy,
		float32(nvgCLOSE),
	})
}

// Circle creates new circle shaped sub-path.
func (c *Context) Circle(cx, cy, r float32) {
	c.Ellipse(cx, cy, r, r)
}

// ClosePath closes current sub-path with a line segment.
func (c *Context) ClosePath() {
	c.appendCommand([]float32{float32(nvgCLOSE)})
}

// PathWinding sets the current sub-path winding, see Winding.
func (c *Context) PathWinding(winding Winding) {
	c.appendCommand([]float32{float32(nvgWINDING), float32(winding)})
}

// DebugDumpPathCache prints cached path information to console
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

// Fill fills the current path with current fill style.
func (c *Context) Fill() {
	state := c.getState()
	fillPaint := state.fill
	c.flattenPaths()

	if c.params.edgeAntiAlias() {
		c.cache.expandFill(c.fringeWidth, Miter, 2.4, c.fringeWidth)
	} else {
		c.cache.expandFill(0.0, Miter, 2.4, c.fringeWidth)
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

// Stroke draws the current path with current stroke style.
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
	for _, path := range c.cache.paths {
		if path.count == 1 {
			panic("")
		}
	}
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

// CreateFont creates font by loading it from the disk from specified file name.
// Returns handle to the font.
func (c *Context) CreateFont(name, filePath string) int {
	return c.fs.AddFont(name, filePath)
}

// CreateFontFromMemory creates image by loading it from the specified memory chunk.
// Returns handle to the font.
func (c *Context) CreateFontFromMemory(name string, data []byte, freeData uint8) int {
	return c.fs.AddFontFromMemory(name, data, freeData)
}

// FindFont finds a loaded font of specified name, and returns handle to it, or -1 if the font is not found.
func (c *Context) FindFont(name string) int {
	return c.fs.GetFontByName(name)
}

// SetFontSize sets the font size of current text style.
func (c *Context) SetFontSize(size float32) {
	if size < 0 {
		panic("Context.SetFontSize: negative font size is invalid")
	}
	c.getState().fontSize = size
}

// FontSize gets the font size of current text style.
func (c *Context) FontSize() float32 {
	return c.getState().fontSize
}

// SetFontBlur sets the font blur of current text style.
func (c *Context) SetFontBlur(blur float32) {
	c.getState().fontBlur = blur
}

// FontBlur gets the font blur of current text style.
func (c *Context) FontBlur() float32 {
	return c.getState().fontBlur
}

// SetTextLetterSpacing sets the letter spacing of current text style.
func (c *Context) SetTextLetterSpacing(spacing float32) {
	c.getState().letterSpacing = spacing
}

// TextLetterSpacing gets the letter spacing of current text style.
func (c *Context) TextLetterSpacing() float32 {
	return c.getState().letterSpacing
}

// SetTextLineHeight sets the line height of current text style.
func (c *Context) SetTextLineHeight(lineHeight float32) {
	c.getState().lineHeight = lineHeight
}

// TextLineHeight gets the line height of current text style.
func (c *Context) TextLineHeight() float32 {
	return c.getState().lineHeight
}

// SetTextAlign sets the text align of current text style.
func (c *Context) SetTextAlign(align Align) {
	c.getState().textAlign = align
}

// TextAlign gets the text align of current text style.
func (c *Context) TextAlign() Align {
	return c.getState().textAlign
}

// SetFontFaceID sets the font face based on specified id of current text style.
func (c *Context) SetFontFaceID(font int) {
	c.getState().fontID = font
}

// FontFaceID gets the font face id of current text style.
func (c *Context) FontFaceID() int {
	return c.getState().fontID
}

// SetFontFace sets the font face based on specified name of current text style.
func (c *Context) SetFontFace(font string) {
	c.getState().fontID = c.fs.GetFontByName(font)
}

// FontFace gets the font face name of current text style.
func (c *Context) FontFace() string {
	return c.fs.GetFontName()
}

// Text draws text string at specified location. If end is specified only the sub-string up to the end is drawn.
func (c *Context) Text(x, y float32, str string) float32 {
	return c.TextRune(x, y, []rune(str))
}

// TextRune is an alternate version of Text that accepts rune slice.
func (c *Context) TextRune(x, y float32, runes []rune) float32 {
	state := c.getState()
	scale := state.getFontScale() * c.devicePxRatio
	invScale := 1.0 / scale
	if state.fontID == fontstashmini.INVALID {
		return 0
	}

	c.fs.SetSize(state.fontSize * scale)
	c.fs.SetSpacing(state.letterSpacing * scale)
	c.fs.SetBlur(state.fontBlur * scale)
	c.fs.SetAlign(fontstashmini.FONSAlign(state.textAlign))
	c.fs.SetFont(state.fontID)

	vertexCount := maxI(2, len(runes)) * 4 // conservative estimate.
	vertexes := c.cache.allocVertexes(vertexCount)

	iter := c.fs.TextIterForRunes(x*scale, y*scale, runes)
	prevIter := iter
	index := 0

	for {
		quad, ok := iter.Next()
		if !ok {
			break
		}
		if iter.PrevGlyph == nil || iter.PrevGlyph.Index == -1 {
			if !c.allocTextAtlas() {
				break // no memory :(
			}
			if index != 0 {
				c.renderText(vertexes[:index])
				index = 0
			}
			iter = prevIter
			quad, _ = iter.Next() // try again
			if iter.PrevGlyph == nil || iter.PrevGlyph.Index == -1 {
				// still can not find glyph?
				break
			}
		}
		prevIter = iter
		// Transform corners.
		c0, c1 := state.xform.TransformPoint(quad.X0*invScale, quad.Y0*invScale)
		c2, c3 := state.xform.TransformPoint(quad.X1*invScale, quad.Y0*invScale)
		c4, c5 := state.xform.TransformPoint(quad.X1*invScale, quad.Y1*invScale)
		c6, c7 := state.xform.TransformPoint(quad.X0*invScale, quad.Y1*invScale)
		//log.Printf("quad(%c) x0=%d, x1=%d, y0=%d, y1=%d, s0=%d, s1=%d, t0=%d, t1=%d\n", iter.CodePoint, int(quad.X0), int(quad.X1), int(quad.Y0), int(quad.Y1), int(1024*quad.S0), int(quad.S1*1024), int(quad.T0*1024), int(quad.T1*1024))
		// Create triangles
		if index+4 <= vertexCount {
			(&vertexes[index]).set(c2, c3, quad.S1, quad.T0)
			(&vertexes[index+1]).set(c0, c1, quad.S0, quad.T0)
			(&vertexes[index+2]).set(c4, c5, quad.S1, quad.T1)
			(&vertexes[index+3]).set(c6, c7, quad.S0, quad.T1)
			index += 4
		}
	}
	c.flushTextTexture()
	c.renderText(vertexes[:index])
	return iter.X
}

// TextBox draws multi-line text string at specified location wrapped at the specified width. If end is specified only the sub-string up to the end is drawn.
// White space is stripped at the beginning of the rows, the text is split at word boundaries or when new-line characters are encountered.
// Words longer than the max width are slit at nearest character (i.e. no hyphenation).
// Draws text string at specified location. If end is specified only the sub-string up to the end is drawn.
func (c *Context) TextBox(x, y, breakRowWidth float32, str string) {
	state := c.getState()
	if state.fontID == fontstashmini.INVALID {
		return
	}
	runes := []rune(str)

	oldAlign := state.textAlign

	var hAlign Align
	if state.textAlign&AlignLeft != 0 {
		hAlign = AlignLeft
	} else if state.textAlign&AlignCenter != 0 {
		hAlign = AlignCenter
	} else if state.textAlign&AlignRight != 0 {
		hAlign = AlignRight
	}
	vAlign := state.textAlign & (AlignTop | AlignMiddle | AlignBottom | AlignBaseline)
	state.textAlign = AlignLeft | vAlign

	_, _, lineH := c.TextMetrics()

	state.textAlign = oldAlign

	for _, row := range c.TextBreakLinesRune(runes, breakRowWidth) {
		text := string(runes[row.StartIndex:row.EndIndex])
		switch hAlign {
		case AlignLeft:
			c.Text(x, y, text)
		case AlignCenter:
			c.Text(x+breakRowWidth*0.5-row.Width*0.5, y, text)
		case AlignRight:
			c.Text(x+breakRowWidth-row.Width, y, text)
		}
		y += lineH * state.lineHeight
	}
}

// TextBounds measures the specified text string. Parameter bounds should be a pointer to float[4],
// if the bounding box of the text should be returned. The bounds value are [xmin,ymin, xmax,ymax]
// Returns the horizontal advance of the measured text (i.e. where the next character should drawn).
// Measured values are returned in local coordinate space.
func (c *Context) TextBounds(x, y float32, str string) (float32, []float32) {
	state := c.getState()
	scale := state.getFontScale() * c.devicePxRatio
	invScale := 1.0 / scale
	if state.fontID == fontstashmini.INVALID {
		return 0, nil
	}

	c.fs.SetSize(state.fontSize * scale)
	c.fs.SetSpacing(state.letterSpacing * scale)
	c.fs.SetBlur(state.fontBlur * scale)
	c.fs.SetAlign(fontstashmini.FONSAlign(state.textAlign))
	c.fs.SetFont(state.fontID)

	width, bounds := c.fs.TextBounds(x*scale, y*scale, str)
	if bounds != nil {
		bounds[1], bounds[3] = c.fs.LineBounds(y * scale)
		bounds[0] *= invScale
		bounds[1] *= invScale
		bounds[2] *= invScale
		bounds[3] *= invScale
	}
	return width * invScale, bounds
}

// TextBoxBounds measures the specified multi-text string. Parameter bounds should be a pointer to float[4],
// if the bounding box of the text should be returned. The bounds value are [xmin,ymin, xmax,ymax]
// Measured values are returned in local coordinate space.
func (c *Context) TextBoxBounds(x, y, breakRowWidth float32, str string) [4]float32 {
	state := c.getState()
	if state.fontID == fontstashmini.INVALID {
		return [4]float32{}
	}
	runes := []rune(str)
	scale := state.getFontScale() * c.devicePxRatio
	invScale := 1.0 / scale

	oldAlign := state.textAlign

	var hAlign Align
	if state.textAlign&AlignLeft != 0 {
		hAlign = AlignLeft
	} else if state.textAlign&AlignCenter != 0 {
		hAlign = AlignCenter
	} else if state.textAlign&AlignRight != 0 {
		hAlign = AlignRight
	}
	vAlign := state.textAlign & (AlignTop | AlignMiddle | AlignBottom | AlignBaseline)
	state.textAlign = AlignLeft | vAlign

	minX := x
	minY := y
	maxX := x
	maxY := y

	_, _, lineH := c.TextMetrics()
	/*c.fs.SetSize(state.fontSize * scale)
	c.fs.SetSpacing(state.letterSpacing * scale)
	c.fs.SetBlur(state.fontBlur * scale)
	c.fs.SetAlign(fontstashmini.FONSAlign(state.textAlign))
	c.fs.SetFont(state.fontId)*/

	rMinY, rMaxY := c.fs.LineBounds(0)
	rMinY *= invScale
	rMaxY *= invScale

	for _, row := range c.TextBreakLinesRune(runes, breakRowWidth) {
		var dx float32
		// Horizontal bounds
		switch hAlign {
		case AlignLeft:
			dx = 0
		case AlignCenter:
			dx = breakRowWidth*0.5 - row.Width*0.5
		case AlignRight:
			dx = breakRowWidth - row.Width
		}
		rMinX := x + row.MinX + dx
		rMaxX := x + row.MaxX + dx
		minX = minF(minX, rMinX)
		maxX = maxF(maxX, rMaxX)
		// Vertical bounds.
		minY = minF(minY, y+rMinY)
		maxY = maxF(maxY, y+rMaxY)
		y += lineH * state.lineHeight
	}

	state.textAlign = oldAlign

	return [4]float32{minX, minY, maxX, maxY}
}

// TextGlyphPositions calculates the glyph x positions of the specified text. If end is specified only the sub-string will be used.
// Measured values are returned in local coordinate space.
func (c *Context) TextGlyphPositions(x, y float32, str string) []GlyphPosition {
	return c.TextGlyphPositionsRune(x, y, []rune(str))
}

// TextGlyphPositionsRune is an alternate version of TextGlyphPositions that accepts rune slice
func (c *Context) TextGlyphPositionsRune(x, y float32, runes []rune) []GlyphPosition {
	state := c.getState()
	scale := state.getFontScale() * c.devicePxRatio
	invScale := 1.0 / scale
	if state.fontID == fontstashmini.INVALID {
		return nil
	}

	c.fs.SetSize(state.fontSize * scale)
	c.fs.SetSpacing(state.letterSpacing * scale)
	c.fs.SetBlur(state.fontBlur * scale)
	c.fs.SetAlign(fontstashmini.FONSAlign(state.textAlign))
	c.fs.SetFont(state.fontID)

	positions := make([]GlyphPosition, 0, len(runes))

	iter := c.fs.TextIterForRunes(x*scale, y*scale, runes)
	prevIter := iter

	for {
		quad, ok := iter.Next()
		if !ok {
			break
		}
		if iter.PrevGlyph.Index == -1 && !c.allocTextAtlas() {
			iter = prevIter
			quad, _ = iter.Next() // try again
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

// TextMetrics returns the vertical metrics based on the current text style.
// Measured values are returned in local coordinate space.
func (c *Context) TextMetrics() (float32, float32, float32) {
	state := c.getState()
	scale := state.getFontScale() * c.devicePxRatio
	invScale := 1.0 / scale
	if state.fontID == fontstashmini.INVALID {
		return 0, 0, 0
	}

	c.fs.SetSize(state.fontSize * scale)
	c.fs.SetSpacing(state.letterSpacing * scale)
	c.fs.SetBlur(state.fontBlur * scale)
	c.fs.SetAlign(fontstashmini.FONSAlign(state.textAlign))
	c.fs.SetFont(state.fontID)

	ascender, descender, lineH := c.fs.VerticalMetrics()
	return ascender * invScale, descender * invScale, lineH * invScale
}

// TextBreakLines breaks the specified text into lines. If end is specified only the sub-string will be used.
// White space is stripped at the beginning of the rows, the text is split at word boundaries or when new-line characters are encountered.
// Words longer than the max width are slit at nearest character (i.e. no hyphenation).
func (c *Context) TextBreakLines(str string, breakRowWidth float32) []TextRow {
	return c.TextBreakLinesRune([]rune(str), breakRowWidth)
}

// TextBreakLinesRune is an alternate version of TextBreakLines that accepts rune slice
func (c *Context) TextBreakLinesRune(runes []rune, breakRowWidth float32) []TextRow {
	state := c.getState()
	scale := state.getFontScale() * c.devicePxRatio
	invScale := 1.0 / scale
	if state.fontID == fontstashmini.INVALID {
		return nil
	}

	currentType := nvgSPACE
	prevType := nvgCHAR

	c.fs.SetSize(state.fontSize * scale)
	c.fs.SetSpacing(state.letterSpacing * scale)
	c.fs.SetBlur(state.fontBlur * scale)
	c.fs.SetAlign(fontstashmini.FONSAlign(state.textAlign))
	c.fs.SetFont(state.fontID)

	breakRowWidth *= scale

	iter := c.fs.TextIterForRunes(0, 0, runes)
	prevIter := iter
	var prevCodePoint rune
	var rows []TextRow

	var rowStartX, rowWidth, rowMinX, rowMaxX, wordStartX, wordMinX, breakWidth, breakMaxX float32
	rowStart := -1
	rowEnd := -1
	wordStart := -1
	breakEnd := -1

	for {
		quad, ok := iter.Next()
		if !ok {
			break
		}
		if iter.PrevGlyph == nil || iter.PrevGlyph.Index == -1 && !c.allocTextAtlas() {
			iter = prevIter
			quad, _ = iter.Next() // try again
		}
		prevIter = iter
		switch iter.CodePoint {
		case 9: // \t
			currentType = nvgSPACE
		case 11: // \v
			currentType = nvgSPACE
		case 12: // \f
			currentType = nvgSPACE
		case 0x00a0: // NBSP
			currentType = nvgSPACE
		case 10: // \n
			if prevCodePoint == 13 {
				currentType = nvgNEWLINE
			} else {
				currentType = nvgSPACE
			}
		case 13: // \r
			if prevCodePoint == 13 {
				currentType = nvgNEWLINE
			} else {
				currentType = nvgSPACE
			}
		case 0x0085: // NEL
			currentType = nvgNEWLINE
		default:
			currentType = nvgCHAR
		}
		if currentType == nvgNEWLINE {
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
				if currentType == nvgCHAR {
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
				if currentType == nvgCHAR {
					rowEnd = iter.NextIndex
					rowWidth = iter.NextX - rowStartX
					rowMaxX = quad.X1 - rowStartX
				}
				// track last end of a word
				if prevType == nvgCHAR && currentType == nvgSPACE {
					breakEnd = iter.CurrentIndex
					breakWidth = rowWidth
					breakMaxX = rowMaxX
				}
				// track last beginning of a word
				if prevType == nvgSPACE && currentType == nvgCHAR {
					wordStart = iter.CurrentIndex
					wordStartX = iter.X
					wordMinX = quad.X0 - rowStartX
				}
				// Break to new line when a character is beyond break width.
				if currentType == nvgCHAR && nextWidth > breakRowWidth {
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
	return rows
}

func createInternal(params nvgParams) (*Context, error) {
	context := &Context{
		params:     params,
		states:     make([]nvgState, 0, nvgMaxStates),
		fontImages: make([]int, nvgMaxFontImages),
		commands:   make([]float32, 0, nvgInitCommandsSize),
		cache: nvgPathCache{
			points:   make([]nvgPoint, 0, nvgInitPointsSize),
			paths:    make([]nvgPath, 0, nvgInitPathsSize),
			vertexes: make([]nvgVertex, 0, nvgInitVertsSize),
		},
	}
	context.Save()
	context.Reset()
	context.setDevicePixelRatio(1.0)
	context.params.renderCreate()

	context.fs = fontstashmini.New(nvgInitFontImageSize, nvgInitFontImageSize)

	context.fontImages[0] = context.params.renderCreateTexture(nvgTextureALPHA, nvgInitFontImageSize, nvgInitFontImageSize, 0, nil)
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

	if nvgCommands(vals[0]) != nvgCLOSE && nvgCommands(vals[0]) != nvgWINDING {
		c.commandX = vals[len(vals)-2]
		c.commandY = vals[len(vals)-1]
	}

	i := 0
	for i < len(vals) {
		switch nvgCommands(vals[i]) {
		case nvgMOVETO:
			vals[i+1], vals[i+2] = xForm.TransformPoint(vals[i+1], vals[i+2])
			i += 3
		case nvgLINETO:
			vals[i+1], vals[i+2] = xForm.TransformPoint(vals[i+1], vals[i+2])
			i += 3
		case nvgBEZIERTO:
			vals[i+1], vals[i+2] = xForm.TransformPoint(vals[i+1], vals[i+2])
			vals[i+3], vals[i+4] = xForm.TransformPoint(vals[i+3], vals[i+4])
			vals[i+5], vals[i+6] = xForm.TransformPoint(vals[i+5], vals[i+6])
			i += 7
		case nvgCLOSE:
			i++
		case nvgWINDING:
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
		case nvgMOVETO:
			cache.addPath()
			cache.addPoint(c.commands[i+1], c.commands[i+2], nvgPtCORNER, c.distTol)
			i += 3
		case nvgLINETO:
			cache.addPoint(c.commands[i+1], c.commands[i+2], nvgPtCORNER, c.distTol)
			i += 3
		case nvgBEZIERTO:
			last := cache.lastPoint()
			if last != nil {
				cache.tesselateBezier(
					last.x, last.y,
					c.commands[i+1], c.commands[i+2],
					c.commands[i+3], c.commands[i+4],
					c.commands[i+5], c.commands[i+6], 0, nvgPtCORNER, c.tessTol, c.distTol)
			}
			i += 7
		case nvgCLOSE:
			cache.closePath()
			i++
		case nvgWINDING:
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
		if ptEquals(p0.x, p0.y, p1.x, p1.y, c.distTol) && path.count > 2 {
			path.count--
			p0 = &points[path.count-1]
			path.closed = true
		}

		// Enforce winding.
		if path.count > 2 {
			area := polyArea(points, path.count)
			if path.winding == Solid && area < 0.0 {
				polyReverse(points, path.count)
			} else if path.winding == Hole && area > 0.0 {
				polyReverse(points, path.count)
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
	if c.fontImageIdx >= nvgMaxFontImages-1 {
		return false
	}
	var iw, ih int
	// if next fontImage already have a texture
	if c.fontImages[c.fontImageIdx+1] != 0 {
		iw, ih, _ = c.ImageSize(c.fontImages[c.fontImageIdx+1])
	} else { // calculate the new font image size and create it.
		iw, ih, _ = c.ImageSize(c.fontImages[c.fontImageIdx])
		if iw > ih {
			ih *= 2
		} else {
			iw *= 2
		}
		if iw > nvgMaxFontImageSize || ih > nvgMaxFontImageSize {
			iw = nvgMaxFontImageSize
			ih = nvgMaxFontImageSize
		}
		c.fontImages[c.fontImageIdx+1] = c.params.renderCreateTexture(nvgTextureALPHA, iw, ih, 0, nil)
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

	c.params.renderTriangleStrip(&paint, &state.scissor, vertexes)

	c.drawCallCount++
	c.textTriCount += len(vertexes) / 3
}
