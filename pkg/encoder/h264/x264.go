package h264

import "C"
import (
	"fmt"
	"github.com/pterm/pterm"
)

// 利用x264进行转换.
type H264 struct {
	ref *T // x264_t

	width      int32
	lumaSize   int32
	chromaSize int32
	csp        int32
	nnals      int32
	nals       []*Nal // 一般一帧只返回一个.

	// keep monotonic pts to suppress warnings
	pts int64
}

func NewEncoder(width, height int, options ...Option) (encoder *H264, err error) {
	libVersion := int(Build)

	if libVersion < 150 {
		return nil, fmt.Errorf("x264: the library version should be newer than v150, you have got version %v", libVersion)
	}

	if libVersion < 160 {
		pterm.FgGreen.Printfln("x264: warning, installed version of libx264 %v is older than minimally supported v160, expect bugs", libVersion)
	}

	opts := &Options{
		Crf:     12,
		Tune:    "zerolatency",
		Preset:  "superfast",
		Profile: "baseline",
	}

	// 初始化没生效.
	for _, opt := range options {
		opt(opts)
	}
	pterm.FgGreen.Printfln("x264: build v%v, opts:%+v", Build, opts)

	param := Param{}
	if opts.Preset != "" && opts.Tune != "" {
		if ParamDefaultPreset(&param, opts.Preset, opts.Tune) < 0 {
			return nil, fmt.Errorf("x264: invalid preset/tune name")
		}
	} else {
		ParamDefault(&param)
	}

	opts.Profile = "baseline"
	if opts.Profile != "" {
		if ParamApplyProfile(&param, opts.Profile) < 0 {
			return nil, fmt.Errorf("x264: invalid profile name")
		}
	}
	pterm.FgLightGreen.Printfln("ParamDefault is :%+v", param)

	// legacy encoder lacks of this param
	param.IBitdepth = 8

	if libVersion > 155 {
		param.ICsp = CspI420
	} else {
		param.ICsp = 1
	}
	param.IWidth = int32(width)
	param.IHeight = int32(height)
	param.ILogLevel = opts.LogLevel

	param.Rc.IRcMethod = RcCrf
	param.Rc.FRfConstant = float32(opts.Crf)

	// use annexb and sps/pps for every key frame default.
	param.BAnnexb = 1
	param.BRepeatHeaders = 1

	encoder = &H264{
		csp:        param.ICsp,
		lumaSize:   int32(width * height),
		chromaSize: int32(width*height) / 4,
		nals:       make([]*Nal, 1),
		width:      int32(width),
	}

	if encoder.ref = EncoderOpen(&param); encoder.ref == nil {
		err = fmt.Errorf("x264: cannot open the encoder")
		return
	}
	return
}

// if yuv is nil, returns buffered frames.
// else, return encoded frame in the queue .
func (e *H264) Encode(yuv []byte) []byte {
	if yuv == nil || len(yuv) == 0 {
		return e.EncodeBuffered()
	}

	var picIn, picOut Picture
	picIn.Img.ICsp = e.csp
	picIn.Img.IPlane = 3
	picIn.Img.IStride[0] = e.width
	picIn.Img.IStride[1] = e.width / 2
	picIn.Img.IStride[2] = e.width / 2

	picIn.Img.Plane[0] = C.CBytes(yuv[:e.lumaSize])
	picIn.Img.Plane[1] = C.CBytes(yuv[e.lumaSize : e.lumaSize+e.chromaSize])
	picIn.Img.Plane[2] = C.CBytes(yuv[e.lumaSize+e.chromaSize:])

	picIn.IPts = e.pts

	// every image is key frame.
	picIn.BKeyframe = 1
	e.pts++

	defer func() {
		picIn.freePlane(0)
		picIn.freePlane(1)
		picIn.freePlane(2)
	}()

	if ret := EncoderEncode(e.ref, e.nals, &e.nnals, &picIn, &picOut); ret > 0 {
		pterm.FgWhite.Printfln("EncoderEncode success ret:%d", ret)
		return C.GoBytes(e.nals[0].PPayload, C.int(ret))
		// ret should be equal to writer writes
	} else if ret == 0 {
		pterm.FgLightGreen.Printfln("EncoderEncode nil ret:%d", ret)
	} else {
		pterm.FgRed.Printfln("EncoderEncode failed ret:%d", ret)
	}
	return []byte{}
}

// outputs SPS/PPS/SEI.
func (e *H264) EncodeHeader() []byte {
	// 返回sps和pps头部.
	if ret := EncoderHeaders(e.ref, e.nals, &e.nnals); ret > 0 {
		pterm.FgWhite.Printfln("EncodeHeader success ret:%d", ret)
		return C.GoBytes(e.nals[0].PPayload, C.int(ret))
		// ret should be equal to writer writes
	} else if ret == 0 {
		pterm.FgLightGreen.Printfln("EncodeHeader nil ret:%d", ret)
	} else {
		pterm.FgRed.Printfln("EncodeHeader failed ret:%d", ret)
	}

	return []byte{}
}

// 获取缓存区的.
func (e *H264) EncodeBuffered() []byte {
	var picOut Picture
	if ret := EncoderDelayedFrames(e.ref); ret > 0 {
		pterm.FgLightGreen.Printfln("EncoderDelayedFrames count:%d", ret)
		if ret := EncoderEncode(e.ref, e.nals, &e.nnals, nil, &picOut); ret > 0 {
			pterm.FgWhite.Printfln("EncodeBuffered success ret:%d", ret)
			return C.GoBytes(e.nals[0].PPayload, C.int(ret))
			// ret should be equal to writer writes
		} else if ret == 0 {
			pterm.FgLightGreen.Printfln("EncodeBuffered nil ret:%d", ret)
		} else {
			pterm.FgRed.Printfln("EncodeBuffered failed ret:%d", ret)
		}
		// return C.GoBytes(e.nals[0].PPayload, C.int(ret))
		// ret should be equal to writer writes
	}

	return []byte{}
}

func (e *H264) Shutdown() error {
	EncoderClose(e.ref)
	return nil
}
