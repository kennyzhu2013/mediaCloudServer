package inpututil

import (
	"common/util/mem"
	"sync"
)

const _maxKeyBuffs  = 100
type pos struct {
	X int
	Y int
}

// for all base inputs.
type Input struct {
	keyPressed         map[Key]bool // 循环扫描每次的健是否有按下...
	// mouseButtonPressed map[MouseButton]bool
	// onceCallback       sync.Once

	// 以下两个
	// This function sets the scroll callback of the specified window, which is called when a scrolling device is used, such as a mouse wheel or scrolling area of a touchpad.
	scrollX            float64
	scrollY            float64

	// mouse moved.
	cursorX            int // 鼠标跟踪...
	cursorY            int
	// runeBuffer         []rune // int32.

	// input list.
	// input channels.
	keyBuffer *mem.Deque
	m sync.RWMutex

	//
}

// NewInput generates a new Input object.
func NewInput() *Input {
	return &Input{ keyBuffer: mem.New(_maxKeyBuffs), }
}

// Keyboards don't work on iOS yet (#1090).
// 用于主循环中判断.
func  (i *Input) IsKeyPressed(key Key) bool {
	if !key.isValid() {
		return false
	}
	inputKey := i.Get()
	if !inputKey.isValid() {
		return false
	}

	var keys []Key
	switch key {
	case KeyAlt:
		keys = []Key{KeyAltLeft, KeyAltRight}
	case KeyControl:
		keys = []Key{KeyControlLeft, KeyControlRight}
	case KeyShift:
		keys = []Key{KeyShiftLeft, KeyShiftRight}
	case KeyMeta:
		keys = []Key{KeyMetaLeft, KeyMetaRight}
	default:
		keys = []Key{Key(key)}
	}
	for _, k := range keys {
		if i.keyPressed == nil {
			i.keyPressed = map[Key]bool{}
			return false
		}

		if i.keyPressed[k] == true {
			return true
		}

	}
	return false
}

// Block until chan is nil.
func (i *Input) InputKey(key Key)  {
	i.m.Lock()
	i.keyBuffer.PushBack(key)
	i.m.Unlock()
}

// Block until chan is nil.
func (i *Input) Get()  Key {
	i.m.Lock()
	defer i.m.Unlock()

	if i.keyBuffer.Len() <= 0 {
		return KeyMax
	}

	return i.keyBuffer.PopFront().(Key)
}

// reset for Update.
func (i *Input) ResetForTick() {
	i.scrollX, i.scrollY = 0, 0
}


//func (i *Input) IsMouseButtonPressed(button MouseButton) bool {
//	if !i.ui.isRunning() {
//		return false
//	}
//
//	i.ui.m.Lock()
//	defer i.ui.m.Unlock()
//	if i.mouseButtonPressed == nil {
//		i.mouseButtonPressed = map[MouseButton]bool{}
//	}
//	for gb, b := range mouseButtonToMouseButton {
//		if b != button {
//			continue
//		}
//		if i.mouseButtonPressed[gb] {
//			return true
//		}
//	}
//	return false
//}
func (i *Input) CursorPosition() (x, y int) {
	return i.cursorX, i.cursorY
}

// set by external
func (i *Input) SetCursorPosition(x, y int) {
	i.cursorX = x
	i.cursorY = y
}

// set by external
func (i *Input) SetScrollOffSet(xoff, yoff float64) {
	i.scrollX = xoff
	i.scrollY = yoff
}

func (i *Input) Wheel() (float64, float64) {
	return i.scrollX, i.scrollY
}

// Update must be called from the main thread.
// not call yet.
// nothing to do.
// 主循环中update, 更新按键状态...
func (i *Input) Update() error {
	if i.keyPressed == nil {
		i.keyPressed = map[Key]bool{}
	}

	i.m.Lock()
	defer i.m.Unlock()

	for key:=KeyA; key < KeyMax; key++ {
		i.keyPressed[key] = false
	}

	// Update get each valid pressed key
	for i.keyBuffer.Len() > 0 {
		gk := i.keyBuffer.PopFront().(Key)
		if gk < KeyMax {
			i.keyPressed[gk] = true
		}
	}

	// gamepad.Update()
	return nil
}
