//go:build pixfmt666

// Dies ist die Implementation des ILI-Farbtyps mit je 6 Bit pro Farbe, resp.
// 3 Bytes pro Pixel.
package adatft

import (
	"image/color"
)

const (
    bytesPerPixel = 3
    pixfmt uint8 = 0x06
)

type ILIColor struct {
	R, G, B uint8
}

func NewILIColor(r, g, b uint8) (ILIColor) {
    return ILIColor{r, g, b}
}

func (c ILIColor) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	g = uint32(c.G)
	g |= g << 8
	b = uint32(c.B)
	b |= b << 8
	a = 0xffff
	return
}

func iliModel(c color.Color) color.Color {
	if _, ok := c.(ILIColor); ok {
		return c
	}
	r, g, b, a := c.RGBA()
	if a == 0xffff {
		return ILIColor{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
	}
	if a == 0x0000 {
		return ILIColor{0, 0, 0}
	}
	r = (r * 0xffff) / a
	g = (g * 0xffff) / a
	b = (b * 0xffff) / a
	return ILIColor{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
}

var (
	ILIModel color.Model = color.ModelFunc(iliModel)
)

