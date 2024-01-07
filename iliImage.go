package adatft

import (
	"image"
	"image/color"
	"log"
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
	dstRect image.Rectangle
}

// Erzeugt einen neuen Buffer, der fuer die Anzeige von image.RGBA Bildern
// zwingend gebraucht wird.
func NewILIImage(r image.Rectangle) *ILIImage {
	b := &ILIImage{}
	b.Pix = make([]uint8, r.Dx()*r.Dy()*bytesPerPixel)
	b.Stride = r.Dx() * bytesPerPixel
	b.Rect = r

	// b.bufLen = width * height
	// b.bufSize = bytesPerPixel * b.bufLen
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

// func (b *Buffer) Clear() {
//     for i, _ := range b.Pix {
//         b.Pix[i] = 0x00
//     }
// }

// Mit dieser Funktion wird ein Bild vom RGBA-Format (image.RGBA) in das
// für den ILI9341 typische 666 (präferiert) oder 565 Format konvertiert.
// Die Grösse von src (Breite, Höhe) muss der Grösse des TFT-Displays
// (d.h. ILI9341_WIDTH x ILI9341_HEIGHT) entsprechen. Allfällige Anpassungen
// sind vorgängig mit anderen Funktionen (bspw. aus dem Package gg oder
// image/draw) durchzuführen. Die Zeitmessung über die Variablen 'ConvTime'
// und 'NumConv' ist in dieser Funktion realisiert.
func (b *ILIImage) Convert(src *image.RGBA) {
	var stride, srcIdx, srcIdx2, dstIdx, dstIdx2 int
    var row, col int

	log.Printf("src.Bounds(): %v", src.Bounds())
	log.Printf("src.Rect    : %v", src.Rect)
	log.Printf("b.Bounds()  : %v", b.Bounds())
    log.Printf("b.Rect      : %v", b.Rect)
	t1 := time.Now()

	b.dstRect = src.Bounds()
    b.Rect = src.Bounds()
	r := src.Bounds()
	stride = r.Dx() * bytesPerPixel
    log.Printf("r: %v", r)

	for row = r.Min.Y; row < r.Max.Y; row++ {
        col = r.Min.X
		srcIdx = (row-r.Min.Y)*src.Stride
		srcIdx2 = src.PixOffset(col, row)
		dstIdx = (row-r.Min.Y)*stride
        dstIdx2 = b.PixOffset(col, row)
        log.Printf("srcIdx, srcIdx2, dstIdx, dstIdx2: %d, %d, %d, %d", srcIdx, srcIdx2, dstIdx, dstIdx2)
		for col = r.Min.X; col < r.Max.X; col++ {
			b.Pix[dstIdx+0] = src.Pix[srcIdx+2]
			b.Pix[dstIdx+1] = src.Pix[srcIdx+1]
			b.Pix[dstIdx+2] = src.Pix[srcIdx+0]
			srcIdx += 4
			dstIdx += bytesPerPixel
		}
	}
	ConvTime += time.Since(t1)
	NumConv++
}
