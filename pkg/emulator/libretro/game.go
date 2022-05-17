package libretro

import (
	"sync/atomic"
	iImage "xmediaEmu/pkg/image"
	"xmediaEmu/pkg/inpututil"
)

// Game defines necessary functions for a game.
type GameUser interface {
	// Update updates a game by one tick. The given argument represents a screen image.
	// Update updates only the game logic and Draw draws the screen.
	//
	// In the first frame, it is ensured that Update is called at least once before Draw. You can use Update
	// to initialize the game state.
	// After the first frame, Update might not be called or might be called once
	// or more for one frame. The frequency is determined by the current TPS (tick-per-second).
	Update() error

	// Draw draws the game screen by one frame.
	//
	// The give argument represents a screen image. The updated content is adopted as the game screen.
	Draw(screen *iImage.Context)

	// Layout accepts a native outside size in device-independent pixels and returns the game's logical screen
	// size.
	// Layout is called almost every frame.
	// You can return a fixed screen size if you don't care, or you can also return a calculated screen size
	// adjusted with the given outside size.
	Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)
}

// 简单的视频逻辑展示.
type GameForUI struct {
	game GameUser

	// 采用两个context循环画输出.
	// offscreen      *iImage.Context
	// screen         *iImage.Context
	isRunGameEnded int32
}

func NewGameForUI(game GameUser) *GameForUI {
	return &GameForUI{
		game:           game,
		isRunGameEnded: int32(0),
	}
}

func (c *GameForUI) Layout(outsideWidth, outsideHeight int) (int, int) {
	ow, oh := c.game.Layout(int(outsideWidth), int(outsideHeight))
	if ow <= 0 || oh <= 0 {
		panic("Layout: Layout must return positive numbers")
	}

	//sw, sh := outsideWidth, outsideHeight
	//if c.screen != nil {
	//	w := c.screen.Width()
	//	h := c.screen.Width()
	//	if w != sw || h != sh {
	//		c.screen.Clear()
	//		c.screen = nil
	//	}
	//}
	//if c.screen == nil {
	//	c.screen = iImage.NewContext(sw, sh)
	//}

	//if c.offscreen != nil {
	//	w := c.screen.Width()
	//	h := c.screen.Width()
	//	if w != ow || h != oh {
	//		c.offscreen.Clear()
	//		c.offscreen = nil
	//	}
	//}
	//if c.offscreen == nil {
	//	c.offscreen = iImage.NewContext(ow, oh)
	//
	//	// TODO: mipmap.
	//	// Keep the offscreen an independent image from an atlas (#1938).
	//	// c.offscreen.mipmap.SetIndependent(true)
	//}

	return ow, oh
}

func (c *GameForUI) Update() error {
	return c.game.Update()
}

// filterEnabled and framebufferYDirection not use
// default screenClearedEveryFrame is true.
func (c *GameForUI) Draw(screenScale float64, offsetX, offsetY float64, needsClearingScreen bool, framebufferYDirection inpututil.YDirection, screen *iImage.Context) error {
	// Even though updateCount == 0, the offscreen is cleared and Draw is called.
	// Draw should not update the game state and then the screen should not be updated without Update, but
	// users might want to process something at Draw with the time intervals of FPS.
	if needsClearingScreen {
		screen.Clear()
	}
	c.game.Draw(screen)
	//if needsClearingScreen {
	//	// This clear is needed for fullscreen mode or some mobile platforms (#622).
	//	c.screen.Clear()
	//}
	//
	////op := &DrawImageOptions{}
	////
	////s := screenScale
	////switch framebufferYDirection {
	////case graphicsdriver.Upward:
	////	op.GeoM.Scale(s, -s)
	////	_, h := c.offscreen.Size()
	////	op.GeoM.Translate(0, float64(h)*s)
	////case graphicsdriver.Downward:
	////	op.GeoM.Scale(s, s)
	////default:
	////	panic(fmt.Sprintf("ebiten: invalid v-direction: %d", framebufferYDirection))
	////}
	//
	////op.GeoM.Translate(offsetX, offsetY)
	////op.CompositeMode = CompositeModeCopy
	////switch {
	////case !filterEnabled:
	////	op.Filter = FilterNearest
	////case math.Floor(s) == s:
	////	op.Filter = FilterNearest
	////case s > 1:
	////	op.Filter = filterScreen
	////default:
	////	// filterScreen works with >=1 scale, but does not well with <1 scale.
	////	// Use regular FilterLinear instead so far (#669).
	////	op.Filter = FilterLinear
	////}
	//c.screen.DrawImage(c.offscreen.Image(), int(offsetX), int(offsetY))
	return nil
}

func (c *GameForUI) IsRunGameEnded() bool {
	return atomic.LoadInt32(&c.isRunGameEnded) != 0
}

// 主循环..
func (c *GameForUI) RunGame(ui *UserInterface) error {
	defer atomic.StoreInt32(&c.isRunGameEnded, 1)

	if err := ui.Run(c); err != nil {
		return err
	}
	return nil
}
