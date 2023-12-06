package adatft

import (
    "errors"
    "image"
    "time"
    . "github.com/stefan-muehlebach/adatft/ili9341"
)

type ChannelDir int

// Konstanten, die bei diesem Display halt einfach so sind wie sie sind...
const (
    toConv ChannelDir = iota
    toDisp
    dspSpeedHz        = 65_000_000
    numBuffers  int   = 2
    initMinimal bool  = false
)

var (
    Width, Height int
)

// Rotationsmoeglichkeiten des Display. Es gibt (logischerweise) 4
// Möglichkeiten das Display zu rotieren. Dies hat Auswirkungen auf die
// Initialisierung des Displays, auf die globalen Variablen Width und Height,
// auf die Konfigurationsdateien, in welchen die Daten für die Transformation
// von Touch-Koordinaten auf Display-Koordianten abgelegt sind, etc.
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
        return "rotate by 0 deg"
    case Rotate090:
        return "rotate by 90 deg"
    case Rotate180:
        return "rotate by 180 deg"
    case Rotate270:
        return "rotate by 270 deg"
    default:
        return "(unknown rotation)"
    }
}

func (rot *RotationType) Set(s string) error {
    switch s {
    case "rot000":
        *rot = Rotate000
    case "rot090":
        *rot = Rotate090
    case "rot180":
        *rot = Rotate180
    case "rot270":
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
    iliParam uint8
    calibDataFile string
    width, height int
}

var (
    rotDat = []RotationData{
        RotationData{0xe0, "rot090.json", ILI9341_SIDE_A, ILI9341_SIDE_B},
        RotationData{0x80, "rot180.json", ILI9341_SIDE_B, ILI9341_SIDE_A},
        RotationData{0x20, "rot270.json", ILI9341_SIDE_A, ILI9341_SIDE_B},
        RotationData{0x40, "rot000.json", ILI9341_SIDE_B, ILI9341_SIDE_A},
    }
)

// Dies ist der Datentyp, welche für die Verbindung zum ILI9341 via SPI
// steht. Im Wesentlichen handelt es sich dabei um den Filedescriptor auf
// das Device-File und um die Channels zu den Go-Routinen, welche
// a) die Konvertierung eines image.RGBA Bildes in ein ILI9341-konformes
//    Format vornimmt und
// b) die Daten via SPI-Bus an den ILI9341 sendet.
type Display struct {
    dspi    ILIInterface
    bufChan []chan *Buffer
    staticBuf *Buffer
    quitQ chan bool
}

var (
    // ConvTime enthält die kumulierte Zeit, welche für das Konvertieren der
    // Bilder vom RGBA-Format in das 565-/666-Format verwendet wird.
    // Misst im Wesentlichen die aktive Zeit der Methode 'Convert'.
    ConvTime  time.Duration
    // NumConv enthält die Anzahl Aufrufe von 'Convert'.
    NumConv   int
    // DispTime enthält die kumulierte Zeit, welche für das Senden der Bilder
    // zum Display verwendet wird. Misst im Wesentlichen die aktive Zeit
    // der Methode 'drawBuffer'.
    DispTime  time.Duration
    // NumDisp enthält die Anzahl Aufrufe von 'drawBuffer'.
    NumDisp   int
    PaintTime time.Duration
    NumPaint  int
)

func OpenDisplay(rot RotationType) (*Display) {
    var dsp *Display

    dsp = &Display{}
    dsp.dspi = OpenILI9341(dspSpeedHz)

    dsp.InitDisplay(rot)
    dsp.InitChannels()

    return dsp
}

// Schliesst die Verbindung zum ILI9341.
func (dsp *Display) Close() {
    close(dsp.bufChan[toDisp])
    <- dsp.quitQ
    dsp.staticBuf.Clear()
    dsp.staticBuf.dstRect = dsp.Bounds()
    dsp.drawBuffer(dsp.staticBuf)
    dsp.dspi.Close()
}

// Initialisiert die Werte im ILI9341. Der Inhalt dieser Funktion ist aus
// unzaehligen Beidspielen im Internet zusammengetragen und wurde durch
// "Trial und Error" ermittelt.
func (dsp *Display) InitDisplay(rot RotationType) {
    var posGamma []uint8 = []uint8{
        0x0f, 0x31, 0x2b, 0x0c, 0x0e, 0x08, 0x4e,
        0xf1,
        0x37, 0x07, 0x10, 0x03, 0x0e, 0x09, 0x00,
    }
    var negGamma []uint8 = []uint8{
        0x00, 0x0e, 0x14, 0x03, 0x11, 0x07, 0x31,
        0xc1,
        0x48, 0x08, 0x0f, 0x0c, 0x31, 0x36, 0x0f,
    }
/*
    var colorLut []uint8
    colorLut = make([]uint8, 128)
    for i := 0; i < 32; i++ {
        colorLut[i] = uint8(2*i)
    }
    for i := 0; i < 64; i++ {
        colorLut[i+32] = uint8(i)
    }
    for i := 0; i < 32; i++ {
        colorLut[i+96] = uint8(2*i)
    }
*/

    madctlParam := rotDat[rot].iliParam

    Width  = rotDat[rot].width
    Height = rotDat[rot].height

    calibDataFile = rotDat[rot].calibDataFile
/*
    macroGamma = make([]uint8, 16)
    for i := 0; i < 16; i++ {
        macroGamma[i] = 0x00
    }
    microGamma = make([]uint8, 64)
    for i := 0; i < 64; i++ {
        microGamma[i] = 0x00
    }
*/

    dsp.dspi.Cmd(ILI9341_DISPOFF) // Display On
    time.Sleep(125 * time.Millisecond)

    dsp.dspi.Cmd(ILI9341_SWRESET) // Reset the chip at the beginning
    time.Sleep(128 * time.Millisecond)

    if !initMinimal {
        dsp.dspi.Cmd(0xEF)
        dsp.dspi.DataArray([]byte{0x03, 0x80, 0x02})

        dsp.dspi.Cmd(ILI9341_PWCTRLB)
        dsp.dspi.DataArray([]byte{0x00, 0xc1, 0x30})
    
        dsp.dspi.Cmd(ILI9341_PWOSEQCTR)
        dsp.dspi.DataArray([]byte{0x64, 0x03, 0x12, 0x81})
    
        dsp.dspi.Cmd(ILI9341_DRVTICTRLA)
        dsp.dspi.DataArray([]byte{0x85, 0x00, 0x78})
    
        dsp.dspi.Cmd(ILI9341_PWCTRLA)
        dsp.dspi.DataArray([]byte{0x39, 0x2c, 0x00, 0x34, 0x02})
    
        dsp.dspi.Cmd(ILI9341_PMPRTCTR)
        dsp.dspi.Data8(0x20)
    
        dsp.dspi.Cmd(ILI9341_DRVTICTRLB)
        dsp.dspi.DataArray([]byte{0x00, 0x00})
    
        dsp.dspi.Cmd(ILI9341_PWCTR1)
        dsp.dspi.Data8(0x23)
    
        dsp.dspi.Cmd(ILI9341_PWCTR2)
        dsp.dspi.Data8(0x10)
    
        dsp.dspi.Cmd(ILI9341_VMCTR1)
        dsp.dspi.DataArray([]byte{0x3e, 0x28})
    
        dsp.dspi.Cmd(ILI9341_VMCTR2)
        dsp.dspi.Data8(0x86)
    }
    
    dsp.dspi.Cmd(ILI9341_MADCTL) // Memory Access Control
    dsp.dspi.Data8(madctlParam)

    if !initMinimal {
        dsp.dspi.Cmd(ILI9341_VSCRSADD)
        dsp.dspi.Data8(0x00)
    }

    dsp.dspi.Cmd(ILI9341_PIXFMT)
    dsp.dspi.Data8(0x66)        // Fuer das 666-Format
    //dsp.dspi.Data8(0x55)        // Fuer das 565-Format

    if !initMinimal {
        //dsp.dspi.Cmd(ILI9341_WRDISBV)
        //dsp.dspi.Data8(0x00)
    
        //dsp.dspi.Cmd(ILI9341_WRCTRLD)
        //dsp.dspi.Data8(0x2c)
    
        dsp.dspi.Cmd(ILI9341_FRMCTR1)
        dsp.dspi.DataArray([]byte{0x00, 0x18})
    
        dsp.dspi.Cmd(ILI9341_DFUNCTR)
        dsp.dspi.DataArray([]byte{0x08, 0x82, 0x27})
    }

    dsp.dspi.Cmd(ILI9341_GAMMA_3G) // Disable 3G (Gamma)
    dsp.dspi.Data8(0x00)

    dsp.dspi.Cmd(ILI9341_GAMMASET) // Set gamma correction to custom
    dsp.dspi.Data8(0x01)           // curve 1

    dsp.dspi.Cmd(ILI9341_GMCTRP1) // Positive Gamma Correction values
    dsp.dspi.DataArray(posGamma)

    dsp.dspi.Cmd(ILI9341_GMCTRN1) // Negative Gamma Correction values
    dsp.dspi.DataArray(negGamma)

    dsp.dspi.Cmd(ILI9341_SLPOUT) // Exit Sleep
    time.Sleep(125 * time.Millisecond)

    dsp.dspi.Cmd(ILI9341_DISPON) // Display On
    time.Sleep(125 * time.Millisecond)
}

// Diese Routine baut die GO-Routinen fuer die parallelisierte Konvertierung
// und Anzeige auf und retourniert einen Channel, auf welchem die Pointer
// auf die RGBA-Images zur Anzeige gesendet werden.
func (dsp *Display) InitChannels() {
    var buf *Buffer

    dsp.bufChan = make([]chan *Buffer, 2)
    for i := 0; i < len(dsp.bufChan); i++ {
        dsp.bufChan[i] = make(chan *Buffer, numBuffers+1)
    }

    for i := 0; i < numBuffers; i++ {
        buf = NewBuffer()
        dsp.bufChan[toConv] <- buf
    }
    dsp.staticBuf = NewBuffer()

    dsp.quitQ = make(chan bool)
    go dsp.displayer()
}

func (dsp *Display) Bounds() (image.Rectangle) {
    return image.Rect(0, 0, Width, Height)
}

func (dsp *Display) DrawSync(dstRect image.Rectangle, img image.Image,
        srcPts image.Point) (error) {
    dsp.staticBuf.Convert(dstRect, img.(*image.RGBA), srcPts)
    dsp.drawBuffer(dsp.staticBuf)
    return nil
}

func (dsp *Display) Draw(dstRect image.Rectangle, img image.Image,
        srcPts image.Point) (error) {
    var buf *Buffer

    buf = <-dsp.bufChan[toConv]
    buf.Convert(dstRect, img.(*image.RGBA), srcPts)
    dsp.bufChan[toDisp] <- buf
    return nil
}

// Mit dieser Funktion wird ein Bild auf dem TFT angezeigt.
func (dsp *Display) drawBuffer(buf *Buffer) {
    t1 := time.Now()

    start := buf.dstRect.Min
    end   := buf.dstRect.Max
    numBytes := buf.dstRect.Dx() * buf.dstRect.Dy() * bytesPerPixel

    dsp.dspi.Cmd(ILI9341_CASET)
    dsp.dspi.Data32(uint32((start.X<<16) | (end.X-1)))
    dsp.dspi.Cmd(ILI9341_PASET)
    dsp.dspi.Data32(uint32((start.Y<<16) | (end.Y-1)))
    dsp.dspi.Cmd(ILI9341_RAMWR)
    dsp.dspi.DataArray(buf.pixBuf[:numBytes])
    DispTime += time.Since(t1)
    NumDisp++
}

// Das ist die Funktion, welche im Hintergrund fuer die Anzeige der Bilder
// zustaendig ist. Sie wird als Go-Routine aufgerufen und wartet bis ueber
// den Channel bufChan[toDisp] Bilder zur Anzeige eintreffen.
func (dsp *Display) displayer() {
    var buf *Buffer
    var ok bool

    for {
        if buf, ok = <-dsp.bufChan[toDisp]; !ok {
            break
        }
        dsp.drawBuffer(buf)
        dsp.bufChan[toConv] <- buf
    }
    close(dsp.bufChan[toConv])
    dsp.quitQ <- true
}

