package nanovgo

import (
	"math"
)

type Color struct {
	R, G, B, A float32
}

func (c Color) TransRGBA(a uint8) Color {
	c.A = float32(a) / 255.0
	return c
}

func (c Color) TransRGBAf(a float32) Color {
	c.A = a
	return c
}

func (c Color) PreMultiply() Color {
	c.R *= c.A
	c.G *= c.A
	c.B *= c.A
	c.A = 1.0
	return c
}

func (c Color) List() []float32 {
	return []float32{c.R, c.G, c.B, c.A}
}

func RGB(r, g, b uint8) Color {
	return RGBA(r, g, b, 255)
}

func RGBf(r, g, b float32) Color {
	return RGBAf(r, g, b, 1.0)
}

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

func RGBA(r, g, b, a uint8) Color {
	return Color{
		R: float32(r) / 255.0,
		G: float32(g) / 255.0,
		B: float32(b) / 255.0,
		A: float32(a) / 255.0,
	}
}

func RGBAf(r, g, b, a float32) Color {
	return Color{r, g, b, a}
}

func HSL(h, s, l float32) Color {
	return HSLA(h, s, l, 255)
}

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
