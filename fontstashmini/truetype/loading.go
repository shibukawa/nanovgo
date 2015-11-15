package truetype

import (
	"errors"
)

// FontInfo is defined publically so you can declare on on the stack or as a
// global or etc, but you should treat it as opaque.
type FontInfo struct {
	data             []byte // contains the .ttf file
	fontStart        int    // offset of start of font
	loca             int    // table location as offset from start of .ttf
	head             int
	glyf             int
	hhea             int
	hmtx             int
	kern             int
	numGlyphs        int // number of glyphs, needed for range checking
	indexMap         int // a cmap mapping for our chosen character encoding
	indexToLocFormat int // format needed to map from glyph index to glyph
}

// Each .ttf/.ttc file may have more than one font. Each font has a sequential
// index number starting from 0. Call this function to get the font offset for
// a given index; it returns -1 if the index is out of range. A regular .ttf
// file will only define one font and it always be at offset 0, so it will return
// '0' for index 0, and -1 for all other indices. You can just skip this step
// if you know it's that kind of font.
func GetFontOffsetForIndex(data []byte, index int) int {
	if isFont(data) {
		if index == 0 {
			return 0
		} else {
			return -1
		}
	}

	// Check if it's a TTC
	if string(data[0:4]) == "ttcf" {
		if u32(data, 4) == 0x00010000 || u32(data, 4) == 0x00020000 {
			n := int(u32(data, 8))
			if index >= n {
				return -1
			}
			return int(u32(data, 12+index*14))
		}
	}
	return -1
}

// Given an offset into the file that defines a font, this function builds the
// necessary cached info for the rest of the system.
func InitFont(data []byte, offset int) (font *FontInfo, err error) {
	if len(data)-offset < 12 {
		err = errors.New("TTF data is too short")
		return
	}
	font = new(FontInfo)
	font.data = data
	font.fontStart = offset

	cmap := findTable(data, offset, "cmap")
	font.loca = findTable(data, offset, "loca")
	font.head = findTable(data, offset, "head")
	font.glyf = findTable(data, offset, "glyf")
	font.hhea = findTable(data, offset, "hhea")
	font.hmtx = findTable(data, offset, "hmtx")
	font.kern = findTable(data, offset, "kern")
	if cmap == 0 || font.loca == 0 || font.head == 0 || font.glyf == 0 || font.hhea == 0 || font.hmtx == 0 {
		err = errors.New("Required table not found")
		return
	}

	t := findTable(data, offset, "maxp")
	if t != 0 {
		font.numGlyphs = int(u16(data, t+4))
	} else {
		font.numGlyphs = 0xfff
	}

	numTables := int(u16(data, cmap+2))
	for i := 0; i < numTables; i++ {
		encodingRecord := cmap + 4 + 8*i
		switch int(u16(data, encodingRecord)) {
		case PLATFORM_ID_MICROSOFT:
			switch int(u16(data, encodingRecord+2)) {
			case MS_EID_UNICODE_FULL, MS_EID_UNICODE_BMP:
				font.indexMap = cmap + int(u32(data, encodingRecord+4))
			}
		}
	}

	if font.indexMap == 0 {
		err = errors.New("Unknown cmap encoding table")
		return
	}

	font.indexToLocFormat = int(u16(data, font.head+50))
	return
}
