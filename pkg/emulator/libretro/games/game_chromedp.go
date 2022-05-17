package games

import (
	"bytes"
	"context"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"image/jpeg"
	// "image/png"
	iImage "xmediaEmu/pkg/image"
)

const timeFPS = 20 // 20 frames per second.
type GameChromeDp struct {
	width  int
	height int

	chromeContext context.Context
	cancel        context.CancelFunc

	fpsCounter int
	// isInit    bool
}

func NewGameChromeDp() *GameChromeDp {
	game := &GameChromeDp{width: screenWidth, height: screenHeight}
	game.chromeContext, game.cancel = chromedp.NewContext(
		context.Background(),
		// chromedp.WithDebugf(log.Printf),
	)
	//if err := chromedp.Run(game.chromeContext, chromedp.Navigate(`https://lab.hakim.se/particles/02/`)); err != nil {
	//	panic(err)
	//}
	// chromedp.EmulateViewport(1920, 2000),  //adjust size and scale, 使用EmulateViewportOption.
	if err := chromedp.Run(game.chromeContext, // chromedp.Emulate(device.IPhone7landscape),
		chromedp.Navigate(`http://htmlpreview.github.io/?https://github.com/Syske/printer/blob/master/textPrinter.html`)); err != nil {
		panic(err)
	}

	return game
}

func (g *GameChromeDp) Layout(outsideWidth, outsideHeight int) (int, int) {
	// return screenWidth, screenHeight
	return g.width, g.height
}

// 更新位置之类的.
func (g *GameChromeDp) Update() error {
	// TODO: 滚动网页.

	return nil
}

// 输出到本地检验是否成功.
func (g *GameChromeDp) Draw(dc *iImage.Context) {
	// 测试性能打开.
	// dc.Clear()

	// 离屏渲染.
	var buf []byte
	// navigate
	//if g.isInit == false {
	//	if err := chromedp.Run(g.chromeContext, chromedp.Navigate(`https://lab.hakim.se/particles/02/`)); err != nil {
	//		panic(err)
	//	}
	//	g.isInit = true
	//}

	// 页面截图.
	if err := chromedp.Run(g.chromeContext, HighQualityScreenShot(90, &buf)); err != nil {
		panic(err)
	}

	// capture entire browser viewport, returning png with quality=90
	//if err := chromedp.Run(g.chromeContext, fullScreenshot(`https://brank.as/`, 90, &buf)); err != nil {
	//	panic(err)
	//}
	reader := bytes.NewReader(buf)
	img, err := jpeg.Decode(reader)
	if err != nil {
		panic(err)
	}

	dc.DrawImage(img, 0, 0)
}

// fullScreenshot takes a screenshot of the entire browser viewport.
// Note: chromedp.FullScreenshot overrides the device's emulation settings. Use
// device.Reset to reset the emulation and viewport settings.
//func fullScreenshot(urlstr string, quality int, res *[]byte) chromedp.Tasks {
//	return chromedp.Tasks{
//		chromedp.Navigate(urlstr),
//		chromedp.FullScreenshot(res, quality),
//	}
//}
func HighQualityScreenShot(quality int, res *[]byte) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		format := page.CaptureScreenshotFormatPng
		if quality != 100 {
			format = page.CaptureScreenshotFormatJpeg
		}

		*res, err = page.CaptureScreenshot().
			// WithCaptureBeyondViewport(true).
			WithFormat(format).
			WithQuality(int64(quality)).
			Do(ctx)
		return err
	})
}
