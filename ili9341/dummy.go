package ili9341

import (
	"periph.io/x/conn/gpio"
	"periph.io/x/conn/physic"
	"periph.io/x/conn/spi"
)

// Dies ist der Datentyp, welche für die Verbindung zum ILI9341 via SPI
// steht. Im Wesentlichen handelt es sich dabei um den Filedescriptor auf
// das Device-File für die SPI-Verbindung und den Pin, welcher für die
// Command/Data-Leitung verwendet wird.
type ILI9341Dummy struct {
	spi spi.Conn
	pin gpio.PinIO
}

// Damit wird die Verbindung zum ILI9341 geöffnet. Die Initialisierung des
// Chips wird in einer separaten Funktion (Init()) durchgeführt!
func OpenDummy(speedHz physic.Frequency) *ILI9341Dummy {
	d := &ILI9341Dummy{}

	return d
}

// Schliesst die Verbindung zum ILI9341.
func (d *ILI9341Dummy) Close() {
	// err := d.spi.Close()
	// check("Close(): error in spi.Close()", err)
}

// Führt die Initialisierung des Chips durch. initParams ist ein Slice
// von Hardware-spezifischen Einstellungen. Beim ILI9341 sind dies:
//
//	{ initMinimal, madctlParam }
func (d *ILI9341Dummy) Init(initParams []any) {

}

// Sende den Befehl in 'cmd' zum ILI9341.
func (d *ILI9341Dummy) Cmd(cmd uint8) {
}

// Sende die Daten in 'value' (1 Byte) als Datenpaket zum ILI9341.
func (d *ILI9341Dummy) Data8(value uint8) {
}

// Sende die Daten in 'value' (4 Bytes) als Datenpaket zum ILI9341.
func (d *ILI9341Dummy) Data32(value uint32) {
}

// Sendet die Daten aus dem Slice 'buf' als Daten zum ILI9341. Dies ist bloss
// eine Hilfsfunktion, damit das Senden von Daten aus einem Slice einfacher
// aufzurufen ist und die ganzen Konvertierungen nicht im Hauptprogramm
// sichtbar sind.
func (d *ILI9341Dummy) DataArray(buf []byte) {
}
