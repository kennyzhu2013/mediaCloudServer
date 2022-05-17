package inpututil

type YDirection int

const (
	Upward YDirection = iota
	Downward
)

type FPSModeType int

const (
	FPSIntOnly FPSModeType = iota
	FPSIntAndKey
	FPSAll
)
