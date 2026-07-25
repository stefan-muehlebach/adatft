package adatft

import (
	"errors"
	"image"

	"github.com/stefan-muehlebach/adatft/iliimg"
	"periph.io/x/conn/v3/physic"
	//hw "github.com/stefan-muehlebach/adatft/ili9341"
	//hw "github.com/stefan-muehlebach/adatft/hx8357"
	hw "github.com/stefan-muehlebach/adatft/gc9503cv"
	"github.com/stefan-muehlebach/gg/geom"
)

//----------------------------------------------------------------------------

// Es ist vorstellbar, den TFT-Display ueber andere Schnittstellen als SPI
// anzusteuern und andere Chips fuer die Ansteuerung als den ILI9341 zu
// verwenden. Dieses Interface beschreibt alle Methoden, welche
// von einer Display-Ansteuerung implementiert werden muessen. Da bisher
// nur der ILI9341 via SPI angesteuert wurde, ist es sehr wahrscheinlich,
// dass dieses Interface noch stark angepasst werden muss.
type DispInterface interface {
	// Damit die Initialisierung so flexibel wie moeglich bleibt, wird der
	// Init-Methode ein Slice von beliebigen Parametern uebergeben. Wie die
	// Werte interpretiert werden, ist Interface-spezifisch.
	Init(rot byte) (w, h int)
	Init2(rot byte) (geom.Rectangle, geom.Rectangle)

	// Schliesst die Verbindung zum ILI-Chip und gibt alle Ressourcen in
	// Zusammenhang mit dieser Verbindung frei.
	Close()

	// Sendet einen Befehl (Command) zum Chip. Das ist in der Regel ein
	// 8 Bit Wert.
	Cmd(cmd uint8)

	// Sendet 8 Bit als Daten zum Chip. In den meisten Fällen ist dies ein
	// Argument eines Befehls, der vorgängig via Cmd gesendet wird.
	Data8(val uint8)

	// Analog Data8, jedoch mit 32 Bit Daten.
	Data32(val uint32)

	// Der gesamte Slice buf wird gesendet.
	DataArray(buf []byte)

	// Diese Methode schliesslich dient dem Anzeigen eines Images auf dem
	// Display.
	Send(img *iliimg.ILIImage)
}

//----------------------------------------------------------------------------

const (
	numBuffers int = 3
)

var (
	dspSpeed      physic.Frequency = 40 * physic.MegaHertz
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
	imgChan            []chan *iliimg.ILIImage
	syncImg, activeImg *iliimg.ILIImage
	quitQ              chan bool
	rect, drawRect     image.Rectangle
}

// OpenDisplay initialisiert die Hardware, damit ein Zeichnen auf dem TFT
// erst möglich wird. Als einziger Parameter muss die gewünschte Rotation des
// Bildschirms angegeben werden.
// Ebenso werden Channels und Go-Routines erstellt, die für das asynchrone
// Anzeigen notwendig sind.
func OpenDisplay(rot RotationType) *Display {
	dsp := &Display{}
	if isRaspberry {
		// for gc9503cv
		dsp.dspi = hw.Open2("", "", "", "", "")
		// for spi tft screens
		//dsp.dspi = hw.Open(dspSpeed)
	} else {
		dsp.dspi = hw.OpenDummy(dspSpeed)
	}
	// for gc9503cv
	dispBounds, drawBounds := dsp.dspi.Init2(byte(rot))
	// for spi tft screens
	//Width, Height = dsp.dspi.Init(byte(rot))

	dsp.imgChan = make([]chan *iliimg.ILIImage, numChannels)
	for i := toConv; i < numChannels; i++ {
		dsp.imgChan[i] = make(chan *iliimg.ILIImage, numBuffers+1)
	}

	dsp.rect = dispBounds.Int()
	dsp.drawRect = drawBounds.Int()
	for i := 0; i < numBuffers; i++ {
		img := iliimg.NewILIImage(dsp.rect)
		img.SetLSBFirst()
		dsp.imgChan[toConv] <- img
	}
	dsp.syncImg = iliimg.NewILIImage(dsp.rect)
	dsp.syncImg.SetLSBFirst()
	dsp.activeImg = iliimg.NewILIImage(dsp.rect)
	dsp.activeImg.SetLSBFirst()

	dsp.dspi.Send(dsp.activeImg)

	dsp.quitQ = make(chan bool)
	go dsp.displayer()

	return dsp
}

// Schliesst die Verbindung zum ILI9341.
func (dsp *Display) Close() {
	close(dsp.imgChan[toDisp])
	<-dsp.quitQ
	dsp.syncImg.Clear()
	dsp.dspi.Send(dsp.syncImg)
	dsp.dspi.Close()
}

// Die Methode Bounds kann verwendet werden, um die Breite und Hoehe des
// Displays zu ermitteln. Das Resultat ist ein image.Rectangle Wert, d.h.
// es wird kein geom.Rectangle Wert zurueckgegeben, da dieser float64-Werte
// enthaelt.
func (dsp *Display) Bounds() image.Rectangle {
	return dsp.rect
}

func (dsp *Display) DrawBounds() image.Rectangle {
	return dsp.drawRect
}

// Damit wird das Bild img auf dem Bildschirm dargestellt. Die Darstellung
// erfolgt synchron, d.h. die Methode wartet so lange, bis alle Bilddaten
// zum TFT gesendet wurden. Wichtig: img muss ein image.RGBA-Typ sein!
func (dsp *Display) DrawSync(img image.Image) error {
	dsp.syncImg.Convert(img.(*image.RGBA))
	rect := dsp.activeImg.Diff(dsp.syncImg)
	dsp.dspi.Send(dsp.syncImg.SubImage(rect).(*iliimg.ILIImage))
	dsp.activeImg, dsp.syncImg = dsp.syncImg, dsp.activeImg
	return nil
}

// Damit wird das Bild img auf dem Bildschirm dargestellt. Die Darstellung
// erfolgt asynchron, d.h. die Methode wartet nur, bis das Bild konvertiert
// wurde. Wichtig: img muss ein image.RGBA-Typ sein!
func (dsp *Display) Draw(img image.Image) error {
	var iliImg *iliimg.ILIImage

	iliImg = <-dsp.imgChan[toConv]
	iliImg.Convert(img.(*image.RGBA))
	dsp.imgChan[toDisp] <- iliImg
	return nil
}

// Mit dieser Funktion wird ein Bild im ILI-Format auf dem TFT dargestellt,
// d.h. die Bilddaten werden via SPI-Bus zum ILI9341 gesendet.
/*
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
		idx0 := 0
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			idx1 := idx0 + bytesPerLine
			dsp.dspi.DataArray(img.Pix[idx0 : idx1 : idx1])
			idx0 += img.Stride
		}
	}
	DispWatch.Stop()
}
*/

// Das ist die Funktion, welche im Hintergrund für die Anzeige der Bilder
// zuständig ist. Sie läuft als Go-Routine und wartet, bis über den Channel
// bufChan[toDisp] Bilder zur Anzeige eintreffen.
func (dsp *Display) displayer() {
	var img *iliimg.ILIImage
	var rect image.Rectangle
	var ok bool

	for {
		if img, ok = <-dsp.imgChan[toDisp]; !ok {
			break
		}
		rect = dsp.activeImg.Diff(img)
		if !rect.Empty() {
			DispWatch.Start()
			dsp.dspi.Send(img.SubImage(rect).(*iliimg.ILIImage))
			DispWatch.Stop()
			dsp.activeImg, img = img, dsp.activeImg
		}
		dsp.imgChan[toConv] <- img
	}
	close(dsp.imgChan[toConv])
	dsp.quitQ <- true
}

// Rotationsmöglichkeiten des Displays. Es gibt (logischerweise) 4
// Möglichkeiten das Display zu rotieren. Dies hat Auswirkungen auf die
// Initialisierung des Displays, auf die globalen Variablen Width und Height
// und auf die Konfigurationsdateien, in welchen die Daten für die
// Transformation von Touch-Koordinaten auf Display-Koordianten abgelegt
// sind, etc.
type RotationType int

const (
	Rotate000 RotationType = iota
	Rotate090
	Rotate180
	Rotate270
)

func (rot RotationType) String() string {
	switch rot {
	case Rotate000:
		return "Rotate000"
	case Rotate090:
		return "Rotate090"
	case Rotate180:
		return "Rotate180"
	case Rotate270:
		return "Rotate270"
	default:
		return "(unknown rotation)"
	}
}

func (rot *RotationType) Set(s string) error {
	switch s {
	case "Rotate000", "000", "0":
		*rot = Rotate000
	case "Rotate090", "090", "90":
		*rot = Rotate090
	case "Rotate180", "180":
		*rot = Rotate180
	case "Rotate270", "270":
		*rot = Rotate270
	default:
		return errors.New("Unknown rotation: " + s)
	}
	return nil
}
