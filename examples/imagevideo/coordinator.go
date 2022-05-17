package main

import (
	"common/rtpengine"
	"common/rtpengine/rtp"
	"common/rtpengine/rtp/codecs"
	"common/web"
	"encoding/json"
	"flag"
	"github.com/pterm/pterm"
	"net"
	"os"
	"strings"
	"time"
	"xmediaEmu/pkg/cws"
	"xmediaEmu/pkg/cws/entity"
	"xmediaEmu/pkg/emulator/rtpua"
	"xmediaEmu/pkg/log"
	"xmediaEmu/pkg/media/h264writer"
)

//
const (
	NThreads         = 100
	CircuitPerThread = 100
)

var (
	Url                     = ""
	FileNames               = ""
	CurrentNThreads         = NThreads
	CurrentCircuitPerThread = CircuitPerThread

	IpLocalAddr = "10.153.90.4:16666"
)

func init() {
	url := flag.String("url", "", "The address for test")
	fileNames := flag.String("f", "", "The filename list")
	nThreads := flag.Int("t", NThreads, "The num of thread")
	circuitPerThread := flag.Int("c", CircuitPerThread, "The num of circuit for per thread")
	flag.Parse()

	if url != nil {
		Url = *url
	}
	if fileNames != nil {
		FileNames = *fileNames
	}
	if nThreads != nil && *nThreads > 0 {
		CurrentNThreads = *nThreads
	}
	if circuitPerThread != nil && *circuitPerThread > 0 {
		CurrentCircuitPerThread = *circuitPerThread
	}
}

// 发送byte.
type WsSender struct {
	url       string
	wsClient  *web.WSocket
	cwsClient *cws.Client
	stop      chan bool
}

func Connect(url string, stop chan bool) *WsSender {
	ws := &WsSender{url: url, wsClient: web.New(url)}
	ws.wsClient.OnConnected = func(socket web.WSocket) {
		pterm.FgWhite.Printfln("Connected to %q server", url)
	}
	//ws.wsClient.OnDisconnected = func(err error, socket web.WSocket) {
	//	pterm.FgWhite.Printfln("Disconnected from %q server", url)
	//}

	ws.stop = stop
	ws.cwsClient = cws.NewClient("testWorkServer", ws.wsClient)
	// 注册回调函数处理响应.
	ws.cwsClient.Receive(entity.RoomStarted, ws.OnHandleRoomStart())
	// ws.cwsClient.Receive(entity.RegisterRoom, ws.OnHandleRoomStart())

	// 开启异步接收协程
	err := ws.cwsClient.Connect()
	if !ws.wsClient.IsConnected {
		pterm.FgRed.Printfln("The ws url %q connect fail:%q", url, err)
		os.Exit(-1)
	}
	return ws
}

func (ws *WsSender) OnHandleRoomClose() cws.PacketHandler {
	return func(resp cws.WSPacket) (req cws.WSPacket) {
		pterm.FgRed.Println("OnHandleRoomClose done")

		// 不用回啥响应
		return cws.EmptyPacket
	}
}

func (ws *WsSender) Send(command, data string) error {
	// callback must be nil
	ws.cwsClient.Send(cws.WSPacket{ID: command, Data: data, PacketID: "1", SessionID: "123456890117"}, nil)
	return nil
}

func (ws *WsSender) Close() {
	// ws.wsClient.SendBinary(websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	ws.wsClient.Close()
}

// 处理roomStart响应
func (ws *WsSender) OnHandleRoomStart() cws.PacketHandler {
	return func(resp cws.WSPacket) (req cws.WSPacket) {
		// resp返回，获取sdp.
		rom := entity.RoomStartRsp{}
		if err := rom.From(resp.Data); err != nil {
			pterm.FgRed.Printfln("OnHandleRoomStart rom.From error: %v", err)
			return cws.EmptyPacket
		}
		pterm.FgWhite.Printfln("OnHandleRoomStart response id: %s sdp: %q", rom.RoomId, rom.Sdp)
		var sdp rtpua.Sdp
		if err := json.Unmarshal([]byte(rom.Sdp), &sdp); err != nil {
			pterm.FgRed.Printfln("OnHandleRoomStart json.Unmarshal error: %v", err)
			return cws.EmptyPacket
		}
		_ = ws.startVideoRtpReceive(sdp)

		// 不用回啥响应
		return cws.EmptyPacket
	}
}

// 接收视频video流10s.
func (ws *WsSender) startVideoRtpReceive(sdp rtpua.Sdp) error {
	ticker := time.NewTicker(time.Second * 40)

	// 创建rtp流.
	localAddr, _ := net.ResolveUDPAddr("udp", IpLocalAddr)
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		pterm.FgRed.Printf("Listen the remote addr fail:%v... \n", err)
		os.Exit(-1)
	}
	_ = conn.SetReadBuffer(5242880)
	defer conn.Close()
	pterm.FgWhite.Printf("startVideoRtpReceive %s\n", localAddr)

	// create h264 file.
	h264File, err := h264writer.New(FileNames)
	if err != nil {
		panic(err)
	}
	defer h264File.Close()
	packet := rtp.Packet{}

	var currentBytes int
	var temp [1500]byte
	for {
		select {
		case <-ticker.C:
			// 20 second 到了停止.
			ws.stop <- true
			return nil
		default:
			// 无限循环.
			// 默认每次读到1200个字节.
			receiveBytes, _, err := conn.ReadFromUDP(temp[0:])
			if err != nil {
				pterm.FgLightRed.Printf("Receive the message fail, the error is %s\n", err.Error())
				return err
			}
			if err := packet.Unmarshal(temp[0:receiveBytes]); err != nil {
				pterm.FgLightRed.Printf("receiveLocalMessage: Unmarshal failed:%v\n. ", err)
			} else {
				if err = h264File.WriteRTP(&packet); err != nil {
					pterm.FgLightRed.Printf("receiveLocalMessage: WriteRTP failed:%v\n. ", err)
				} else {
					pterm.FgWhite.Printf("hWriter.WriteFile count:%d done. \n", receiveBytes)
				}
			}

			// 接收完decode.
			currentBytes += receiveBytes
			pterm.FgWhite.Printf("Receive the message success, the total length is %d\n", currentBytes)
		}
	}
	return nil
}

// 采用rtptrack接收rtp
func (ws *WsSender) startVideoRtpTrack(sdp rtpua.Sdp) error {
	ticker := time.NewTicker(time.Second * 20)
	// create h264 file.
	h264File, err := h264writer.New(FileNames)
	if err != nil {
		panic(err)
	}
	defer h264File.Close()

	// 用track进行接收====================================.
	trackLocal := rtpengine.NewTrackLocal("testTrack4321", IpLocalAddr)
	binds := rtpengine.SenderBinding{RightAddr: "127.0.0.1:6440", PayloadType: 101, Payloader: &codecs.H264Payloader{}}
	trackLocal.Bind(binds)
	trackLocal.SetSamples(16000)
	if err := trackLocal.StartSession(rtpengine.Config{}); err != nil {
		pterm.FgLightRed.Printf("StartSession failed:%v\n. ", err)
		return err
	}

	r, ssrc, err := trackLocal.AcceptStream()
	if err != nil {
		pterm.FgLightRed.Printf("AcceptStream failed:%v\n. ", err)
		return err
	}
	pterm.FgWhite.Printf("AcceptStream successed, ssrc is:%d\n. ", ssrc)

	// start to receive bytes.
	// defer localConn.Close()
	var currentBytes int
	var temp [1500]byte

	//enc, err := h264.NewEncoder(640, 480, h264.Preset("fast"), h264.LogLevel(3), h264.Crf(20), h264.Tune("stillimage"), h264.Profile("baseline"))
	//if err != nil {
	//	panic(err)
	//}
	//defer enc.Shutdown()

	// 直接写入sps和pps.
	//h264Heads := enc.EncodeHeader()
	//if err := h264File.WriteFile(h264Heads); err != nil {
	//	pterm.FgLightRed.Printfln("h264File.WriteFile failed:%v\n", err)
	//} else {
	//	pterm.FgWhite.Printf("EncodeHeader write header length:%d done. \n", len(h264Heads))
	//}
	for {
		select {
		case <-ticker.C:
			// 20 second 到了停止.
			ws.stop <- true
			return nil
		default:
			// 无限循环.
			// 默认每次读到1200个字节.
			// 无限循环.
			// 默认每次读到1200个字节.
			_ = r.SetReadDeadline(time.Now().Add(time.Second))
			packet, err := r.ReadRTP(temp[0:])
			if err != nil {
				if strings.Contains(err.Error(), "timeout") {
					// normal read done
					pterm.FgRed.Printf("Timeout, the error is %s", err.Error())
					return err
				} else {
					pterm.FgRed.Printf("Receive the message fail, the error is %s", err.Error())
					return err
				}
			}
			if err = h264File.WriteRTP(packet); err != nil {
				pterm.FgLightRed.Printf("receiveLocalMessage: WriteRTP failed:%v\n. ", err)
			} else {
				pterm.FgWhite.Printf("hWriter.WriteFile Payload count:%d done. \n", len(packet.Payload))
			}
			currentBytes += len(packet.Payload)
			pterm.FgWhite.Printf("Receive the message success, the current lenth is %d\n", currentBytes)
		}
	}
}

// 模拟xmedia分发到client.
// 转成h264流后给到ffmpeg读取播放.
// act as ws client.
func main() {
	log.Init("1", log.Config{Level: int32(0), Dir: "./", Path: "coordinator.log", FileNum: 10})

	if Url == "" {
		Url = "ws://10.153.90.4:9999/room"
	}

	if FileNames == "" {
		FileNames = "coordinator.h264"
	}
	done := make(chan bool, 1)
	sender := Connect(Url, done)
	defer sender.Close()
	// waitGroup := sync.WaitGroup{}

	// 先测试一个文件
	//fileNameList := strings.Split(FileNames, ",")
	//for _, fileName := range fileNameList {
	//file, err := os.Create(FileNames)
	//if err != nil {
	//	pterm.FgRed.Printfln("The file %q create failed:%v\n", FileNames, err)
	//	os.Exit(-2)
	//}
	//defer file.Close()
	//if err != nil {
	//	pterm.FgRed.Println("obtainFileContent fail")
	//	os.Exit(-1)
	//}
	start := time.Now()
	pterm.FgWhite.Printfln("The %q file start receiving from ws\n", FileNames)
	//for i := 0; i < conf.CurrentNThreads; i++ {
	//	waitGroup.Add(1)
	//	go func() {
	//		defer waitGroup.Done()
	//		for j := 0; j < conf.CurrentCircuitPerThread; j++ {
	//			ws.Send(content)
	//		}
	//	}()
	//}
	//waitGroup.Wait()
	// 创建房间交互 Begin:================================================================================================
	// text or image or chromedp
	rom := entity.RoomStartCall{Name: "chromedp", AudioPayloadType: 97, VideoPayloadType: 101, Zone: "udp", Addr: IpLocalAddr}
	data, err := rom.To()
	if err = sender.Send(entity.RoomStart, data); err != nil {
		pterm.FgRed.Printfln("sender.Send fail: %v", err)
		os.Exit(-3)
	}
	// 创建房间交互 End:================================================================================================
	<-done
	pterm.FgLightYellow.Printfln("The %s file end sending, result is %v, duration is %f\n", FileNames, err, time.Since(start).Seconds())
	time.Sleep(time.Second * 2)
	// }
}
