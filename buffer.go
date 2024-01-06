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
func NewBuffer(width, height int) *Buffer {
    buf := &Buffer{}
    buf.bufLen = width * height
    buf.bufSize = bytesPerPixel * buf.bufLen
    buf.strideLen = width * bytesPerPixel
    buf.pixBuf = make([]uint8, buf.bufSize)
    return buf
}

func (buf *Buffer) Clear() {
    for i, _ := range buf.pixBuf {
        buf.pixBuf[i] = 0x00
    }
}

// Mit dieser Funktion wird ein Bild vom RGBA-Format (image.RGBA) in das
// für den ILI9341 typische 666 (präferiert) oder 565 Format konvertiert.
// Die Grösse von src (Breite, Höhe) muss der Grösse des TFT-Displays
// (d.h. ILI9341_WIDTH x ILI9341_HEIGHT) entsprechen. Allfällige Anpassungen
// sind vorgängig mit anderen Funktionen (bspw. aus dem Package gg oder
// image/draw) durchzuführen. Die Zeitmessung über die Variablen 'ConvTime'
// und 'NumConv' ist in dieser Funktion realisiert.
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
