// +build !js

package main

import (
	"github.com/goxjs/gl"
	"github.com/goxjs/glfw"
	"github.com/shibukawa/nanovgo"
	"github.com/shibukawa/nanovgo/perfgraph"
)

const (
	iconSEARCH       = 0x1F50D
	iconCIRCLEDCROSS = 0x2716
	iconCHEVRONRIGHT = 0xE75E
	iconCHECK        = 0x2713
	iconLOGIN        = 0xE740
	iconTRASH        = 0xE729
)

var blowup bool
var screenshot bool
var premult bool

func key(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	} else if key == glfw.KeySpace && action == glfw.Press {
		blowup = !blowup
	} else if key == glfw.KeyS && action == glfw.Press {
		screenshot = true
	} else if key == glfw.KeyP && action == glfw.Press {
		premult = !premult
	}
}

func renderDemo(ctx *nanovgo.Context, mx, my, width, height, t float32, data *DemoData) {
	drawEyes(ctx, width-250, 50, 150, 100, mx, my, t)
	drawParagraph(ctx, width-450, 50, 150, 100, mx, my)
	drawGraph(ctx, 0, height/2, width, height/2, t)
	drawColorWheel(ctx, width-300, height-300, 250.0, 250.0, t)

	// Line joints
	drawLines(ctx, 120, height-50, 600, 50, t)

	// Line widths
	drawWidths(ctx, 10, 50, 30)

	// Line caps
	drawCaps(ctx, 10, 300, 30)

	drawScissor(ctx, 50, height-80, t)

	ctx.Save()
	defer ctx.Restore()

	if blowup {
		ctx.Rotate(sinF(t*0.3) * 5.0 / 180.0 * nanovgo.PI)
		ctx.Scale(2.0, 2.0)
	}

	// Widgets
	drawWindow(ctx, "Widgets `n Stuff", 50, 50, 300, 400)
	var x float32 = 60.0
	var y float32 = 95.0
	drawSearchBox(ctx, "Search", x, y, 280, 25)
	y += 40
	drawDropDown(ctx, "Effects", x, y, 280, 28)
	popy := y + 14
	y += 45

	// Form
	drawLabel(ctx, "Login", x, y, 280, 20)
	y += 25
	drawEditBox(ctx, "Email", x, y, 280, 28)
	y += 35
	drawEditBox(ctx, "Password", x, y, 280, 28)
	y += 38
	drawCheckBox(ctx, "Remember me", x, y, 140, 28)
	drawButton(ctx, iconLOGIN, "Sign in", x+138, y, 140, 28, nanovgo.RGBA(0, 96, 128, 255))
	y += 45

	// Slider
	drawLabel(ctx, "Diameter", x, y, 280, 20)
	y += 25
	drawEditBoxNum(ctx, "123.00", "px", x+180, y, 100, 28)
	drawSlider(ctx, 0.4, x, y, 170, 28)
	y += 55

	drawButton(ctx, iconTRASH, "Delete", x, y, 160, 28, nanovgo.RGBA(128, 16, 8, 255))
	drawButton(ctx, 0, "Cancel", x+170, y, 110, 28, nanovgo.RGBA(0, 0, 0, 0))

	// Thumbnails box
	drawThumbnails(ctx, 365, popy-30, 160, 300, data.images, t)
}

func main() {
	err := glfw.Init(gl.ContextWatcher)
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	// demo MSAA
	glfw.WindowHint(glfw.Samples, 4)

	window, err := glfw.CreateWindow(1000, 600, "NanoVGo", nil, nil)
	if err != nil {
		panic(err)
	}
	window.SetKeyCallback(key)
	window.MakeContextCurrent()

	//ctx, err := nanovgo.NewContext( nanovgo.ANTIALIAS | nanovgo.STENCIL_STROKES | nanovgo.DEBUG)
	ctx, err := nanovgo.NewContext(nanovgo.StencilStrokes | nanovgo.Debug)
	defer ctx.Delete()

	if err != nil {
		panic(err)
	}

	demoData := &DemoData{}
	demoData.loadData(ctx)

	glfw.SwapInterval(0)

	fps := perfgraph.NewPerfGraph("Frame Time", "sans")

	for !window.ShouldClose() {
		t, _ := fps.UpdateGraph()

		//time.Sleep(time.Second*time.Duration(0.016666 - dt))

		fbWidth, fbHeight := window.GetFramebufferSize()
		winWidth, winHeight := window.GetSize()
		mx, my := window.GetCursorPos()

		pixelRatio := float32(fbWidth) / float32(winWidth)
		gl.Viewport(0, 0, fbWidth, fbHeight)
		if premult {
			gl.ClearColor(0, 0, 0, 0)
		} else {
			gl.ClearColor(0.3, 0.3, 0.32, 1.0)
		}
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		gl.Enable(gl.CULL_FACE)
		gl.Disable(gl.DEPTH_TEST)

		ctx.BeginFrame(winWidth, winHeight, pixelRatio)

		renderDemo(ctx, float32(mx), float32(my), float32(winWidth), float32(winHeight), t, demoData)
		fps.RenderGraph(ctx, 5, 5)

		ctx.EndFrame()

		gl.Enable(gl.DEPTH_TEST)
		window.SwapBuffers()
		glfw.PollEvents()
	}

	demoData.freeData(ctx)
}
