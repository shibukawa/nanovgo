// This library processes TrueType files:
//  - parse files
//  - extract glyph metrics
//  - extract glyph shapes
//  - render glyphs to one-channel bitmaps with antialiasing (box filter)
package truetype

import (
	"container/list"
	"math"
	"sort"
)

const (
	PLATFORM_ID_UNICODE int = iota
	PLATFORM_ID_MAC
	PLATFORM_ID_ISO
	PLATFORM_ID_MICROSOFT
)

const (
	MS_EID_SYMBOL       int = 0
	MS_EID_UNICODE_BMP      = 1
	MS_EID_SHIFTJIS         = 2
	MS_EID_UNICODE_FULL     = 10
)

const (
	vmove uint8 = iota + 1
	vline
	vcurve
)

const (
	tt_FIXSHIFT uint = 10
	tt_FIX           = (1 << tt_FIXSHIFT)
	tt_FIXMASK       = (tt_FIX - 1)
)

type Vertex struct {
	X       int
	Y       int
	CX      int
	CY      int
	Type    uint8
	Padding byte
}

func (font *FontInfo) ScaleForPixelHeight(height float64) float64 {
	fheight := float64(u16(font.data, font.hhea+4) - u16(font.data, font.hhea+6))
	return height / fheight
}

func (font *FontInfo) GetGlyphBitmapBox(glyph int, scaleX, scaleY float64) (int, int, int, int) {
	return font.GetGlyphBitmapBoxSubpixel(glyph, scaleX, scaleY, 0, 0)
}

func (font *FontInfo) GetCodepointHMetrics(codepoint int) (int, int) {
	return font.GetGlyphHMetrics(font.FindGlyphIndex(codepoint))
}

func (font *FontInfo) GetFontVMetrics() (int, int, int) {
	return int(int16(u16(font.data, font.hhea+4))), int(int16(u16(font.data, font.hhea+6))), int(int16(u16(font.data, font.hhea+8)))
}

func (font *FontInfo) GetGlyphHMetrics(glyphIndex int) (int, int) {
	numOfLongHorMetrics := int(u16(font.data, font.hhea+34))
	if glyphIndex < numOfLongHorMetrics {
		return int(int16(u16(font.data, font.hmtx+4*glyphIndex))), int(int16(u16(font.data, font.hmtx+4*glyphIndex+2)))
	}
	return int(int16(u16(font.data, font.hmtx+4*(numOfLongHorMetrics-1)))), int(int16(u16(font.data, font.hmtx+4*numOfLongHorMetrics+2*(glyphIndex-numOfLongHorMetrics))))
}

func (font *FontInfo) GetFontBoundingBox() (int, int, int, int) {
	return int(int16(u16(font.data, font.head+36))),
		int(int16(u16(font.data, font.head+38))),
		int(int16(u16(font.data, font.head+40))),
		int(int16(u16(font.data, font.head+42)))
}

func (font *FontInfo) GetCodepointBitmapBox(codepoint int, scaleX, scaleY float64) (int, int, int, int) {
	return font.GetCodepointBitmapBoxSubpixel(codepoint, scaleX, scaleY, 0, 0)
}

func (font *FontInfo) GetCodepointBitmapBoxSubpixel(codepoint int, scaleX, scaleY, shiftX, shiftY float64) (int, int, int, int) {
	return font.GetGlyphBitmapBoxSubpixel(font.FindGlyphIndex(codepoint), scaleX, scaleY, shiftX, shiftY)
}

func (font *FontInfo) GetCodepointBitmap(scaleX, scaleY float64, codePoint, xoff, yoff int) ([]byte, int, int) {
	return font.GetCodepointBitmapSubpixel(scaleX, scaleY, 0., 0., codePoint, xoff, yoff)
}

func (font *FontInfo) GetCodepointBitmapSubpixel(scaleX, scaleY, shiftX, shiftY float64, codePoint, xoff, yoff int) ([]byte, int, int) {
	return font.GetGlyphBitmapSubpixel(scaleX, scaleY, shiftX, shiftY, font.FindGlyphIndex(codePoint), xoff, yoff)
}

type Bitmap struct {
	W      int
	H      int
	Stride int
	Pixels []byte
}

func (font *FontInfo) GetGlyphBitmapSubpixel(scaleX, scaleY, shiftX, shiftY float64, glyph, xoff, yoff int) ([]byte, int, int) {
	var gbm Bitmap
	var width, height int
	vertices := font.GetGlyphShape(glyph)
	if scaleX == 0 {
		scaleX = scaleY
	}
	if scaleY == 0 {
		if scaleX == 0 {
			return nil, 0, 0
		}
		scaleY = scaleX
	}

	ix0, iy0, ix1, iy1 := font.GetGlyphBitmapBoxSubpixel(glyph, scaleX, scaleY, shiftX, shiftY)

	// now we get the size
	gbm.W = ix1 - ix0
	gbm.H = iy1 - iy0
	gbm.Pixels = nil

	width = gbm.W
	height = gbm.H
	xoff = ix0
	yoff = iy0

	if gbm.W != 0 && gbm.H != 0 {
		gbm.Pixels = make([]byte, gbm.W*gbm.H)
		gbm.Stride = gbm.W

		Rasterize(&gbm, 0.35, vertices, scaleX, scaleY, shiftX, shiftY, ix0, iy0, true)
	}

	return gbm.Pixels, width, height
}

type point struct {
	x float64
	y float64
}

func Rasterize(result *Bitmap, flatnessInPixels float64, vertices []Vertex, scaleX, scaleY, shiftX, shiftY float64, xOff, yOff int, invert bool) {
	var scale float64
	if scaleX > scaleY {
		scale = scaleY
	} else {
		scale = scaleX
	}
	windings, windingLengths, windingCount := FlattenCurves(vertices, flatnessInPixels/scale)
	if windings != nil {
		tt_rasterize(result, windings, windingLengths, windingCount, scaleX, scaleY, shiftX, shiftY, xOff, yOff, invert)
	}
}

func FlattenCurves(vertices []Vertex, objspaceFlatness float64) ([]point, []int, int) {
	var contourLengths []int
	points := []point{}

	objspaceFlatnessSquared := objspaceFlatness * objspaceFlatness
	n := 0
	start := 0

	for _, vertex := range vertices {
		if vertex.Type == vmove {
			n++
		}
	}
	numContours := n

	if n == 0 {
		return nil, nil, 0
	}

	contourLengths = make([]int, n)

	var x, y float64
	n = -1
	for _, vertex := range vertices {
		switch vertex.Type {
		case vmove:
			if n >= 0 {
				contourLengths[n] = len(points) - start
			}
			n++
			start = len(points)

			x = float64(vertex.X)
			y = float64(vertex.Y)
			points = append(points, point{x, y})
		case vline:
			x = float64(vertex.X)
			y = float64(vertex.Y)
			points = append(points, point{x, y})
		case vcurve:
			tesselateCurve(&points, x, y, float64(vertex.CX), float64(vertex.CY), float64(vertex.X), float64(vertex.Y), objspaceFlatnessSquared, 0)
			x = float64(vertex.X)
			y = float64(vertex.Y)
		}
		contourLengths[n] = len(points) - start
	}
	return points, contourLengths, numContours
}

// tesselate until threshold p is happy... @TODO warped to compensate for non-linear stretching
func tesselateCurve(points *[]point, x0, y0, x1, y1, x2, y2, objspaceFlatnessSquared float64, n int) int {
	// midpoint
	mx := (x0 + 2*x1 + x2) / 4
	my := (y0 + 2*y1 + y2) / 4
	// versus directly drawn line
	dx := (x0+x2)/2 - mx
	dy := (y0+y2)/2 - my
	if n > 16 {
		return 1
	}
	if dx*dx+dy*dy > objspaceFlatnessSquared { // half-pixel error allowed... need to be smaller if AA
		tesselateCurve(points, x0, y0, (x0+x1)/2, (y0+y1)/2, mx, my, objspaceFlatnessSquared, n+1)
		tesselateCurve(points, mx, my, (x1+x2)/2, (y1+y2)/2, x2, y2, objspaceFlatnessSquared, n+1)
	} else {
		*points = append(*points, point{x2, y2})
	}
	return 1
}

type Edge struct {
	x0     float64
	y0     float64
	x1     float64
	y1     float64
	invert bool
}

type Edges []Edge

func (e Edges) Len() int      { return len(e) }
func (e Edges) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e Edges) Less(i, j int) bool {
	return e[i].y0 < e[j].y0
}

func tt_rasterize(result *Bitmap, pts []point, wcount []int, windings int, scaleX, scaleY, shiftX, shiftY float64, offX, offY int, invert bool) {
	var yScaleInv float64
	if invert {
		yScaleInv = -scaleY
	} else {
		yScaleInv = scaleY
	}
	var vsubsample int
	if result.H < 8 {
		vsubsample = 15
	} else {
		vsubsample = 5
	}
	// vsubsample should divide 255 evenly; otherwise we won't reach full opacity

	// now we have to blow out the windings into explicit edge lists
	n := 0
	for i := 0; i < windings; i++ {
		n += wcount[i]
	}

	e := make([]Edge, n+1)
	n = 0

	m := 0
	for i := 0; i < windings; i++ {
		winding := wcount[i]
		p := pts[m:]
		m += winding
		j := winding - 1
		for k := 0; k < winding; k++ {
			a := k
			b := j
			// skip the edge if horizontal
			if p[j].y == p[k].y {
				j = k
				continue
			}
			// add edge from j to k to the list
			e[n].invert = false
			if invert {
				if p[j].y > p[k].y {
					e[n].invert = true
					a = j
					b = k
				}
			} else {
				if p[j].y < p[k].y {
					e[n].invert = true
					a = j
					b = k
				}
			}
			e[n].x0 = p[a].x*scaleX + shiftX
			e[n].y0 = p[a].y*yScaleInv*float64(vsubsample) + shiftY
			e[n].x1 = p[b].x*scaleX + shiftX
			e[n].y1 = p[b].y*yScaleInv*float64(vsubsample) + shiftY
			n++
			j = k
		}
	}

	// now sort the edges by their highest point (should snap to integer, and then by x
	sort.Sort(Edges(e[:n]))

	// now, traverse the scanlines and find the intersections on each scanline, use xor winding rule
	rasterizeSortedEdges(result, e, n, vsubsample, offX, offY)
}

type activeEdge struct {
	x     int
	dx    int
	ey    float64
	valid int
}

func newActive(e *Edge, offX int, startPoint float64) *activeEdge {
	z := &activeEdge{}
	dxdy := (e.x1 - e.x0) / (e.y1 - e.y0)
	if dxdy < 0 {
		z.dx = -int(math.Floor(tt_FIX * -dxdy))
	} else {
		z.dx = int(math.Floor(tt_FIX * dxdy))
	}
	z.x = int(math.Floor(tt_FIX * (e.x0 + dxdy*(startPoint-e.y0))))
	z.x -= offX * tt_FIX
	z.ey = e.y1
	if e.invert {
		z.valid = 1
	} else {
		z.valid = -1
	}
	return z
}

func rasterizeSortedEdges(result *Bitmap, e []Edge, n, vsubsample, offX, offY int) {
	var scanline []byte
	active := list.New()
	maxWeight := (255 / vsubsample) // weight per vertical scanline

	dataLength := 512
	if result.W > 512 {
		dataLength = result.W
	}

	y := offY * vsubsample
	e[n].y0 = float64(offY+result.H)*float64(vsubsample) + 1
	var j float64
	var i int

	for j < float64(result.H) {
		scanline = make([]byte, dataLength)
		for s := 0; s < vsubsample; s++ {
			// find center of pixel for this scanline
			scanY := float64(y) + 0.5

			// update all active edges;
			// remove all active edges that terminate before the center of this scanline
			var next *list.Element
			for step := active.Front(); step != nil; step = next {
				z := step
				if z.Value.(*activeEdge).ey <= scanY {
					next = z.Next()
					active.Remove(z)
				} else {
					z.Value.(*activeEdge).x += z.Value.(*activeEdge).dx
					next = z.Next()
				}
			}

			// resort the list if needed
			for {
				changed := false
				for step := active.Front(); step != nil && step.Next() != nil; step = step.Next() {
					if step.Value.(*activeEdge).x > step.Next().Value.(*activeEdge).x {
						active.MoveBefore(step.Next(), step)
						changed = true
						step = step.Prev()
					}
				}
				if !changed {
					break
				}
			}

			// insert all edges that start before the center of this scanline -- omit ones that also end on this scanline
			for e[i].y0 <= scanY {
				if e[i].y1 > scanY {
					z := newActive(&e[i], offX, scanY)
					if active.Len() == 0 {
						active.PushBack(z)
					} else if z.x < active.Front().Value.(*activeEdge).x {
						active.PushFront(z)
					} else {
						p := active.Front()
						for p.Next() != nil && p.Next().Value.(*activeEdge).x < z.x {
							p = p.Next()
						}
						active.InsertAfter(z, p)
					}
				}
				i++
			}

			// now process all active edges in XOR fashion
			if active.Len() > 0 {
				scanline = fillActiveEdges(scanline, result.W, active, maxWeight)
			}

			y++
		}
		copy(result.Pixels[int(j)*result.Stride:], scanline[:result.W])
		// result.Pixels = append(result.Pixels[:int(j)*result.Stride], scanline[:result.W]...)
		j++
	}
}

// note: this routine clips fills that extend off the edges... ideally this
// wouldn't happen, but it could happen if the truetype glyph bounding boxes
// are wrong, or if the user supplies a too-small bitmap
func fillActiveEdges(scanline []byte, length int, e *list.List, maxWeight int) []byte {
	// non-zero winding fill
	x0 := 0
	w := 0

	for p := e.Front(); p != nil; p = p.Next() {
		if w == 0 {
			// if we're currently at zero, we need to record the edge start point
			x0 = p.Value.(*activeEdge).x
			w += p.Value.(*activeEdge).valid
		} else {
			x1 := p.Value.(*activeEdge).x
			w += p.Value.(*activeEdge).valid
			// if we went to zero, we need to draw
			if w == 0 {
				i := (x0 >> tt_FIXSHIFT)
				j := (x1 >> tt_FIXSHIFT)

				if i < length && j >= 0 {
					if i == j {
						// x0, x1 are the same pixel, so compute combined coverage
						scanline[i] = scanline[i] + uint8((x1-x0)*maxWeight>>tt_FIXSHIFT)
					} else {
						if i >= 0 { // add antialiasing for x0
							scanline[i] = scanline[i] + uint8(((tt_FIX-(x0&tt_FIXMASK))*maxWeight)>>tt_FIXSHIFT)
						} else {
							i = -1 // clip
						}

						if j < length { // add antialiasing for x1
							scanline[j] = scanline[j] + uint8(((x1&tt_FIXMASK)*maxWeight)>>tt_FIXSHIFT)
						} else {
							j = length // clip
						}

						for i++; i < j; i++ { // fill pixels between x0 and x1
							scanline[i] = scanline[i] + uint8(maxWeight)
						}
					}
				}
			}
		}
	}
	return scanline
}

func (font *FontInfo) GetGlyphBitmapBoxSubpixel(glyph int, scaleX, scaleY, shiftX, shiftY float64) (ix0, iy0, ix1, iy1 int) {
	result, x0, y0, x1, y1 := font.GetGlyphBox(glyph)
	if !result {
		x0 = 0
		y0 = 0
		x1 = 0
		y1 = 0
	}
	ix0 = int(math.Floor(float64(x0)*scaleX + shiftX))
	iy0 = -int(math.Ceil(float64(y1)*scaleY + shiftY))
	ix1 = int(math.Ceil(float64(x1)*scaleX + shiftX))
	iy1 = -int(math.Floor(float64(y0)*scaleY + shiftY))
	return
}

func (font *FontInfo) GetGlyphBox(glyph int) (result bool, x0, y0, x1, y1 int) {
	g := font.GetGlyphOffset(glyph)
	if g < 0 {
		result = false
		return
	}

	x0 = int(int16(u16(font.data, g+2)))
	y0 = int(int16(u16(font.data, g+4)))
	x1 = int(int16(u16(font.data, g+6)))
	y1 = int(int16(u16(font.data, g+8)))
	result = true
	return
}

func (font *FontInfo) GetGlyphShape(glyphIndex int) []Vertex {
	data := font.data
	g := font.GetGlyphOffset(glyphIndex)
	if g < 0 {
		return nil
	}
	var vertices []Vertex

	numberOfContours := int(int16(u16(data, g)))
	numVertices := 0

	if numberOfContours > 0 {
		var flags uint8
		endPtsOfContours := g + 10
		ins := int(u16(data, g+10+numberOfContours*2))
		points := g + 10 + numberOfContours*2 + 2 + ins

		n := 1 + int(u16(data[endPtsOfContours:], numberOfContours*2-2))

		m := n + 2*numberOfContours
		vertices = make([]Vertex, m)

		nextMove := 0
		flagcount := 0

		// in first pass, we load uninterpreted data into the allocated array
		// above, shifted to the end of the array so we won't overwrite it when
		// we create our final data starting from the front

		off := m - n // starting offset for uninterpreted data, regardless of how m ends up being calculated

		// first load flags

		for i := 0; i < n; i++ {
			if flagcount == 0 {
				flags = uint8(data[points])
				points++
				if flags&8 != 0 {
					flagcount = int(data[points])
					points++
				}
			} else {
				flagcount--
			}
			vertices[off+i].Type = flags
		}

		// now load x coordinates
		x := 0
		for i := 0; i < n; i++ {
			flags = vertices[off+i].Type
			if flags&2 != 0 {
				dx := int(data[points])
				points++
				// ???
				if flags&16 != 0 {
					x += dx
				} else {
					x -= dx
				}
			} else {
				if flags&16 == 0 {
					x = x + int(int16(data[points])*256+int16(data[points+1]))
					points += 2
				}
			}
			vertices[off+i].X = x
		}

		// now load y coordinates
		y := 0
		for i := 0; i < n; i++ {
			flags = vertices[off+i].Type
			if flags&4 != 0 {
				dy := int(data[points])
				points++
				// ???
				if flags&32 != 0 {
					y += dy
				} else {
					y -= dy
				}
			} else {
				if flags&32 == 0 {
					y = y + int(int16(data[points])*256+int16(data[points+1]))
					points += 2
				}
			}
			vertices[off+i].Y = y
		}

		// now convert them to our format
		numVertices = 0
		var sx, sy, cx, cy, scx, scy int
		var wasOff, startOff bool
		var j int
		for i := 0; i < n; i++ {
			flags = vertices[off+i].Type
			x = vertices[off+i].X
			y = vertices[off+i].Y

			if nextMove == i {
				if i != 0 {
					numVertices = closeShape(vertices, numVertices, wasOff, startOff, sx, sy, scx, scy, cx, cy)
				}

				// now start the new one
				startOff = flags&1 == 0
				if startOff {
					// if we start off with an off-curve point, then when we need to find a point on the curve
					// where we can start, and we need to save some state for when we wrap around.
					scx = x
					scy = y
					if vertices[off+i+1].Type&1 == 0 {
						// next point is also a curve point, so interpolate an on-point curve
						sx = (x + vertices[off+i+1].X) >> 1
						sy = (y + vertices[off+i+1].Y) >> 1
					} else {
						// otherwise just use the next point as our start point
						sx = vertices[off+i+1].X
						sy = vertices[off+i+1].Y
						i++
					}
				} else {
					sx = x
					sy = y
				}
				vertices[numVertices] = Vertex{Type: vmove, X: sx, Y: sy, CX: 0, CY: 0}
				numVertices++
				wasOff = false
				nextMove = 1 + int(u16(data[endPtsOfContours:], j*2))
				j++
			} else {
				if flags&1 == 0 { // if it's a curve
					if wasOff { // two off-curve control points in a row means interpolate an on-curve midpoint
						vertices[numVertices] = Vertex{Type: vcurve, X: (cx + x) >> 1, Y: (cy + y) >> 1, CX: cx, CY: cy}
						numVertices++
					}
					cx = x
					cy = y
					wasOff = true
				} else {
					if wasOff {
						vertices[numVertices] = Vertex{Type: vcurve, X: x, Y: y, CX: cx, CY: cy}
						numVertices++
					} else {
						vertices[numVertices] = Vertex{Type: vline, X: x, Y: y, CX: 0, CY: 0}
						numVertices++
					}
					wasOff = false
				}
			}
		}
		numVertices = closeShape(vertices, numVertices, wasOff, startOff, sx, sy, scx, scy, cx, cy)
	} else if numberOfContours == -1 {
		// Compound shapes.
		more := true
		comp := g + 10
		numVertices = 0
		vertices = nil
		for more {
			var mtx = [6]float64{1, 0, 0, 1, 0, 0}

			flags := int(u16(data, comp))
			comp += 2
			gidx := int(u16(data, comp))
			comp += 2

			if flags&2 != 0 { // XY values
				if flags&1 != 0 { // shorts
					mtx[4] = float64(u16(data, comp))
					comp += 2
					mtx[5] = float64(u16(data, comp))
					comp += 2
				} else {
					mtx[4] = float64(data[comp])
					comp++
					mtx[5] = float64(data[comp])
					comp++
				}
			} else {
				// @TODO handle matching point
				panic("Handle matching point")
			}
			if flags&(1<<3) != 0 { // WE_HAVE_A_SCALE
				mtx[3] = float64(u16(data, comp)) / 16384.
				comp += 2
				mtx[0] = mtx[3]
				mtx[1] = 0
				mtx[2] = 0
			} else if flags&(1<<6) != 0 { // WE_HAVE_AN_X_AND_YSCALE
				mtx[0] = float64(u16(data, comp)) / 16384.
				comp += 2
				mtx[1] = 0
				mtx[2] = 0
				mtx[3] = float64(u16(data, comp)) / 16384.
				comp += 2
			} else if flags&(1<<7) != 0 { // WE_HAVE_A_TWO_BY_TWO
				mtx[0] = float64(u16(data, comp)) / 16384.
				comp += 2
				mtx[1] = float64(u16(data, comp)) / 16384.
				comp += 2
				mtx[2] = float64(u16(data, comp)) / 16384.
				comp += 2
				mtx[3] = float64(u16(data, comp)) / 16384.
				comp += 2
			}

			// Find transformation scales.
			m := math.Sqrt(mtx[0]*mtx[0] + mtx[1]*mtx[1])
			n := math.Sqrt(mtx[2]*mtx[2] + mtx[3]*mtx[3])

			// Get indexed glyph.
			compVerts := font.GetGlyphShape(gidx)
			compNumVerts := len(compVerts)
			if compNumVerts > 0 {
				// Transform vertices.
				for i := 0; i < compNumVerts; i++ {
					v := i
					x := compVerts[v].X
					y := compVerts[v].Y
					compVerts[v].X = int(m * (mtx[0]*float64(x) + mtx[2]*float64(y) + mtx[4]))
					compVerts[v].Y = int(n * (mtx[1]*float64(x) + mtx[3]*float64(y) + mtx[5]))
					x = compVerts[v].CX
					y = compVerts[v].CY
					compVerts[v].CX = int(m * (mtx[0]*float64(x) + mtx[2]*float64(y) + mtx[4]))
					compVerts[v].CY = int(n * (mtx[1]*float64(x) + mtx[3]*float64(y) + mtx[5]))
				}
				vertices = append(vertices, compVerts...)
				numVertices += compNumVerts
			}
			// More components?
			more = flags&(1<<5) != 0
		}
	} else if numberOfContours < 0 {
		// @TODO other compound variations?
		panic("Possibly other compound variations")
	} // numberOfContours == 0, do nothing
	return vertices[:numVertices]
}

func closeShape(vertices []Vertex, numVertices int, wasOff, startOff bool, sx, sy, scx, scy, cx, cy int) int {
	if startOff {
		if wasOff {
			vertices[numVertices] = Vertex{Type: vcurve, X: (cx + scx) >> 1, Y: (cy + scy) >> 1, CX: cx, CY: cy}
			numVertices++
		}
		vertices[numVertices] = Vertex{Type: vcurve, X: sx, Y: sy, CX: scx, CY: scy}
		numVertices++
	} else {
		if wasOff {
			vertices[numVertices] = Vertex{Type: vcurve, X: sx, Y: sy, CX: cx, CY: cy}
			numVertices++
		} else {
			vertices[numVertices] = Vertex{Type: vline, X: sx, Y: sy, CX: 0, CY: 0}
			numVertices++
		}
	}
	return numVertices
}

func (font *FontInfo) GetGlyphOffset(glyphIndex int) int {
	if glyphIndex >= font.numGlyphs {
		// Glyph index out of range
		return -1
	}
	if font.indexToLocFormat >= 2 {
		// Unknown index-glyph map format
		return -1
	}

	var g1, g2 int

	if font.indexToLocFormat == 0 {
		g1 = font.glyf + int(u16(font.data, font.loca+glyphIndex*2))*2
		g2 = font.glyf + int(u16(font.data, font.loca+glyphIndex*2+2))*2
	} else {
		g1 = font.glyf + int(u32(font.data, font.loca+glyphIndex*4))
		g2 = font.glyf + int(u32(font.data, font.loca+glyphIndex*4+4))
	}

	if g1 == g2 {
		// length is 0
		return -1
	}
	return g1
}

func (font *FontInfo) MakeCodepointBitmap(output []byte, outW, outH, outStride int, scaleX, scaleY float64, codepoint int) []byte {
	return font.MakeCodepointBitmapSubpixel(output, outW, outH, outStride, scaleX, scaleY, 0, 0, codepoint)
}

func (font *FontInfo) MakeCodepointBitmapSubpixel(output []byte, outW, outH, outStride int, scaleX, scaleY, shiftX, shiftY float64, codepoint int) []byte {
	return font.MakeGlyphBitmapSubpixel(output, outW, outH, outStride, scaleX, scaleY, shiftX, shiftY, font.FindGlyphIndex(codepoint))
}

func (font *FontInfo) MakeGlyphBitmap(output []byte, outW, outH, outStride int, scaleX, scaleY float64, glyph int) []byte {
	return font.MakeGlyphBitmapSubpixel(output, outW, outH, outStride, scaleX, scaleY, 0, 0, glyph)
}

func (font *FontInfo) MakeGlyphBitmapSubpixel(output []byte, outW, outH, outStride int, scaleX, scaleY, shiftX, shiftY float64, glyph int) []byte {
	var gbm Bitmap
	vertices := font.GetGlyphShape(glyph)

	ix0, iy0, _, _ := font.GetGlyphBitmapBoxSubpixel(glyph, scaleX, scaleY, shiftX, shiftY)
	gbm.W = outW
	gbm.H = outH
	gbm.Stride = outStride

	if gbm.W > 0 && gbm.H > 0 {
		gbm.Pixels = output
		Rasterize(&gbm, 0.35, vertices, scaleX, scaleY, shiftX, shiftY, ix0, iy0, true)
	}
	return gbm.Pixels
}

func (font *FontInfo) GetCodepointKernAdvance(ch1, ch2 int) int {
	if font.kern == 0 {
		return 0
	}
	return font.GetGlyphKernAdvance(font.FindGlyphIndex(ch1), font.FindGlyphIndex(ch2))
}

func (font *FontInfo) GetGlyphKernAdvance(glyph1, glyph2 int) int {
	data := font.kern

	// we only look at the first table. it must be 'horizontal' and format 0.
	if font.kern == 0 {
		return 0
	}
	if u16(font.data, data+2) < 1 { // number of tables, need at least 1
		return 0
	}
	if u16(font.data, data+8) != 1 { // horizontal flag must be set in format
		return 0
	}

	l := 0
	r := int(u16(font.data, data+10)) - 1
	needle := uint(glyph1)<<16 | uint(glyph2)
	for l <= r {
		m := (l + r) >> 1
		straw := uint(u32(font.data, data+18+(m*6))) // note: unaligned read
		if needle < straw {
			r = m - 1
		} else if needle > straw {
			l = m + 1
		} else {
			return int(int16(u16(font.data, data+22+(m*6))))
		}
	}
	return 0
}

func (font *FontInfo) FindGlyphIndex(unicodeCodepoint int) int {
	data := font.data
	indexMap := font.indexMap

	format := int(u16(data, indexMap))
	if format == 0 { // apple byte encoding
		numBytes := int(u16(data, indexMap+2))
		if unicodeCodepoint < numBytes-6 {
			return int(data[indexMap+6+unicodeCodepoint])
		}
		return 0
	} else if format == 6 {
		first := int(u16(data, indexMap+6))
		count := int(u16(data, indexMap+8))
		if unicodeCodepoint >= first && unicodeCodepoint < first+count {
			return int(u16(data, indexMap+10+(unicodeCodepoint-first)*2))
		}
		return 0
	} else if format == 2 {
		panic("TODO: high-byte mapping for japanese/chinese/korean")
		return 0
	} else if format == 4 {
		segcount := int(u16(data, indexMap+6) >> 1)
		searchRange := int(u16(data, indexMap+8) >> 1)
		entrySelector := int(u16(data, indexMap+10))
		rangeShift := int(u16(data, indexMap+12) >> 1)

		endCount := indexMap + 14
		search := endCount

		if unicodeCodepoint > 0xffff {
			return 0
		}

		if unicodeCodepoint >= int(u16(data, search+rangeShift*2)) {
			search += rangeShift * 2
		}

		search -= 2
		for entrySelector > 0 {
			searchRange >>= 1
			// start := int(u16(data, search+2+segcount*2+2))
			// end := int(u16(data, search+2))
			// start := int(u16(data, search+searchRange*2+segcount*2+2))
			end := int(u16(data, search+searchRange*2))
			if unicodeCodepoint > end {
				search += searchRange * 2
			}
			entrySelector--
		}
		search += 2

		item := ((search - endCount) >> 1)

		if !(unicodeCodepoint <= int(u16(data, endCount+2*item))) {
			panic("unicode codepoint doesn't match")
		}
		start := int(u16(data, indexMap+14+segcount*2+2+2*item))
		// end := int(u16(data, indexMap+14+2+2*item))
		if unicodeCodepoint < start {
			return 0
		}

		offset := int(u16(data, indexMap+14+segcount*6+2+2*item))
		if offset == 0 {
			return unicodeCodepoint + int(int16(u16(data, indexMap+14+segcount*4+2+2*item)))
		}
		return int(u16(data, offset+(unicodeCodepoint-start)*2+indexMap+14+segcount*6+2+2*item))
	} else if format == 12 || format == 13 {
		ngroups := int(u32(data, indexMap+12))
		low := 0
		high := ngroups
		for low < high {
			mid := low + ((high - low) >> 1)
			startChar := int(u32(data, indexMap+16+mid*12))
			endChar := int(u32(data, indexMap+16+mid*12+4))
			if unicodeCodepoint < startChar {
				high = mid
			} else if unicodeCodepoint > endChar {
				low = mid + 1
			} else {
				startGlyph := int(u32(data, indexMap+16+mid*12+8))
				if format == 12 {
					return startGlyph + unicodeCodepoint - startChar
				} else { // format == 13
					return startGlyph
				}
			}
		}
		return 0 // not found
	}
	panic("Glyph not found!")
	return 0
}

func findTable(data []byte, offset int, tag string) int {
	numTables := int(u16(data, offset+4))
	tableDir := offset + 12
	for i := 0; i < numTables; i++ {
		loc := tableDir + 16*i
		if string(data[loc:loc+4]) == tag {
			return int(u32(data, loc+8))
		}
	}
	return 0
}

func u32(b []byte, i int) uint32 {
	return uint32(b[i])<<24 | uint32(b[i+1])<<16 | uint32(b[i+2])<<8 | uint32(b[i+3])
}

// u16 returns the big-endian uint16 at b[i:].
func u16(b []byte, i int) uint16 {
	return uint16(b[i])<<8 | uint16(b[i+1])
}

func isFont(data []byte) bool {
	if tag4(data, '1', 0, 0, 0) {
		return true
	}
	if string(data[0:4]) == "typ1" {
		return true
	}
	if string(data[0:4]) == "OTTO" {
		return true
	}
	if tag4(data, 0, 1, 0, 0) {
		return true
	}
	return false
}

func tag4(data []byte, c0, c1, c2, c3 byte) bool {
	return data[0] == c0 && data[1] == c1 && data[2] == c2 && data[3] == c3
}
