// +build js

package nanovgo

import (
	"encoding/binary"
	"honnef.co/go/js/console"
	"math"
)

var shaderHeader string = `
#version 100
#define NANOVG_GL2 1
#define UNIFORMARRAY_SIZE 11
`

func prepareTextureBuffer(data []byte, w, h, bpp int) []byte {
	// gl.TexImage2D on WebGL doesn't allow nil as input
	if data == nil {
		data = make([]byte, w*h*bpp)
	} else if len(data) < w*h*bpp {
		data = append(data, make([]byte, w*h*bpp-len(data))...)
	}
	return data
}

func castFloat32ToByte(vertexes []float32) []byte {
	b := make([]byte, len(vertexes)*4)
	for i, v := range vertexes {
		binary.LittleEndian.PutUint32(b[4*i:], math.Float32bits(v))
	}
	return b
}

func dumpLog(values ...interface{}) {
	console.Log(values...)
}
