package nanovgo

import (
	"math"
)

type Paint struct {
	xform      TransformMatrix
	extent     [2]float32
	radius     float32
	feather    float32
	innerColor Color
	outerColor Color
	image      int
}

func (p *Paint) setPaintColor(color Color) {
	p.xform.Identity()
	p.extent[0] = 0.0
	p.extent[1] = 0.0
	p.radius = 0.0
	p.feather = 1.0
	p.innerColor = color
	p.outerColor = color
	p.image = 0
}

func LinearGradient(sx, sy, ex, ey float32, iColor, oColor Color) Paint {
	var large float32 = 1e5
	dx := ex - sx
	dy := ey - sy
	d := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if d > 0.0001 {
		dx /= d
		dy /= d
	} else {
		dx = 0.0
		dy = 1.0
	}

	return Paint{
		xform:      TransformMatrix{dy, -dx, dx, dy, sx - dx*large, sy - dy*large},
		extent:     [2]float32{large, large + d*0.5},
		radius:     0.0,
		feather:    maxF(1.0, d),
		innerColor: iColor,
		outerColor: oColor,
	}
}

func RadialGradient(cx, cy, inR, outR float32, iColor, oColor Color) Paint {
	r := (inR + outR) * 0.5
	f := outR - inR

	return Paint{
		xform:      TransformMatrixTranslate(cx, cy),
		extent:     [2]float32{r, r},
		radius:     0.0,
		feather:    maxF(1.0, f),
		innerColor: iColor,
		outerColor: oColor,
	}
}

func BoxGradient(x, y, w, h, r, f float32, iColor, oColor Color) Paint {
	return Paint{
		xform:      TransformMatrixTranslate(x+w*0.5, y+h*0.5),
		extent:     [2]float32{w * 0.5, h * 0.5},
		radius:     r,
		feather:    maxF(1.0, f),
		innerColor: iColor,
		outerColor: oColor,
	}
}

func ImagePattern(cx, cy, w, h, angle float32, img int, alpha float32) Paint {
	xform := TransformMatrixRotate(angle)
	xform[4] = cx
	xform[5] = cy
	color := RGBAf(1, 1, 1, alpha)
	return Paint{
		xform:      xform,
		extent:     [2]float32{w, h},
		image:      img,
		innerColor: color,
		outerColor: color,
	}
}
