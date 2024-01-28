package adatft

import (
    "image"
    "time"

    ili "github.com/stefan-muehlebach/adatft/ili9341"
)

type channelDir int

// Konstanten, die bei diesem Display halt einfach so sind wie sie sind.
const (
    toConv channelDir = iota
    toDisp
    dspSpeedHz       = 65_000_000
    numBuffers  int  = 3
    initMinimal bool = false
)

var (
    Width, Height int
)


//type BufChanItem struct {
//    img  *ILIImage
//    rect image.Rectangle
//}

// Dies ist der Datentyp, welche für die Verbindung zum ILI9341 via SPI
// steht. Im Wesentlichen handelt es sich dabei um den Filedescriptor auf
// das Device-File und um die Channels zu den Go-Routinen, welche
// die Konvertierung eines image.RGBA Bildes in ein ILI9341-konformes
// Format vornehmen und die Daten via SPI-Bus an den ILI9341 sendet.
type Display struct {
    dspi         DispInterface
    //bufChan      []chan *BufChanItem
    imgChan      []chan *ILIImage
    syncImg      *ILIImage
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
    // der Methode 'sendImage'.
    DispTime time.Duration
    // NumDisp enthält die Anzahl Aufrufe von 'sendImage'.
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
    var rect image.Rectangle
    var img *ILIImage

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

    dsp.imgChan = make([]chan *ILIImage, 2)
    for i := 0; i < len(dsp.imgChan); i++ {
        dsp.imgChan[i] = make(chan *ILIImage, numBuffers+1)
    }

    rect = image.Rect(0, 0, Width, Height)
    for i := 0; i < numBuffers; i++ {
        img = NewILIImage(rect)
        dsp.imgChan[toConv] <- img
    }
    dsp.syncImg = NewILIImage(rect)

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

// func (dsp *Display) InitChannels() {
//     var buf *ILIImage

//     dsp.bufChan = make([]chan *ILIImage, 2)
//     for i := 0; i < len(dsp.bufChan); i++ {
//         dsp.bufChan[i] = make(chan *ILIImage, numBuffers+1)
//     }

//     for i := 0; i < numBuffers; i++ {
//         buf = NewILIImage(image.Rect(0, 0, Width, Height))
//         dsp.bufChan[toConv] <- buf
//     }
//     dsp.staticBuf = NewILIImage(image.Rect(0, 0, Width, Height))

//     dsp.quitQ = make(chan bool)
//     go dsp.displayer()
// }

func (dsp *Display) Bounds() image.Rectangle {
    return image.Rect(0, 0, Width, Height)
}

// Damit wird das Bild img auf dem Bildschirm dargestellt. Die Darstellung
// erfolgt synchron, d.h. die Methode wartet so lange, bis alle Bilddaten
// zum TFT gesendet wurden. Wichtig: img muss ein image.RGBA-Typ sein!
func (dsp *Display) DrawSync(img image.Image) error {
    dsp.syncImg.Convert(img.(*image.RGBA))
    dsp.sendImage(dsp.syncImg)
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

// Mit dieser Funktion wird ein Bild auf dem TFT angezeigt.
func (dsp *Display) sendImage(img *ILIImage) {
    t1 := time.Now()
    start, end := img.Rect.Min, img.Rect.Max
    bytesPerLine := img.Rect.Dx() * bytesPerPixel

    dsp.dspi.Cmd(ili.ILI9341_CASET)
    dsp.dspi.Data32(uint32((start.X << 16) | (end.X - 1)))
    dsp.dspi.Cmd(ili.ILI9341_PASET)
    dsp.dspi.Data32(uint32((start.Y << 16) | (end.Y - 1)))
    dsp.dspi.Cmd(ili.ILI9341_RAMWR)

    if bytesPerLine == img.Stride {
        dsp.dspi.DataArray(img.Pix[:img.Rect.Dy() * img.Stride])
    } else {
        idx := 0
        for y := start.Y; y < end.Y; y++ {
            dsp.dspi.DataArray(img.Pix[idx : idx+bytesPerLine])
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
    var img, lastImg *ILIImage
    var rect image.Rectangle
    var ok bool

    for {
        if img, ok = <-dsp.imgChan[toDisp]; !ok {
            break
        }
        if lastImg != nil {
            rect = img.Diff(lastImg)
            dsp.sendImage(img.SubImage(rect).(*ILIImage))
            dsp.imgChan[toConv] <- lastImg
        } else {
            dsp.sendImage(img)
        }
        lastImg = img
    }
    close(dsp.imgChan[toConv])
    dsp.quitQ <- true
}

