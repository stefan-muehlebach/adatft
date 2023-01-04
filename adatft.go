package adatft

// #cgo CFLAGS:  -I${SRCDIR}/tftlib -Ofast -fomit-frame-pointer -DPIXFMT_565
// #cgo LDFLAGS: -Wl,-rpath,${SRCDIR}/tftlib
// #cgo LDFLAGS: -L${SRCDIR}/tftlib
// #cgo LDFLAGS: -lili9341 -lm -lrt -lbcm_host
// #include "ili9341.h"
import "C"

import (
    "fmt"
    "image"
    "log" 
    "time"
    "unsafe"
)

//-----------------------------------------------------------------------------
//
// Konstanten, die bei diesem Display halt einfach so sind wie sie sind...
//
const (
    Width  int = C.TFT_WIDTH
    Height int = C.TFT_HEIGHT
)

//-----------------------------------------------------------------------------
//
// Ein paar oeffentliche Variablen:
//
// DebugFlag - Wenn 'true' werden viele Meldungen ausgegeben, mit welchen das
//             Funktionieren des Packages ueberprueft werden kann.
//             Default ist 'false'
// ConvTime  - Kumulierte Zeit, welche fuer das Konvertieren der Bilder
//             vom RGBA in das 565-Format verwendet wird. Misst im Wesentlichen
//             die aktive Zeit der Go-Routine 'converter()'.
// DispTime  - Kumulierte Zeit, welche fuer das Senden der Bilder zum
//             Display verwendet wird. Misst im Wesentlichen die aktive Zeit
//             der Go-Routine 'displayer()'.
//
var (
    DebugFlag bool = false
    ConvTime  time.Duration
    DispTime  time.Duration
)

//-----------------------------------------------------------------------------
//
// Type: 'TFT' --
//
type TFT struct {
    imgChan chan *image.RGBA
}

// Das sind die einzigen nach aussen sichtbaren Funktionen - das Interface
// gewissermassen.
//
func OpenTFT() (*TFT) {
    tft := &TFT{}
    tft.imgChan = initialize()
    return tft
}

func (tft *TFT) Rotate(rotate bool) {

}

func (tft *TFT) Show(img image.Image) {
    tft.imgChan <- img.(*image.RGBA)
    <- tft.imgChan
}

func (tft *TFT) Close() {
    tft.imgChan <- nil
    <- tft.imgChan
    close(tft.imgChan)
}

//-----------------------------------------------------------------------------
//
// Fuer das Pixelformat 666:
//
//type rgbType struct {
//    r, g, b uint8
//}

// Fuer das Pixelformat 565:
//
type rgbType struct {
    lo, hi uint8
}

//-----------------------------------------------------------------------------
//
// Type: 'device' --
//
// Dieser Record steht fuer das TFT Display und enthaelt folgende Felder:
//
//   fd            - Filedescriptor fuer das SPI device file
//
type device struct {
    fd C.int
}

func openDevice() (*device) {
    var err error

    dev := &device{}
    dev.fd, err = C.ILI_open()
    check("C.ILI_open()", err)

    _, err = C.ILI_init(dev.fd)
    check("C.ILI_init()", err)

    return dev
}

func (dev *device) close() {
    _, err := C.ILI_close(dev.fd)
    check("C.ILI_close()", err)
}

//-----------------------------------------------------------------------------
//
// Type: 'buffer' --
//
// Dieser Record wird fuer die Konvertierung der Bilddaten in ein von
// ILI9341 unterstuetztes Format verwendet.
//
//   pixBuf        - Ein Array mit den einzelnen Pixeln
//   width, height - Breite und Hoehe des Displays in Pixel
//   bufLen        - Anzahl Pixel des Display (Laenge des Arrays pixBuf)
//   bufSize       - Groesse von pixBuf in Bytes.
//   rotate        - Soll die Ausgabe um 180 Grad gedreht werden (Default: f)
//
type buffer struct {
    pixBuf []rgbType
    bufLen, bufSize int
    rotate bool
}

func createBuffer() (*buffer) {
    buf := &buffer{}

    buf.bufLen  = Width * Height
    buf.bufSize = int(unsafe.Sizeof(buf.pixBuf[0])) * buf.bufLen
    buf.rotate  = true

    buf.pixBuf  = make([]rgbType, buf.bufLen)

    return buf
}

func (buf *buffer) Rotate(rotate bool) {
    buf.rotate = rotate
}

//-----------------------------------------------------------------------------
//
// Mit dieser Funktion wird ein Bild vom RGBA-Format (image.RGBA) in das
// fuer den ILI9341 typische 666 oder (praeferiert) 565 Format konvertiert.
// Die Zeitmessung ueber die Variable 'ConvTime' ist in dieser Funktion
// realisiert.
//
func Convert(img *image.RGBA, buf *buffer) {
    var idx1, idx2, idx2Step int
    var r, gh, gl, b uint8

    t1 := time.Now()
    idx1 = 0
    if buf.rotate {
        idx2 = buf.bufLen-1
        idx2Step = -1
    } else {
        idx2 = 0
        idx2Step = +1
    }
    for y:=0; y<Height; y++ {
        for x:=0; x<Width; x++ {

            // Fuer Pixelformat 565
            //
            r  =  img.Pix[idx1+0] & 0xf8
            gh =  img.Pix[idx1+1] >> 5
            gl = (img.Pix[idx1+1] << 3) & 0xe0
            b  =  img.Pix[idx1+2] >> 3
            buf.pixBuf[idx2].lo = r | gh
            buf.pixBuf[idx2].hi = gl | b

            // Fuer Pixelformat 666
            //
//            buf.pixBuf[idx2].r = img.Pix[idx1+0] & 0xfc
//            buf.pixBuf[idx2].g = img.Pix[idx1+1] & 0xfc
//            buf.pixBuf[idx2].b = img.Pix[idx1+2] & 0xfc

            idx1 += 4
            idx2 += idx2Step
        }
    }
    t2 := time.Now()
    ConvTime += t2.Sub(t1)
}

//-----------------------------------------------------------------------------
//
// Das ist die GO-Routine, welche nur fur das Konvertieren der Bilddaten
// in ein von ILI9341 unterstuetztes Format zustaendig ist.
//
// Der Channel imgChan ist fuer den Austausch mit dem Animations-Threads
// vorgesehen. Ueber diesen Channel wird nur ein (1) Pointer auf ein
// RGBA-Image ausgetauscht, der als Token zwischen dem Animations-Thread
// und dem Konvertier-Thread verwendet wird.
//
// bufInChan ist der Channel für den Ruecklauf der Buffer-Pointer vom 
// Anzeige-Thread zum Konverter und bufOutChan fuer den Fluss der Buffer
// zum Konvertierungs-Thread (werden auch als Token verstanden).
//
func converter(imgChan chan *image.RGBA, bufInChan <-chan *buffer,
        bufOutChan chan<- *buffer) {
    var buf *buffer
    var ok bool
    var img *image.RGBA

    debugOut("[Conv] Starting\n")
    for true {
        debugOut("[Conv] Get image pointer from channel\n")
        img = <-imgChan
        if img == nil {
            break
        }
        debugOut("[Conv] Get buffer pointer from channel\n")
        buf = <-bufInChan

        debugOut("[Conv] Convert the pixels\n")
        Convert(img, buf)

        debugOut("[Conv] Put back the buffer pointer\n")
        bufOutChan <- buf
        debugOut("[Conv] Put back the image pointer\n")
        imgChan <- img
    }
    debugOut("[Conv] Close channel to displayer\n")
    close(bufOutChan)
    debugOut("[Conv] Wait for displayer to terminate\n")
    for true {
        _, ok = <-bufInChan
        if !ok { break }
    }
    debugOut("[Conv] Terminate\n")
    imgChan <- nil
}

//-----------------------------------------------------------------------------
//
// Mit dieser Funktion wird ein Bild auf dem TFT angezeigt.
//
func Display(dev *device, buf *buffer) {
    var pixBufPtr *C.uint8_t

    t1 := time.Now()
    pixBufPtr = (*C.uint8_t) (unsafe.Pointer(&buf.pixBuf[0]))
    C.ILI_cmd(dev.fd, 0x2a)
    C.ILI_data32(dev.fd, C.uint(Width-1))
    C.ILI_cmd(dev.fd, 0x2b)
    C.ILI_data32(dev.fd, C.uint(Height-1))
    C.ILI_cmd(dev.fd, 0x2c)
    C.ILI_data(dev.fd, pixBufPtr, C.uint(buf.bufSize))
    t2 := time.Now()
    DispTime += t2.Sub(t1)
}

//-----------------------------------------------------------------------------
//
// Dies ist die GO-Routine, welche 'nur' fuer das Anzeigen der Bilder
// zustaendig ist.
//
// Ueber den Channel bufInChan erhaelt diese Routine einen Pointer auf das
// naechste anzuzeigende Bild. Sobald das Bild dargestellt ist, wird der
// Pointer ueber den Channel bufOutChan wieder dem Konverter zugesandt.
//
const (
    dispRate time.Duration = 33 * time.Millisecond
)

func displayer(bufInChan <-chan *buffer, bufOutChan chan<- *buffer) {
    var dev *device
    var buf *buffer
    var ok bool

    debugOut("[Disp] Starting\n")
    dev = openDevice()

    for true {
        debugOut("[Disp] Get buffer pointer from channel\n")
        buf, ok = <-bufInChan
        if !ok {
            break
        }
        debugOut("[Disp] Display the content of the buffer\n")
        Display(dev, buf)
        
        debugOut("[Disp] Put buffer back to channel\n")
        bufOutChan <- buf
    }
    debugOut("[Disp] Close device\n")
    dev.close()
    debugOut("[Disp] Close channel to converter\n")
    close(bufOutChan)
    debugOut("[Disp] Terminate\n")
}

//-----------------------------------------------------------------------------
//
// Diese Routine baut die GO-Routinen fuer die parallelisierte Konvertierung
// und Anzeige auf und retourniert einen Channel, auf welchem die Pointer
// auf die RGBA-Images zur Anzeige gesendet werden.
//
const numBuffers int = 2

func initialize() (chan *image.RGBA) {
    var imgChan chan *image.RGBA
    var toConvChan, toDispChan chan *buffer
    var buf *buffer

    debugOut("[Init] Create image channel\n")
    imgChan = make(chan *image.RGBA)
    debugOut("[Init] Create buffer channels\n")
    toConvChan = make(chan *buffer, numBuffers+1)
    toDispChan = make(chan *buffer, numBuffers+1)

    debugOut("[Init] Start convert thread\n")
    go converter(imgChan, toConvChan, toDispChan)
    debugOut("[Init] Start display thread\n")
    go displayer(toDispChan, toConvChan)

    debugOut("[Init] Create %d buffers and inject them into channels\n",
            numBuffers)
    for i:=0; i<numBuffers; i++ {
        buf = createBuffer()
        toConvChan <- buf
    }

    debugOut("[Init] Return image channel\n")
    return imgChan
}

// Interne Check-Funktion, welche bei gravierenden Fehlern das Programm
// beendet.
//
func check(fncName string, err error) {
    if err != nil {
        log.Fatalf("%s: %s", fncName, C.GoString(C.ILI_errStr))
    }
}

// Funktioenchen fuer die Ausgabe von Debug-Meldungen. In C wuerde diese
// Routine als Praeprozessor-Makro erstellt, die in einem produktiven Code
// einfach weggelassen wird. Go kennt leider keine solchen Dinge, resp. nur
// unter Zuhilfenahme des C-Praeprozessors.
//
func debugOut(format string, a ...any) {
    if DebugFlag {
        fmt.Printf(format, a...)
    }
}

