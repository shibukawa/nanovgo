package nanovgo

import (
	"math"
)

// Color utils
//
// Colors in NanoVGo are stored as unsigned ints in ABGR format.
type Color struct {
	R, G, B, A float32
}

// TransRGBA sets transparency of a color value.
func (c Color) TransRGBA(a uint8) Color {
	c.A = float32(a) / 255.0
	return c
}

// TransRGBAf sets transparency of a color value.
func (c Color) TransRGBAf(a float32) Color {
	c.A = a
	return c
}

// PreMultiply preset alpha to each color.
func (c Color) PreMultiply() Color {
	c.R *= c.A
	c.G *= c.A
	c.B *= c.A
	return c
}

// List returns color value as array.
func (c Color) List() []float32 {
	return []float32{c.R, c.G, c.B, c.A}
}

// Convert To HSLA
func (c Color) HSLA() (h, s, l, a float32) {
	max := maxFs(c.R, c.G, c.B)
	min := minFs(c.R, c.G, c.B)

	l = (max + min) * 0.5

	if max == min {
		h = 0
		s = 0
	} else {
		if max == c.R {
			h = ((c.G - c.B) / (max - min)) * 1.0 / 6.0
		} else if max == c.G {
			h = ((c.B-c.R)/(max-min))*1.0/6.0 + 1.0/3.0
		} else {
			h = ((c.R-c.G)/(max-min))*1.0/6.0 + 2.0/3.0
		}
		h = float32(math.Mod(float64(h), 1.0))
		if l <= 0.5 {
			s = (max - min) / (max + min)
		} else {
			s = (max - min) / (2.0 - max - min)
		}
	}
	a = c.A
	return
}

// Calc luminance value
func (c Color) Luminance() float32 {
	return c.R*0.299 + c.G*0.587 + c.B*0.144
}

// Calc constraint color
func (c Color) ContrastingColor() Color {
	if c.Luminance() < 0.5 {
		return MONO(255, 255)
	}
	return MONO(0, 255)
}

// RGB returns a color value from red, green, blue values. Alpha will be set to 255 (1.0f).
func RGB(r, g, b uint8) Color {
	return RGBA(r, g, b, 255)
}

// RGBf returns a color value from red, green, blue values. Alpha will be set to 1.0f.
func RGBf(r, g, b float32) Color {
	return RGBAf(r, g, b, 1.0)
}

// RGBA returns a color value from red, green, blue and alpha values.
func RGBA(r, g, b, a uint8) Color {
	return Color{
		R: float32(r) / 255.0,
		G: float32(g) / 255.0,
		B: float32(b) / 255.0,
		A: float32(a) / 255.0,
	}
}

// RGBAf returns a color value from red, green, blue and alpha values.
func RGBAf(r, g, b, a float32) Color {
	return Color{r, g, b, a}
}

// HSL returns color value specified by hue, saturation and lightness.
// HSL values are all in range [0..1], alpha will be set to 255.
func HSL(h, s, l float32) Color {
	return HSLA(h, s, l, 255)
}

// HSLA returns color value specified by hue, saturation and lightness and alpha.
// HSL values are all in range [0..1], alpha in range [0..255]
func HSLA(h, s, l float32, a uint8) Color {
	h = float32(math.Mod(float64(h), 1.0))
	if h < 0.0 {
		h += 1.0
	}
	s = clampF(s, 0.0, 1.0)
	l = clampF(l, 0.0, 1.0)
	var m2 float32
	if l <= 0.5 {
		m2 = l * (1 + s)
	} else {
		m2 = l + s - l*s
	}
	m1 := 2*l - m2
	return Color{
		R: clampF(hue(h+1.0/3.0, m1, m2), 0.0, 1.0),
		G: clampF(hue(h, m1, m2), 0.0, 1.0),
		B: clampF(hue(h-1.0/3.0, m1, m2), 0.0, 1.0),
		A: float32(a) / 255.0,
	}
}

// MONO returns color value specified by intensity value.
func MONO(i, alpha uint8) Color {
	return RGBA(i, i, i, alpha)
}

// MONOf returns color value specified by intensity value.
func MONOf(i, alpha float32) Color {
	return RGBAf(i, i, i, alpha)
}

// LerpRGBA linearly interpolates from color c0 to c1, and returns resulting color value.
func LerpRGBA(c0, c1 Color, u float32) Color {
	u = clampF(u, 0.0, 1.0)
	oneMinus := 1 - u
	return Color{
		R: c0.R*oneMinus + c1.R*u,
		G: c0.G*oneMinus + c1.G*u,
		B: c0.B*oneMinus + c1.B*u,
		A: c0.A*oneMinus + c1.A*u,
	}
}
