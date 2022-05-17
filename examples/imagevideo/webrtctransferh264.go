package main

import (
	"common/rtpengine"
	"common/rtpengine/media"
	"common/rtpengine/rtp"
	"common/rtpengine/rtp/codecs"
	"flag"
	"fmt"
	"github.com/pterm/pterm"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
	"xmediaEmu/pkg/encoder"
	"xmediaEmu/pkg/encoder/h264"
	"xmediaEmu/pkg/image"
	"xmediaEmu/pkg/log"
	"xmediaEmu/pkg/media/h264writer"
)

var seeds = rand.New(rand.NewSource(time.Now().UnixNano())).Uint32()
var network = "udp"
var localAddress = "10.153.90.4:6666"
var remoteAddress = "10.153.90.4:6440"

// 将图片转移的h264流使用rtp传输.
func main() {
	log.Init("1", log.Config{Level: int32(0), Dir: "./", Path: "playVideo.log", FileNum: 10})

	// 启动网络连接...

	//addr,_ := net.ResolveUDPAddr(network, "10.153.90.4:6666")
	//serverSocket,err := udp.Listen(network, addr)
	//if err != nil { panic(err) }
	//defer serverSocket.Close()
	//
	//// 等待网络连接有才监听数控,此时客户端client udp必须发送数据才行.
	//clientConn, err := serverSocket.Accept()
	//if err != nil { panic(err) }
	//pterm.FgWhite.Printfln("serverSocket.Accept client from:%s.", clientConn.RemoteAddr().String())
	//trackLocal := getLocalTrack(clientConn)
	//if trackLocal == nil { panic("trackLocal is nil") }
	// current := flag.Int("t", defaultConcurrent, "Total Concurrent")
	flag.Parse()
	stop := make(chan interface{})
	// The remote listen the local message： listen and print message
	testTrackLocalReceive(stop) // remoteAddress
	// go receiveLocalMessage(stop)

	trackLocal := rtpengine.NewTrackLocal("testTrack1234", localAddress)
	binds := rtpengine.SenderBinding{RightAddr: remoteAddress, PayloadType: 101, Payloader: &codecs.H264Payloader{}, Options: rtpengine.SenderBindingOptions{Mtu: 1388}} //codecs.PCMAPayLoadLength}}
	trackLocal.Bind(binds)
	trackLocal.SetSamples(8000)
	if err := trackLocal.StartSession(rtpengine.Config{}); err != nil {
		fmt.Printf("StartSession failed:%v\n. ", err)
		return
	}

	const W = 640
	const H = 480
	// log.Init("1", log.Config{Level: int32(0), Dir: "./", Path: "playVideo.log", FileNum: 10})
	enc, err := h264.NewEncoder(W, H, h264.Preset("fast"), h264.LogLevel(3), h264.Crf(20), h264.Tune("stillimage"), h264.Profile("baseline"))
	if err != nil {
		panic(err)
	}
	vPipe := encoder.NewVideoPipe(enc, W, H) // 目的大小和原始大小一致.
	einput, eoutput := vPipe.Input, vPipe.Output

	// 启动异步解码.
	go vPipe.Start()

	done := make(chan bool, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				pterm.FgLightRed.Println("Recovered when sent to close Image Channel\n")
			}
		}()
		// 改成直接发送.
		// 直接写入sps和pps.
		// h264Heads := enc.EncodeHeader()
		startTime := time.Now()
		//dataSamples := media.Sample{Data: h264Heads, PrevDroppedPackets: 0}
		//if err := trackLocal.WriteSample(dataSamples); err != nil {
		//	pterm.FgLightRed.Printfln("h264File.WriteFile failed:%v\n", err)
		//} else {
		//	pterm.FgWhite.Printf("EncodeHeader write header length:%d done. \n", len(h264Heads))
		//}

		// send to every one who connected.
		// fan-out Screen, write file directly.
		count := 0
		for data := range eoutput {
			dataSamples := media.Sample{Data: data.Data, PrevDroppedPackets: 0}
			if err := trackLocal.WriteSample(dataSamples); err != nil {
				pterm.FgLightRed.Printfln("h264File.WriteFile failed:%v\n", err)
			} else {
				pterm.FgWhite.Printf("e-output write image length:%d done. \n", len(data.Data))
			}
			count++
		}
		pterm.FgWhite.Printf("Ending send rtp, concurrent is 1, frame count is %d, duration is %fs\n", count, time.Since(startTime).Seconds())
		done <- true
	}()

	// TODO: 1,没有nalu； 2,第一个循环前面13帧必定失败
	// 采用vPipe进行h264转码.
	for j := 0; j < 2; j++ {
		for i := 1; i < 25; i++ {
			filename := "outdir/" + fmt.Sprintf("%05d.jpg", i)
			im, err := image.LoadJPG(filename)
			if err != nil {
				pterm.FgLightRed.Printf("load file failed:%s\n. err:%v", filename, err)
				panic(err)
			}

			//index := 0
			// im转h264.
			for {
				if len(einput) < cap(einput) {
					// time.Sleep(time.Millisecond * 20)
					einput <- encoder.InFrame{Image: image.ImageToRGBA(im), Timestamp: uint32(time.Now().UnixNano()/8333) + seeds}
					break
					//if index >= 20 {
					//	break
					//}
					//index++
				} else {
					// pterm.FgLightRed.Printf("einput capacity is full\n. ")
					time.Sleep(time.Millisecond * 20)
				}
			}
			pterm.FgWhite.Printf("put image index:%s done. \n", filename)
		}
	}

	for len(einput) > 0 {
		time.Sleep(time.Millisecond * 20)
	}
	pterm.FgLightGreen.Print("vPipe stopped. \n")

	// wait for all done.
	vPipe.Stop()
	<-done
	time.Sleep(time.Second * 2)
}

// 获取连接的客户端.
// use rptua to send h264.
func getLocalTrack(clientConn net.Conn) *rtpengine.TrackLocal {
	// bindings 生成.
	trackLocal := rtpengine.NewTrackLocal("local", clientConn.LocalAddr().String())
	binds := rtpengine.SenderBinding{PayloadType: 101, Payloader: &codecs.G711Payloader{}, Options: rtpengine.SenderBindingOptions{Mtu: codecs.PCMAPayLoadLength}}
	trackLocal.Bind(binds)
	trackLocal.SetSamples(8000)
	if err := trackLocal.StartSession(rtpengine.Config{}); err != nil {
		pterm.FgLightRed.Printf("StartSession failed:%v\n. ", err)
		return nil
	}
	return trackLocal
}

// 直接udp接收
func receiveLocalMessage(stop chan interface{}) {
	remoteAddr, _ := net.ResolveUDPAddr("udp", remoteAddress)
	conn, err := net.ListenUDP("udp", remoteAddr)
	if err != nil {
		pterm.FgLightRed.Printf("Listen the remote addr fail:%v... \n", err)
		os.Exit(-1)
	}
	_ = conn.SetReadBuffer(5242880)
	defer close(stop)
	defer conn.Close()

	hWriter, _ := h264writer.New("receiver.h264")
	defer hWriter.Close()
	packet := rtp.Packet{}

	var currentBytes int
	var temp [1500]byte
	for {
		// 无限循环.
		// 默认每次读到1200个字节.
		receiveBytes, _, err := conn.ReadFromUDP(temp[0:])
		if err != nil {
			fmt.Printf("Receive the message fail, the error is %s\n", err.Error())
			return
		}
		if err := packet.Unmarshal(temp[0:receiveBytes]); err != nil {
			pterm.FgLightRed.Printf("receiveLocalMessage: Unmarshal failed:%v\n. ", err)
		} else {
			if err = hWriter.WriteRTP(&packet); err != nil {
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

// 接收协程, 接收rtp流并写入本地文件.
// 验证h264有效性.
func testTrackLocalReceive(stop chan interface{}) {
	// 用track进行接收.
	trackLocal := rtpengine.NewTrackLocal("testTrack4321", remoteAddress)
	binds := rtpengine.SenderBinding{RightAddr: localAddress, PayloadType: 101, Payloader: &codecs.H264Payloader{}}
	trackLocal.Bind(binds)
	trackLocal.SetSamples(8000)
	if err := trackLocal.StartSession(rtpengine.Config{}); err != nil {
		fmt.Printf("StartSession failed:%v\n. ", err)
		return
	}

	go receiveMessageFromTrack(stop, trackLocal)
}

func receiveMessageFromTrack(stop chan interface{}, trackLocal *rtpengine.TrackLocal) {
	defer close(stop)
	r, ssrc, err := trackLocal.AcceptStream()
	if err != nil {
		fmt.Printf("AcceptStream failed:%v\n. ", err)
		return
	}
	fmt.Printf("AcceptStream successed, ssrc is:%d\n. ", ssrc)
	hWriter, _ := h264writer.New("receiver.h264")
	defer hWriter.Close()

	var currentBytes int
	var temp [1500]byte
	// start to receive.
	for {
		// 无限循环.
		// 默认每次读到1200个字节.
		_ = r.SetReadDeadline(time.Now().Add(time.Second))
		packet, err := r.ReadRTP(temp[0:])
		if err != nil {
			if strings.Contains(err.Error(), "timeout") {
				// normal read done
				pterm.FgRed.Printf("Timeout, the error is %s", err.Error())
				return
			} else {
				pterm.FgRed.Printf("Receive the message fail, the error is %s", err.Error())
				return
			}
		}
		if err = hWriter.WriteRTP(packet); err != nil {
			pterm.FgLightRed.Printf("receiveLocalMessage: WriteRTP failed:%v\n. ", err)
		} else {
			pterm.FgWhite.Printf("hWriter.WriteFile Payload count:%d done. \n", len(packet.Payload))
		}
		currentBytes += len(packet.Payload)
		fmt.Printf("Receive the message success, the current lenth is %d\n", currentBytes)
	}
}
