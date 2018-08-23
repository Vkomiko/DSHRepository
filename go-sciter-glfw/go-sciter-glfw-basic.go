package main

import (
	"runtime"
	"syscall"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/sciter-sdk/go-sciter"
	"github.com/lxn/win"
	"unsafe"
	"github.com/go-gl/gl/v4.6-core/gl"
	"fmt"
)

//TODO: sciter_sdk_patch.go:
//TODO: 必须用skia版本的sciter.dll !!!

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

var former uintptr = 0
var sciterWindowProc = syscall.NewCallback(
	func (hwnd syscall.Handle, uMsg uint32, wParam ,lParam uintptr) uintptr {
		if uMsg != win.WM_PAINT {
			if lr, handled := sciter.ProcND(win.HWND(hwnd), uint(uMsg), wParam, lParam); handled {
				return uintptr(lr)
			}
		}
		return win.CallWindowProc(former, win.HWND(hwnd), uMsg, wParam, lParam)
	})

func getHwnd (window *glfw.Window) unsafe.Pointer {
	// #ifdef WINDOWS
	return unsafe.Pointer(window.GetWin32Window())
	// #endif
}

func attach (window *glfw.Window) bool {
	// enable inspector.exe on this window
	//sciter.SetOption(sciter.SCITER_SET_DEBUG_MODE, 1)
	//sciter.SetOption(sciter.SCITER_SET_GFX_LAYER, sciter.GFX_LAYER_SKIA_OPENGL)
	//sciter.SetOption(sciter.SCITER_SET_GFX_LAYER, sciter.GFX_LAYER_SKIA)

	hw := win.HWND(getHwnd(window))

	// subclass the window
	former = win.GetWindowLongPtr(hw, win.GWLP_WNDPROC)
	win.SetWindowLongPtr(hw, win.GWLP_WNDPROC, sciterWindowProc)
	sciter.ProcND(hw, win.WM_CREATE, 0, 0)
	// sciter::attach_dom_event_handler(hw, &the_dom_events_handler);
	return true
}

func main() {
	// #ifdef WINDOWS
	win.OleInitialize()

	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)

	window, err := glfw.CreateWindow(640, 480, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()
	// glfw.SwapInterval(1)

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	gl.ClearColor(1.0, 1.0, 1.0, 1.0)
	width, height := window.GetFramebufferSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	window.SetFramebufferSizeCallback(func(w *glfw.Window, width int, height int) {
		gl.Viewport(0, 0, int32(width), int32(height))
		//ratio := float32(width) / float32(height)
	})

	// SCITER
	w := sciter.Wrap2(getHwnd(window))
	var back_layer *sciter.Element
	var fore_layer *sciter.Element
	if attach(window) {
		var errSel error
		w.LoadFile("go-sciter-glfw/sciter-glfw-basic-facade.htm")
		root, err := w.GetRootElement()
		if err != nil {
			println("hhhhh")
			println(err.Error())
		}

		if back_layer, errSel = root.SelectFirst("section#back-layer"); errSel != nil {
			println("23333333")
			println(errSel.Error())
		}
		if fore_layer, errSel = root.SelectFirst("section#fore-layer"); errSel != nil {
			println("213213")
			println(errSel.Error())
		}
		//fore_layer = root
	}
	// END SCITER

	for !window.ShouldClose() {
		// Do OpenGL stuff.
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)

		// SCITER
		pc := w.XMsgPaint(back_layer,false)
		w.ProcX(pc)
		// END SCITER

		//gl.Disable(gl.BLEND)
		// SCITER
		// draw foreground layer
		w.ProcX(w.XMsgPaint(fore_layer,true))
		// END SCITER

		//gl.Flush()

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
