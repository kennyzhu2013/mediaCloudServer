package libretro

import (
	"errors"
	"xmediaEmu/pkg/mainthread"
)

// RegularTermination represents a regular termination.
// Run can return this error, and if this error is received,
// the game loop should be terminated as soon as possible.
var RegularTermination = errors.New("regular termination")

func (u *UserInterface) Run(game Game) error {
	u.context = newContextImpl(game)

	// Initialize the main thread first so the thread is available at u.run (#809).
	u.t = mainthread.NewOSThread()
	// graphicscommand.SetRenderingThread(u.t)

	ch := make(chan error, 1)
	go func() {
		defer u.t.Stop()

		defer close(ch)

		// main image init.
		u.t.Call(func() {
			u.init()
		})

		if err := u.loop(); err != nil {
			ch <- err
			return
		}
	}()

	// 主线程干啥需要进步一步确定.
	u.SetRunning(true)
	u.t.Loop()
	u.SetRunning(false)
	return <-ch
}

// InternalImageSize returns a nearest appropriate size as an internal image.
func InternalImageSize(x int) int {
	// minInternalImageSize is the minimum size of internal images (texture/framebuffer).
	//
	// For example, the image size less than 15 is not supported on some iOS devices.
	const minInternalImageSize = 16

	if x <= 0 {
		panic("graphics: x must be positive")
	}
	if x < minInternalImageSize {
		return minInternalImageSize
	}
	r := 1
	for r < x {
		r <<= 1
	}
	return r
}
