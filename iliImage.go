package adatft

import (
    "image"
    "image/color"
    "time"
)

//----------------------------------------------------------------------------

// Dies ist die Implementation des ILI-Farbtyps mit je 6 Bit pro Farbe, resp.
// 3 Bytes pro Pixel.
type ILIColor struct {
    R, G, B uint8
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

func iliModel(c color.Color) (color.Color) {
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

// Dies ist die Implementation des 565-Farbtyps.
//
// type ILIcolor struct {
//     R, G, B uint8
// }

// func (c ILIcolor) RGBA() (r, g, b, a uint32) {
//     r = uint32(c.R)
//     r |= r << 8
//     g = uint32(c.G)
//     g |= g << 8
//     b = uint32(c.B)
//     b |= b << 8
//     a = 0xffff
//     return
// }

// func iliModel(c color.Color) (color.Color) {
//     if _, ok := c.(ILIcolor); ok {
//         return c
//     }
//     r, g, b, a := c.RGBA()
//     if a == 0xffff {
//         return ILIcolor{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
//     }
//     if a == 0x0000 {
//         return ILIcolor{0, 0, 0}
//     }
// 	r = (r * 0xffff) / a
// 	g = (g * 0xffff) / a
// 	b = (b * 0xffff) / a
//     return ILIcolor{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
// }

//-----------------------------------------------------------------------------

const (
    //bytesPerPixel = 2
    bytesPerPixel = 3
)

// Diese Datenstruktur stellt ein Bild dar, welches auf dem TFT direkt
// dargestellt werden kann und implementiert alle Interfaces, welche Go
// fuer Bild-Typen kennt.
type ILIImage struct {
    Pix    []uint8
    Stride int
    Rect   image.Rectangle
}

func NewILIImage(r image.Rectangle) *ILIImage {
    stride := r.Dx() * bytesPerPixel
    length := r.Dy() * stride
    return &ILIImage{
        Pix:    make([]uint8, length),
        Stride: stride,
        Rect:   r,
    }
}

// ColorModel, Bounds und At werden vom Interface image.Image gefordert.
func (p *ILIImage) ColorModel() color.Model {
    return ILIModel
}
func (p *ILIImage) Bounds() image.Rectangle {
    return p.Rect
}
func (p *ILIImage) At(x, y int) color.Color {
    return p.ILIAt(x, y)
}

// Set wird ausserdem von draw.Image gefordert.
func (p *ILIImage) Set(x, y int, c color.Color) {
    if !(image.Point{x, y}.In(p.Rect)) {
        return
    }
    i := p.PixOffset(x, y)
    c1 := ILIModel.Convert(c).(ILIColor)
    s := p.Pix[i : i+3 : i+3]
    s[0] = c1.R
    s[1] = c1.G
    s[2] = c1.B
}

func (p *ILIImage) PixOffset(x, y int) int {
    return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*bytesPerPixel
}

func (p *ILIImage) ILIAt(x, y int) ILIColor {
    if !(image.Point{x, y}.In(p.Rect)) {
        return ILIColor{}
    }
    i := p.PixOffset(x, y)
    s := p.Pix[i : i+3 : i+3]
    return ILIColor{s[0], s[1], s[2]}
}

func (p *ILIImage) SetILI(x, y int, c ILIColor) {
    if !(image.Point{x, y}.In(p.Rect)) {
        return
    }
    i := p.PixOffset(x, y)
    s := p.Pix[i : i+3 : i+3]
    s[0] = c.R
    s[1] = c.G
    s[2] = c.B
}

func (p *ILIImage) SubImage(r image.Rectangle) image.Image {
    r = r.Intersect(p.Rect)
    if r.Empty() {
        return &ILIImage{}
    }
    i := p.PixOffset(r.Min.X, r.Min.Y)
    return &ILIImage{
        Pix:    p.Pix[i:len(p.Pix):len(p.Pix)],
        Stride: p.Stride,
        Rect:   r,
    }
}

func (p *ILIImage) Opaque() bool {
    return true
}

//----------------------------------------------------------------------------

func (p *ILIImage) Diff(img *ILIImage) image.Rectangle {
    var xMin, xMax, yMin, yMax int

    xMin, xMax = p.Rect.Dx(), 0
    yMin, yMax = p.Rect.Dy(), 0

Loop1:
    for y := 0; y < p.Rect.Dy(); y++ {
        idx := y * p.Stride
        s := p.Pix[idx : idx+p.Stride : idx+p.Stride]
        d := img.Pix[idx : idx+img.Stride : idx+img.Stride]
        for i := 0; i < p.Stride; i++ {
            if s[i] != d[i] {
                yMin = y
                yMax = y
                break Loop1
            }
        }
    }
    if yMin > yMax {
        return image.Rectangle{}
    }
Loop2:
    for y := p.Rect.Dy() - 1; y > yMin; y-- {
        idx := y * p.Stride
        s := p.Pix[idx : idx+p.Stride : idx+p.Stride]
        d := img.Pix[idx : idx+img.Stride : idx+img.Stride]
        for i := 0; i < p.Stride; i++ {
            if s[i] != d[i] {
                yMax = y+1
                break Loop2
            }
        }
    }
Loop3:
    for x := 0; x < p.Rect.Dx(); x++ {
        idx := x * bytesPerPixel
        for i := yMin*p.Stride; i < yMax*p.Stride; i += p.Stride {
            if p.Pix[idx+i] != img.Pix[idx+i] {
                xMin = x
                xMax = x
                break Loop3
            }
        }
    }
Loop4:
    for x := p.Rect.Dx() - 1; x > xMin; x-- {
        idx := x * bytesPerPixel
        for i := yMin*p.Stride; i < yMax*p.Stride; i += p.Stride {
            if p.Pix[idx+i] != img.Pix[idx+i] {
                xMax = x+1
                break Loop4
            }
        }
    }
    return image.Rectangle{image.Point{xMin, yMin}, image.Point{xMax, yMax}}
}

func (p *ILIImage) Clear() {
    for i := range p.Pix {
        p.Pix[i] = 0x00
    }
}

// Konvertierung, welche vom Rect (d.h. Bounds()) des darzustellenden Bildes
// (img) abhaengig ist.
func (p *ILIImage) Convert(src *image.RGBA) {
    var row, col int
    var srcBaseIdx, srcIdx, dstBaseIdx, dstIdx int

    t1 := time.Now()

    srcBaseIdx = 0
    dstBaseIdx = src.Rect.Min.Y*p.Stride + src.Rect.Min.X*bytesPerPixel
    for row = src.Rect.Min.Y; row < src.Rect.Max.Y; row++ {
        srcIdx = srcBaseIdx
        dstIdx = dstBaseIdx

        for col = src.Rect.Min.X; col < src.Rect.Max.X; col++ {
            s := src.Pix[srcIdx : srcIdx+3 : srcIdx+3]
            d := p.Pix[dstIdx : dstIdx+3 : dstIdx+3]
            d[0] = s[2]
            d[1] = s[1]
            d[2] = s[0]
            srcIdx += 4
            dstIdx += bytesPerPixel
        }

        srcBaseIdx += src.Stride
        dstBaseIdx += p.Stride
    }
    ConvTime += time.Since(t1)
    NumConv++
}

// Konvertierung des gesamten Bildes. Und ohne Zeitmessung.
// func (p *ILIImage) ConvertFull(src *image.RGBA) {
//     var row, col int
//     var srcIdx, dstIdx int
// 
//     srcIdx = 0
//     dstIdx = 0
//     for row = 0; row < Height; row++ {
//         for col = 0; col < Width; col++ {
//             s := src.Pix[srcIdx : srcIdx+3 : srcIdx+3]
//             d := p.Pix[dstIdx : dstIdx+3 : dstIdx+3]
//             d[0] = s[2]
//             d[1] = s[1]
//             d[2] = s[0]
//             srcIdx += 4
//             dstIdx += bytesPerPixel
//         }
//     }
// }
// 
