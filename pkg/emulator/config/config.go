package config

// 以下信息从没配置文件读取.
type Config struct {
	Encoder EncoderConfig
	Width   int
	Height  int

	LocalMediaIp string
}

type EncoderConfig struct {
	Audio          AudioConfig
	Video          VideoConfig
	BUseUnixSocket bool // bUseUnixSocket, true will
}

// Codec没有就表示对应的音频或视频没有.
type AudioConfig struct {
	Codec     string
	Channels  int
	Frame     int
	Frequency int
}

type VideoConfig struct {
	Codec string
	H264  struct {
		Crf     uint8
		Preset  string
		Profile string
		Tune    string
	}
	Vpx struct {
		Bitrate          uint
		KeyframeInterval uint
	}
}

func (a *AudioConfig) GetFrameDuration() int {
	return a.Frequency * a.Frame / 1000 * a.Channels
}

type VideoCodec string
type AudioCodec string

const (
	H264 VideoCodec = "h264"

	// TODO:支持vpx.

	// 音频.
	G711  AudioCodec = "h264"
	AMRNB AudioCodec = "AmrNb"
	AMRWB AudioCodec = "Amrwb"
	PCM   AudioCodec = "PCM" // pcm原始码流.
)
