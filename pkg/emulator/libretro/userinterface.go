package libretro

import (
	"errors"
	"image"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
	"xmediaEmu/pkg/hooks"
	iImage "xmediaEmu/pkg/image"
	"xmediaEmu/pkg/inpututil"
	"xmediaEmu/pkg/log"
	"xmediaEmu/pkg/mainthread"
)

// default single UI Object.
var (
	seed  = rand.New(rand.NewSource(time.Now().UnixNano())).Uint32()
	theUI = NewUserInterface(inpututil.DefaultMgr())
)

func init() {
}

func Get() *UserInterface {
	return theUI
}

func NewUserInterface(input *inpututil.InputManager) *UserInterface {
	return &UserInterface{
		runnableOnUnfocused:   true,
		bSuspended:            false,
		initWindowWidthInDIP:  640,
		initWindowHeightInDIP: 480,
		fpsMode:               inpututil.FPSIntAndKey, // 默认只接受整数序和字符串输入.
		input:                 input,
	}
}

// 输出.
type UserInterface struct {
	context *contextImpl
	title   string // game name.

	runnableOnUnfocused bool
	fpsModeInited       bool
	fpsMode             inpututil.FPSModeType   // 输入模式，默认只允许输入整数.
	input               *inpututil.InputManager // new one

	bSuspended bool

	initWindowWidthInDIP  int
	initWindowHeightInDIP int
	// initFullscreenWidthInDIP  int
	// initFullscreenHeightInDIP int
	windowOutSideWidth  int
	windowOutSideHeight int

	// err must be accessed from the main thread.
	err error

	// window image
	iwindow *iImage.Context

	running uint32

	// t is the main thread == the rendering thread.
	t mainthread.Thread
	m sync.RWMutex

	// 输出到channel通道.
	imageChannel chan<- GameFrame
}

func (u *UserInterface) SetRunnableOnUnfocused(runnableOnUnfocused bool) {
	u.runnableOnUnfocused = runnableOnUnfocused
}

func (u *UserInterface) IsRunnableOnUnfocused() bool {
	return u.runnableOnUnfocused
}

func (u *UserInterface) GetInputMgr() *inpututil.InputManager {
	return u.input
}

func (u *UserInterface) GetViewPort() image.Image {
	return u.iwindow.Image()
}

// ops only once.
func (u *UserInterface) SetWindowSize(w, h int) bool {
	if u.iwindow != nil {
		return false
	}
	u.windowOutSideWidth = w
	u.windowOutSideHeight = h
	return true
}

// TODO: 常规尺寸设定函数: 480*640. 根据目标尺寸进行对应调整.
func (u *UserInterface) AdjustSize(winWidth, winHeight int) {

}

func (u *UserInterface) SetSuspended(b bool) {
	u.bSuspended = b
}

func (u *UserInterface) SetImageChannel(imageChannel chan<- GameFrame) {
	u.imageChannel = imageChannel
}

// ops only once.
func (u *UserInterface) SetWindowTitle(t string) bool {
	if u.iwindow == nil {
		u.iwindow = iImage.NewContext(u.initWindowWidthInDIP, u.initWindowHeightInDIP)
	}
	u.title = t
	return true
}

func (u *UserInterface) GetGameName() string {
	return u.title
}

// init all, not use open-gl.
func (u *UserInterface) init() {
	if u.iwindow == nil {
		if u.initWindowWidthInDIP == 0 {
			u.initWindowWidthInDIP = 640
			u.initWindowHeightInDIP = 480
		}
		u.iwindow = iImage.NewContext(u.initWindowWidthInDIP, u.initWindowHeightInDIP)
	}
}

func (u *UserInterface) SetFPSMode(mode inpututil.FPSModeType) {
	if !u.IsRunning() {
		u.m.Lock()
		u.fpsMode = mode
		u.m.Unlock()
		return
	}
	u.t.Call(func() {
		if !u.fpsModeInited {
			u.fpsMode = mode
			return
		}
		u.input.SetFpsMode(u.fpsMode)
	})
}

func (u *UserInterface) IsRunning() bool {
	return atomic.LoadUint32(&u.running) != 0
}

func (u *UserInterface) SetRunning(running bool) {
	if running {
		atomic.StoreUint32(&u.running, 1)
	} else {
		atomic.StoreUint32(&u.running, 0)
	}
}

// update must be called from the main thread.
// 更新大小和input
func (u *UserInterface) update() (int, int, error) {
	if u.err != nil {
		return 0, 0, u.err
	}

	if u.iwindow == nil || u.iwindow.Image() == nil {
		log.Logger.Debug("update nil")
		return 0, 0, RegularTermination
	}
	// not support full screen

	// Initialize vsync after SetMonitor is called. See the comment in updateVsync.
	// Calling this inside setWindowSize didn't work (#1363).
	// 输入模式设置.
	if !u.fpsModeInited {
		// 根据if u.fpsMode 模式初始化InputManager.
		// u.input =
		u.fpsModeInited = true
		u.input.SetFpsMode(u.fpsMode)
	}

	// 手机如果调转size会变化, 调整size由函数SetWindowSize设定.
	outsideWidth, outsideHeight := u.iwindow.Width(), u.iwindow.Height()
	if u.windowOutSideWidth != outsideWidth || u.windowOutSideHeight != outsideHeight {
		u.iwindow = iImage.NewContext(outsideWidth, outsideHeight)
	} else {
		// Call updateVsync even though fpsMode is not updated.
		// The vsync state might be changed in other places (e.g., the SetSizeCallback).
		// Also, when toggling to fullscreen, vsync state might be reset unexpectedly (#1787).
		// VSync 信号负责调度从 Back Buffer 到 Frame Buffer 的复制操作.
		// u.updateVsync()
		u.iwindow.Clear() // = updateVsync
	}

	// events callback.
	// inputs call back
	if err := u.input.Update(); err != nil {
		return 0, 0, err
	}

	// focus失效，静音处理.
	// image1 unchanged.
	for u.bSuspended {
		if err := hooks.SuspendAudio(); err != nil {
			return 0, 0, err
		}
		// Wait for an arbitrary period to avoid busy loop.
		time.Sleep(time.Second / 60)

		// 可以接收inputs.
		// glfw.PollEvents()
	}
	if err := hooks.ResumeAudio(); err != nil {
		return 0, 0, err
	}

	return outsideWidth, outsideHeight, nil
}

func (u *UserInterface) SetMaxFPS(fps int) error {
	return u.context.SetMaxTPS(fps)
}

func (u *UserInterface) MaxTPS() int {
	return u.context.MaxTPS()
}

// main thread.
func (u *UserInterface) loop() error {
	defer u.t.Call(u.iwindow.Clear)
	u.SetRunning(true)

	// fps 设置问题.
	ticker := time.NewTicker(time.Second / time.Duration(u.context.MaxTPS()))
	for range ticker.C {
		if u.IsRunning() == false {
			return errors.New("Game end. ")
		}

		// TODO:
		//var outsideWidth, outsideHeight int
		//var err error
		//if u.t.Call(func() {
		//	outsideWidth, outsideHeight, err = u.update()
		//	// 分辨率.
		//	// deviceScaleFactor = u.deviceScaleFactor(u.currentMonitor())
		//}); err != nil {
		//	return err
		//}

		if err := u.context.updateFrameImpl(u.windowOutSideWidth, u.windowOutSideHeight, u.iwindow); err != nil {
			return err
		}

		// 直接将结果写入channel.
		// swapBuffers also checks IsGL, so this condition is redundant.
		// However, (*thread).Call is not good for performance due to channels.
		// Let's avoid this whenever possible (#1367).
		// 实际渲染到屏幕.
		u.t.Call(u.swapBuffers)
	}
	return nil
}

func (u *UserInterface) resetForTick() {
	u.input.GetInput().ResetForTick()
}

// swapBuffers must be called from the main thread.
func (u *UserInterface) swapBuffers() {
	// Timestamp = current?.
	u.imageChannel <- GameFrame{Image: u.iwindow.ImageRgba(), Timestamp: uint32(time.Now().UnixNano()/8333) + seed}
	log.Logger.Debugf("swapBuffers write pixels length: %d, width:%d, length:%d\n", len(u.iwindow.ImageRgba().Pix), u.windowOutSideWidth, u.windowOutSideHeight)
}
