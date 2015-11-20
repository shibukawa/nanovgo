package nanovgo

type CreateFlags int

const (
	ANTIALIAS       CreateFlags = 1 << 0
	STENCIL_STROKES CreateFlags = 1 << 1
	DEBUG           CreateFlags = 1 << 2
)

const (
	KAPPA90 float32 = 0.5522847493 // Length proportional to radius of a cubic bezier handle for 90deg arcs.)
	PI      float32 = 3.14159265358979323846264338327
)

type Direction int

const (
	CCW Direction = 1
	CW  Direction = 2
)

type LineCap int

const (
	BUTT LineCap = iota
	ROUND
	SQUARE
	BEVEL
	MITER
)

type Align int

const (
	// Horizontal align
	ALIGN_LEFT   Align = 1 << 0 // Default, align text horizontally to left.
	ALIGN_CENTER Align = 1 << 1 // Align text horizontally to center.
	ALIGN_RIGHT  Align = 1 << 2 // Align text horizontally to right.
	// Vertical align
	ALIGN_TOP      Align = 1 << 3 // Align text vertically to top.
	ALIGN_MIDDLE   Align = 1 << 4 // Align text vertically to middle.
	ALIGN_BOTTOM   Align = 1 << 5 // Align text vertically to bottom.
	ALIGN_BASELINE Align = 1 << 6 // Default, align text vertically to baseline.
)

type ImageFlags int

const (
	IMAGE_GENERATE_MIPMAPS ImageFlags = 1 << 0 // Generate mipmaps during creation of the image.
	IMAGE_REPEATX          ImageFlags = 1 << 1 // Repeat image in X direction.
	IMAGE_REPEATY          ImageFlags = 1 << 2 // Repeat image in X direction.
	IMAGE_FLIPY            ImageFlags = 1 << 3 // Flips (inverses) image in Y direction when rendered.
	IMAGE_PREMULTIPLIED    ImageFlags = 1 << 4 // Image data has premultiplied alpha.
)

type Winding int

const (
	SOLID    Winding = 1
	NON_ZERO Winding = 1
	HOLE     Winding = 2
	EVEN_ODD Winding = 2
)
