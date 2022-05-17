package games

import (
	"fmt"
	"github.com/pterm/pterm"
	iImage "xmediaEmu/pkg/image"
)

// 实现一个text的游戏逻辑.
// TODO:
// 启动过程ebiten..
type GameImage struct {
	weight int
	height int
	// text   string

	//
	Index     int
	loopCount int
	//counter        int
	//kanjiText      []rune
	//kanjiTextColor color.RGBA
	//glyphs         []text.Glyph
}

func (g *GameImage) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// 更新位置之类的.
func (g *GameImage) Update() error {
	//if g.text == "" {
	//	g.text = "Hello, world,你好世界!"
	//}
	g.Index++
	if g.Index >= 25 {
		g.Index = 1
		g.loopCount++
	}

	return nil
}

// 输出到本地检验是否成功.
func (g *GameImage) Draw(dc *iImage.Context) {
	//if g.loopCount >= 3 {
	//	return
	//}
	// dc.Clear()
	filename := "outdir/" + fmt.Sprintf("%05d.jpg", g.Index)
	im, err := iImage.LoadJPG(filename)
	if err != nil {
		pterm.FgLightRed.Printf("load file failed:%s\n. err:%v", filename, err)

		// TODO:正式代码中逻辑不能出现panic.
		panic(err)
	}
	dc.DrawImage(im, 0, 0)
	pterm.FgWhite.Printf("put image index:%s done. \n", filename)
}
