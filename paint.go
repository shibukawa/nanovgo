package nanovgo

import (
	"math"
)

// Paint structure represent paint information including gradient and image painting.
// Context.SetFillPaint() and Context.SetStrokePaint() accept this instance.
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
	p.xform = IdentityMatrix()
	p.extent[0] = 0.0
	p.extent[1] = 0.0
	p.radius = 0.0
	p.feather = 1.0
	p.innerColor = color
	p.outerColor = color
	p.image = 0
}

// LinearGradient creates and returns a linear gradient. Parameters (sx,sy)-(ex,ey) specify the start and end coordinates
// of the linear gradient, icol specifies the start color and ocol the end color.
// The gradient is transformed by the current transform when it is passed to Context.FillPaint() or Context.StrokePaint().
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

// RadialGradient creates and returns a radial gradient. Parameters (cx,cy) specify the center, inr and outr specify
// the inner and outer radius of the gradient, icol specifies the start color and ocol the end color.
// The gradient is transformed by the current transform when it is passed to Context.FillPaint() or Context.StrokePaint().
func RadialGradient(cx, cy, inR, outR float32, iColor, oColor Color) Paint {
	r := (inR + outR) * 0.5
	f := outR - inR

	return Paint{
		xform:      TranslateMatrix(cx, cy),
		extent:     [2]float32{r, r},
		radius:     0.0,
		feather:    maxF(1.0, f),
		innerColor: iColor,
		outerColor: oColor,
	}
}

// BoxGradient creates and returns a box gradient. Box gradient is a feathered rounded rectangle, it is useful for rendering
// drop shadows or highlights for boxes. Parameters (x,y) define the top-left corner of the rectangle,
// (w,h) define the size of the rectangle, r defines the corner radius, and f feather. Feather defines how blurry
// the border of the rectangle is. Parameter icol specifies the inner color and ocol the outer color of the gradient.
// The gradient is transformed by the current transform when it is passed to Context.FillPaint() or Context.StrokePaint().
func BoxGradient(x, y, w, h, r, f float32, iColor, oColor Color) Paint {
	return Paint{
		xform:      TranslateMatrix(x+w*0.5, y+h*0.5),
		extent:     [2]float32{w * 0.5, h * 0.5},
		radius:     r,
		feather:    maxF(1.0, f),
		innerColor: iColor,
		outerColor: oColor,
	}
}

// ImagePattern creates and returns an image patter. Parameters (ox,oy) specify the left-top location of the image pattern,
// (ex,ey) the size of one image, angle rotation around the top-left corner, image is handle to the image to render.
// The gradient is transformed by the current transform when it is passed to Context.FillPaint() or Context.StrokePaint().
func ImagePattern(cx, cy, w, h, angle float32, img int, alpha float32) Paint {
	xform := RotateMatrix(angle)
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
