package main

import (
	"fmt"
	"github.com/pterm/pterm"
	"math/rand"
	"time"
	"xmediaEmu/pkg/encoder"
	"xmediaEmu/pkg/encoder/h264"
	"xmediaEmu/pkg/image"
	"xmediaEmu/pkg/log"
	"xmediaEmu/pkg/media/h264writer"
)

var seed = rand.New(rand.NewSource(time.Now().UnixNano())).Uint32()

// 测试图片转视频函数.
// 功能，
func main() {
	const W = 640
	const H = 480

	log.Init("1", log.Config{Level: int32(0), Dir: "./", Path: "playVideo.log", FileNum: 10})

	// https://blog.csdn.net/A199222/article/details/85785198.
	//var video config.VideoConfig
	//video.Codec = "h264"
	//video.H264.Preset = "fast"
	////video.H264.Tune = "stillimage"
	//video.H264.Profile = "baseline" // 只支持I和P帧.
	//video.H264.Crf = 20             // 数的取值范围为0~51，其中0为无损模式，数值越大，画质越差，生成的文件却越小, 18-28比较合适.
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
		h264File, err := h264writer.New("output.h264")
		if err != nil {
			panic(err)
		}

		// 直接写入sps和pps.
		h264Heads := enc.EncodeHeader()
		if err := h264File.WriteFile(h264Heads); err != nil {
			pterm.FgLightRed.Printfln("h264File.WriteFile failed:%v\n", err)
		} else {
			pterm.FgWhite.Printf("EncodeHeader write header length:%d done. \n", len(h264Heads))
		}

		// fan-out Screen, write file directly.
		for data := range eoutput {
			if err := h264File.WriteFile(data.Data); err != nil {
				pterm.FgLightRed.Printfln("h264File.WriteFile failed:%v\n", err)
			} else {
				pterm.FgWhite.Printf("e-output write image length:%d done. \n", len(data.Data))
			}
		}
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

			// im转h264.
			for {
				if len(einput) < cap(einput) {
					// time.Sleep(time.Millisecond * 20)
					einput <- encoder.InFrame{Image: image.ImageToRGBA(im), Timestamp: uint32(time.Now().UnixNano()/8333) + seed}
					break
				} else {
					// pterm.FgLightRed.Printf("einput capacity is full\n. ")
					time.Sleep(time.Millisecond * 20)
				}
			}
			pterm.FgWhite.Printf("put image index:%s done. \n", filename)
		}
	}

	//for i := 1; i < 24; i++ {
	//	filename := "outdir/" + fmt.Sprintf("%05d.jpg", i)
	//	im, err := image.LoadJPG(filename)
	//	if err != nil {
	//		pterm.FgLightRed.Printf("load file failed:%s\n. err:%v", filename, err)
	//		panic(err)
	//	}
	//
	//	// im转h264.
	//	for {
	//		if len(einput) < cap(einput) {
	//			einput <- encoder.InFrame{Image: image.ImageToRGBA(im), Timestamp: uint32(time.Now().UnixNano()/8333) + seed}
	//			break
	//		} else {
	//			// pterm.FgLightRed.Printf("einput capacity is full\n. ")
	//			time.Sleep(time.Millisecond * 20)
	//		}
	//	}
	//	pterm.FgGreen.Printf("put image index:%s done. \n", filename)
	//}
	for len(einput) > 0 {
		time.Sleep(time.Millisecond * 20)
	}
	pterm.FgLightGreen.Print("vPipe stopped. \n")
	vPipe.Stop()
	<-done
}
