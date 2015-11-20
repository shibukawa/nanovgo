package nanovgo

const (
	nvg_INIT_FONTIMAGE_SIZE = 512
	nvg_MAX_FONTIMAGE_SIZE  = 2048
	nvg_MAX_FONTIMAGES      = 4

	nvg_INIT_COMMANDS_SIZE = 256
	nvg_INIT_POINTS_SIZE   = 128
	nvg_INIT_PATHS_SIZE    = 16
	nvg_INIT_VERTS_SIZE    = 256
	nvg_MAX_STATES         = 32
)

type nvgCommands int

const (
	nvg_MOVETO nvgCommands = iota
	nvg_LINETO
	nvg_BEZIERTO
	nvg_CLOSE
	nvg_WINDING
)

type nvgPointFlags int

const (
	nvg_PT_CORNER     nvgPointFlags = 0x01
	nvg_PT_LEFT       nvgPointFlags = 0x02
	nvg_PT_BEVEL      nvgPointFlags = 0x04
	nvg_PR_INNERBEVEL nvgPointFlags = 0x08
)

type nvgTextureType int

const (
	nvg_TEXTURE_ALPHA nvgTextureType = 1
	nvg_TEXTURE_RGBA  nvgTextureType = 2
)

type nvgCodePointSize int

const (
	nvg_NEWLINE nvgCodePointSize = iota
	nvg_SPACE
	nvg_CHAR
)
