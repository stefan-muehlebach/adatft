package adatft

import (
    "image"
    "image/color"
//    "time"
)

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
    p := &ILIImage{}
    stride := r.Dx() * bytesPerPixel
    length := r.Dy() * stride
    p.Pix = make([]uint8, length)
    p.Stride = stride
    p.Rect = r
    p.Clear()
    return p
}

// ColorModel, Bounds und At werden vom Interface image.Image gefordert.
func (p *ILIImage) ColorModel() color.Model {
    return ILIModel
}
func (p *ILIImage) Bounds() image.Rectangle {
    return p.Rect
}
func (p *ILIImage) At(x, y int) color.Color {
    return p.ILIColorAt(x, y)
}

// Set wird ausserdem von draw.Image gefordert. Damit wird ein bestimmtes
// Pixel des Bildes auf den Farbwert c gesetzt.
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

// Ermittelt den Offset des Pixels mit Koordinaten x und y in p.Pix.
func (p *ILIImage) PixOffset(x, y int) int {
    return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*bytesPerPixel
}


func (p *ILIImage) ILIColorAt(x, y int) ILIColor {
    if !(image.Point{x, y}.In(p.Rect)) {
        return ILIColor{}
    }
    i := p.PixOffset(x, y)
    s := p.Pix[i : i+3 : i+3]
    return ILIColor{s[0], s[1], s[2]}
}

func (p *ILIImage) SetILIColor(x, y int, c ILIColor) {
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

// Mit Diff wird das kleinstmoegliche Rechteck ermittelt, welches alle
// Differenzen zwischen den Bildern p und img umschliesst..
func (p *ILIImage) Diff(img *ILIImage) image.Rectangle {
    var xMin, xMax, yMin, yMax int

    xMin, xMax = p.Rect.Dx(), 0
    yMin, yMax = p.Rect.Dy(), 0

    s := p.Pix[0 : len(p.Pix) : len(p.Pix)]
    d := img.Pix[0 : len(p.Pix) : len(p.Pix)]
    for i, pix := range s {
        if pix != d[i] {
            yMin = i / p.Stride
            yMax = yMin
            xMin = (i % p.Stride) / bytesPerPixel
            break
        }
    }
    if yMin > yMax {
        return image.Rectangle{}
    }
    
    s = p.Pix[yMin * p.Stride : len(p.Pix) : len(p.Pix)]
    d = img.Pix[yMin * img.Stride : len(p.Pix) : len(p.Pix)]
    for i := len(s)-1; i >= 0; i-- {
        if s[i] != d[i] {
            yMax = yMin + i / p.Stride + 1
            xMax = (i % p.Stride) / bytesPerPixel
            break
        }
    }
    if xMin > xMax {
        xMin, xMax = xMax, xMin
    }
    
Loop3:
    for x := 0; x < xMin; x++ {
        idx := x * bytesPerPixel
        for i := yMin*p.Stride; i < yMax*p.Stride; i += p.Stride {
            idxStart := idx + i
            idxEnd := idxStart + bytesPerPixel
            s := p.Pix[idxStart : idxEnd : idxEnd]
            d := img.Pix[idxStart : idxEnd : idxEnd]
            if s[0] != d[0] || s[1] != d[1] || s[2] != d[2] {
                xMin = x
                break Loop3
            }
        }
    }
    
Loop4:
    for x := p.Rect.Dx() - 1; x >= xMax; x-- {
        idx := x * bytesPerPixel
        for i := yMin*p.Stride; i < yMax*p.Stride; i += p.Stride {
            idxStart := idx + i
            idxEnd := idxStart + bytesPerPixel
            s := p.Pix[idxStart : idxEnd : idxEnd]
            d := img.Pix[idxStart : idxEnd : idxEnd]
            if s[0] != d[0] || s[1] != d[1] || s[2] != d[2] {
                xMax = x+1
                break Loop4
            }
        }
    }

    return image.Rect(xMin, yMin, xMax, yMax)
}

// Loescht das Bild hinter p, resp. setzt alle Bytes auf 0x00.
func (p *ILIImage) Clear() {
    for i := range p.Pix {
        p.Pix[i] = 0x00
    }
}

// Konvertiert die Bilddaten des Bildes hinter src (RGBA-Image) in ein
// ILI-spezifisches Bild. Dabei kann mit Rect (d.h. Bounds()) bestimmt werden
// welcher Bereich konvertiert werden soll.
func (p *ILIImage) Convert(src *image.RGBA) {
    var row, col int
    var srcBaseIdx, srcIdx, dstBaseIdx, dstIdx int

//    t1 := time.Now()
    ConvWatch.Start()

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
    ConvWatch.Stop()
//    ConvTime += time.Since(t1)
//    NumConv++
}

