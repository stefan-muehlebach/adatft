package adatft

import (
	"image/color"
    "image"
    "time"
)

//-----------------------------------------------------------------------------

// Dieser Record wird fuer die Konvertierung der Bilddaten in ein von
// ILI9341 unterstuetztes Format verwendet.
const (
    //bytesPerPixel = 2
    bytesPerPixel = 3
)

type Buffer struct {
    Pix []uint8
    Stride int
    Rect image.Rectangle
    // bufLen, bufSize int
    dstRect image.Rectangle
}

// Erzeugt einen neuen Buffer, der fuer die Anzeige von image.RGBA Bildern
// zwingend gebraucht wird.
func NewBuffer(r image.Rectangle) (*Buffer) {
    b := &Buffer{}
    b.Pix = make([]uint8, r.Dx() * r.Dy() * bytesPerPixel)
    b.Stride = r.Dx() * bytesPerPixel
    b.Rect = r

    // b.bufLen = width * height
    // b.bufSize = bytesPerPixel * b.bufLen
    return b
}

func (b *Buffer) ColorModel() (color.Model) {
    return ILIModel
}

func (b *Buffer) Bounds() (image.Rectangle) {
    return b.Rect
}

func (b *Buffer) At(x, y int) (color.Color) {
    if !(image.Point{x, y}.In(b.Rect)) {
        return ILIColor{}
    }
    i := (y-b.Rect.Min.Y)*b.Stride + (x-b.Rect.Min.X)*bytesPerPixel
    s := b.Pix[i : i+3 : i+3]
    return ILIColor{s[0], s[1], s[2]}
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
func (b *Buffer) Convert(src *image.RGBA) {
    var stride, srcIdx, dstIdx int

    t1 := time.Now()
    b.dstRect = src.Bounds()
    stride = b.dstRect.Dx() * bytesPerPixel

    for row:=b.dstRect.Min.Y; row<b.dstRect.Max.Y; row++ {
        srcIdx = (row-b.dstRect.Min.Y)*src.Stride
        dstIdx = (row-b.dstRect.Min.Y)*stride
        for col:=b.dstRect.Min.X; col<b.dstRect.Max.X; col++ {
            b.Pix[dstIdx+0]  =  src.Pix[srcIdx+2]
            b.Pix[dstIdx+1]  =  src.Pix[srcIdx+1]
            b.Pix[dstIdx+2]  =  src.Pix[srcIdx+0]
            srcIdx += 4
            dstIdx += bytesPerPixel
        }
    }
    ConvTime += time.Since(t1)
    NumConv++
}
