package rtpua

import (
	encoder "xmediaEmu/pkg/emulator/config"
)

const audioPayload  = 101
const videoPayload  = 102

type Config struct {
	Encoder encoder.EncoderConfig
	NetWork string
	// UdpAddr *net.UDPAddr
	UdpAddr string // "IP:Port"
	SessionId  string //
}

