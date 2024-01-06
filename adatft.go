/*
Mit diesem Package kann das 2.8'' TFT-Display mit Touchscreen von Adafruit
via Go angesprochen werden.

Das Package besteht im Wesentlichen aus 2 Teilen

 * Einer Sammlung von Typen und Funktionen für das Ansteuern des Bildschirms
 * und einem Teil für die Ansteuerung des Touchscreens.

Jeder Teil ist dabei in eine hardwarenahe Implementation und ein etwas
abstraketeres API unterteilt. Konkret:

 * ili9341.go, ili9341-spi.go: enthalten alles, was für die direkte
   Ansteuerung des Display-Chips benötigt wird.
 * display.go:

Interfaces: der SPI-Verbindung zum
Display-Chip ILI9341 und der SPI-Verbindung zum Touchscreen-Chip STMPE610.
Die Packages 'adatft/ili9341' und 'adatft/stmpe610' enthalten Typen und
Methoden für den direkten Zugang zu diesen Hardware-Komponenten
(low level API).

Darauf aufbauend enthält das Package die Typen 'Display' und 'Touch', welche
ein "high level API" anbieten.

*/
package adatft

import (
	"io/fs"
	"errors"
    "fmt"
    "log"
    "os"
    "path/filepath"
    //"periph.io/x/host/v3"
    "periph.io/x/host"
)

const (
    // Enthält den Namen des Verzeichnisses, in welchem alle
    // AdaTFT-spezifischen Konfigurationsdateien abgelegt werden.
    confDirName = "adatft"
)

var (
    // Enthält den absoluten Pfad des AdaTFT-spezifischen Konfigurations-
    // verzeichnisses.
    confDir string
)

// Damit wird die 'periph.io'-Umgebung initialisiert. Diese Funktion muss
// immer als erstes aufgerufen werden, noch bevor irgendwelche Devices
// geoeffnet werden.
func Init() {
    var userConfDir string
    var err error

    if _, err = host.Init(); err != nil {
        log.Fatalf("host.Init(): %v", err)
    }

    if userConfDir, err = os.UserConfigDir(); err != nil {
        log.Fatalf("os.UserConfigDir(): %v", err)
    }

    confDir = filepath.Join(userConfDir, confDirName)
    if err = os.MkdirAll(confDir, 0755); err != nil && errors.Is(err, fs.ErrNotExist) {
        log.Fatalf("os.MdirAll(): %v", err)
    }
}

func PrintStat() {
    fmt.Printf("total:\n")
    fmt.Printf("  %d frames\n", NumConv)
    fmt.Printf("buffer conversion:\n")
    fmt.Printf("  %v total\n", ConvTime)
    fmt.Printf("  %.3f ms / frame\n", float64(ConvTime.Milliseconds()) / float64(NumConv))
    fmt.Printf("application painting:\n")
    fmt.Printf("  %v total\n", PaintTime)
    fmt.Printf("  %.3f ms / frame\n", float64(PaintTime.Milliseconds()) / float64(NumPaint))
    fmt.Printf("sending to SPI:\n")
    fmt.Printf("  %v total\n", DispTime)
    fmt.Printf("  %.3f ms / frame\n", float64(DispTime.Milliseconds()) / float64(NumDisp))
}
