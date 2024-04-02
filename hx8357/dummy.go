package hx8357

import (
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
)

// Dies ist der Datentyp, welche für die Verbindung zum ILI9341 via SPI
// steht. Im Wesentlichen handelt es sich dabei um den Filedescriptor auf
// das Device-File für die SPI-Verbindung und den Pin, welcher für die
// Command/Data-Leitung verwendet wird.
type HX8357Dummy struct {
	spi spi.Conn
	pin gpio.PinIO
}

// Damit wird die Verbindung zum HX8357 geöffnet. Die Initialisierung des
// Chips wird in einer separaten Funktion (Init()) durchgeführt!
func OpenDummy(speedHz physic.Frequency) *HX8357Dummy {
	d := &HX8357Dummy{}

	return d
}

// Schliesst die Verbindung zum HX8357.
func (d *HX8357Dummy) Close() {
	// err := d.spi.Close()
	// check("Close(): error in spi.Close()", err)
}

// Führt die Initialisierung des Chips durch. initParams ist ein Slice
// von Hardware-spezifischen Einstellungen. Beim HX8357 sind dies:
//
//	{ initMinimal, madctlParam }
func (d *HX8357Dummy) Init(initParams []any) {

}

// Sende den Befehl in 'cmd' zum HX8357.
func (d *HX8357Dummy) Cmd(cmd uint8) {
}

// Sende die Daten in 'value' (1 Byte) als Datenpaket zum HX8357.
func (d *HX8357Dummy) Data8(value uint8) {
}

// Sende die Daten in 'value' (4 Bytes) als Datenpaket zum HX8357.
func (d *HX8357Dummy) Data32(value uint32) {
}

// Sendet die Daten aus dem Slice 'buf' als Daten zum HX8357. Dies ist bloss
// eine Hilfsfunktion, damit das Senden von Daten aus einem Slice einfacher
// aufzurufen ist und die ganzen Konvertierungen nicht im Hauptprogramm
// sichtbar sind.
func (d *HX8357Dummy) DataArray(buf []byte) {
}
