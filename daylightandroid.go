// +build android

package main

import (
	"github.com/DayLightProject/go-daylight/packages/daylight"
	"image"
	"log"
	"time"
	_ "image/png"
	"github.com/c-darwin/mobile/app"
	"github.com/c-darwin/mobile/asset"
	"github.com/c-darwin/mobile/event/size"
	"github.com/c-darwin/mobile/event/paint"
	"github.com/c-darwin/mobile/exp/f32"
	"github.com/c-darwin/mobile/exp/sprite"
	"github.com/c-darwin/mobile/exp/sprite/clock"
	"github.com/c-darwin/mobile/exp/sprite/glsprite"
	"github.com/c-darwin/mobile/get_files_dir"
	"fmt"
	"runtime"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

/*
#include <stdio.h>
#include <stdlib.h>
*/
import "C"

var (
	startTime = time.Now()
	eng       = glsprite.Engine()
	scene     *sprite.Node
	cfg size.Event
)

func main() {

	var dir string
	//dir := C.GoString(C.getenv(C.CString("FILESDIR")))
	if runtime.GOOS=="android" {
		dir = get_files_dir.GetFilesDir();
	} else {
		dir = *utils.Dir
	}
	fmt.Println("dir::", dir)

	go daylight.Start(dir, nil)

	app.Main(func(a app.App) {

		for e := range a.Events() {
			fmt.Println("e:", e)
			switch e := app.Filter(e).(type) {
				case size.Event:
				cfg = e
				case paint.Event:
				onPaint(cfg)
				a.EndPaint(e)
			}
		}
	})
}

func onPaint(c size.Event) {
	loadScene()
	now := clock.Time(time.Since(startTime) * 60 / time.Second)
	eng.Render(scene, now, c)
}

func newNode() *sprite.Node {
	n := &sprite.Node{}
	eng.Register(n)
	scene.AppendChild(n)
	return n
}

func loadScene() {
	texs := loadTextures()
	scene = &sprite.Node{}
	eng.Register(scene)
	eng.SetTransform(scene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})

	//var n *sprite.Node
	new_w := float32(cfg.WidthPt)
	new_h := new_w*1.77
	if float32(cfg.WidthPt)/float32(cfg.HeightPt) > 1 {
		new_w = float32(cfg.WidthPt)
		new_h = float32(cfg.WidthPt)*0.5625

	}
	n := newNode()
	eng.SetSubTex(n, texs)
	eng.SetTransform(n, f32.Affine{
		{new_w, 0, 0},
		{0, new_h, 0},
	})
}

const (
	texBooks = iota
	texFire
	texGopherR
	texGopherL
)

func loadTextures() sprite.SubTex {
	imgPath := "mobile.png"
	w := 1080
	h := 1920
	if float32(cfg.WidthPt)/float32(cfg.HeightPt) > 1 {
		imgPath = "mobile-landscape.png"
		w = 1920
		h = 1080
	}
	a, err := asset.Open(imgPath)

	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	img, _, err := image.Decode(a)
	if err != nil {
		log.Fatal(err)
	}
	t, err := eng.LoadTexture(img)
	if err != nil {
		log.Fatal(err)
	}

	return sprite.SubTex{t, image.Rect(0, 0, w, h)}
}

type arrangerFunc func(e sprite.Engine, n *sprite.Node, t clock.Time)

func (a arrangerFunc) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) { a(e, n, t) }
