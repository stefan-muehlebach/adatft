package adatft

import (
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
    pixBuf []uint8
    bufLen, bufSize, strideLen int
    dstRect image.Rectangle
}

// Erzeugt einen neuen Buffer, der fuer die Anzeige von image.RGBA Bildern
// zwingend gebraucht wird.
func NewBuffer() *Buffer {
    buf := &Buffer{}
    buf.bufLen = Width * Height
    buf.bufSize = bytesPerPixel * buf.bufLen
    buf.strideLen = Width * bytesPerPixel
    buf.pixBuf = make([]uint8, buf.bufSize)
    return buf
}

func (buf *Buffer) Clear() {
    for i, _ := range buf.pixBuf {
        buf.pixBuf[i] = 0x00
    }
}

// Mit dieser Funktion wird ein Bild vom RGBA-Format (image.RGBA) in das
// fuer den ILI9341 typische 666 oder (praeferiert) 565 Format konvertiert.
// Die Masse von src muessen den Massen des TFT-Displays
// (d.h. ILI9341_WIDTH x ILI9341_HEIGHT) entsprechen. Allfaellige Anpassungen
// sind vorgaengig mit anderen Funktionen (bspw. aus dem Package gg oder
// image/draw) durchzufuehren. Die Zeitmessung ueber die Variable 'ConvTime'
// ist in dieser Funktion realisiert.
func (buf *Buffer) Convert(src *image.RGBA) {
    var srcIdx, dstIdx int

    t1 := time.Now()
    buf.dstRect = src.Bounds()
    buf.strideLen = buf.dstRect.Dx() * bytesPerPixel

    for row:=buf.dstRect.Min.Y; row<buf.dstRect.Max.Y; row++ {
        srcIdx = (row-buf.dstRect.Min.Y)*src.Stride
        dstIdx = (row-buf.dstRect.Min.Y)*buf.strideLen
        for col:=buf.dstRect.Min.X; col<buf.dstRect.Max.X; col++ {
            buf.pixBuf[dstIdx+0]  =  src.Pix[srcIdx+2]
            buf.pixBuf[dstIdx+1]  =  src.Pix[srcIdx+1]
            buf.pixBuf[dstIdx+2]  =  src.Pix[srcIdx+0]
            srcIdx += 4
            dstIdx += bytesPerPixel
        }
    }
    ConvTime += time.Since(t1)
    NumConv++
}

/*
func (buf *Buffer) Convert(dstRect image.Rectangle, src *image.RGBA,
        srcPt image.Point) {
    var idx1, idx2, idx1Step, idx2Step int

    t1 := time.Now()
    buf.dstRect = dstRect
    idx1, idx1Step = 0, 4
    idx2, idx2Step = 0, bytesPerPixel
    for row:=srcPt.Y; row<srcPt.Y+dstRect.Dy(); row++ {
        for col:=srcPt.X; col<srcPt.X+dstRect.Dx(); col++ {
            idx1 = idx1Step * (row * src.Bounds().Dx() + col)

            buf.pixBuf[idx2+0]  =  src.Pix[idx1+2]
            buf.pixBuf[idx2+1]  =  src.Pix[idx1+1]
            buf.pixBuf[idx2+2]  =  src.Pix[idx1+0]

            idx2 += idx2Step
        }
    }
    ConvTime += time.Since(t1)
    NumConv++
}
*/
