package libretro

import (
	"math"
	"sync"
	iImage "xmediaEmu/pkg/image"
	"xmediaEmu/pkg/inpututil"
	"xmediaEmu/pkg/log"
)

// TODO: fot test, write 2.
const DefaultTPS = 25 // 25 default.
type Game interface {
	Layout(outsideWidth, outsideHeight int) (int, int) // deviceScaleFactor不支持缩放暂时.
	Update() error
	Draw(screenScale float64, offsetX, offsetY float64, needsClearingScreen bool, framebufferYDirection inpututil.YDirection, screen *iImage.Context) error
}

// run context.
type contextImpl struct {
	game Game

	// The following members must be protected by the mutex m.
	// 支持手机旋转.
	outsideWidth  int
	outsideHeight int
	screenWidth   int
	screenHeight  int

	// cl *clock.Clock
	*globalState

	m sync.Mutex
}

func newContextImpl(game Game) *contextImpl {
	return &contextImpl{
		game: game,
		// cl: clock.NewClock(),
		globalState: newGlobalState(),
	}
}

func (c *contextImpl) layoutGame(outsideWidth, outsideHeight int) (int, int) {
	c.m.Lock()
	defer c.m.Unlock()

	c.outsideWidth = outsideWidth
	c.outsideHeight = outsideHeight
	w, h := c.game.Layout(outsideWidth, outsideHeight)
	c.screenWidth = w
	c.screenHeight = h
	return w, h
}

func (c *contextImpl) updateFrameImpl(outsideWidth, outsideHeight int, screen *iImage.Context) error {
	if err := c.globalState.err(); err != nil {
		return err
	}

	// The given outside size can be 0 e.g. just after restoring from the fullscreen mode on Windows (#1589)
	// Just ignore such cases. Otherwise, creating a zero-sized framebuffer causes a panic.
	if outsideWidth == 0 || outsideHeight == 0 {
		return nil
	}

	// ForceUpdate can be invoked even if the context is not initialized yet (#1591).
	if w, h := c.layoutGame(outsideWidth, outsideHeight); w == 0 || h == 0 {
		return nil
	}
	// log.Logger.Debug("----\n")

	// buffered frame， put all images into atlas.
	//if err := buffered.BeginFrame(); err != nil {
	//	return err
	//}

	log.Logger.Debug("updateFrameImpl: Update 1 per frame")

	// Update the game.
	// for i := 0; i < updateCount; i++ {
	//if err := hooks.RunBeforeUpdateHooks(); err != nil {
	//	return err
	//}
	if err := c.game.Update(); err != nil {
		return err
	}
	Get().resetForTick()
	//}

	// Draw the game.
	//screenScale, offsetX, offsetY := c.screenScaleAndOffsets()
	//if err := c.game.Draw(screenScale, offsetX, offsetY, true, inpututil.Upward, c.globalState.isScreenClearedEveryFrame(), c.globalState.isScreenFilterEnabled()); err != nil {
	//	return err
	//}

	// TODO: 测试先写死
	if err := c.game.Draw(1.0, 0, 0, c.globalState.isScreenClearedEveryFrame(), inpututil.Upward, screen); err != nil {
		return err
	}
	return nil
}

// ignore *deviceScaleFactor
func (c *contextImpl) screenScaleAndOffsets() (float64, float64, float64) {
	if c.screenWidth == 0 || c.screenHeight == 0 {
		return 0, 0, 0
	}

	scaleX := float64(c.outsideWidth) / float64(c.screenWidth)
	scaleY := float64(c.outsideHeight) / float64(c.screenHeight)
	scale := math.Min(scaleX, scaleY)
	width := float64(c.screenWidth) * scale
	height := float64(c.screenHeight) * scale
	x := (float64(c.outsideWidth) - width) / 2
	y := (float64(c.outsideHeight) - height) / 2
	return scale, x, y
}
