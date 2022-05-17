package worker

import (
	"xmediaEmu/pkg/emulator/config"
	"xmediaEmu/pkg/emulator/rtpua"
	"xmediaEmu/pkg/encoder"
	"xmediaEmu/pkg/encoder/h264"
	"xmediaEmu/pkg/log"
)

// startVideo processes imageChannel images with an encoder (codec) then pushes the result to WebRTC.
func (r *Room) startVideo(width, height int, video config.VideoConfig) {
	var enc *h264.H264 // encoder.Encoder
	var err error

	log.Logger.Debug("Video codec:", video.Codec)
	if video.Codec != string(config.H264) {
		log.Logger.Error("ignore unknown codec:", video.Codec)
	}
	enc, err = h264.NewEncoder(width, height, h264.WithOptions(h264.Options{
		Crf:      video.H264.Crf,
		Tune:     video.H264.Tune,
		Preset:   video.H264.Preset,
		Profile:  video.H264.Profile,
		LogLevel: 3, // debug
	}))
	if err != nil {
		log.Logger.Error("error create new encoder", err)
		return
	}
	log.Logger.Debugf("startVideo: create new encoder:%v", enc)

	r.vPipe = encoder.NewVideoPipe(enc, width, height)
	einput, eoutput := r.vPipe.Input, r.vPipe.Output

	go r.vPipe.Start()
	defer r.vPipe.Stop()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Logger.Warn("Recovered when sent to close Image Channel")
			}
		}()

		// Test: TODO:测试代码，待删除.
		//h264File, err := h264writer.New("worker.h264")
		//if err != nil {
		//	panic(err)
		//}

		// fanout Screen, send to result rtp to .
		for data := range eoutput {
			// TODO: r.rtcSessions is rarely updated. Lock will hold down perf
			for _, webRTC := range r.rtcSessions {
				if !webRTC.IsConnected() {
					log.Logger.Debugf("webRTC disconnect, ignored. ")
					continue
				}
				// encode frame
				// fanout imageChannel
				// NOTE: can block here
				webRTC.ImageChannel <- rtpua.WebFrame{Data: data.Data, Timestamp: data.Timestamp}
				log.Logger.Debugf("startVideo done, write ImageChannel: %d", len(data.Data))

				// Test:
				//if err := h264File.WriteFile(data.Data); err != nil {
				//	pterm.FgLightRed.Printfln("startVideo:test  h264File.WriteFile failed:%v\n", err)
				//} else {
				//	pterm.FgWhite.Printf("startVideo:test write image length:%d done. \n", len(data.Data))
				//}
			}
		}
	}()

	// imageChannel来自图片的接收输入流.
	for image := range r.imageChannel {
		if len(einput) < cap(einput) {
			einput <- encoder.InFrame{Image: image.Image, Timestamp: image.Timestamp}
			log.Logger.Debugf("startVideo done, einput length of image: %d", len(image.Image.Pix))
		} else {
			log.Logger.Info("startVideo done, einput queue is full")
		}
	}
	log.Logger.Fatal("Room ", r.ID, " video channel closed")
}
