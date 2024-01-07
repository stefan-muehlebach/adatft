package adatft

import (
	"image"
	"image/color"
	"time"
)

//-----------------------------------------------------------------------------

// Dieser Record wird fuer die Konvertierung der Bilddaten in ein von
// ILI9341 unterstuetztes Format verwendet.
const (
	//bytesPerPixel = 2
	bytesPerPixel = 3
)

type ILIImage struct {
	Pix    []uint8
	Stride int
	Rect   image.Rectangle
	// bufLen, bufSize int
	// dstRect image.Rectangle
}

// Erzeugt einen neuen Buffer, der fuer die Anzeige von image.RGBA Bildern
// zwingend gebraucht wird.
func NewILIImage(r image.Rectangle) *ILIImage {
	b := &ILIImage{}
	b.Pix = make([]uint8, r.Dx()*r.Dy()*bytesPerPixel)
	b.Stride = r.Dx() * bytesPerPixel
	b.Rect = r
	return b
}

func (b *ILIImage) ColorModel() color.Model {
	return ILIModel
}

func (b *ILIImage) Bounds() image.Rectangle {
	return b.Rect
}

func (b *ILIImage) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(b.Rect)) {
		return ILIColor{}
	}
	i := b.PixOffset(x, y)
	s := b.Pix[i : i+3 : i+3]
	return ILIColor{s[0], s[1], s[2]}
}

func (b *ILIImage) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(b.Rect)) {
		return
	}
	i := b.PixOffset(x, y)
	s := b.Pix[i : i+3 : i+3]
	c1 := ILIModel.Convert(c).(ILIColor)
	s[0] = c1.R
	s[1] = c1.G
	s[2] = c1.B
}

func (b *ILIImage) PixOffset(x, y int) int {
	return (y-b.Rect.Min.Y)*b.Stride + (x-b.Rect.Min.X)*bytesPerPixel
}

func (b *ILIImage) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(b.Rect)
	if r.Empty() {
		return &ILIImage{}
	}
	i := b.PixOffset(r.Min.X, r.Min.Y)
	return &ILIImage{
		Pix:    b.Pix[i:],
		Stride: b.Stride,
		Rect:   r,
	}
}

func (b *ILIImage) Diff(img *ILIImage) image.Rectangle {
	var xMin, xMax, yMin, yMax int

	xMin, xMax = b.Rect.Dx(), 0
	yMin, yMax = b.Rect.Dy(), 0

Loop1:
	for y := 0; y < b.Rect.Dy(); y++ {
		idx := y * b.Stride
		for i, pix := range b.Pix[idx : idx+b.Stride] {
			// log.Printf("idx, i: %d, %d", idx, i)
			if pix != img.Pix[idx+i] {
				yMin = y
				break Loop1
			}
		}
	}
Loop2:
	for y := b.Rect.Dy() - 1; y > yMin; y-- {
		idx := y * b.Stride
		for i, pix := range b.Pix[idx : idx+b.Stride] {
			// log.Printf("idx, i: %d, %d", idx, i)
			if pix != img.Pix[idx+i] {
				yMax = y
				break Loop2
			}
		}
	}
Loop3:
	for x := 0; x < b.Rect.Dx(); x++ {
		idx := x * bytesPerPixel
		for i := 0; i < b.Rect.Dy()*b.Stride; i += b.Stride {
			if b.Pix[idx+i] != img.Pix[idx+i] {
				xMin = x
				break Loop3
			}
		}
	}
Loop4:
	for x := b.Rect.Dx() - 1; x > xMin; x-- {
		idx := x * bytesPerPixel
		for i := 0; i < b.Rect.Dy()*b.Stride; i += b.Stride {
			if b.Pix[idx+i] != img.Pix[idx+i] {
				xMax = x
				break Loop4
			}
		}
	}

	return image.Rect(xMin, yMin, xMax, yMax)
}

func (b *ILIImage) Clear() {
	for i := range b.Pix {
		b.Pix[i] = 0x00
	}
}

func (b *ILIImage) Convert(src *image.RGBA) {
	var row, col int
	var srcBaseIdx, srcIdx, dstBaseIdx, dstIdx int

	t1 := time.Now()

	// log.Printf("src.Bounds(): %v", src.Bounds())
	// log.Printf("src.Rect    : %v", src.Rect)
	// log.Printf("dst.Bounds(): %v", dst.Bounds())
	// log.Printf("dst.Rect    : %v", dst.Rect)

	// dst.Rect = src.Rect

	srcBaseIdx = 0
	dstBaseIdx = src.Rect.Min.Y*b.Stride + src.Rect.Min.X*bytesPerPixel
	for row = src.Rect.Min.Y; row < src.Rect.Max.Y; row++ {
		srcIdx = srcBaseIdx
		dstIdx = dstBaseIdx

		for col = src.Rect.Min.X; col < src.Rect.Max.X; col++ {
			b.Pix[dstIdx+0] = src.Pix[srcIdx+2]
			b.Pix[dstIdx+1] = src.Pix[srcIdx+1]
			b.Pix[dstIdx+2] = src.Pix[srcIdx+0]
			srcIdx += 4
			dstIdx += bytesPerPixel
		}

		srcBaseIdx += src.Stride
		dstBaseIdx += b.Stride
	}
	ConvTime += time.Since(t1)
	NumConv++
}
