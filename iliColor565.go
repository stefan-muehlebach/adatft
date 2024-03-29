//go:build pixfmt565

// Dies ist die Implementation des 565-Farbtyps.
package adatft

import (
	"image/color"
)

type ILIColor struct {
    HB, LB uint8
}

func (c ILIColor) RGBA() (r, g, b, a uint32) {
    r = uint32(c.HB & 0xF8)
    r |= r << 8
    g = uint32(((c.HB & 0x07) << 5) | ((c.LB & 0xE0) >> 3))
    g |= g << 8
    b = uint32((c.LB & 0x1F) << 3)
    b |= b << 8
    a = 0xffff
    return
}

func iliModel(c color.Color) (color.Color) {
    if _, ok := c.(ILIColor); ok {
        return c
    }
    r, g, b, a := c.RGBA()
    if a == 0xffff {
        r = (r >> 8) & 0xF8
        g = (g >> 8) & 0xFC
        b = (b >> 8) & 0xF8
        return ILIColor{uint8(r | (g >> 5)), uint8((g << 3) | (b >> 3))}
    }
    if a == 0x0000 {
        return ILIColor{0, 0}
    }
	r = (r * 0xffff) / a
    r = (r >> 8) & 0x1F
	g = (g * 0xffff) / a
    g = (g >> 8) & 0x3F
	b = (b * 0xffff) / a
    b = (b >> 8) & 0x1F
    return ILIColor{uint8((r << 3) | (g >> 5)), uint8((g << 5) | b)}
}

var (
    ILIModel color.Model = color.ModelFunc(iliModel)
)

