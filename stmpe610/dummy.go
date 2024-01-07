package stmpe610

import (
	"periph.io/x/conn/gpio"
	"periph.io/x/conn/physic"
	"periph.io/x/conn/spi"
)

type STMPE610Dummy struct {
	spi spi.Conn
	pin gpio.PinIn
}

// Oeffnet eine Verbindung zum Touchscreen-Controller STMPE610 ueber den
// zweiten Kanal des SPI-Interfaces. Mit speedHz kann die Frequenz der
// Verbindung in Hertz angegen werden. Fuer den STMPE610 ist eine max.
// Uebertragungsgeschwindigkeit von 1 MHz angegeben. Das Resultat
// ist ein Pointer auf eine STMPE610 Struktur.
//
// Beim auftreten eines Fehlers wird das Programm abgebrochen. Ausserdem
// wird der Pin fuer das Empfangen von Interrupts konfiguriert.
func OpenDummy(speedHz physic.Frequency) *STMPE610Dummy {
	var d *STMPE610Dummy
	d = &STMPE610Dummy{}
	return d
}

// Schliesst die Verbindung zum STMPE610 und gibt alle damit verbundenen
// Ressourcen wieder frei.
func (d *STMPE610Dummy) Close() {
}

// Initialisierung des Touchscreens. Diese Einstellungen wurden (wie auch
// für das Display) aus vielen Code-Vorlagen und Dokumentationen aus
// dem Internet zusammenorchestriert - geschmückt mit vielen Stunden
// 'try and error'. Verbesserungen und Vorschläge sind jederzeit herzlich
// willkommen.
func (d *STMPE610Dummy) Init(initParams []any) {
}

func (d *STMPE610Dummy) ReadReg8(addr uint8) uint8 {
	var rxBuf []byte = []byte{0x00, 0x00}
	return rxBuf[1]
}

func (d *STMPE610Dummy) WriteReg8(addr uint8, value uint8) {
}

func (d *STMPE610Dummy) ReadReg16(addr uint8) uint16 {
	var rxBuf []byte = []byte{0x00, 0x00, 0x00}
	return (uint16(rxBuf[1]) << 8) | uint16(rxBuf[2])
}

func (d *STMPE610Dummy) WriteReg16(addr uint8, value uint16) {
	// Nicht implementiert
}

func (d *STMPE610Dummy) ReadData() (x, y uint16 /*, z uint8*/) {
	return 0, 0
}

func (d *STMPE610Dummy) SetCallback(cbFunc func(any), cbData any) {

}
