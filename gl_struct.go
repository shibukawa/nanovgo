package nanovgo

import (
	"github.com/goxjs/gl"
)

const (
	NSVG_SHADER_FILLGRAD = iota
	NSVG_SHADER_FILLIMG
	NSVG_SHADER_SIMPLE
	NSVG_SHADER_IMG
)

type GLNVGcallType int

const (
	GLNVG_NONE GLNVGcallType = iota
	GLNVG_FILL
	GLNVG_CONVEXFILL
	GLNVG_STROKE
	GLNVG_TRIANGLES
)

type GLCall struct {
	callType       GLNVGcallType
	image          int
	pathOffset     int
	pathCount      int
	triangleOffset int
	triangleCount  int
	uniformOffset  int
}

type GLPath struct {
	fillOffset   int
	fillCount    int
	strokeOffset int
	strokeCount  int
}

type GLFragUniforms [44]float32

func (u *GLFragUniforms) reset() {
	for i := 0; i < 44; i++ {
		u[i] = 0
	}
}

func (u *GLFragUniforms) setScissorMat(mat []float32) {
	copy(u[0:12][:], mat[0:12])
}

func (u *GLFragUniforms) setPaintMat(mat []float32) {
	copy(u[12:24], mat[0:12])
}

func (u *GLFragUniforms) setInnerColor(color Color) {
	copy(u[24:28], color.List())
}

func (u *GLFragUniforms) setOuterColor(color Color) {
	copy(u[28:32], color.List())
}

func (u *GLFragUniforms) setScissorExt(a, b float32) {
	u[32] = a
	u[33] = b
}

func (u *GLFragUniforms) setScissorScale(a, b float32) {
	u[34] = a
	u[35] = b
}

func (u *GLFragUniforms) setExtent(ext [2]float32) {
	copy(u[36:38], ext[:])
}

func (u *GLFragUniforms) setRadius(radius float32) {
	u[38] = radius
}

func (u *GLFragUniforms) setFeather(feather float32) {
	u[39] = feather
}

func (u *GLFragUniforms) setStrokeMult(strokeMult float32) {
	u[40] = strokeMult
}

func (u *GLFragUniforms) setStrokeThr(strokeThr float32) {
	u[41] = strokeThr
}

func (u *GLFragUniforms) setTexType(texType float32) {
	u[42] = texType
}

func (u *GLFragUniforms) setType(typeCode float32) {
	u[43] = typeCode
}

type GLTexture struct {
	id            int
	tex           gl.Texture
	width, height int
	texType       nvgTextureType
	flags         ImageFlags
}
