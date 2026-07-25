package gc9503cv

import (
	"github.com/stefan-muehlebach/adatft/iliimg"
	"github.com/stefan-muehlebach/gg/geom"
	"periph.io/x/conn/v3/physic"
)

// Dies ist der Datentyp, welche für die Verbindung zum ILI9341 via SPI
// steht. Im Wesentlichen handelt es sich dabei um den Filedescriptor auf
// das Device-File für die SPI-Verbindung und den Pin, welcher für die
// Command/Data-Leitung verwendet wird.
type GC9503CVDummy struct {
}

// Damit wird die Verbindung zum GC9503CV geöffnet. Die Initialisierung des
// Chips wird in einer separaten Funktion (Init()) durchgeführt!
func OpenDummy(speedHz physic.Frequency) *GC9503CVDummy {
	d := &GC9503CVDummy{}

	return d
}

// Schliesst die Verbindung zum GC9503CV.
func (d *GC9503CVDummy) Close() {
	// err := d.spi.Close()
	// check("Close(): error in spi.Close()", err)
}

// Führt die Initialisierung des Chips durch. initParams ist ein Slice
// von Hardware-spezifischen Einstellungen. Beim GC9503CV sind dies:
//
//	{ initMinimal, madctlParam }
func (d *GC9503CVDummy) Init(rot byte) (w, h int) {
	return 0, 0
}

func (d *GC9503CVDummy) Init2(rot byte) (geom.Rectangle, geom.Rectangle) {
	return geom.Rectangle{}, geom.Rectangle{}
}

// Sende den Befehl in 'cmd' zum GC9503CV.
func (d *GC9503CVDummy) Cmd(cmd uint8) {
}

// Sende die Daten in 'value' (1 Byte) als Datenpaket zum GC9503CV.
func (d *GC9503CVDummy) Data8(value uint8) {
}

// Sende die Daten in 'value' (4 Bytes) als Datenpaket zum GC9503CV.
func (d *GC9503CVDummy) Data32(value uint32) {
}

// Sendet die Daten aus dem Slice 'buf' als Daten zum GC9503CV. Dies ist bloss
// eine Hilfsfunktion, damit das Senden von Daten aus einem Slice einfacher
// aufzurufen ist und die ganzen Konvertierungen nicht im Hauptprogramm
// sichtbar sind.
func (d *GC9503CVDummy) DataArray(buf []byte) {
}

func (d *GC9503CVDummy) Send(img *iliimg.ILIImage) {
}
