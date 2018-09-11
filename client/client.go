package main

import (
	"log"
	"github.com/sciter-sdk/go-sciter"
	"github.com/sciter-sdk/go-sciter/window"
	"fmt"
	"runtime"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	fmt.Println(sciter.VersionAsString())
	w, err := window.New(sciter.DefaultWindowCreateFlag, sciter.DefaultRect)
	if err != nil {
		log.Fatal(err)
	}
	w.LoadFile("main.html")
	w.SetTitle("DSHRepo Client")
	w.Show()

	pixmap := make([]byte, 800*600*4, 800*600*4)
	img, _ := sciter.CreatImageFromPixmap(800, 600, sciter.True, &pixmap[0])
	fmt.Println(img)

	v0 := sciter.NewValue("abc")
	fmt.Println(v0)

	v, _ := img.ToValue()
	fmt.Println(v)

	w.DefineFunction("test_func", func(args ...*sciter.Value) *sciter.Value {
		fmt.Println(args)
		gfx := args[0]
		fmt.Println(gfx.String())
		sciter.TTT_gfc(gfx)

		//v, _ := img.ToValue()
		//fmt.Println(v)

		return sciter.NewValue("hhh")
	})

	w.Run()

}


