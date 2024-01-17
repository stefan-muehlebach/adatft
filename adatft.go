// Alles für die Ansteuerung des 2.8" TFT-Display/Touchscreen von AdaFruit.
//
// Mit diesem Package kann das 2.8” TFT-Display mit Touchscreen von Adafruit
// via Go angesprochen werden.
// 
// Das Package besteht im Wesentlichen aus 2 Teilen
// 
// - Einer Sammlung von Typen und Funktionen für das Ansteuern des Bildschirms.
// - Und einem Teil für die Ansteuerung des Touchscreens.
//
// Jeder Teil ist dabei in eine hardwarenahe Implementation und ein etwas
// abstraketeres API unterteilt. Konkret:
// 
// - ili9341/ili9341.go: enthält alles, was für die Ansteuerung dieses
//   konkreten Chips via SPI notwendig ist.
// - display.go: enthält den Typ 'Display', der ein "high level API" anbietet.
//
// - stmpe610/stmpe610.go: "low level API" für die Ansteuerung des Touchscreens
//   über diesen Chip via SPI.
// - touch.go: enthält den Typ 'Touch', der ein "high level API" anbietet.
package adatft

import (
    "errors"
    "fmt"
    "io/fs"
    "log"
    "os"
    "path/filepath"

    //"periph.io/x/host/v3"
    "periph.io/x/conn/driver/driverreg"
    "periph.io/x/host"
)

const (
    // Enthält den Namen des aktuellen Packages. Dieser Name wird u.a. fuer
    // Verzeichnisse verwendet, die applikationsspezifische Konfigurations-
    // und Logdateien enthalten.
    applName = "adatft"
)

var (
    // Enthält den absoluten Pfad des AdaTFT-spezifischen Konfigurations-
    // verzeichnisses.
    confDir string

    // In dieser Variable wird festgehalten, ob es sich bei der aktuellen
    // Plattform um einen RaspberryPi oder um einen PC handelt.
    isRaspberry bool

    // Der Logger, welcher von diesem Package verwendet wird.
    adalog *log.Logger
)

// Damit wird die 'periph.io'-Umgebung initialisiert. Diese Funktion muss
// immer als erstes aufgerufen werden, noch bevor irgendwelche Devices
// geoeffnet werden.
func init() {
    var userConfDir, userLogDir, logDir, logFile string
    var fh *os.File
    var driverStates *driverreg.State
    var err error

    // Erstellt ggf. Konfigurations- und Log-Verzeichnisse.
    if userConfDir, err = os.UserConfigDir(); err != nil {
        log.Fatalf("os.UserConfigDir(): %v", err)
    }
    confDir = filepath.Join(userConfDir, applName)
    if err = os.MkdirAll(confDir, 0755); err != nil &&
            errors.Is(err, fs.ErrNotExist) {
        log.Fatalf("os.MdirAll(): %v", err)
    }

    if userLogDir, err = os.UserCacheDir(); err != nil {
        log.Fatalf("os.UserCacheDir(): %v", err)
    }
    logDir = filepath.Join(userLogDir, applName)
    if err = os.MkdirAll(logDir, 0755); err != nil &&
            errors.Is(err, fs.ErrNotExist) {
        log.Fatalf("os.MkdirAll(): %v", err)
    }
    logFile = filepath.Join(logDir, "adatft.log")
    if fh, err = os.OpenFile(logFile, os.O_RDWR | os.O_CREATE, 0644);
            err != nil {
        log.Fatalf("os.OpenFile(): %v", err)
    }
    adalog = log.New(fh, "", log.Ltime | log.Lshortfile)
    
    // Initialisiere die 'periph.io'-Umgebung und halte fest, ob wir
    // auf einem rechten RaspberryPi laufen.
    isRaspberry = false
    if driverStates, err = host.Init(); err != nil {
        log.Fatalf("host.Init(): %v", err)
    }
    for _, drv := range driverStates.Loaded {
        if drv.String() == "rpi" {
            isRaspberry = true
            break
        }
    }
}

// Gibt eine Reihe von Messdaten aus, mit denen die Performance der Umgebung
// eingeschaetzt werden kann. Drei Bereiche werden gemessen:
//   - Die applikatorische Zeit (NumPaint, PaintTime).
//   - Die Zeit, welche fuer die Konvertierung der Bilder ins ILI-spezifische
//     Format benoetigt wird (NumConf, ConvTime).
//   - Die Zeit, welche fuer die Darstellung der Bilder auf dem TFT benoetigt
//     wird (NumDisp, DispTime).
// Als Daumenregel gilt: wenn die applikatorische Zeit pro Frame
// (PaintTime / NumPaint) groesser ist als die Zeit, welche fuer die
// Darstellung benoetigt wird (DispTime / NumDisp), dann besteht Bedarf nach
// Optimierung.
func PrintStat() {
    fmt.Printf("total:\n")
    fmt.Printf("  %d frames\n", NumConv)
    if NumPaint != 0 {
        fmt.Printf("application painting:\n")
        fmt.Printf("  %v total\n", PaintTime)
        fmt.Printf("  %.3f ms / frame\n", float64(PaintTime.Milliseconds())/float64(NumPaint))
    }
    fmt.Printf("buffer conversion:\n")
    fmt.Printf("  %v total\n", ConvTime)
    fmt.Printf("  %.3f ms / frame\n", float64(ConvTime.Milliseconds())/float64(NumConv))
    fmt.Printf("sending to SPI:\n")
    fmt.Printf("  %v total\n", DispTime)
    fmt.Printf("  %.3f ms / frame\n", float64(DispTime.Milliseconds())/float64(NumDisp))
}
