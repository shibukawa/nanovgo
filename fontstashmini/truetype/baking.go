package truetype

import (
	"bytes"
	"errors"
	"math"
)

type BakedChar struct {
	x0, y0, x1, y1       uint16 // coordinates of bbox in bitmap
	xoff, yoff, xadvance float64
}

type AlignedQuad struct {
	X0, Y0, S0, T0 float32 // top-left
	X1, Y1, S1, T1 float32 // bottom-right
}

// Call GetBakedQuad with charIndex = 'character - firstChar', and it creates
// the quad you need to draw and advances the current position.
//
// The coordinate system used assumes y increases downwards.
//
// Characters will extend both above and below the current position.
func GetBakedQuad(chardata []*BakedChar, pw, ph, charIndex int, xpos, ypos float64, openglFillRule bool) (float64, *AlignedQuad) {
	q := &AlignedQuad{}
	d3dBias := float32(-0.5)
	if openglFillRule {
		d3dBias = 0
	}
	ipw := 1 / float32(pw)
	iph := 1 / float32(ph)
	b := chardata[charIndex]
	roundX := float32(math.Floor(xpos + b.xoff + 0.5))
	roundY := float32(math.Floor(ypos + b.yoff + 0.5))

	q.X0 = roundX + d3dBias
	q.Y0 = roundY + d3dBias
	q.X1 = roundX + float32(b.x1-b.x0) + d3dBias
	q.Y1 = roundY + float32(b.y1-b.y0) + d3dBias

	q.S0 = float32(b.x0) * ipw
	q.T0 = float32(b.y0) * iph
	q.S1 = float32(b.x1) * ipw
	q.T1 = float32(b.y1) * iph

	return xpos + b.xadvance, q
}

// offset is the font location (use offset=0 for plain .ttf), pixelHeight is the height of font in pixels. pixels is the bitmap to be filled in characters to bake. This uses a very crappy packing.
func BakeFontBitmap(data []byte, offset int, pixelHeight float64, pixels []byte, pw, ph, firstChar, numChars int) (chardata []*BakedChar, err error, bottomY int, rtPixels []byte) {
	f, err := InitFont(data, offset)
	if err != nil {
		return
	}
	chardata = make([]*BakedChar, 96)
	// background of 0 around pixels
	copy(pixels, bytes.Repeat([]byte{0}, pw*ph))
	x := 1
	y := 1
	bottomY = 1

	scale := f.ScaleForPixelHeight(pixelHeight)

	for i := 0; i < numChars; i++ {
		g := f.FindGlyphIndex(firstChar + i)
		advance, _ := f.GetGlyphHMetrics(g)
		x0, y0, x1, y1 := f.GetGlyphBitmapBox(g, scale, scale)
		gw := x1 - x0
		gh := y1 - y0
		if x+gw+1 >= pw {
			// advance to next row
			y = bottomY
			x = 1
		}
		if y+gh+1 >= ph {
			// check if it fits vertically AFTER potentially moving to next row
			err = errors.New("Doesn't fit")
			bottomY = -i
			return
		}
		if !(x+gw < pw) {
			err = errors.New("Error x+gw<pw")
			return
		}
		if !(y+gh < ph) {
			err = errors.New("Error y+gh<ph")
			return
		}
		tmp := f.MakeGlyphBitmap(pixels[x+y*pw:], gw, gh, pw, scale, scale, g)
		copy(pixels[x+y*pw:], tmp)
		if chardata[i] == nil {
			chardata[i] = &BakedChar{}
		}
		chardata[i].x0 = uint16(x)
		chardata[i].y0 = uint16(y)
		chardata[i].x1 = uint16(x + gw)
		chardata[i].y1 = uint16(y + gh)
		chardata[i].xadvance = scale * float64(advance)
		chardata[i].xoff = float64(x0)
		chardata[i].yoff = float64(y0)
		x = x + gw + 2
		if y+gh+2 > bottomY {
			bottomY = y + gh + 2
		}
	}
	rtPixels = pixels
	return
}
