package rtpua

import (
	"common/rtpengine"
	"common/rtpengine/media"
	"common/rtpengine/rtp/codecs"
	"common/util/process"
	"encoding/json"
	"sync/atomic"
	"time"
	"xmediaEmu/pkg/log"
)

type WebFrame struct {
	Data      []byte
	Timestamp uint32
}

// RtpUa connection, 只考虑单边接收和发送回去。不考虑来回透传.
// rtp 的ua行为.
// rtpua负责发送rtp流，和asr负责接收同一级别.
type RtpUa struct { // = WebRtc.
	ID string // SessionId

	// all for connection
	paddr string

	//
	// singleConnection       *net.UDPConn // incoming connections.TODO:暂时只考虑视频流或者音频流.
	singleTrack *rtpengine.TrackLocal // 只支持一路.目前只做测试只有视频.
	cfg         Config

	globalVideoFrameTimestamp uint32 // TODO:分段剪辑视频时用.
	isConnected               bool   // 目前只支持一端.
	// for yuvI420 image
	ImageChannel chan WebFrame
	AudioChannel chan []byte

	// input event?.
	InputChannel chan []byte // ws 接口输入当做DataChannel,

	Done bool

	RoomID      string
	PlayerIndex int

	audioPayLoad int
	videoPayLoad int
}

type OnIceCallback func(candidate string)

// 为了安全点要base加密.
// Encode encodes the input in base64
func Encode(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	// return base64.StdEncoding.EncodeToString(b), nil
	return string(b), nil
}

// Decode decodes the input from base64
func Decode(in string, obj interface{}) error {
	//b, err := base64.StdEncoding.DecodeString(in)
	//if err != nil {
	//	return err
	//}

	err := json.Unmarshal([]byte(in), obj)
	if err != nil {
		return err
	}

	return nil
}

// NewWebRTC create, 创建RtpUa请求，不支持重协商.
func NewWebRTC(conf Config, aPayload, vPayload int) (*RtpUa, error) {
	if aPayload == 0 {
		aPayload = audioPayload
	}
	if vPayload == 0 {
		vPayload = videoPayload
	}

	w := &RtpUa{
		ID: conf.SessionId,

		ImageChannel: make(chan WebFrame, 30),
		AudioChannel: make(chan []byte, 1),
		//VoiceInChannel:  make(chan []byte, 1),
		//VoiceOutChannel: make(chan []byte, 1),
		InputChannel: make(chan []byte, 100),
		cfg:          conf,
		audioPayLoad: aPayload,
		videoPayLoad: vPayload,
	}
	w.singleTrack = rtpengine.NewTrackLocal(conf.NetWork, conf.UdpAddr)
	log.Logger.Debugf("NewWebRTC:%s %s", conf.NetWork, conf.UdpAddr)

	// w.laddr = *conf.UdpAddr
	// w.network = conf.NetWork
	// conn, err := net.DialUDP(conf.NetWork, conf.UdpAddr)
	//if err != nil {
	//	return nil, err
	//}
	//w.listen = conn
	return w, nil
}

// StartClient start rtp ua.
// accept some clients.
// 不支持媒体协商.
// rtp等待开始连接，返回连接地址, 等待对端地址带过来...
// 对端ip和端口发送给对端.
//
func (w *RtpUa) StartClient(peerAddr string) (string, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Logger.Error(err)
			w.StopClient()
		}
	}()
	var err error
	w.paddr = peerAddr

	// reset client
	if w.isConnected {
		w.StopClient()
		time.Sleep(2 * time.Second)
	}
	log.Logger.Debug("=== RtpUa: StartClient ===")

	// 编解码, 这里以h264视频为例.
	binds := rtpengine.SenderBinding{RightAddr: peerAddr, PayloadType: rtpengine.PayloadType(w.videoPayLoad), Payloader: &codecs.H264Payloader{}}
	w.singleTrack.Bind(binds)
	w.singleTrack.SetFrequence(30)  // 视频30一帧？..
	w.singleTrack.SetSamples(90000) // 视频采样例90k.

	// 连接打通接口，复用rtp端口，音视频都用同一个接口.
	// 使用默认缓存.
	if err := w.singleTrack.StartSession(rtpengine.Config{}); err != nil {
		log.Logger.Debugf("StartClient: StartSession failed:%v", err)
		return "", err
	}
	w.isConnected = true

	// 不断输入的指令在ws接口解决.
	// add audio rtp connection.
	//opusTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "game-audio")
	//if err != nil {
	//	return "", err
	//}
	//_, err = w.connection.AddTrack(opusTrack)
	//if err != nil {
	//	return "", err
	//}
	//
	////_, err = w.connection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RtpTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly})
	//
	//// create data channel for input, and register callbacks
	//// order: true, negotiated: false, id: random
	//inputTrack, err := w.connection.CreateDataChannel("game-input", nil)
	//if err != nil {
	//	return "", err
	//}
	//
	//inputTrack.OnOpen(func() {
	//	log.Printf("Data channel '%s'-'%d' open.\n", inputTrack.Label(), inputTrack.ID())
	//})
	//
	//// Register text message handling
	//inputTrack.OnMessage(func(msg webrtc.DataChannelMessage) {
	//	// TODO: Can add recover here
	//	w.InputChannel <- msg.Data
	//})

	// RtpUa state callback
	//w.connection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
	//	log.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	//	if connectionState == webrtc.ICEConnectionStateConnected {
	//		go func() {
	//			w.isConnected = true
	//			log.Println("ConnectionStateConnected")
	//			w.startStreaming(videoTrack, opusTrack)
	//		}()
	//
	//	}
	//	if connectionState == webrtc.ICEConnectionStateFailed || connectionState == webrtc.ICEConnectionStateClosed || connectionState == webrtc.ICEConnectionStateDisconnected {
	//		w.StopClient()
	//	}
	//})
	w.startVideoOrAudioStreaming(true)

	offer := Sdp{Address: w.cfg.UdpAddr, AudioPayLoad: w.audioPayLoad, VideoPayLoad: w.videoPayLoad}
	localSession, err := Encode(offer)
	if err != nil {
		log.Logger.Debugf("StartClient: Sdp Encode failed:%v", err)
		return "", err
	}

	return localSession, nil
}

// 直接返回h264.
//func (w *RtpUa) getVideoCodec() string {
//	switch w.cfg.Encoder.Video.Codec {
//	case string(encoder.H264):
//		return webrtc.MimeTypeH264
//	case string(codec.VPX):
//		return webrtc.MimeTypeVP8
//	default:
//		return webrtc.MimeTypeH264
//	}
//}
func (w *RtpUa) AttachRoomID(roomID string) {
	w.RoomID = roomID
}

func (w *RtpUa) GetVideoCodec() string {
	return w.cfg.Encoder.Video.Codec
}

func (w *RtpUa) GetAudioCodec() string {
	return w.cfg.Encoder.Audio.Codec
}

// StopClient disconnect
// 手动拆除rtp连接.
func (w *RtpUa) StopClient() {
	// if stopped, bypass
	if !w.isConnected {
		return
	}

	w.isConnected = false
	if w.singleTrack != nil {
		if err := w.singleTrack.Close(); err != nil {
			log.Logger.Errorf("error: couldn't close RtpUa connection, %v", err)
		}
	}
	w.singleTrack = nil
	//close(w.InputChannel)
	// webrtc is producer, so we close
	// NOTE: ImageChannel is waiting for input. Close in writer is not correct for this
	close(w.ImageChannel)
	close(w.AudioChannel)
	//close(w.VoiceInChannel)
	//close(w.VoiceOutChannel)
	log.Logger.Debug("===StopClient===")
}

func (w *RtpUa) IsConnected() bool { return w.isConnected }

// 音视频发送, TODO: 多路复用一个连接.
func (w *RtpUa) startVideoOrAudioStreaming(isVideo bool) {
	log.Logger.Debug("Start streaming")
	// receive frame buffer
	if isVideo {
		go func() {
			defer func() {
				if v := recover(); v != nil {
					process.DefaultPanicReport.RecoverFromPanic(w.ID, "RtpUa:startStreaming", v)
				}
			}()

			// TODO:为啥channel没传输到?..
			for data := range w.ImageChannel {
				atomic.StoreUint32(&w.globalVideoFrameTimestamp, data.Timestamp)
				if err := w.singleTrack.WriteSample(media.Sample{Data: data.Data}); err != nil {
					w.StopClient()
					log.Logger.Error("WriteSample: Err write sample: ", err)
					break
				} else {
					log.Logger.Debugf("WriteSample done, length: %d", len(data.Data))
				}
			}
		}()
	} else {
		// send audio
		go func() {
			defer func() {
				if v := recover(); v != nil {
					process.DefaultPanicReport.RecoverFromPanic(w.ID, "RtpUa:startStreaming", v)
				}
			}()

			// audioDuration := time.Duration(w.cfg.Encoder.Audio.Frame) * time.Millisecond
			for data := range w.AudioChannel {
				if !w.isConnected {
					return
				}
				err := w.singleTrack.WriteSample(media.Sample{Data: data})
				if err != nil {
					log.Logger.Error("Warn: Err write sample: ", err)
				}
			}
		}()
	}

	//// send voice
	//go func() {
	//	defer func() {
	//		if r := recover(); r != nil {
	//			fmt.Println("Recovered from err", r)
	//			log.Println(debug.Stack())
	//		}
	//	}()
	//
	//	for data := range w.VoiceOutChannel {
	//		if !w.isConnected {
	//			return
	//		}
	//		// !to pass duration from the input
	//		err := opusTrack.WriteSample(media.Sample{Data: data})
	//		if err != nil {
	//			log.Println("Warn: Err write sample: ", err)
	//		}
	//	}
	//}()
}
