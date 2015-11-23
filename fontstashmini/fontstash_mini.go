package fontstashmini

import (
	"github.com/shibukawa/nanovgo/fontstashmini/truetype"
	"io/ioutil"
	"math"
)

const (
	FONS_VERTEX_COUNT     = 1024
	FONS_SCRATCH_BUF_SIZE = 16000
	FONS_INIT_FONTS       = 4
	FONS_INIT_GLYPHS      = 256
	FONS_INIT_ATLAS_NODES = 256
	INVALID = -1
)

type FONSAlign int

const (
	// Horizontal Align
	ALIGN_LEFT   FONSAlign = 1 << 0 // Default
	ALIGN_CENTER           = 1 << 1
	ALIGN_RIGHT            = 1 << 2
	// Vertical Align
	ALIGN_TOP      = 1 << 3
	ALIGN_MIDDLE   = 1 << 4
	ALIGN_BOTTOM   = 1 << 5
	ALIGN_BASELINE = 1 << 6 // Default
)

type Params struct {
	width, height int
}

type State struct {
	font    int
	align   FONSAlign
	size    float32
	blur    float32
	spacing float32
}

type GlyphKey struct {
	codePoint  rune
	size, blur int16
}

type Glyph struct {
	codePoint        rune
	Index            int
	size, blur       int16
	x0, y0, x1, y1   int16
	xAdv, xOff, yOff int16
}

type Font struct {
	font      *truetype.FontInfo
	name      string
	data      []byte
	freeData  uint8
	ascender  float32
	descender float32
	lineh     float32
	glyphs    map[GlyphKey]*Glyph
	lut       []int
}

type Quad struct {
	X0, Y0, S0, T0 float32
	X1, Y1, S1, T1 float32
}

type TextIterator struct {
	stash *FontStash
	font  *Font

	X, Y, NextX, NextY, Scale, Spacing float32
	CodePoint                          rune
	Size, Blur                         int
	PrevGlyph                          *Glyph
	CurrentIndex                       int
	NextIndex                          int
	End                                int
	Runes                              []rune
}

type FontStash struct {
	params      Params
	itw, ith    float32
	textureData []byte
	dirtyRect   [4]int
	fonts       []*Font
	atlas       *Atlas
	verts       []float32
	tcoords     []float32
	scratch     []byte
	nscratch    int
	state       State
}

func New(width, height int) *FontStash {
	params := Params{
		width:  width,
		height: height,
	}
	stash := &FontStash{
		params:      params,
		atlas:       newAtlas(params.width, params.height, FONS_INIT_ATLAS_NODES),
		fonts:       make([]*Font, 0, 4),
		itw:         1.0 / float32(params.width),
		ith:         1.0 / float32(params.height),
		textureData: make([]byte, params.width*params.height),
		verts:       make([]float32, 0, FONS_VERTEX_COUNT*2),
		tcoords:     make([]float32, 0, FONS_VERTEX_COUNT*2),
		dirtyRect:   [4]int{params.width, params.height, 0, 0},
		state: State{
			size:    12.0,
			font:    0,
			blur:    0.0,
			spacing: 0.0,
			align:   ALIGN_LEFT | ALIGN_BASELINE,
		},
	}
	stash.addWhiteRect(2, 2)

	return stash
}

func (stash *FontStash) AddFont(name, path string) int {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return INVALID
	}
	return stash.AddFontFromMemory(name, data, 1)
}

func (stash *FontStash) AddFontFromMemory(name string, data []byte, freeData uint8) int {
	fontInstance, err := truetype.InitFont(data, 0)
	if err != nil {
		return INVALID
	}
	ascent, descent, lineGap := fontInstance.GetFontVMetrics()
	fh := float32(ascent - descent)

	font := &Font{
		glyphs:    make(map[GlyphKey]*Glyph),
		name:      name,
		data:      data,
		freeData:  freeData,
		font:      fontInstance,
		ascender:  float32(ascent) / fh,
		descender: float32(descent) / fh,
		lineh:     (fh + float32(lineGap)) / fh,
	}
	stash.fonts = append(stash.fonts, font)
	return len(stash.fonts) - 1
}

func (stash *FontStash) GetFontByName(name string) int {
	for i, font := range stash.fonts {
		if font.name == name {
			return i
		}
	}
	return INVALID
}

func (stash *FontStash) SetSize(size float32) {
	stash.state.size = size
}

func (stash *FontStash) SetSpacing(spacing float32) {
	stash.state.spacing = spacing
}

func (stash *FontStash) SetBlur(blur float32) {
	stash.state.blur = blur
}

func (stash *FontStash) SetAlign(align FONSAlign) {
	stash.state.align = align
}

func (stash *FontStash) SetFont(font int) {
	stash.state.font = font
}

func (stash *FontStash) GetFontName() string {
	return stash.fonts[stash.state.font].name
}

func (stash *FontStash) VerticalMetrics() (float32, float32, float32) {
	state := stash.state
	if len(stash.fonts) < state.font+1 {
		return -1, -1, -1
	}
	font := stash.fonts[state.font]
	iSize := float32(int16(state.size * 10.0))
	return font.ascender * iSize / 10.0, font.descender * iSize / 10.0, font.lineh * iSize / 10.0
}

func (stash *FontStash) LineBounds(y float32) (minY, maxY float32) {
	state := stash.state
	if len(stash.fonts) < state.font+1 {
		return -1, -1
	}
	font := stash.fonts[state.font]
	iSize := float32(int16(state.size * 10.0))

	y += stash.getVerticalAlign(font, state.align, iSize)

	// FontStash mini support only ZERO_TOPLEFT
	minY = y - font.ascender*iSize/10.0
	maxY = minY + font.lineh*iSize/10.0
	return
}

func (stash *FontStash) ValidateTexture() []int {
	if stash.dirtyRect[0] < stash.dirtyRect[2] && stash.dirtyRect[1] < stash.dirtyRect[3] {
		dirty := make([]int, 4)
		copy(dirty[0:4], stash.dirtyRect[:])
		stash.dirtyRect[0] = stash.params.width
		stash.dirtyRect[1] = stash.params.height
		stash.dirtyRect[2] = 0
		stash.dirtyRect[3] = 0
		return dirty
	}
	return nil
}

func (stash *FontStash) GetTextureData() ([]byte, int, int) {
	return stash.textureData, stash.params.width, stash.params.height
}

func (stash *FontStash) ResetAtlas(width, height int) {
	// Flush pending glyphs
	stash.flush()
	// Reset atlas
	stash.atlas.reset(width, height)
	// Clear texture data
	stash.textureData = make([]byte, width*height)
	// Reset dirty rect
	stash.dirtyRect[0] = width
	stash.dirtyRect[1] = height
	stash.dirtyRect[2] = 0
	stash.dirtyRect[3] = 0
	// reset cached glyphs
	for _, font := range stash.fonts {
		font.glyphs = make(map[GlyphKey]*Glyph)
	}
	stash.params.width = width
	stash.params.height = height
	stash.itw = 1.0 / float32(width)
	stash.ith = 1.0 / float32(height)
	// Add white rect at 0, 0 for debug drawing
	stash.addWhiteRect(2, 2)
}

func (stash *FontStash) TextBounds(x, y float32, str string) (float32, []float32) {
	return stash.TextBoundsOfRunes(x, y, []rune(str))
}

func (stash *FontStash) TextBoundsOfRunes(x, y float32, runes []rune) (float32, []float32) {
	state := stash.state
	prevGlyphIndex := -1
	size := int(state.size * 10.0)
	blur := int(state.blur)

	if len(stash.fonts) < state.font+1 {
		return 0, nil
	}
	font := stash.fonts[state.font]

	scale := font.getPixelHeightScale(state.size)
	y += stash.getVerticalAlign(font, state.align, float32(size))

	minX := x
	maxX := x
	minY := y
	maxY := y
	startX := x

	for _, codePoint := range runes {
		glyph := stash.getGlyph(font, codePoint, size, blur)
		if glyph != nil {
			var quad Quad
			quad, x, y = stash.getQuad(font, prevGlyphIndex, glyph, scale, state.spacing, x, y)
			if quad.X0 < minX {
				minX = quad.X0
			}
			if quad.X1 > maxX {
				maxX = quad.X1
			}
			if quad.Y0 < minY {
				minY = quad.Y0
			}
			if quad.Y1 > maxY {
				maxY = quad.Y1
			}
			prevGlyphIndex = glyph.Index
		} else {
			prevGlyphIndex = -1
		}
	}

	advance := x - startX

	if (state.align & ALIGN_LEFT) != 0 {
		// do nothing
	} else if (state.align & ALIGN_RIGHT) != 0 {
		minX -= advance
		maxX -= advance
	} else if (state.align & ALIGN_CENTER) != 0 {
		minX -= advance * 0.5
		maxX -= advance * 0.5
	}
	bounds := []float32{minX, minY, maxX, maxY}
	return advance, bounds
}

func (stash *FontStash) TextIter(x, y float32, str string) *TextIterator {
	return stash.TextIterForRunes(x, y, []rune(str))
}

func (stash *FontStash) TextIterForRunes(x, y float32, runes []rune) *TextIterator {
	state := stash.state
	if len(stash.fonts) < state.font+1 {
		return nil
	}
	font := stash.fonts[state.font]
	if (state.align & ALIGN_LEFT) != 0 {
		// do nothing
	} else if (state.align & ALIGN_RIGHT) != 0 {
		width, _ := stash.TextBoundsOfRunes(x, y, runes)
		x -= width
	} else if (state.align & ALIGN_CENTER) != 0 {
		width, _ := stash.TextBoundsOfRunes(x, y, runes)
		x -= width * 0.5
	}
	y += stash.getVerticalAlign(font, state.align, state.size * 10.0)
	iter := &TextIterator{
		stash:        stash,
		font:         font,
		X:            x,
		Y:            y,
		NextX:        x,
		NextY:        y,
		Spacing:      state.spacing,
		Size:         int(state.size * 10.0),
		Blur:         int(state.blur),
		Scale:        font.getPixelHeightScale(state.size),
		CurrentIndex: 0,
		NextIndex:    0,
		End:          len(runes),
		CodePoint:    0,
		PrevGlyph:    nil,
		Runes:        runes,
	}
	return iter
}

func (iter *TextIterator) Next() (quad Quad, ok bool) {
	iter.CurrentIndex = iter.NextIndex
	if iter.CurrentIndex == iter.End {
		return Quad{}, false
	}
	current := iter.NextIndex
	stash := iter.stash
	font := iter.font

	iter.CodePoint = iter.Runes[current]
	current++
	iter.X = iter.NextX
	iter.Y = iter.NextY
	glyph := stash.getGlyph(font, iter.CodePoint, iter.Size, iter.Blur)
	prevGlyphIndex := -1
	if iter.PrevGlyph != nil {
		prevGlyphIndex = iter.PrevGlyph.Index
	}
	if glyph != nil {
		quad, iter.NextX, iter.NextY = iter.stash.getQuad(font, prevGlyphIndex, glyph, iter.Scale, iter.Spacing, iter.NextX, iter.NextY)
	}
	iter.PrevGlyph = glyph
	iter.NextIndex = current
	return quad, true
}

func (stash *FontStash) flush() {
	// Flush texture
	stash.ValidateTexture()
	// Flush triangles
	if len(stash.verts) > 0 {
		stash.verts = make([]float32, 0, FONS_VERTEX_COUNT*2)
	}
}

func (stash *FontStash) addWhiteRect(w, h int) {
	gx, gy, err := stash.atlas.addRect(w, h)
	if err != nil {
		return
	}
	gr := gx + w
	gb := gy + h

	for y := gy; y < gb; y++ {
		for x := gx; x < gr; x++ {
			stash.textureData[x+y*stash.params.width] = 0xff
		}
	}

	stash.dirtyRect[0] = fons__mini(stash.dirtyRect[0], gx)
	stash.dirtyRect[1] = fons__mini(stash.dirtyRect[1], gy)
	stash.dirtyRect[2] = fons__maxi(stash.dirtyRect[2], gr)
	stash.dirtyRect[3] = fons__maxi(stash.dirtyRect[3], gb)
}

func (stash *FontStash) getVerticalAlign(font *Font, align FONSAlign, iSize float32) float32 {
	// FontStash mini support only ZERO_TOPLEFT
	if (align & ALIGN_BASELINE) != 0 {
		return 0.0
	} else if (align & ALIGN_TOP) != 0 {
		return font.ascender * iSize / 10.0
	} else if (align & ALIGN_MIDDLE) != 0 {
		return (font.ascender + font.descender) / 2.0 * iSize / 10.0
	} else if (align & ALIGN_BOTTOM) != 0 {
		return font.descender * iSize / 10.0
	}
	return 0.0
}

func (stash *FontStash) getGlyph(font *Font, codePoint rune, size, blur int) *Glyph {
	if size < 0 {
		return nil
	}
	if blur > 20 {
		blur = 20
	}
	pad := blur + 2
	glyphKey := GlyphKey{
		codePoint: codePoint,
		size:      int16(size),
		blur:      int16(blur),
	}
	glyph, ok := font.glyphs[glyphKey]
	if ok {
		return glyph
	}
	scale := font.getPixelHeightScale(float32(size) / 10.0)
	index := font.getGlyphIndex(codePoint)
	advance, _, x0, y0, x1, y1 := font.buildGlyphBitmap(index, scale)
	gw := x1 - x0 + pad*2
	gh := y1 - y0 + pad*2
	gx, gy, err := stash.atlas.addRect(gw, gh)
	if err != nil {
		return nil
	}
	gr := gx + gw
	gb := gy + gh
	width := stash.params.width
	glyph = &Glyph{
		codePoint: codePoint,
		Index:     index,
		size:      int16(size),
		blur:      int16(blur),
		x0:        int16(gx),
		y0:        int16(gy),
		x1:        int16(gr),
		y1:        int16(gb),
		xAdv:      int16(scale * float32(advance) * 10.0),
		xOff:      int16(x0 - pad),
		yOff:      int16(y0 - pad),
	}
	font.glyphs[glyphKey] = glyph
	// Rasterize
	font.renderGlyphBitmap(stash.textureData, gx+pad, gy+pad, x1-x0, y1-y0, width, scale, scale, index)
	// Make sure there is one pixel empty border
	for y := gy; y < gb; y++ {
		stash.textureData[gx+y*width] = 0
		stash.textureData[gr-1+y*width] = 0
	}
	for x := gx; x < gr; x++ {
		stash.textureData[x+gy*width] = 0
		stash.textureData[x+(gb-1)*width] = 0
	}
	if blur > 0 {
		stash.nscratch = 0
		stash.blur(gx, gy, gw, gh, blur)
	}

	stash.dirtyRect[0] = fons__mini(stash.dirtyRect[0], gx)
	stash.dirtyRect[1] = fons__mini(stash.dirtyRect[1], gy)
	stash.dirtyRect[2] = fons__maxi(stash.dirtyRect[2], gr)
	stash.dirtyRect[3] = fons__maxi(stash.dirtyRect[3], gb)

	return glyph
}

func (stash *FontStash) getQuad(font *Font, prevGlyphIndex int, glyph *Glyph, scale, spacing float32, originalX, originalY float32) (quad Quad, x, y float32) {
	x = originalX
	y = originalY
	if prevGlyphIndex != -1 {
		adv := float32(font.getGlyphKernAdvance(prevGlyphIndex, glyph.Index)) * scale
		x += float32(int(adv + spacing + 0.5))
	}
	xOff := float32(int(glyph.xOff + 1))
	yOff := float32(int(glyph.yOff + 1))
	x0 := float32(int(glyph.x0 + 1))
	y0 := float32(int(glyph.y0 + 1))
	x1 := float32(int(glyph.x1 - 1))
	y1 := float32(int(glyph.y1 - 1))
	// only support FONS_ZERO_TOPLEFT
	rx := float32(int(x + xOff))
	ry := float32(int(y + yOff))

	quad = Quad{
		X0: rx,
		Y0: ry,
		X1: rx + x1 - x0,
		Y1: ry + y1 - y0,
		S0: x0 * stash.itw,
		T0: y0 * stash.ith,
		S1: x1 * stash.itw,
		T1: y1 * stash.ith,
	}
	x += float32(int(float32(glyph.xAdv)/10.0 + 0.5))
	return
}

const (
	APREC = 16
	ZPREC = 7
)

func (stash *FontStash) blurCols(x0, y0, w, h, alpha int) {
	b := y0 + h
	r := x0 + w
	texture := stash.textureData
	textureWidth := stash.params.width
	for y := y0; y < b; y++ {
		z := 0 // force zero border
		yOffset := y * textureWidth
		for x := 1 + x0; x < r; x++ {
			offset := x + yOffset
			z += (alpha * ((int(texture[offset]) << ZPREC) - z)) >> APREC
			texture[offset] = byte(z >> ZPREC)
		}
		texture[r-1+yOffset] = 0 // force zero border
		z = 0
		for x := r - 2; x >= x0; x-- {
			offset := x + yOffset
			z += (alpha * ((int(texture[offset]) << ZPREC) - z)) >> APREC
			texture[offset] = byte(z >> ZPREC)
		}
		texture[x0+yOffset] = 0
	}
}

func (stash *FontStash) blurRows(x0, y0, w, h, alpha int) {
	b := y0 + h
	r := x0 + w
	texture := stash.textureData
	textureWidth := stash.params.width
	for x := x0; x < r; x++ {
		z := 0 // force zero border
		for y := 1 + y0; y < b; y++ {
			offset := x + y*textureWidth
			z += (alpha * ((int(texture[offset]) << ZPREC) - z)) >> APREC
			texture[offset] = byte(z >> ZPREC)
		}
		texture[x+(b-1)*textureWidth] = 0 // force zero border
		z = 0
		for y := b - 2; y >= y0; y-- {
			offset := x + y*textureWidth
			z += (alpha * ((int(texture[offset]) << ZPREC) - z)) >> APREC
			texture[offset] = byte(z >> ZPREC)
		}
		texture[x+y0*textureWidth] = 0
	}
}

func (stash *FontStash) blur(x, y, width, height, blur int) {
	sigma := float64(blur) * 0.57735 // 1 / sqrt(3)
	alpha := int(float64(1<<APREC) * (1.0 - math.Exp(-2.3/(sigma+1.0))))
	stash.blurRows(x, y, width, height, alpha)
	stash.blurCols(x, y, width, height, alpha)
	stash.blurRows(x, y, width, height, alpha)
	stash.blurCols(x, y, width, height, alpha)
}

func fons__maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func fons__mini(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (font *Font) getPixelHeightScale(size float32) float32 {
	return float32(font.font.ScaleForPixelHeight(float64(size)))
}

func (font *Font) getGlyphKernAdvance(glyph1, glyph2 int) int {
	return font.font.GetGlyphKernAdvance(glyph1, glyph2)
}

func (font *Font) getGlyphIndex(codePoint rune) int {
	return font.font.FindGlyphIndex(int(codePoint))
}

func (font *Font) buildGlyphBitmap(index int, scale float32) (advance, lsb, x0, y0, x1, y1 int) {
	advance, lsb = font.font.GetGlyphHMetrics(index)
	x0, y0, x1, y1 = font.font.GetGlyphBitmapBoxSubpixel(index, float64(scale), float64(scale), 0, 0)
	return
}

func (font *Font) renderGlyphBitmap(data []byte, offsetX, offsetY, outWidth, outHeight, outStride int, scaleX, scaleY float32, index int) {
	font.font.MakeGlyphBitmapSubpixel(data[offsetY*outStride+offsetX:], outWidth, outHeight, outStride, float64(scaleX), float64(scaleY), 0, 0, index)
}
