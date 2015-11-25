package nanovgo

// CreateFlags is used when NewContext() to create NanoVGo context.
type CreateFlags int

const (
	// AntiAlias sets NanoVGo to use AA
	AntiAlias CreateFlags = 1 << 0
	// StencilStrokes sets NanoVGo to use stencil buffer to draw strokes
	StencilStrokes CreateFlags = 1 << 1
	// Debug shows OpenGL errors to console
	Debug CreateFlags = 1 << 2
)

const (
	// Kappa90 is length proportional to radius of a cubic bezier handle for 90deg arcs.)
	Kappa90 float32 = 0.5522847493
	// PI of float32
	PI float32 = 3.14159265358979323846264338327
)

// Direction is used with Context.Arc
type Direction int

const (
	// CounterClockwise specify Arc curve direction
	CounterClockwise Direction = 1
	// Clockwise specify Arc curve direction
	Clockwise Direction = 2
)

// LineCap is used for line cap and joint
type LineCap int

const (
	// Butt is used for line cap (default value)
	Butt LineCap = iota
	// Round is used for line cap and joint
	Round
	// Square is used for line cap
	Square
	// Bevel is used for joint
	Bevel
	// Miter is used for joint (default value)
	Miter
)

// Align is used for text location
type Align int

const (
	// AlignLeft (default) is used for horizontal align. Align text horizontally to left.
	AlignLeft Align = 1 << 0
	// AlignCenter is used for horizontal align. Align text horizontally to center.
	AlignCenter Align = 1 << 1
	// AlignRight is used for horizontal align. Align text horizontally to right.
	AlignRight Align = 1 << 2
	// AlignTop is used for vertical align. Align text vertically to top.
	AlignTop Align = 1 << 3
	// AlignMiddle is used for vertical align. Align text vertically to middle.
	AlignMiddle Align = 1 << 4
	// AlignBottom is used for vertical align. Align text vertically to bottom.
	AlignBottom Align = 1 << 5
	// AlignBaseline (default) is used for vertical align. Align text vertically to baseline.
	AlignBaseline Align = 1 << 6
)

// ImageFlags is used for setting image object
type ImageFlags int

const (
	// ImageGenerateMipmaps generates mipmaps during creation of the image.
	ImageGenerateMipmaps ImageFlags = 1 << 0
	// ImageRepeatX repeats image in X direction.
	ImageRepeatX ImageFlags = 1 << 1
	// ImageRepeatY repeats image in X direction.
	ImageRepeatY ImageFlags = 1 << 2
	// ImageFlippy flips (inverses) image in Y direction when rendered.
	ImageFlippy ImageFlags = 1 << 3
	// ImagePreMultiplied specifies image data has premultiplied alpha.
	ImagePreMultiplied ImageFlags = 1 << 4
)

// Winding is used for changing filling strategy
type Winding int

const (
	// Solid fills internal hole
	Solid Winding = 1
	// Hole keeps internal hole
	Hole Winding = 2
)
