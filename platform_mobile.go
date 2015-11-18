// +build darwin linux
// +build arm arm64

package nanovgo

var shaderHeader string = `
#version 100
#define NANOVG_GL2 1
#define UNIFORMARRAY_SIZE 11
`
