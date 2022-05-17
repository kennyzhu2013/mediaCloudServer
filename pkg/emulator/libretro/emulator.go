package libretro

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"sync"
	"xmediaEmu/pkg/emulator/libretro/games"
	"xmediaEmu/pkg/log"
)

// 防止game.GameForUI指针指示错误.
// type GameForUI game.GameForUI

// 基于模拟器框架实现图片转视频.
// 只做run ui的外包封装.
type NaEmulator struct {
	sync.Mutex

	imageChannel chan<- GameFrame
	audioChannel chan<- []int16

	// 模拟器输入事件..
	// 直接调用.
	inputChannel  <-chan InputEvent
	game          *UserInterface
	videoExporter *VideoExporter

	// meta            emulator.Metadata
	gamePath        string
	roomID          string
	gameName        string
	isSavingLoading bool

	// 每次电话都是无状态的.
	// storage         Storage

	// 支持多方进入（一个player就是一个通话方）..
	// players Players // // 暂时不考虑任何player相关处理.

	// 所有游戏注册入口.
	// gameMap  map [string]*game.GameForUI
	// gameMap  map [string]*GameForUI

	// done chan struct{}
}

// game在外面创建好传入.
// 支持声音输入，可以实时流输出pcm码流和h264流.
func NewNAEmulator(roomID string, inputChannel <-chan InputEvent, game *UserInterface) (*NaEmulator, chan GameFrame, chan []int16) {
	imageChannel := make(chan GameFrame, 30)
	audioChannel := make(chan []int16, 30)

	// 目前只支持视频channel.
	game.SetImageChannel(imageChannel)

	return &NaEmulator{
		game:         game,
		imageChannel: imageChannel,
		audioChannel: audioChannel,
		// players:      NewPlayerSessionInput(),
		inputChannel: inputChannel,
		roomID:       roomID,
		// gameMap:      map [string]*GameForUI{},
		// done:         make(chan struct{}, 1),
	}, imageChannel, audioChannel
}

// NewVideoExporter creates new video Exporter that produces to net or unix socket
// unix套接字单独线程传输h264.
func NewVideoExporter(roomID string, imgChannel chan GameFrame) *VideoExporter {
	// sockAddr地址无效时本地socket导出?.
	network := "unix"
	sockAddr := fmt.Sprintf("/tmp/cloudretro-retro-%s.sock", roomID)

	go func(sockAddr string) {
		log.Logger.Debug("Dialing to ", sockAddr)
		conn, err := net.Dial(network, sockAddr)
		if err != nil {
			log.Logger.Error("accept error: ", err)
		}

		defer conn.Close()

		for img := range imgChannel {
			reqBodyBytes := new(bytes.Buffer)
			_ = gob.NewEncoder(reqBodyBytes).Encode(img)
			//fmt.Printf("%+v %+v %+v \n", img.Image.Stride, img.Image.Rect.Max.X, len(img.Image.Pix))
			// conn.Write(img.Image.Pix)
			b := reqBodyBytes.Bytes()
			log.Logger.Debugf("Bytes %d\n", len(b))
			if _, err := conn.Write(b); err != nil {
				log.Logger.Errorf("NewVideoExporter: conn.Write error %v", err)
			}
		}
	}(sockAddr)

	return &VideoExporter{imageChannel: imgChannel}
}

// 解决输入问题，如文字或按键.
// listenInput handles user input.
// The user input is encoded as bitmap that we decode
// and send into the game emulator.
// 处理当前会话进入按键需求，参考游戏方案的player状态.
func (na *NaEmulator) listenInput() {
	for in := range na.inputChannel {
		// terminate.
		//if in.Raw == InputTerminate {
		//	na.players.session.close(in.ConnID)
		//	continue
		//} else {
		//	bitmap := in.tryBitmap()
		//	if bitmap != 0 {
		//		na.players.session.setInput(in.ConnID, in.PlayerIdx, bitmap, in.Raw.([]byte))
		//	} else {
		// 非player的正常输入,直接发送.
		if err := na.game.GetInputMgr().SendInput(in.Raw); err != nil {
			log.Logger.Error("listenInput error: ", err)
		}
		//	}
		//}
	}
}

// 基于room的形式创建emulator.
func New(roomID string, withImageChannel bool, inputChannel <-chan InputEvent, ui *UserInterface) (*NaEmulator, chan GameFrame, chan []int16) {
	emu, imageChannel, audioChannel := NewNAEmulator(roomID, inputChannel, ui)
	// Set to global NAEmulator
	// NAEmulator = emu
	//if !withImageChannel { // 控制是否产生图片帧到本地unix套接字.
	//	emu.videoExporter = NewVideoExporter(roomID, imageChannel)
	//}

	go emu.listenInput()

	return emu, imageChannel, audioChannel
}

// 重新调整窗口大小..
func (na *NaEmulator) SetViewport(width int, height int) {
	// outputImg is tmp img used for decoding and reuse in encoding flow
	// TODO: 暂时不支持运行中调增大小.
	na.game.SetWindowSize(width, height)
	// outputImg = image.NewRGBA(image.Rect(0, 0, width, height))
}

// 模拟器开始.暂时不支持游戏保存功能.
func (na *NaEmulator) Start() error {
	gameLogic, err := na.LoadGame()
	if err != nil {
		log.Logger.Errorf("error: couldn't load a save, %v", err)
		return err
	}

	// fps 设置问题.
	// 启动game.
	if err := gameLogic.RunGame(na.game); err != nil {
		log.Logger.Errorf("Run game failed: %v.\n ", err)
		return err
	}
	na.Close()
	log.Logger.Info("Closed Director")
	return nil
}

// 所有游戏根据名字在此增加入口.
func (na *NaEmulator) LoadGame() (*GameForUI, error) {
	switch na.game.GetGameName() {
	case "text":
		return NewGameForUI(&games.Game{}), nil
	case "image":
		return NewGameForUI(&games.GameImage{Index: 0}), nil
	case "chromedp": // chromedp应用.
		return NewGameForUI(games.NewGameChromeDp()), nil
	default: // 演示用的.
		return NewGameForUI(&games.Game{}), nil
	}
	// return nil, errors.New("")
}

func (na *NaEmulator) Close() {
	na.game.SetRunning(false)
	close(na.imageChannel)
	close(na.audioChannel)
}

// 获取当前显示图像内容.
func (na *NaEmulator) GetViewport() interface{} {
	return na.game.GetViewPort()
}
