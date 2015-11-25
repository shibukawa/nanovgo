// +build !arm !arm64
// +build !js

package nanovgo

var shaderHeader = `
#define NANOVG_GL2 1
#define UNIFORMARRAY_SIZE 11
`
