// +build js

package main

import (
	"fmt"
	"github.com/goxjs/gl"
	"github.com/goxjs/glfw"
	"github.com/shibukawa/nanovgo"
	"github.com/shibukawa/nanovgo/perfgraph"
	"github.com/shibukawa/nanovgo/sample/demo"
	"io/ioutil"
	"log"
	"net/http"
)

var blowup bool
var premult bool

func key(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	} else if key == glfw.KeySpace && action == glfw.Press {
		blowup = !blowup
	} else if key == glfw.KeyP && action == glfw.Press {
		premult = !premult
	}
}

func main() {
	err := glfw.Init(gl.ContextWatcher)
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.StencilBits, 1)
	glfw.WindowHint(glfw.Samples, 4)

	window, err := glfw.CreateWindow(1000*0.6, 600*0.6, "NanoVGo", nil, nil)
	if err != nil {
		panic(err)
	}
	window.SetKeyCallback(key)
	window.MakeContextCurrent()

	ctx, err := nanovgo.NewContext(0)
	defer ctx.Delete()

	if err != nil {
		panic(err)
	}

	demoData := LoadDemo(ctx)

	glfw.SwapInterval(0)

	fps := perfgraph.NewPerfGraph("Frame Time", "sans")

	for !window.ShouldClose() {
		t, _ := fps.UpdateGraph()

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

		demo.RenderDemo(ctx, float32(mx), float32(my), float32(winWidth), float32(winHeight), t, blowup, demoData)
		fps.RenderGraph(ctx, 5, 5)

		ctx.EndFrame()

		gl.Enable(gl.DEPTH_TEST)
		window.SwapBuffers()
		glfw.PollEvents()
	}

	demoData.FreeData(ctx)
}

func LoadDemo(ctx *nanovgo.Context) *demo.DemoData {
	d := &demo.DemoData{}
	for i := 0; i < 12; i++ {
		path := fmt.Sprintf("assets/image%d.jpg", i+1)
		d.Images = append(d.Images, ctx.CreateImageFromMemory(0, demo.MustAsset(path)))
		if d.Images[i] == 0 {
			log.Fatalf("Could not load %s", path)
		}
	}

	d.FontIcons = ctx.CreateFontFromMemory("icons", demo.MustAsset("assets/entypo.ttf"), 0)
	if d.FontIcons == -1 {
		log.Fatalln("Could not add font icons.")
	}
	d.FontNormal = ctx.CreateFontFromMemory("sans", demo.MustAsset("assets/Roboto-Regular.ttf"), 0)
	if d.FontNormal == -1 {
		log.Fatalln("Could not add font italic.")
	}
	d.FontBold = ctx.CreateFontFromMemory("sans-bold", demo.MustAsset("assets/Roboto-Bold.ttf"), 0)
	if d.FontBold == -1 {
		log.Fatalln("Could not add font bold.")
	}
	return d
}

func readFile(path string) ([]byte, error) {
	resp, err := http.Get("/nanovgo/" + path)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("/nanovgo/" + path + ": " + resp.Status)
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
