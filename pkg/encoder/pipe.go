package encoder

import (
	"github.com/pterm/pterm"
	"xmediaEmu/pkg/encoder/yuv"
	"xmediaEmu/pkg/log"
)

type VideoPipe struct {
	Input  chan InFrame
	Output chan OutFrame
	done   chan struct{}

	encoder Encoder

	// frame size
	w, h int
}

// NewVideoPipe returns new video encoder pipe.
// By default it waits for RGBA images on the input channel,
// converts them into YUV I420 format,
// encodes with provided video encoder, and
// puts the result into the output channel.
func NewVideoPipe(enc Encoder, w, h int) *VideoPipe {
	return &VideoPipe{
		Input:  make(chan InFrame, 1),
		Output: make(chan OutFrame, 2),
		done:   make(chan struct{}),

		encoder: enc,

		w: w,
		h: h,
	}
}

// Start begins video encoding pipe.
// Should be wrapped into a goroutine.
func (vp *VideoPipe) Start() {
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Error("Warn: Recovered panic in encoding ", r)
		}
		close(vp.Output)
		close(vp.done)
	}()

	yuvProc := yuv.NewYuvImgProcessor(vp.w, vp.h)
	for img := range vp.Input {

		// for test.
		yCbCr := yuvProc.Process(img.Image).Get()
		frame := vp.encoder.Encode(yCbCr)
		if len(frame) > 0 {
			pterm.FgGreen.Printf("VideoPipe Encode success, Image length:%d, yCbCr length:%d, frame length:%d. \n", len(img.Image.Pix), len(yCbCr), len(frame))
			vp.Output <- OutFrame{Data: frame, Timestamp: img.Timestamp}
		} else if len(frame) == 0 {
			pterm.FgWhite.Printf("VideoPipe Encode image nil, may be buffer, img.Image length:%d, yCbCr length:%d, frame length:%d. \n", len(img.Image.Pix), len(yCbCr), len(frame))
		} else {
			pterm.FgWhite.Printf("VideoPipe Encode image error, img.Image length:%d, yCbCr length:%d, frame length:%d. \n", len(img.Image.Pix), len(yCbCr), len(frame))
		}
	}

	// 输出缓存的.
	for frame := vp.encoder.Encode(nil); len(frame) > 0; {
		pterm.FgGreen.Printf("VideoPipe Encode delayed buff success, frame length:%d. \n", len(frame))
		vp.Output <- OutFrame{Data: frame, Timestamp: 0} // 时间戳丢失算了.
		frame = vp.encoder.Encode(nil)
	}

	// else to do?.
}

func (vp *VideoPipe) Stop() {
	close(vp.Input)
	<-vp.done
	if err := vp.encoder.Shutdown(); err != nil {
		log.Logger.Error("error: failed to close the encoder")
	}
}
