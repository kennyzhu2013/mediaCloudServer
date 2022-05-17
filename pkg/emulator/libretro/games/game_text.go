package games

import (
	iImage "xmediaEmu/pkg/image"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

// 实现一个text的游戏逻辑.
// TODO:
// 启动过程ebiten..
type Game struct {
	weight int
	height int
	text   string
	index  int

	// bFontLoaded bool
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// 更新位置之类的.
func (g *Game) Update() error {
	if g.text == "" {
		g.text = "Hello, world,你好世界!"
	}
	g.index++

	return nil
}

// 输出到本地检验是否成功.
func (g *Game) Draw(dc *iImage.Context) {
	// 文字需要clear..
	dc.Clear()
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	if err := dc.LoadFontFace("resources/msyh.ttc", 48); err != nil {
		panic(err)
	}
	dc.DrawStringAnchored("Hello, world,你好世界!", float64(dc.Width()/2), float64(dc.Height()/2), 0.5, 0.5)

	if g.index < 10 {
		dc.SavePNG("out.png")
	}
}
