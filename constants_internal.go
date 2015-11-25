package nanovgo

const (
	nvgInitFontImageSize = 512
	nvgMaxFontImageSize  = 2048
	nvgMaxFontImages     = 4

	nvgInitCommandsSize = 256
	nvgInitPointsSize   = 128
	nvgInitPathsSize    = 16
	nvgInitVertsSize    = 256
	nvgMaxStates        = 32
)

type nvgCommands int

const (
	nvgMOVETO nvgCommands = iota
	nvgLINETO
	nvgBEZIERTO
	nvgCLOSE
	nvgWINDING
)

type nvgPointFlags int

const (
	nvgPtCORNER     nvgPointFlags = 0x01
	nvgPtLEFT       nvgPointFlags = 0x02
	nvgPtBEVEL      nvgPointFlags = 0x04
	nvgPrINNERBEVEL nvgPointFlags = 0x08
)

type nvgTextureType int

const (
	nvgTextureALPHA nvgTextureType = 1
	nvgTextureRGBA  nvgTextureType = 2
)

type nvgCodePointSize int

const (
	nvgNEWLINE nvgCodePointSize = iota
	nvgSPACE
	nvgCHAR
)
