package adatft

import (
	"errors"
	"image"
	"time"

	ili "github.com/stefan-muehlebach/adatft/ili9341"
)

type ChannelDir int

// Konstanten, die bei diesem Display halt einfach so sind wie sie sind.
const (
	toConv ChannelDir = iota
	toDisp
	dspSpeedHz       = 65_000_000
	numBuffers  int  = 2
	initMinimal bool = false
)

var (
	Width, Height int
)

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
	case "Rotate000":
		*rot = Rotate000
	case "Rotate090":
		*rot = Rotate090
	case "Rotate180":
		*rot = Rotate180
	case "Rotate270":
		*rot = Rotate270
	default:
		return errors.New("Unknown rotation: " + s)
	}
	return nil
}

// In RotationData sind nun alle von der Rotation abhängigen Einstellungen
// abgelegt. Es ist ein interner Datentyp, der wohl verwendet, aber nicht
// verändert werden kann.
type RotationData struct {
	iliParam      uint8
	calibDataFile string
	width, height int
}

var (
	rotDat = []RotationData{
		RotationData{0xe0, "Rotate000.json", ili.ILI9341_SIDE_A, ili.ILI9341_SIDE_B},
		RotationData{0x80, "Rotate090.json", ili.ILI9341_SIDE_B, ili.ILI9341_SIDE_A},
		RotationData{0x20, "Rotate180.json", ili.ILI9341_SIDE_A, ili.ILI9341_SIDE_B},
		RotationData{0x40, "Rotate270.json", ili.ILI9341_SIDE_B, ili.ILI9341_SIDE_A},
	}
)

type BufChanItem struct {
	img  *ILIImage
	rect image.Rectangle
}

// Dies ist der Datentyp, welche für die Verbindung zum ILI9341 via SPI
// steht. Im Wesentlichen handelt es sich dabei um den Filedescriptor auf
// das Device-File und um die Channels zu den Go-Routinen, welche
// die Konvertierung eines image.RGBA Bildes in ein ILI9341-konformes
// Format vornehmen und die Daten via SPI-Bus an den ILI9341 sendet.
type Display struct {
	dspi         DispInterface
	bufChan      []chan *BufChanItem
	staticImg    []*ILIImage
	staticImgIdx int
	lastImg      *ILIImage
	quitQ        chan bool
}

var (
	// ConvTime enthält die kumulierte Zeit, welche für das Konvertieren der
	// Bilder vom RGBA-Format in das 565-/666-Format verwendet wird.
	// Misst im Wesentlichen die aktive Zeit der Methode 'Convert'.
	ConvTime time.Duration
	// NumConv enthält die Anzahl Aufrufe von 'Convert'.
	NumConv int
	// DispTime enthält die kumulierte Zeit, welche für das Senden der Bilder
	// zum Display verwendet wird. Misst im Wesentlichen die aktive Zeit
	// der Methode 'drawBuffer'.
	DispTime time.Duration
	// NumDisp enthält die Anzahl Aufrufe von 'drawBuffer'.
	NumDisp int
	// PaintTime kann von der Applikation verwendet werden, um die kumulierte
	// Zeit zu erfassen, die von der Applikation selber zum Zeichnen des
	// Bildschirms verwendet wird.
	PaintTime time.Duration
	// In NumPaint kann die Applikation festhalten, wie oft der Bildschirm-
	// inhalt (oder Teile davon) neu gezeichnet wird.
	NumPaint int
)

// OpenDisplay initialisiert die Hardware, damit ein Zeichnen auf dem TFT
// erst möglich wird. Als einziger Parameter muss die gewünschte Rotation des
// Bildschirms angegeben werden.
// Ebenso werden Channels und Go-Routines erstellt, die für das asynchrone
// Anzeigen notwendig sind.
func OpenDisplay(rot RotationType) *Display {
	var dsp *Display

	Width = rotDat[rot].width
	Height = rotDat[rot].height
	calibDataFile = rotDat[rot].calibDataFile

	dsp = &Display{}
    if isRaspberry {
	    dsp.dspi = ili.Open(dspSpeedHz)
    } else {
	    dsp.dspi = ili.OpenDummy(dspSpeedHz)
    }
	dsp.dspi.Init([]any{false, rotDat[rot].iliParam})

	dsp.InitChannels()

	return dsp
}

// Schliesst die Verbindung zum ILI9341.
func (dsp *Display) Close() {
	close(dsp.bufChan[toDisp])
	<-dsp.quitQ
	dsp.staticImg[0].Clear()
	dsp.drawBuffer(dsp.staticImg[0])
	dsp.dspi.Close()
}

// Diese Routine baut die GO-Routinen fuer die parallelisierte Konvertierung
// und Anzeige auf und retourniert einen Channel, auf welchem die Pointer
// auf die RGBA-Images zur Anzeige gesendet werden.
func (dsp *Display) InitChannels() {
	var bufItem *BufChanItem
	var rect image.Rectangle

	dsp.bufChan = make([]chan *BufChanItem, 2)
	for i := 0; i < len(dsp.bufChan); i++ {
		dsp.bufChan[i] = make(chan *BufChanItem, numBuffers+1)
	}

	rect = image.Rect(0, 0, Width, Height)
	for i := 0; i < numBuffers; i++ {
		bufItem = &BufChanItem{}

		bufItem.img = NewILIImage(rect)
		bufItem.rect = rect
		dsp.bufChan[toConv] <- bufItem
	}
	dsp.staticImg = make([]*ILIImage, 2)
	dsp.staticImg[0] = NewILIImage(image.Rect(0, 0, Width, Height))
	dsp.staticImg[1] = NewILIImage(image.Rect(0, 0, Width, Height))
	dsp.staticImgIdx = 0

	dsp.lastImg = nil

	dsp.quitQ = make(chan bool)
	go dsp.displayer()
}

// func (dsp *Display) InitChannels() {
// 	var buf *ILIImage

// 	dsp.bufChan = make([]chan *ILIImage, 2)
// 	for i := 0; i < len(dsp.bufChan); i++ {
// 		dsp.bufChan[i] = make(chan *ILIImage, numBuffers+1)
// 	}

// 	for i := 0; i < numBuffers; i++ {
// 		buf = NewILIImage(image.Rect(0, 0, Width, Height))
// 		dsp.bufChan[toConv] <- buf
// 	}
// 	dsp.staticBuf = NewILIImage(image.Rect(0, 0, Width, Height))

// 	dsp.quitQ = make(chan bool)
// 	go dsp.displayer()
// }

func (dsp *Display) Bounds() image.Rectangle {
	return image.Rect(0, 0, Width, Height)
}

// Damit wird das Bild img auf dem Bildschirm dargestellt. Die Darstellung
// erfolgt synchron, d.h. die Methode wartet so lange, bis alle Bilddaten
// zum TFT gesendet wurden. Wichtig: img muss ein image.RGBA-Typ sein!
func (dsp *Display) DrawSync(img image.Image) error {
	// log.Printf("DrawSync(): img.Bounds(): %v", img.Bounds())
	dsp.staticImg[0].Convert(img.(*image.RGBA))
	rect := dsp.staticImg[0].Diff(dsp.staticImg[1])
	if rect.Empty() {
		return nil
		// rect = img.Bounds()
	}
	// log.Printf("DrawSync(): rect: %v", rect)
	dsp.drawBuffer(dsp.staticImg[0].SubImage(rect).(*ILIImage))
	dsp.staticImg[0], dsp.staticImg[1] = dsp.staticImg[1], dsp.staticImg[0]
	return nil
}

// Damit wird das Bild img auf dem Bildschirm dargestellt. Die Darstellung
// erfolgt asynchron, d.h. die Methode wartet nur, bis das Bild konvertiert
// wurde. Wichtig: img muss ein image.RGBA-Typ sein!
func (dsp *Display) Draw(img image.Image) error {
	var bufItem *BufChanItem

	bufItem = <-dsp.bufChan[toConv]
	bufItem.img.Convert(img.(*image.RGBA))
	if dsp.lastImg != nil {
		bufItem.rect = bufItem.img.Diff(dsp.lastImg)
		if bufItem.rect.Empty() {
			dsp.bufChan[toConv] <- bufItem
			return nil
		}
	} else {
		bufItem.rect = img.Bounds()
	}
	dsp.lastImg = bufItem.img
	dsp.bufChan[toDisp] <- bufItem
	return nil
}

// Mit dieser Funktion wird ein Bild auf dem TFT angezeigt.
func (dsp *Display) drawBuffer(img *ILIImage) {
	t1 := time.Now()
	start, end := img.Rect.Min, img.Rect.Max
	numBytes := img.Rect.Dx() * bytesPerPixel

	dsp.dspi.Cmd(ili.ILI9341_CASET)
	dsp.dspi.Data32(uint32((start.X << 16) | (end.X - 1)))
	dsp.dspi.Cmd(ili.ILI9341_PASET)
	dsp.dspi.Data32(uint32((start.Y << 16) | (end.Y - 1)))
	dsp.dspi.Cmd(ili.ILI9341_RAMWR)

	if numBytes == img.Stride {
		dsp.dspi.DataArray(img.Pix[:])
	} else {
		idx := 0
		for y := start.Y; y < end.Y; y++ {
			dsp.dspi.DataArray(img.Pix[idx : idx+numBytes])
			idx += img.Stride
		}
	}
	DispTime += time.Since(t1)
	NumDisp++
}

// Das ist die Funktion, welche im Hintergrund für die Anzeige der Bilder
// zuständig ist. Sie läuft als Go-Routine und wartet, bis über den Channel
// bufChan[toDisp] Bilder zur Anzeige eintreffen.
func (dsp *Display) displayer() {
	var bufItem *BufChanItem
	var ok bool

	for {
		if bufItem, ok = <-dsp.bufChan[toDisp]; !ok {
			break
		}
		dsp.drawBuffer(bufItem.img.SubImage(bufItem.rect).(*ILIImage))
		dsp.bufChan[toConv] <- bufItem
	}
	close(dsp.bufChan[toConv])
	dsp.quitQ <- true
}
