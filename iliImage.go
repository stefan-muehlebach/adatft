package adatft

import (
	"image"
	"image/color"
)

// Diese Datenstruktur stellt ein Bild dar, welches auf dem TFT direkt
// dargestellt werden kann und implementiert alle Interfaces, welche Go
// fuer Bild-Typen kennt.
type ILIImage struct {
	Rect   image.Rectangle
	Stride int
	Pix    []uint8
}

func NewILIImage(r image.Rectangle) *ILIImage {
	p := &ILIImage{
		Rect:   r,
		Stride: r.Dx() * bytesPerPixel,
		Pix:    make([]uint8, r.Dx()*r.Dy()*bytesPerPixel),
	}
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

// Ermittelt den Offset des Pixels mit Koordinaten x und y in p.Pix.
func (p *ILIImage) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*bytesPerPixel
}

func (p *ILIImage) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(p.Rect)
	if r.Empty() {
		return &ILIImage{}
	}
	idx := p.PixOffset(r.Min.X, r.Min.Y)
	return &ILIImage{
		Rect:   r,
		Stride: p.Stride,
		Pix:    p.Pix[idx:],
	}
}

func (p *ILIImage) Opaque() bool {
	return true
}

// Loescht das Bild hinter p, resp. setzt alle Bytes auf 0x00.
func (p *ILIImage) Clear() {
	for i := range p.Pix {
		p.Pix[i] = 0x00
	}
}
