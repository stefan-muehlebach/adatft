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
// Die Masse von img muessen den Massen des TFT-Displays
// (d.h. ILI9341_WIDTH x ILI9341_HEIGHT) entsprechen. Allfaellige Anpassungen
// sind vorgaengig mit anderen Funktionen (bspw. aus dem Package gg oder
// image/draw) durchzufuehren. Die Zeitmessung ueber die Variable 'ConvTime'
// ist in dieser Funktion realisiert.
func (buf *Buffer) Convert(dstRect image.Rectangle, img *image.RGBA,
        srcPts image.Point) {
    var idx1, idx2, idx1Step, idx2Step int

    t1 := time.Now()
    buf.dstRect = dstRect
    idx1, idx1Step = 0, 4
    idx2, idx2Step = 0, bytesPerPixel
    for row:=srcPts.Y; row<srcPts.Y+dstRect.Dy(); row++ {
        for col:=srcPts.X; col<srcPts.X+dstRect.Dx(); col++ {
            idx1 = idx1Step * (row * img.Bounds().Dx() + col)

            buf.pixBuf[idx2+0]  =  img.Pix[idx1+2]
            buf.pixBuf[idx2+1]  =  img.Pix[idx1+1]
            buf.pixBuf[idx2+2]  =  img.Pix[idx1+0]

            idx2 += idx2Step
        }
    }
    ConvTime += time.Since(t1)
    NumConv++
}

/*
func (buf *Buffer) Convert(img *image.RGBA) {
    var idx1, idx2, idx1Step, idx2Step int

    t1 := time.Now()
    idx1, idx1Step = 0, 4
    idx2, idx2Step = 0, bytesPerPixel
    for i := 0; i < buf.bufLen; i++ {
        // Für Pixelformat 565
        //val := img.Pix[idx1] & 0xf8
        //val |= img.Pix[idx1+1] >> 5
        //buf.pixBuf[idx2] = val
        //val  = (img.Pix[idx1+1] << 3) & 0xe0
        //val |= img.Pix[idx1+2] >> 3
        //buf.pixBuf[idx2+1] = val

        // Alternative für 565, läuft jedoch langsamer...
        //buf.pixBuf[idx2+0] = img.Pix[idx1] & 0xf8
        //buf.pixBuf[idx2+0] |= img.Pix[idx1+1] >> 5
        //buf.pixBuf[idx2+1] = (img.Pix[idx1+1] << 3) & 0xe0
        //buf.pixBuf[idx2+1] |= img.Pix[idx1+2] >> 3

        // Für das Pixelformat 666
        buf.pixBuf[idx2+0]  =  img.Pix[idx1+2]
        buf.pixBuf[idx2+1]  =  img.Pix[idx1+1]
        buf.pixBuf[idx2+2]  =  img.Pix[idx1+0]

        idx1 += idx1Step
        idx2 += idx2Step
    }
    ConvTime += time.Since(t1)
    NumConv++
}
*/


