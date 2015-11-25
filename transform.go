package nanovgo

import (
	"math"
)

// The following functions can be used to make calculations on 2x3 transformation matrices.

// TransformMatrix is a 2x3 matrix is represented as float[6].
type TransformMatrix [6]float32

// IdentityMatrix makes the transform to identity matrix.
func IdentityMatrix() TransformMatrix {
	return TransformMatrix{1.0, 0.0, 0.0, 1.0, 0.0, 0.0}
}

// TranslateMatrix makes the transform to translation matrix matrix.
func TranslateMatrix(tx, ty float32) TransformMatrix {
	return TransformMatrix{1.0, 0.0, 0.0, 1.0, tx, ty}
}

// ScaleMatrix makes the transform to scale matrix.
func ScaleMatrix(sx, sy float32) TransformMatrix {
	return TransformMatrix{sx, 0.0, 0.0, sy, 0.0, 0.0}
}

// RotateMatrix makes the transform to rotate matrix. Angle is specified in radians.
func RotateMatrix(a float32) TransformMatrix {
	sin, cos := math.Sincos(float64(a))
	sinF := float32(sin)
	cosF := float32(cos)
	return TransformMatrix{cosF, sinF, -sinF, cosF, 0.0, 0.0}
}

// SkewXMatrix makes the transform to skew-x matrix. Angle is specified in radians.
func SkewXMatrix(a float32) TransformMatrix {
	return TransformMatrix{1.0, 0.0, float32(math.Tan(float64(a))), 1.0, 0.0, 0.0}
}

// SkewYMatrix makes the transform to skew-y matrix. Angle is specified in radians.
func SkewYMatrix(a float32) TransformMatrix {
	return TransformMatrix{1.0, float32(math.Tan(float64(a))), 0.0, 1.0, 0.0, 0.0}
}

// Multiply makes the transform to the result of multiplication of two transforms, of A = A*B.
func (t TransformMatrix) Multiply(s TransformMatrix) TransformMatrix {
	t0 := t[0]*s[0] + t[1]*s[2]
	t2 := t[2]*s[0] + t[3]*s[2]
	t4 := t[4]*s[0] + t[5]*s[2] + s[4]
	t[1] = t[0]*s[1] + t[1]*s[3]
	t[3] = t[2]*s[1] + t[3]*s[3]
	t[5] = t[4]*s[1] + t[5]*s[3] + s[5]
	t[0] = t0
	t[2] = t2
	t[4] = t4
	return t
}

// PreMultiply makes the transform to the result of multiplication of two transforms, of A = B*A.
func (t TransformMatrix) PreMultiply(s TransformMatrix) TransformMatrix {
	return s.Multiply(t)
}

// Inverse makes the destination to inverse of specified transform.
// Returns 1 if the inverse could be calculated, else 0.
func (t TransformMatrix) Inverse() TransformMatrix {
	t0 := float64(t[0])
	t1 := float64(t[1])
	t2 := float64(t[2])
	t3 := float64(t[3])
	det := t0*t3 - t2*t1
	if det > -1e-6 && det < 1e-6 {
		return IdentityMatrix()
	}
	t4 := float64(t[4])
	t5 := float64(t[5])
	invdet := 1.0 / det
	return TransformMatrix{
		float32(t3 * invdet),
		float32(-t1 * invdet),
		float32(-t2 * invdet),
		float32(t0 * invdet),
		float32((t2*t5 - t3*t4) * invdet),
		float32((t1*t4 - t0*t5) * invdet),
	}
}

// TransformPoint transforms a point by given TransformMatrix.
func (t TransformMatrix) TransformPoint(sx, sy float32) (dx, dy float32) {
	dx = sx*t[0] + sy*t[2] + t[4]
	dy = sx*t[1] + sy*t[3] + t[5]
	return
}

// ToMat3x4 makes 3x4 matrix.
func (t TransformMatrix) ToMat3x4() []float32 {
	return []float32{
		t[0], t[1], 0.0, 0.0,
		t[2], t[3], 0.0, 0.0,
		t[4], t[5], 1.0, 0.0,
	}
}

func (t TransformMatrix) getAverageScale() float32 {
	sx := math.Sqrt(float64(t[0]*t[0] + t[2]*t[2]))
	sy := math.Sqrt(float64(t[1]*t[1] + t[3]*t[3]))
	return float32((sx + sy) * 0.5)
}
