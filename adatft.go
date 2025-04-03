// Mit diesem Package kann das 2.8” TFT-Display mit Touchscreen von Adafruit
// via Go angesprochen werden.
//
// Das Package besteht im Wesentlichen aus 2 Teilen
//
//   - Einer Sammlung von Typen und Funktionen für das Ansteuern des
//     Bildschirms.
//   - Und einem Teil für die Ansteuerung des Touchscreens.
//
// Jeder Teil ist dabei in eine hardwarenahe Implementation und ein etwas
// abstraketeres API unterteilt. Konkret:
//
//   - ili9341/ili9341.go: enthält alles, was für die Ansteuerung dieses
//     konkreten Chips via SPI notwendig ist.
//
//   - hx8357/hx8357.go: analoge Sammlung für den leistungsfähigeren Chip.
//
//   - stmpe610/stmpe610.go: "low level API" für die Ansteuerung des
//     Touchscreens über diesen Chip via SPI.
//
//   - display.go: enthält den Typ 'Display', der ein "high level API" anbietet.
//
//   - touch.go: enthält den Typ 'Touch', der ein "high level API" anbietet.
package adatft

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/host/v3"
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

// Damit wird die 'periph.io'-Umgebung und diverse globale Variablen
// initialisiert.
func Init() {
	var userConfDir, userLogDir, logDir, logFile string
	var fh *os.File
	var driverStates *driverreg.State
	var err error

	// Erstellt Verzeichnisse fuer Konfigurations- und Log-Dateien (falls
	// noch nicht vorhanden).
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
	if fh, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE, 0644); err != nil {
		log.Fatalf("os.OpenFile(): %v", err)
	}
	adalog = log.New(fh, "", log.Ltime|log.Lshortfile)

	// Initialisiere die 'periph.io'-Umgebung und halte fest, ob wir
	// auf einem echten RaspberryPi laufen.
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

func init() {
	Init()
}
