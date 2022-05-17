package rtpua

type Sdp struct {
	// SessionId    string `binding:"required"`
	Address      string
	AudioPayLoad int
	VideoPayLoad int
}
