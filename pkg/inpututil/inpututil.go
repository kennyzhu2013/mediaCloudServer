package inpututil

import (
	"errors"
	"sync"
	"xmediaEmu/pkg/hooks"
)

// 支持整数，字符串和结构体输入.
// 按键时长统计，
type InputManager struct {
	fpsMode FPSModeType // 输入模式，默认只允许输入整数.

	// key采用整数, 只有模拟按键有持续时间.
	keyDurations     []int
	prevKeyDurations []int

	// for all basic input
	input *Input

	// 字符串和struct输入.
	stringInput *InputSequence

	// struct输入, TODO:暂时用不到.
	structInput *InputSequence

	m sync.RWMutex // 所有输入的锁.
}

// input管理默认只要有一个.
var theInputState = NewInputMgr()

func init() {
	// 自动开启？..
	hooks.AppendHookOnBeforeUpdate(func() error {
		theInputState.Update()
		return nil
	})
}

// DefaultMgr
func DefaultMgr() *InputManager {
	return theInputState
}

func NewInputMgr() *InputManager {
	return &InputManager{
		keyDurations:     make([]int, KeyMax+1),
		prevKeyDurations: make([]int, KeyMax+1),
		fpsMode:          FPSIntOnly, // 默认只接受整数序列输入.

		input:       NewInput(),
		stringInput: NewInputSequence(),
		structInput: NewInputSequence(),
	}
}

func (i *InputManager) GetInput() *Input {
	return i.input
}

func (i *InputManager) GetStringInput() *InputSequence {
	return i.stringInput
}

func (i *InputManager) SetFpsMode(mode FPSModeType) {
	i.fpsMode = mode
}

// 统一接口.
func (i *InputManager) SendInput(input interface{}) error {
	switch input.(type) {
	case int:
		i.input.InputKey(input.(Key))
	case string:
		i.stringInput.InputKeySeq(input)
	case struct{}:
		i.structInput.InputKeySeq(input)
	default:
		return errors.New("Unknown inputs. ")
	}
	return nil
}

// all input, run before Update.
func (i *InputManager) Update() error {
	i.m.Lock()
	defer i.m.Unlock()

	if err := i.input.Update(); err != nil {
		return err
	}
	if i.fpsMode > FPSIntOnly {
		if err := i.stringInput.Update(); err != nil {
			return err
		}
		if i.fpsMode == FPSAll {
			if err := i.structInput.Update(); err != nil {
				return err
			}
		}
	}

	// Keyboard
	copy(i.prevKeyDurations[:], i.keyDurations[:])
	for k := Key(0); k <= KeyMax; k++ {
		if i.input.IsKeyPressed(k) {
			i.keyDurations[k]++
		} else {
			i.keyDurations[k] = 0
		}
	}

	// string and struct?..
	return nil
}

// AppendPressedKeys append currently pressed keyboard keys to keys and returns the extended buffer.
// Giving a slice that already has enough capacity works efficiently.
// AppendPressedKeys is concurrent safe.
func (i *InputManager) AppendPressedKeys(keys []Key) []Key {
	i.m.RLock()
	defer i.m.RUnlock()

	for index, d := range i.keyDurations {
		if d == 0 {
			continue
		}
		keys = append(keys, Key(index))
	}
	return keys
}

// IsKeyJustPressed returns a boolean value indicating
// whether the given key is pressed just in the current frame.
//
// IsKeyJustPressed is concurrent safe.
func (i *InputManager) IsKeyJustPressed(key Key) bool {
	return i.KeyPressDuration(key) == 1
}

// IsKeyJustReleased returns a boolean value indicating
// whether the given key is released just in the current frame.
//
// IsKeyJustReleased is concurrent safe.
func (i *InputManager) IsKeyJustReleased(key Key) bool {
	i.m.RLock()
	r := i.keyDurations[key] == 0 && i.prevKeyDurations[key] > 0
	i.m.RUnlock()
	return r
}

// KeyPressDuration returns how long the key is pressed in frames.
//
// KeyPressDuration is concurrent safe.
func (i *InputManager) KeyPressDuration(key Key) int {
	i.m.RLock()
	s := i.keyDurations[key]
	i.m.RUnlock()
	return s
}

// AppendJustPressedTouchIDs append touch IDs that are created just in the current frame to touchIDs,
// and returns the extended buffer.
// Giving a slice that already has enough capacity works efficiently.
//
