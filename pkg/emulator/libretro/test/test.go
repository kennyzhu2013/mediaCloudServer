package main

import (
	"fmt"
	"xmediaEmu/pkg/emulator/libretro"
	"xmediaEmu/pkg/emulator/libretro/games"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

//func TextMain() {
//	const S = 1024
//	dc := image.NewContext(S, S)
//	dc.SetRGB(1, 1, 1)
//	dc.Clear()
//	dc.SetRGB(0, 0, 0)
//	if err := dc.LoadFontFace("resources/msyh.ttc", 96); err != nil {
//		panic(err)
//	}
//	dc.DrawStringAnchored("Hello, world,你好世界!", S/2, S/2, 0.5, 0.5)
//	dc.SavePNG("out.png")
//}

// 以默认theUI为例.
func gameMain() {
	// FPS如何定义.
	ui := libretro.NewGameForUI(&games.Game{})
	libretro.Get().SetWindowSize(screenWidth, screenHeight)
	if err := ui.RunGame(libretro.Get()); err != nil {
		fmt.Printf("Run game failed: %v.\n ", err)
	}
}
