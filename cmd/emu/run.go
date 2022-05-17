package emu

import (
	"xmediaEmu/pkg/mainthread"
)

// main thread run
func Run(run func()) {
	// 初始化和终结进程.
	//err := glfw.Init()
	//if err != nil {
	//	panic(errors.Wrap(err, "failed to initialize GLFW"))
	//}
	//defer glfw.Terminate()
	mainthread.Run(run)
}
