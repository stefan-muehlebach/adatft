//go:build pixfmt565

package adatft

import (
	"image"
	"image/color"
)

const (
	bytesPerPixel = 2
    pixfmt uint8 = 0x05
)

// Set wird ausserdem von draw.Image gefordert. Damit wird ein bestimmtes
// Pixel des Bildes auf den Farbwert c gesetzt.
func (p *ILIImage) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	idx := p.PixOffset(x, y)
	c1 := ILIModel.Convert(c).(ILIColor)
	s := p.Pix[idx : idx+bytesPerPixel : idx+bytesPerPixel]
	s[0] = c1.HB
	s[1] = c1.LB
}

func (p *ILIImage) ILIColorAt(x, y int) ILIColor {
	if !(image.Point{x, y}.In(p.Rect)) {
		return ILIColor{}
	}
	idx := p.PixOffset(x, y)
	s := p.Pix[idx : idx+bytesPerPixel : idx+bytesPerPixel]
	return ILIColor{s[0], s[1]}
}

func (p *ILIImage) SetILIColor(x, y int, c ILIColor) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	idx := p.PixOffset(x, y)
	s := p.Pix[idx : idx+bytesPerPixel : idx+bytesPerPixel]
	s[0] = c.HB
	s[1] = c.LB
}

// Mit Diff wird das kleinstmoegliche Rechteck ermittelt, welches alle
// Differenzen zwischen den Bildern p und img umschliesst..
func (p *ILIImage) Diff(img *ILIImage) image.Rectangle {
	var xMin, xMax, yMin, yMax int

	xMin, xMax = p.Rect.Dx(), 0
	yMin, yMax = p.Rect.Dy(), 0

	s := p.Pix[0:len(p.Pix):len(p.Pix)]
	d := img.Pix[0:len(p.Pix):len(p.Pix)]
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

	s = p.Pix[yMin*p.Stride : len(p.Pix) : len(p.Pix)]
	d = img.Pix[yMin*img.Stride : len(p.Pix) : len(p.Pix)]
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != d[i] {
			yMax = yMin + i/p.Stride + 1
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
		for i := yMin * p.Stride; i < yMax*p.Stride; i += p.Stride {
			idxStart := idx + i
			idxEnd := idxStart + bytesPerPixel
			s := p.Pix[idxStart:idxEnd:idxEnd]
			d := img.Pix[idxStart:idxEnd:idxEnd]
			if s[0] != d[0] || s[1] != d[1] {
				xMin = x
				break Loop3
			}
		}
	}

Loop4:
	for x := p.Rect.Dx() - 1; x >= xMax; x-- {
		idx := x * bytesPerPixel
		for i := yMin * p.Stride; i < yMax*p.Stride; i += p.Stride {
			idxStart := idx + i
			idxEnd := idxStart + bytesPerPixel
			s := p.Pix[idxStart:idxEnd:idxEnd]
			d := img.Pix[idxStart:idxEnd:idxEnd]
			if s[0] != d[0] || s[1] != d[1] {
				xMax = x + 1
				break Loop4
			}
		}
	}

	return image.Rect(xMin, yMin, xMax, yMax)
}

// Konvertiert die Bilddaten des Bildes hinter src (RGBA-Image) in ein
// ILI-spezifisches Bild. Dabei kann mit Rect (d.h. Bounds()) bestimmt werden
// welcher Bereich konvertiert werden soll.
func (p *ILIImage) Convert(src *image.RGBA) {
	var row, col int
	var srcBaseIdx, srcIdx, dstBaseIdx, dstIdx int

	ConvWatch.Start()

	srcBaseIdx = 0
	dstBaseIdx = src.Rect.Min.Y*p.Stride + src.Rect.Min.X*bytesPerPixel
	for row = src.Rect.Min.Y; row < src.Rect.Max.Y; row++ {
		srcIdx = srcBaseIdx
		dstIdx = dstBaseIdx

		for col = src.Rect.Min.X; col < src.Rect.Max.X; col++ {
			s := src.Pix[srcIdx : srcIdx+3 : srcIdx+3]
			d := p.Pix[dstIdx : dstIdx+bytesPerPixel : dstIdx+bytesPerPixel]
            r := s[0] & 0xF8
            g := s[1] & 0xFC
            b := s[2] & 0xF8
			d[0] = (r)      | (g >> 5)
			d[1] = (g << 3) | (b >> 3)
			srcIdx += 4
			dstIdx += bytesPerPixel
		}
		srcBaseIdx += src.Stride
		dstBaseIdx += p.Stride
	}
	ConvWatch.Stop()
}
