package worker

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"xmediaEmu/pkg/emulator/config"
	"xmediaEmu/pkg/emulator/libretro"
	"xmediaEmu/pkg/emulator/rtpua"
	// "xmediaEmu/pkg/emulator/run"
	"xmediaEmu/pkg/encoder"
	"xmediaEmu/pkg/log"
)

const (
	bufSize        = 245969 // 缓存游戏输出的视频语音帧.
	SocketAddrTmpl = "/tmp/cloudretro-retro-%s.sock"
)

// Room is a game session. multi webRTC sessions can connect to a same game.
// A room stores all the channel for interaction between all webRTCs session and emulator
type Room struct {
	ID string

	// imageChannel is image stream received from director
	imageChannel <-chan libretro.GameFrame
	// audioChannel is audio stream received from director
	audioChannel <-chan []int16

	// inputChannel is input stream send to director.
	inputChannel chan<- libretro.InputEvent

	// voiceInChannel is voice stream received from users
	// voiceInChannel chan []byte
	// voiceOutChannel is voice stream routed to all users
	// voiceOutChannel chan []byte
	// voiceSample     [][]byte
	// State of room
	IsRunning bool
	// Done channel is to fire exit event when room is closed
	Done chan struct{}

	// List of peer connections in the room
	// ws流还是纯rtp流？...
	rtcSessions []*rtpua.RtpUa

	// NOTE: Not in use, lock rtcSessions
	// sessionsLock *sync.Mutex

	// 先直接引用.
	director *libretro.NaEmulator

	vPipe *encoder.VideoPipe
}

// TODO:

// NewVideoImporter return image Channel from stream
// 从游戏后台不停接收数据好发送.
// TODO: 通过channel接收：...中间层跳过直接游戏原始channel不更好?.
func NewVideoImporter(roomID string) chan libretro.GameFrame {
	sockAddr := fmt.Sprintf(SocketAddrTmpl, roomID)
	imgChan := make(chan libretro.GameFrame)

	l, err := net.Listen("unix", sockAddr)
	if err != nil {
		log.Logger.Fatal("listen error:", err)
	}

	log.Logger.Info("Creating uds server", sockAddr)
	go func(l net.Listener) {
		defer l.Close()

		conn, err := l.Accept()
		if err != nil {
			log.Logger.Fatal("Accept error: ", err)
		}
		defer conn.Close()
		log.Logger.Info("Received new conn: Spawn Importer: ")

		// 这种方式效率低下.
		fullBuf := make([]byte, bufSize*2)
		fullBuf = fullBuf[:0]
		for {
			// TODO: Not reallocate
			// 参考:https://www.cnblogs.com/sparkdev/p/10704614.html.
			// length and capability = bufSize
			buf := make([]byte, bufSize)
			l, err := conn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Logger.Errorf("error: %v", err)
				}
				continue
			}

			buf = buf[:l]
			fullBuf = append(fullBuf, buf...)
			if len(fullBuf) >= bufSize { // 重新分配：
				buff := bytes.NewBuffer(fullBuf)
				dec := gob.NewDecoder(buff)

				frame := libretro.GameFrame{}
				err := dec.Decode(&frame)
				if err != nil {
					log.Logger.Errorf("%v", err)
				}
				imgChan <- frame
				fullBuf = fullBuf[bufSize:]
			}
		}
	}(l)

	return imgChan
}

// new room

// NewRoom creates a new room
// 目前直接根据应用名字加载对应的game,直接采用注册机制实现.
func NewRoom(roomID string, gameName string, bUseUnixSocket bool, config config.Config) *Room {
	if roomID == "" {
		roomID = GenerateRoomID(gameName)
	}

	// 创建对应的director和UserInterface.
	log.Logger.Infof("New room: %s, gameName:%s", roomID, gameName)
	inputChannel := make(chan libretro.InputEvent, 100)

	room := &Room{
		ID: roomID,

		inputChannel: inputChannel,
		imageChannel: nil,
		//voiceInChannel:  make(chan []byte, 1),
		//voiceOutChannel: make(chan []byte, 1),
		rtcSessions: []*rtpua.RtpUa{},
		IsRunning:   true,

		Done: make(chan struct{}, 1),
	}

	// Check if room is on local storage, if not, pull from GCS to local storage
	go func(game, roomID string) {
		// 逻辑不存.
		//store := nanoarch.Storage{
		//	Path:     cfg.Emulator.Storage,
		//	MainSave: roomID,
		//}

		// Check room is on local or fetch from server
		// If not then load room or create room from local.
		log.Logger.Infof("Room %s started. GameName: %s, bUseUnixSocket: %t", roomID, game, bUseUnixSocket)

		// Spawn new emulator and plug-in all channels
		gameHandleName := game

		// TODO: run.Get()用的是默认输入管理.
		libretro.Get().SetWindowSize(config.Width, config.Height)
		libretro.Get().SetWindowTitle(gameHandleName)

		// 创建userInterface.
		if bUseUnixSocket {
			// Run without game, image stream is communicated over a unix socket
			imageChannel := NewVideoImporter(roomID)
			director, _, audioChannel := libretro.New(roomID, false, inputChannel, libretro.Get())
			room.imageChannel = imageChannel
			room.director = director
			room.audioChannel = audioChannel
		} else {
			// Run without game, image stream is communicated over image channel
			director, imageChannel, audioChannel := libretro.New(roomID, true, inputChannel, libretro.Get())
			room.imageChannel = imageChannel
			room.director = director
			room.audioChannel = audioChannel
		}

		// gameMeta := room.director.LoadMeta(filepath.Join(game.Base, game.Path))
		log.Logger.Infof("Viewport custom size is disabled, base size will be used instead %dx%d", config.Width, config.Height)

		// set game frame size considering its orientation
		//encoderW, encoderH := nwidth, nheight
		//if gameMeta.Rotation.IsEven {
		//	encoderW, encoderH = nheight, nwidth
		//}

		room.director.SetViewport(config.Width, config.Height)

		// Spawn video and audio encoding for rtp
		go room.startVideo(config.Width, config.Height, config.Encoder.Video)

		// TODO audio: 711 or amr.
		// go room.startAudio(8000, cfg.Encoder.Audio)
		//go room.startVoice()
		room.director.Start()
	}(gameName, roomID)
	return room
}

func (r *Room) IsRunningSessions() bool {
	// If there is running session
	for _, s := range r.rtcSessions {
		if s.IsConnected() {
			return true
		}
	}

	return false
}

// TODO: Reuse for remove Session
func (r *Room) IsPCInRoom(w *rtpua.RtpUa) bool {
	if r == nil {
		return false
	}
	for _, s := range r.rtcSessions {
		if s.ID == w.ID {
			return true
		}
	}
	return false
}

func (r *Room) UpdatePlayerIndex(peerconnection *rtpua.RtpUa, playerIndex int) {
	log.Logger.Info("Updated player Index to: ", playerIndex)
	peerconnection.PlayerIndex = playerIndex
}

func (r *Room) AddConnectionToRoom(peerconnection *rtpua.RtpUa) {
	peerconnection.AttachRoomID(r.ID)
	r.rtcSessions = append(r.rtcSessions, peerconnection)

	go r.startRtpSession(peerconnection)
}

// 开启rtp session.
func (r *Room) startRtpSession(peerconnection *rtpua.RtpUa) {
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Fatal("Warn: Recovered when sent to close inputChannel")
		}
	}()

	log.Logger.Info("Start Rtp session")

	// bug: when input channel here = nil, skip and finish
	for input := range peerconnection.InputChannel {
		// NOTE: when room is no longer running. InputChannel needs to have extra event to go inside the loop
		if peerconnection.Done || !peerconnection.IsConnected() || !r.IsRunning {
			break
		}

		if peerconnection.IsConnected() {
			select {
			case r.inputChannel <- libretro.InputEvent{Raw: input, PlayerIdx: peerconnection.PlayerIndex, ConnID: peerconnection.ID}:
			default:
			}
		}
	}
	log.Logger.Info("[worker] peer connection is done")
}

func (r *Room) IsEmpty() bool { return len(r.rtcSessions) == 0 }

// RemoveSession removes a peerconnection from room and return true if there is no more room
func (r *Room) RemoveSession(w *rtpua.RtpUa) {
	log.Logger.Info("Cleaning session: ", w.ID)
	// TODO: get list of r.rtcSessions in lock
	for i, s := range r.rtcSessions {
		log.Logger.Info("found session: ", w.ID)
		if s.ID == w.ID {
			r.rtcSessions = append(r.rtcSessions[:i], r.rtcSessions[i+1:]...)
			s.RoomID = ""
			log.Logger.Info("Removed session ", s.ID, " from room: ", r.ID)
			break
		}
	}
	// Detach input. Send end signal
	select {
	// 关闭byte.
	case r.inputChannel <- libretro.InputEvent{Raw: []byte{0xFF, 0xFF}, ConnID: w.ID}:
	default:
	}
}

func (r *Room) Close() {
	if !r.IsRunning {
		return
	}

	r.IsRunning = false
	log.Logger.Info("Closing room and director of room ", r.ID)
	r.director.Close()
	log.Logger.Info("Closing input of room ", r.ID)
	close(r.inputChannel)
	//close(r.voiceOutChannel)
	//close(r.voiceInChannel)
	close(r.Done)
	// Close here is a bit wrong because this read channel
	// Just dont close it, let it be gc
	//close(r.imageChannel)
	//close(r.audioChannel)
}
