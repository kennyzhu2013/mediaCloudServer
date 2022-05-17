package libretro

import "sync"

////Deprecated
// 暂时不考虑任何player相关处理.
const (
	// how many axes on the D-pad
	dpadAxesNum = 4
	// the upper limit on how many controllers (players)
	// are possible for one play session (emulator instance)
	// 所有玩家最大数量..
	controllersNum = 8
)

const (
	InputTerminate = 0xFFFF
)

//Deprecated
type Players struct {
	session playerSession
}

//Deprecated
type playerSession struct {
	sync.RWMutex

	state map[string][]controllerState
}

type controllerState struct {
	keyState uint16             // 按键.
	axes     [dpadAxesNum]int16 // 4个控制按钮.
}

func NewPlayerSessionInput() Players {
	return Players{
		session: playerSession{
			state: map[string][]controllerState{},
		},
	}
}

// close terminates user input session.
func (ps *playerSession) close(id string) {
	ps.Lock()
	defer ps.Unlock()

	delete(ps.state, id)
}

//Deprecated
func (ps *playerSession) setInput(id string, player int, buttons uint16, dpad []byte) {
	ps.Lock()
	defer ps.Unlock()

	if _, ok := ps.state[id]; !ok {
		ps.state[id] = make([]controllerState, controllersNum)
	}

	ps.state[id][player].keyState = buttons
	for i, axes := 0, len(dpad); i < dpadAxesNum && (i+1)*2+1 < axes; i++ {
		axis := (i + 1) * 2
		ps.state[id][player].axes[i] = int16(dpad[axis+1])<<8 + int16(dpad[axis])
	}
}

//Deprecated 游戏回调接口.
func (p *Players) isKeyPressed(player uint, key int) (pressed bool) {
	p.session.RLock()
	defer p.session.RUnlock()

	for k := range p.session.state {
		if ((p.session.state[k][player].keyState >> uint(key)) & 1) == 1 {
			return true
		}
	}
	return
}

//Deprecated ： 游戏回调接口.
func (p *Players) isDpadTouched(player uint, axis uint) (shift int16) {
	p.session.RLock()
	defer p.session.RUnlock()

	for k := range p.session.state {
		value := p.session.state[k][player].axes[axis]
		if value != 0 {
			return value
		}
	}
	return
}

// 输入事件: 字符串...
type InputEvent struct {
	Raw       interface{} //
	PlayerIdx int         // 玩家索引.
	ConnID    string      // sessionid
}

// 游戏按键.
func (ie InputEvent) tryBitmap() uint16 {
	switch ie.Raw.(type) {
	case []byte:
		bitmap := ie.Raw.([]byte)
		return uint16(bitmap[1])<<8 + uint16(bitmap[0])
	default:
		return 0
	}
}
