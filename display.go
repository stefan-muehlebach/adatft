package adatft

import (
	"image"
	"periph.io/x/conn/v3/physic"
	// hw "github.com/stefan-muehlebach/adatft/ili9341"
	hw "github.com/stefan-muehlebach/adatft/hx8357"
)

const (
	numBuffers  int  = 3
	initMinimal bool = false
)

var (
	SPISpeedHz physic.Frequency = 45_000_000
	//SPISpeedHz physic.Frequency = 65_000_000
	Width, Height int
)

// Es gibt zwei Channels, welche fuer darzustellende Bilder verwendet
// werden: einen der vom Converter zur Displayer fuehrt und einen, der vom
// Displayer wieder zurueck zum Converter fuehrt. Mit dem Typ channelDir und
// den Konstanten toConv und toDisp werden die beiden Channels angesprochen.
type channelDir int

const (
	toConv channelDir = iota
	toDisp
	numChannels
)

// Dies ist der Datentyp, welche für die Verbindung zum ILI9341 via SPI
// steht. Im Wesentlichen handelt es sich dabei um den Filedescriptor auf
// das Device-File und um die Channels zu den Go-Routinen, welche
// die Konvertierung eines image.RGBA Bildes in ein ILI9341-konformes
// Format vornehmen und die Daten via SPI-Bus an den ILI9341 sendet.
type Display struct {
	dspi               DispInterface
	imgChan            []chan *ILIImage
	syncImg, activeImg *ILIImage
	quitQ              chan bool
}

// OpenDisplay initialisiert die Hardware, damit ein Zeichnen auf dem TFT
// erst möglich wird. Als einziger Parameter muss die gewünschte Rotation des
// Bildschirms angegeben werden.
// Ebenso werden Channels und Go-Routines erstellt, die für das asynchrone
// Anzeigen notwendig sind.
func OpenDisplay(rot RotationType) *Display {
	var dsp *Display
	var rect image.Rectangle
	var img *ILIImage

	Width = rotDat[rot].width
	Height = rotDat[rot].height
	//calibDataFile = rotDat[rot].calibDataFile

	dsp = &Display{}
	if isRaspberry {
		dsp.dspi = hw.Open(SPISpeedHz)
	} else {
		dsp.dspi = hw.OpenDummy(SPISpeedHz)
	}
	dsp.dspi.Init([]any{false, rotDat[rot].madctlParam, pixfmt})

	dsp.imgChan = make([]chan *ILIImage, numChannels)
	for i := 0; i < len(dsp.imgChan); i++ {
		dsp.imgChan[i] = make(chan *ILIImage, numBuffers+1)
	}

	rect = image.Rect(0, 0, Width, Height)
	for i := 0; i < numBuffers; i++ {
		img = NewILIImage(rect)
		dsp.imgChan[toConv] <- img
	}
	dsp.syncImg = NewILIImage(rect)
	dsp.activeImg = NewILIImage(rect)
	dsp.sendImage(dsp.activeImg)

	dsp.quitQ = make(chan bool)
	go dsp.displayer()

	return dsp
}

// Schliesst die Verbindung zum ILI9341.
func (dsp *Display) Close() {
	close(dsp.imgChan[toDisp])
	<-dsp.quitQ
	dsp.syncImg.Clear()
	dsp.sendImage(dsp.syncImg)
	dsp.dspi.Close()
}

// Die Methode Bounds kann verwendet werden, um die Breite und Hoehe des
// Displays zu ermitteln. Das Resultat ist ein image.Rectangle Wert, d.h.
// es wird kein geom.Rectangle Wert zurueckgegeben, da dieser float64-Werte
// enthaelt.
func (dsp *Display) Bounds() image.Rectangle {
	return image.Rect(0, 0, Width, Height)
}

// Damit wird das Bild img auf dem Bildschirm dargestellt. Die Darstellung
// erfolgt synchron, d.h. die Methode wartet so lange, bis alle Bilddaten
// zum TFT gesendet wurden. Wichtig: img muss ein image.RGBA-Typ sein!
func (dsp *Display) DrawSync(img image.Image) error {
	dsp.syncImg.Convert(img.(*image.RGBA))
	rect := dsp.activeImg.Diff(dsp.syncImg)
	dsp.sendImage(dsp.syncImg.SubImage(rect).(*ILIImage))
	dsp.activeImg, dsp.syncImg = dsp.syncImg, dsp.activeImg
	return nil
}

// Damit wird das Bild img auf dem Bildschirm dargestellt. Die Darstellung
// erfolgt asynchron, d.h. die Methode wartet nur, bis das Bild konvertiert
// wurde. Wichtig: img muss ein image.RGBA-Typ sein!
func (dsp *Display) Draw(img image.Image) error {
	var iliImg *ILIImage

	iliImg = <-dsp.imgChan[toConv]
	iliImg.Convert(img.(*image.RGBA))
	dsp.imgChan[toDisp] <- iliImg
	return nil
}

// Mit dieser Funktion wird ein Bild im ILI-Format auf dem TFT dargestellt,
// d.h. die Bilddaten werden via SPI-Bus zum ILI9341 gesendet.
func (dsp *Display) sendImage(img *ILIImage) {
	var len int

	DispWatch.Start()
	rect := img.Rect
	bytesPerLine := rect.Dx() * bytesPerPixel

	dsp.dspi.Cmd(hw.PASET)
	dsp.dspi.Data32(uint32((rect.Min.Y << 16) | (rect.Max.Y - 1)))
	dsp.dspi.Cmd(hw.CASET)
	dsp.dspi.Data32(uint32((rect.Min.X << 16) | (rect.Max.X - 1)))
	dsp.dspi.Cmd(hw.RAMWR)

	if bytesPerLine == img.Stride {
		len = rect.Dy() * img.Stride
		dsp.dspi.DataArray(img.Pix[:len:len])
	} else {
		idx := 0
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			dsp.dspi.DataArray(img.Pix[idx : idx+bytesPerLine : idx+bytesPerLine])
			idx += img.Stride
		}
	}
	DispWatch.Stop()
}

// Das ist die Funktion, welche im Hintergrund für die Anzeige der Bilder
// zuständig ist. Sie läuft als Go-Routine und wartet, bis über den Channel
// bufChan[toDisp] Bilder zur Anzeige eintreffen.
func (dsp *Display) displayer() {
	var img *ILIImage
	var rect image.Rectangle
	var ok bool

	for {
		if img, ok = <-dsp.imgChan[toDisp]; !ok {
			break
		}
		rect = dsp.activeImg.Diff(img)
		if !rect.Empty() {
			dsp.sendImage(img.SubImage(rect).(*ILIImage))
			dsp.activeImg, img = img, dsp.activeImg
		}
		dsp.imgChan[toConv] <- img
	}
	close(dsp.imgChan[toConv])
	dsp.quitQ <- true
}
