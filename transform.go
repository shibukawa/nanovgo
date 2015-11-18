package nanovgo

import (
	"math"
)

type TransformMatrix [6]float32

func (t *TransformMatrix) Identity() {
	t[0] = 1.0
	t[1] = 0.0
	t[2] = 0.0
	t[3] = 1.0
	t[4] = 0.0
	t[5] = 0.0
}

func TransformMatrixTranslate(tx, ty float32) TransformMatrix {
	return TransformMatrix{1.0, 0.0, 0.0, 1.0, tx, ty}
}

func TransformMatrixScale(sx, sy float32) TransformMatrix {
	return TransformMatrix{sx, 0.0, 0.0, sy, 0.0, 0.0}
}

func TransformMatrixRotate(a float32) TransformMatrix {
	sin, cos := math.Sincos(float64(a))
	sinF := float32(sin)
	cosF := float32(cos)
	return TransformMatrix{cosF, sinF, -sinF, -cosF, 0.0, 0.0}
}

func TransformMatrixSkewX(a float32) TransformMatrix {
	return TransformMatrix{1.0, 0.0, float32(math.Tan(float64(a))), 1.0, 0.0, 0.0}
}

func TransformMatrixSkewY(a float32) TransformMatrix {
	return TransformMatrix{1.0, float32(math.Tan(float64(a))), 0.0, 1.0, 0.0, 0.0}
}

func (t *TransformMatrix) Multiply(s TransformMatrix) {
	t0 := t[0]*s[0] + t[1]*s[2]
	t2 := t[2]*s[0] + t[3]*s[2]
	t4 := t[4]*s[0] + t[5]*s[2] + s[4]
	t[1] = t[0]*s[1] + t[1]*s[3]
	t[3] = t[2]*s[1] + t[3]*s[3]
	t[5] = t[4]*s[1] + t[5]*s[3] + s[5]
	t[0] = t0
	t[2] = t2
	t[4] = t4
}

func (t *TransformMatrix) PreMultiply(s TransformMatrix) {
	s.Multiply(*t)
	*t = s
}

func (t TransformMatrix) Inverse() TransformMatrix {
	inv := TransformMatrix{}
	t0 := float64(t[0])
	t1 := float64(t[1])
	t2 := float64(t[2])
	t3 := float64(t[3])
	t4 := float64(t[4])
	t5 := float64(t[5])
	det := t0*t3 - t2*t1
	if det > -1e-6 && det < 1e-6 {
		inv.Identity()
	} else {
		invdet := 1.0 / det
		inv[0] = float32(t3 * invdet)
		inv[2] = float32(-t2 * invdet)
		inv[4] = float32((t2*t5 - t3*t4) * invdet)
		inv[1] = float32(-t1 * invdet)
		inv[3] = float32(t0 * invdet)
		inv[5] = float32((t1*t4 - t0*t5) * invdet)
	}
	return inv
}

func (t TransformMatrix) Point(sx, sy float32) (dx, dy float32) {
	dx = sx*t[0] + sy*t[2] + t[4]
	dy = sx*t[1] + sy*t[3] + t[5]
	return
}

func (t TransformMatrix) ToMat3x4() []float32 {
	return []float32{
		t[0], t[1], 0.0, 0.0,
		t[2], t[3], 0.0, 0.0,
		t[4], t[5], 1.0, 0.0,
	}
}
