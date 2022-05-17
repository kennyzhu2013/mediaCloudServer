package inpututil

import (
	"common/util/mem"
	"sync"
)

// for all base inputs.
type InputSequence struct {
	stringsInput         []string // 待优化，slice固定内存块存储string.
	// input list.
	// input channels.
	seqBuffer *mem.Deque
	m sync.RWMutex

	//
}


// NewInput generates a new Input object.
func NewInputSequence() *InputSequence {
	return &InputSequence{ seqBuffer: mem.New(_maxKeyBuffs), stringsInput: make([]string, 0, _maxKeyBuffs)}
}

// Block until chan is nil.
func (i *InputSequence) InputKeySeq(seq interface{})  {
	i.m.Lock()
	i.seqBuffer.PushBack(seq)
	i.m.Unlock()
}

// Block until chan is nil.
func (i *InputSequence) GetSeq()  interface{} {
	i.m.Lock()
	defer i.m.Unlock()

	if i.seqBuffer.Len() <= 0 {
		return nil
	}

	return i.seqBuffer.PopFront()
}

// AppendJustPressedTouchIDs is concurrent safe.
func (i *InputSequence) GetLastInputString() string {
	if i.seqBuffer.Len() < 1 {
		return ""
	}
	return i.seqBuffer.Front().(string)
}

// AppendJustPressedTouchIDs is concurrent safe.
func (i *InputSequence) GetFirstInputString() string {
	if i.seqBuffer.Len() < 1 {
		return ""
	}
	return i.seqBuffer.Back().(string)
}


// 主循环中update, 更新按键状态...
func (i *InputSequence) Update() error {
	i.m.Lock()
	defer i.m.Unlock()

	i.stringsInput = []string{}

	// Update get each valid pressed key
	for i.seqBuffer.Len() > 0 {
		str := i.seqBuffer.PopFront().(string)
		if str != "" {
			i.stringsInput = append(i.stringsInput, str)
		}
	}

	// gamepad.Update()
	return nil
}


// TouchPressDuration returns how long the touch remains in frames.
//
// TouchPressDuration is concurrent safe.
func (i *InputSequence) InputStrings() []string {
	return i.stringsInput
}
