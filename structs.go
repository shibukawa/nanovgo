package nanovgo

import (
	"github.com/shibukawa/nanovgo/fontstashmini"
)

type nvgParams interface {
	edgeAntiAlias() bool
	renderCreate() error
	renderCreateTexture(texType nvgTextureType, w, h int, flags ImageFlags, data []byte) int
	renderDeleteTexture(image int) error
	renderUpdateTexture(image, x, y, w, h int, data []byte) error
	renderGetTextureSize(image int) (int, int, error)
	renderViewport(width, height int)
	renderCancel()
	renderFlush()
	renderFill(paint *Paint, scissor *nvgScissor, fringe float32, bounds [4]float32, paths []nvgPath)
	renderStroke(paint *Paint, scissor *nvgScissor, fringe float32, strokeWidth float32, paths []nvgPath)
	renderTriangles(paint *Paint, scissor *nvgScissor, vertexes []nvgVertex)
	renderTriangleStrip(paint *Paint, scissor *nvgScissor, vertexes []nvgVertex)
	renderDelete()
}

type nvgPoint struct {
	x, y     float32
	dx, dy   float32
	len      float32
	dmx, dmy float32
	flags    nvgPointFlags
}

type nvgVertex struct {
	x, y, u, v float32
}

func (vtx *nvgVertex) set(x, y, u, v float32) {
	vtx.x = x
	vtx.y = y
	vtx.u = u
	vtx.v = v
}

type nvgPath struct {
	first   int
	count   int
	closed  bool
	nBevel  int
	fills   []nvgVertex
	strokes []nvgVertex
	winding Winding
	convex  bool
}

type nvgScissor struct {
	xform  TransformMatrix
	extent [2]float32
}

type nvgState struct {
	fill, stroke  Paint
	strokeWidth   float32
	miterLimit    float32
	lineJoin      LineCap
	lineCap       LineCap
	alpha         float32
	xform         TransformMatrix
	scissor       nvgScissor
	fontSize      float32
	letterSpacing float32
	lineHeight    float32
	fontBlur      float32
	textAlign     Align
	fontID        int
}

func (s *nvgState) reset() {
	s.fill.setPaintColor(RGBA(255, 255, 255, 255))
	s.stroke.setPaintColor(RGBA(0, 0, 0, 255))
	s.strokeWidth = 1.0
	s.miterLimit = 10.0
	s.lineCap = Butt
	s.lineJoin = Miter
	s.alpha = 1.0
	s.xform = IdentityMatrix()
	s.scissor.xform = IdentityMatrix()
	s.scissor.xform[0] = 0.0
	s.scissor.xform[3] = 0.0
	s.scissor.extent[0] = -1.0
	s.scissor.extent[1] = -1.0

	s.fontSize = 16.0
	s.letterSpacing = 0.0
	s.lineHeight = 1.0
	s.fontBlur = 0.0
	s.textAlign = AlignLeft | AlignBaseline
	s.fontID = fontstashmini.INVALID
}

func (s *nvgState) getFontScale() float32 {
	return minF(quantize(s.xform.getAverageScale(), 0.01), 4.0)
}

type nvgPathCache struct {
	points   []nvgPoint
	paths    []nvgPath
	vertexes []nvgVertex
	bounds   [4]float32
}

func (c *nvgPathCache) allocVertexes(n int) []nvgVertex {
	offset := len(c.vertexes)
	c.vertexes = append(c.vertexes, make([]nvgVertex, n)...)
	return c.vertexes[offset:]
}

func (c *nvgPathCache) clearPathCache() {
	c.points = c.points[:0]
	c.paths = c.paths[:0]
	c.vertexes = c.vertexes[:0]
}

func (c *nvgPathCache) lastPath() *nvgPath {
	if len(c.paths) > 0 {
		return &c.paths[len(c.paths)-1]
	}
	return nil
}

func (c *nvgPathCache) addPath() {
	c.paths = append(c.paths, nvgPath{first: len(c.points), winding: Solid})
}

func (c *nvgPathCache) lastPoint() *nvgPoint {
	if len(c.points) > 0 {
		return &c.points[len(c.points)-1]
	}
	return nil
}

func (c *nvgPathCache) addPoint(x, y float32, flags nvgPointFlags, distTol float32) {
	path := c.lastPath()

	if path.count > 0 && len(c.points) > 0 {
		lastPoint := c.lastPoint()
		if ptEquals(lastPoint.x, lastPoint.y, x, y, distTol) {
			lastPoint.flags |= flags
			return
		}
	}

	c.points = append(c.points, nvgPoint{
		x:     x,
		y:     y,
		dx:    0,
		dy:    0,
		len:   0,
		dmx:   0,
		dmy:   0,
		flags: flags,
	})
	path.count++
}

func (c *nvgPathCache) closePath() {
	path := c.lastPath()
	if path != nil {
		path.closed = true
	}
}

func (c *nvgPathCache) pathWinding(winding Winding) {
	path := c.lastPath()
	if path != nil {
		path.winding = winding
	}
}

func (c *nvgPathCache) tesselateBezier(x1, y1, x2, y2, x3, y3, x4, y4 float32, level int, flags nvgPointFlags, tessTol, distTol float32) {
	if level > 10 {
		return
	}
	dx := x4 - x1
	dy := y4 - y1
	d2 := absF(((x2-x4)*dy - (y2-y4)*dx))
	d3 := absF(((x3-x4)*dy - (y3-y4)*dx))

	if (d2+d3)*(d2+d3) < tessTol*(dx*dx+dy*dy) {
		c.addPoint(x4, y4, flags, distTol)
		return
	}

	x12 := (x1 + x2) * 0.5
	y12 := (y1 + y2) * 0.5
	x23 := (x2 + x3) * 0.5
	y23 := (y2 + y3) * 0.5
	x34 := (x3 + x4) * 0.5
	y34 := (y3 + y4) * 0.5
	x123 := (x12 + x23) * 0.5
	y123 := (y12 + y23) * 0.5
	x234 := (x23 + x34) * 0.5
	y234 := (y23 + y34) * 0.5
	x1234 := (x123 + x234) * 0.5
	y1234 := (y123 + y234) * 0.5
	c.tesselateBezier(x1, y1, x12, y12, x123, y123, x1234, y1234, level+1, 0, tessTol, distTol)
	c.tesselateBezier(x1234, y1234, x234, y234, x34, y34, x4, y4, level+1, flags, tessTol, distTol)
}

func (c *nvgPathCache) calculateJoins(w float32, lineJoin LineCap, miterLimit float32) {
	var iw float32
	if w > 0.0 {
		iw = 1.0 / w
	}
	// Calculate which joins needs extra vertices to append, and gather vertex count.
	for i := 0; i < len(c.paths); i++ {
		path := &c.paths[i]
		points := c.points[path.first:]
		p0 := &points[path.count-1]
		p1 := &points[0]
		nLeft := 0
		path.nBevel = 0
		p1Index := 0

		for j := 0; j < path.count; j++ {
			dlx0 := p0.dy
			dly0 := -p0.dx
			dlx1 := p1.dy
			dly1 := -p1.dx

			// Calculate extrusions
			p1.dmx = (dlx0 + dlx1) * 0.5
			p1.dmy = (dly0 + dly1) * 0.5
			dmr2 := p1.dmx*p1.dmx + p1.dmy*p1.dmy
			if dmr2 > 0.000001 {
				scale := minF(1.0/dmr2, 600.0)
				p1.dmx *= scale
				p1.dmy *= scale
			}

			// Clear flags, but keep the corner.
			if p1.flags&nvgPtCORNER != 0 {
				p1.flags = nvgPtCORNER
			} else {
				p1.flags = 0
			}

			// Keep track of left turns.
			cross := p1.dx*p0.dy - p0.dx*p1.dy
			if cross > 0.0 {
				nLeft++
				p1.flags |= nvgPtLEFT
			}

			// Calculate if we should use bevel or miter for inner join.
			limit := maxF(1.0, minF(p0.len, p1.len)*iw)
			if dmr2*limit*limit < 1.0 {
				p1.flags |= nvgPrINNERBEVEL
			}

			// Check to see if the corner needs to be beveled.
			if p1.flags&nvgPtCORNER != 0 {
				if dmr2*miterLimit*miterLimit < 1.0 || lineJoin == Bevel || lineJoin == Round {
					p1.flags |= nvgPtBEVEL
				}
			}

			if p1.flags&(nvgPtBEVEL|nvgPrINNERBEVEL) != 0 {
				path.nBevel++
			}

			p1Index++
			p0 = p1
			if len(points) != p1Index {
				p1 = &points[p1Index]
			}
		}
		path.convex = (nLeft == path.count)
	}
}

func (c *nvgPathCache) expandStroke(w float32, lineCap, lineJoin LineCap, miterLimit, fringeWidth, tessTol float32) {
	aa := fringeWidth
	// Calculate divisions per half circle.
	nCap := curveDivs(w, PI, tessTol)
	c.calculateJoins(w, lineJoin, miterLimit)

	// Calculate max vertex usage.
	countVertex := 0
	for i := 0; i < len(c.paths); i++ {
		path := &c.paths[i]
		if lineJoin == Round {
			countVertex += (path.count + path.nBevel*(nCap+2) + 1) * 2 // plus one for loop
		} else {
			countVertex += (path.count + path.nBevel*5 + 1) * 2 // plus one for loop
		}
		if !path.closed {
			// space for caps
			if lineCap == Round {
				countVertex += (nCap*2 + 2) * 2
			} else {
				countVertex += (3 + 3) * 2
			}
		}
	}

	dst := c.allocVertexes(countVertex)

	for i := 0; i < len(c.paths); i++ {
		path := &c.paths[i]
		points := c.points[path.first:]

		path.fills = path.fills[:0]

		// Calculate fringe or stroke
		index := 0
		var p0, p1 *nvgPoint
		var s, e, p1Index int

		if path.closed {
			// Looping
			p0 = &points[path.count-1]
			p1 = &points[0]
			s = 0
			e = path.count
			p1Index = 0
		} else {
			// Add cap
			p0 = &points[0]
			p1 = &points[1]
			s = 1
			e = path.count - 1
			p1Index = 1

			dx := p1.x - p0.x
			dy := p1.y - p0.y
			_, dx, dy = normalize(dx, dy)
			switch lineCap {
			case Butt:
				index = buttCapStart(dst, index, p0, dx, dy, w, -aa*0.5, aa)
			case Square:
				index = buttCapStart(dst, index, p0, dx, dy, w, w-aa, aa)
			case Round:
				index = roundCapStart(dst, index, p0, dx, dy, w, nCap, aa)
			}
		}

		for j := s; j < e; j++ {
			if p1.flags&(nvgPtBEVEL|nvgPrINNERBEVEL) != 0 {
				if lineJoin == Round {
					index = roundJoin(dst, index, p0, p1, w, w, 0, 1, nCap, aa)
				} else {
					index = bevelJoin(dst, index, p0, p1, w, w, 0, 1, aa)
				}
			} else {
				(&dst[index]).set(p1.x+p1.dmx*w, p1.y+p1.dmy*w, 0, 1)
				(&dst[index+1]).set(p1.x-p1.dmx*w, p1.y-p1.dmy*w, 1, 1)
				index += 2
			}
			p1Index++
			p0 = p1
			if len(points) != p1Index {
				p1 = &points[p1Index]
			}
		}

		if path.closed {
			(&dst[index]).set(dst[0].x, dst[0].y, 0, 1)
			(&dst[index+1]).set(dst[1].x, dst[1].y, 1, 1)
			index += 2
		} else {
			dx := p1.x - p0.x
			dy := p1.y - p0.y
			_, dx, dy = normalize(dx, dy)
			switch lineCap {
			case Butt:
				index = buttCapEnd(dst, index, p1, dx, dy, w, -aa*0.5, aa)
			case Square:
				index = buttCapEnd(dst, index, p1, dx, dy, w, w-aa, aa)
			case Round:
				index = roundCapEnd(dst, index, p1, dx, dy, w, nCap, aa)
			}
		}

		path.strokes = dst[0:index]
		dst = dst[index:]
	}
}

func (c *nvgPathCache) expandFill(w float32, lineJoin LineCap, miterLimit, fringeWidth float32) {
	aa := fringeWidth
	fringe := w > 0.0

	// Calculate max vertex usage.
	c.calculateJoins(w, lineJoin, miterLimit)
	countVertex := 0
	for i := 0; i < len(c.paths); i++ {
		path := &c.paths[i]
		countVertex += path.count + path.nBevel + 1
		if fringe {
			countVertex += (path.count + path.nBevel*5 + 1) * 2 // plus one for loop
		}
	}

	dst := c.allocVertexes(countVertex)

	convex := len(c.paths) == 1 && c.paths[0].convex

	for i := 0; i < len(c.paths); i++ {
		path := &c.paths[i]
		points := c.points[path.first:]

		// Calculate shape vertices.
		wOff := 0.5 * aa
		index := 0

		if fringe {
			p0 := &points[path.count-1]
			p1 := &points[0]
			p1Index := 0
			for j := 0; j < path.count; j++ {
				if p1.flags&nvgPtBEVEL != 0 {
					dlx0 := p0.dy
					dly0 := -p0.dx
					dlx1 := p1.dy
					dly1 := -p1.dx
					if p1.flags&nvgPtLEFT != 0 {
						lx := p1.x + p1.dmx*wOff
						ly := p1.y + p1.dmy*wOff
						(&dst[index]).set(lx, ly, 0.5, 1)
						index++
					} else {
						lx0 := p1.x + dlx0*wOff
						ly0 := p1.y + dly0*wOff
						lx1 := p1.x + dlx1*wOff
						ly1 := p1.y + dly1*wOff
						(&dst[index]).set(lx0, ly0, 0.5, 1)
						(&dst[index+1]).set(lx1, ly1, 0.5, 1)
						index += 2
					}
				} else {
					lx := p1.x + p1.dmx*wOff
					ly := p1.y + p1.dmy*wOff
					(&dst[index]).set(lx, ly, 0.5, 1)
					index++
				}

				p1Index++
				p0 = p1
				if len(points) != p1Index {
					p1 = &points[p1Index]
				}
			}
		} else {
			for j := 0; j < path.count; j++ {
				point := &points[j]
				(&dst[index]).set(point.x, point.y, 0.5, 1)
				index++
			}
		}
		path.fills = dst[0:index]
		dst = dst[index:]

		// Calculate fringe
		if fringe {
			lw := w + wOff
			rw := w - wOff
			var lu float32
			var ru float32 = 1.0

			// Create only half a fringe for convex shapes so that
			// the shape can be rendered without stenciling.
			if convex {
				lw = wOff // This should generate the same vertex as fill inset above.
				lu = 0.5  // Set outline fade at middle.
			}
			p0 := &points[path.count-1]
			p1 := &points[0]
			p1Index := 0
			index := 0

			// Looping
			for j := 0; j < path.count; j++ {
				if p1.flags&(nvgPtBEVEL|nvgPrINNERBEVEL) != 0 {
					index = bevelJoin(dst, index, p0, p1, lw, rw, lu, ru, fringeWidth)
				} else {
					(&dst[index]).set(p1.x+(p1.dmx*lw), p1.y+(p1.dmy*lw), lu, 1)
					(&dst[index+1]).set(p1.x+(p1.dmx*lw), p1.y+(p1.dmy*lw), lu, 1)
					index += 2
				}
				p1Index++
				p0 = p1
				if len(points) != p1Index {
					p1 = &points[p1Index]
				}
			}
			// Loop it
			(&dst[index]).set(dst[0].x, dst[0].y, lu, 1)
			(&dst[index+1]).set(dst[1].x, dst[1].y, ru, 1)
			index += 2

			path.strokes = dst[0:index]
			dst = dst[index:]
		} else {
			path.strokes = path.strokes[:0]
		}
	}
}

// GlyphPosition keeps glyph location information
type GlyphPosition struct {
	Index      int // Position of the glyph in the input string.
	Runes      []rune
	X          float32 // The x-coordinate of the logical glyph position.
	MinX, MaxX float32 // The bounds of the glyph shape.
}

// TextRow keeps row geometry information
type TextRow struct {
	Runes      []rune  // The input string.
	StartIndex int     // Index to the input text where the row starts.
	EndIndex   int     // Index to the input text where the row ends (one past the last character).
	NextIndex  int     // Index to the beginning of the next row.
	Width      float32 // Logical width of the row.
	MinX, MaxX float32 // Actual bounds of the row. Logical with and bounds can differ because of kerning and some parts over extending.
}
