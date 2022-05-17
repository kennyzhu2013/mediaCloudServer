package libretro

import (
	"errors"
	"sync/atomic"
	"xmediaEmu/pkg/clock"
)

type FPSSyncType int

const (
	FPSSyncWithFps FPSSyncType = iota
	FPSSyncWithUser
)

// globalState represents a global state in this package.
// This is available even before the game loop starts.
type globalState struct {
	err_                       atomic.Value
	syncWithFps                int32 // 0 or 1.
	maxTPS_                    int32
	isScreenClearedEveryFrame_ int32
	// screenFilterEnabled_       int32
}

func newGlobalState() *globalState {
	return &globalState{
		maxTPS_:                    DefaultTPS,
		isScreenClearedEveryFrame_: 0, //TODO: 测试时改成0.
		// screenFilterEnabled_:       1,
	}
}

func (g *globalState) err() error {
	err, ok := g.err_.Load().(error)
	if !ok {
		return nil
	}
	return err
}

func (g *globalState) setError(err error) {
	if g.err_.Load() != nil {
		return
	}
	g.err_.Store(err)
}

func (g *globalState) fpsMode() FPSSyncType {
	return FPSSyncType(atomic.LoadInt32(&g.syncWithFps))
}

func (g *globalState) setFPSMode(fpsMode FPSSyncType) {
	atomic.StoreInt32(&g.syncWithFps, int32(fpsMode))
}

func (g *globalState) MaxTPS() int {
	return int(atomic.LoadInt32(&g.maxTPS_))
}

func (g *globalState) SetMaxTPS(tps int) error {
	if tps < 0 && tps != clock.SyncWithFPS {
		return errors.New("globalState: tps must be >= 0 or SyncWithFPS. ")
	}
	atomic.StoreInt32(&g.maxTPS_, int32(tps))
	return nil
}

func (g *globalState) isScreenClearedEveryFrame() bool {
	return atomic.LoadInt32(&g.isScreenClearedEveryFrame_) != 0
}

// TODO: 每帧clear需要优化下.
func (g *globalState) setScreenClearedEveryFrame(cleared bool) {
	v := int32(0)
	if cleared {
		v = 1
	}
	atomic.StoreInt32(&g.isScreenClearedEveryFrame_, v)
}

//
//func (g *globalState) isScreenFilterEnabled() bool {
//	return atomic.LoadInt32(&g.screenFilterEnabled_) != 0
//}
//
//func (g *globalState) setScreenFilterEnabled(enabled bool) {
//	v := int32(0)
//	if enabled {
//		v = 1
//	}
//	atomic.StoreInt32(&g.screenFilterEnabled_, v)
//}
