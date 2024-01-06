package stmpe610

import (
	_ "fmt"
	"log"
	"time"

	//"periph.io/x/conn/v3/gpio"
	//"periph.io/x/conn/v3/gpio/gpioreg"
	//"periph.io/x/conn/v3/physic"
	//"periph.io/x/conn/v3/spi"
	//"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/conn/gpio"
	"periph.io/x/conn/gpio/gpioreg"
	"periph.io/x/conn/physic"
	"periph.io/x/conn/spi"
	"periph.io/x/conn/spi/spireg"
)

// -----------------------------------------------------------------------------
//
// # Konstanten
//
// Viele Konstanten, welche fuer Register des STMPE610 oder fuer einzelne
// Steuerbits in den Registern stehen. Der obere Abschnitt enthaelt die
// Namen und Konstanten der Register. Im unteren Abschnitt sind die
// Konstanten zu den einzelnen Bits enthalten.
const (

	// Registerbezeichnungen
	//
	STMPE610_CHIP_ID        = 0x00
	STMPE610_ID_VER         = 0x02
	STMPE610_SYS_CTRL1      = 0x03
	STMPE610_SYS_CTRL2      = 0x04
	STMPE610_INT_CTRL       = 0x09
	STMPE610_INT_EN         = 0x0A
	STMPE610_INT_STA        = 0x0B
	STMPE610_ADC_CTRL1      = 0x20
	STMPE610_ADC_CTRL2      = 0x21
	STMPE610_ADC_CAPT       = 0x22
	STMPE610_TSC_CTRL       = 0x40
	STMPE610_TSC_CFG        = 0x41
	STMPE610_FIFO_TH        = 0x4A
	STMPE610_FIFO_STA       = 0x4B
	STMPE610_FIFO_SIZE      = 0x4C
	STMPE610_TSC_FRACTION_Z = 0x56
	STMPE610_TSC_I_DRIVE    = 0x58
	STMPE610_TSC_SHIELD     = 0x59

	// Bitmuster fuer die Registerinhalte
	//
	STMPE610_SYS_CTRL1_RESET = 0x02

	STMPE610_INT_CTRL_POL_HIGH = 0x04
	STMPE610_INT_CTRL_EDGE     = 0x02
	STMPE610_INT_CTRL_ENABLE   = 0x01
	STMPE610_INT_FIFO_EMPTY    = 0x10
	STMPE610_INT_FIFO_FULL     = 0x08
	STMPE610_INT_FIFO_OFLOW    = 0x04
	STMPE610_INT_FIFO_TH       = 0x02
	STMPE610_INT_TOUCH_DET     = 0x01

	STMPE610_ADC_CTRL1_10BIT  = 0x00
	STMPE610_ADC_CTRL1_12BIT  = 0x08
	STMPE610_ADC_CTRL1_36CLK  = 0x00
	STMPE610_ADC_CTRL1_44CLK  = 0x10
	STMPE610_ADC_CTRL1_56CLK  = 0x20
	STMPE610_ADC_CTRL1_64CLK  = 0x30
	STMPE610_ADC_CTRL1_80CLK  = 0x40
	STMPE610_ADC_CTRL1_96CLK  = 0x50
	STMPE610_ADC_CTRL1_124CLK = 0x60

	STMPE610_ADC_CTRL2_1_625MHZ = 0x00
	STMPE610_ADC_CTRL2_3_25MHZ  = 0x01
	STMPE610_ADC_CTRL2_6_5MHZ   = 0x02
	STMPE610_ADC_CTRL2_6_5_MHZ  = 0x03

	STMPE610_ADC_CAPT_ALL = 0xff

	STMPE610_TSC_CTRL_STATUS   = 0x80
	STMPE610_TSC_CTRL_WTRK_OFF = 0x00
	STMPE610_TSC_CTRL_WTRK4    = 0x10
	STMPE610_TSC_CTRL_WTRK8    = 0x20
	STMPE610_TSC_CTRL_WTRK16   = 0x30
	STMPE610_TSC_CTRL_WTRK32   = 0x40
	STMPE610_TSC_CTRL_WTRK64   = 0x50
	STMPE610_TSC_CTRL_WTRK92   = 0x60
	STMPE610_TSC_CTRL_WTRK127  = 0x70
	STMPE610_TSC_CTRL_XYZ      = 0x00
	STMPE610_TSC_CTRL_XY       = 0x02
	STMPE610_TSC_CTRL_X        = 0x04
	STMPE610_TSC_CTRL_Y        = 0x06
	STMPE610_TSC_CTRL_Z        = 0x08
	STMPE610_TSC_CTRL_EN       = 0x01

	STMPE610_TSC_CFG_1SAMPLE      = 0x00
	STMPE610_TSC_CFG_2SAMPLE      = 0x40
	STMPE610_TSC_CFG_4SAMPLE      = 0x80
	STMPE610_TSC_CFG_8SAMPLE      = 0xC0
	STMPE610_TSC_CFG_DELAY_10US   = 0x00
	STMPE610_TSC_CFG_DELAY_50US   = 0x08
	STMPE610_TSC_CFG_DELAY_100US  = 0x10
	STMPE610_TSC_CFG_DELAY_500US  = 0x18
	STMPE610_TSC_CFG_DELAY_1MS    = 0x20
	STMPE610_TSC_CFG_DELAY_5MS    = 0x28
	STMPE610_TSC_CFG_DELAY_10MS   = 0x30
	STMPE610_TSC_CFG_DELAY_50MS   = 0x38
	STMPE610_TSC_CFG_SETTLE_10US  = 0x00
	STMPE610_TSC_CFG_SETTLE_100US = 0x01
	STMPE610_TSC_CFG_SETTLE_500US = 0x02
	STMPE610_TSC_CFG_SETTLE_1MS   = 0x03
	STMPE610_TSC_CFG_SETTLE_5MS   = 0x04
	STMPE610_TSC_CFG_SETTLE_10MS  = 0x05
	STMPE610_TSC_CFG_SETTLE_50MS  = 0x06
	STMPE610_TSC_CFG_SETTLE_100MS = 0x07

	STMPE610_FIFO_STA_RESET = 0x01

	STMPE610_TSC_FRACT_Z_8_0 = 0x00
	STMPE610_TSC_FRACT_Z_7_1 = 0x01
	STMPE610_TSC_FRACT_Z_6_2 = 0x02
	STMPE610_TSC_FRACT_Z_5_3 = 0x03
	STMPE610_TSC_FRACT_Z_4_4 = 0x04
	STMPE610_TSC_FRACT_Z_3_5 = 0x05
	STMPE610_TSC_FRACT_Z_2_6 = 0x06
	STMPE610_TSC_FRACT_Z_1_7 = 0x07

	STMPE610_TSC_I_DRIVE_20MA = 0x00
	STMPE610_TSC_I_DRIVE_50MA = 0x01

	STMPE610_TSC_GROUND_X_P = 0x08
	STMPE610_TSC_GROUND_X_N = 0x04
	STMPE610_TSC_GROUND_Y_P = 0x02
	STMPE610_TSC_GROUND_Y_N = 0x01

	// Dies schlussendlich sind Konstanten, welche in Zusammenhang mit einer
	// konkrete Verwendung des Adafruit TFT-Display auf einem RaspberryPi
	// oder ASUS TinkerBoard stehen.
	//
	SpiDevFile = "/dev/spidev0.1"
	IntPin     = "GPIO24"
)

type STMPE610 struct {
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
func Open(speedHz physic.Frequency) (*STMPE610) {
	var err error
	var d *STMPE610
	var p spi.PortCloser

	d = &STMPE610{}
	p, err = spireg.Open(SpiDevFile)
	check("OpenSTMPE610(): error on spireg.Open()", err)

	d.spi, err = p.Connect(speedHz*physic.Hertz, spi.Mode1, 8)
	check("OpenSTMPE610(): error on port.Connect()", err)

	d.pin = gpioreg.ByName(IntPin)
	if d.pin == nil {
		log.Fatal("OpenSTMPE610(): gpio io pin not found")
	}
	err = d.pin.In(gpio.PullNoChange, gpio.FallingEdge)
	check("OpenSTMPE610(): couldn't configure interrupt pin", err)

	return d
}

// Schliesst die Verbindung zum STMPE610 und gibt alle damit verbundenen
// Ressourcen wieder frei.
func (d *STMPE610) Close() {
	d.pin.Halt()
	// err := d.spi.Close()
	// check("spi.Close()", err)
}

// Initialisierung des Touchscreens. Diese Einstellungen wurden (wie auch
// für das Display) aus vielen Code-Vorlagen und Dokumentationen aus
// dem Internet zusammenorchestriert - geschmückt mit vielen Stunden
// 'try and error'. Verbesserungen und Vorschläge sind jederzeit herzlich
// willkommen.
func (d *STMPE610) Init(initParams []any) {

	// System Register (STMPE610_SYS_XXX)
	//
	d.WriteReg8(STMPE610_SYS_CTRL1,
		STMPE610_SYS_CTRL1_RESET)
	time.Sleep(10 * time.Millisecond)
	for i := uint8(0); i < 65; i++ {
		d.ReadReg8(i)
	}
	d.WriteReg8(STMPE610_SYS_CTRL2,
		0x00)

	// Touchscreen Register (STMPE610_TSC_XXX)
	//
	// Touchscreen Controller Control
	// - set the window tracking feature to 8 pixels
	// - acquire X, Y data
	// - enable touch screen control
	//
	d.WriteReg8(STMPE610_TSC_CTRL,
		STMPE610_TSC_CTRL_WTRK8|
			STMPE610_TSC_CTRL_XY|
			STMPE610_TSC_CTRL_EN)

	// Analog Digital Converter Register (STMPE610_ADC_XXX)
	// (Wozu braucht es diese?)
	//
	d.WriteReg8(STMPE610_ADC_CTRL1,
		STMPE610_ADC_CTRL1_12BIT|
			STMPE610_ADC_CTRL1_96CLK) // Ada
	//STMPE610_ADC_CTRL1_36CLK)

	d.WriteReg8(STMPE610_ADC_CTRL2,
		STMPE610_ADC_CTRL2_6_5MHZ)

	d.WriteReg8(STMPE610_ADC_CAPT,
		STMPE610_ADC_CAPT_ALL)

	// Touchscreen Controller Configuration
	// - average 8 samples
	// - set a touch detect delay of 1ms
	// - set a settling time of 5ms
	//
	d.WriteReg8(STMPE610_TSC_CFG,
		STMPE610_TSC_CFG_8SAMPLE|
			STMPE610_TSC_CFG_DELAY_1MS|
			STMPE610_TSC_CFG_SETTLE_5MS)
	// STMPE610_TSC_CFG_8SAMPLE |
	// STMPE610_TSC_CFG_DELAY_500US |
	// STMPE610_TSC_CFG_SETTLE_500US)

	// Don't collect any Z data since we cannot relay on this feature!
	//d.WriteReg8(STMPE610_TSC_FRACTION_Z,
	//        STMPE610_TSC_FRACT_Z_2_6)

	// FIFO Register (STMPE610_FIFO_XXX)
	//
	d.WriteReg8(STMPE610_FIFO_TH, 1)
	d.WriteReg8(STMPE610_FIFO_STA,
		STMPE610_FIFO_STA_RESET)
	d.WriteReg8(STMPE610_FIFO_STA, 0x00)

	// Interrupt Register (STMPE610_INT_XXX)
	//
	// Wir abonnieren uns auf zwei Events: das Drücken, respl. Loslassen
	// des Bildschirms (beide Ereignisse generieren das gleiche Event) sowie
	// das Erreichen eines bestimmten Schwellwertes bei der FIFO-Queue
	d.WriteReg8(STMPE610_INT_EN,
		STMPE610_INT_TOUCH_DET |
			STMPE610_INT_FIFO_TH)

	d.WriteReg8(STMPE610_TSC_I_DRIVE,
		STMPE610_TSC_I_DRIVE_50MA)

	//d.WriteReg8(STMPE610_TSC_SHIELD,
	//        STMPE610_TSC_GROUND_X_P |
	//        STMPE610_TSC_GROUND_X_N |
	//        STMPE610_TSC_GROUND_Y_P |
	//        STMPE610_TSC_GROUND_Y_N)

	// Reset all interupts to begin with
	d.WriteReg8(STMPE610_INT_STA, 0xFF)

	// Mit diesem Register schliesslich, wird das Interrupt-System aktiviert.
	d.WriteReg8(STMPE610_INT_CTRL,
		STMPE610_INT_CTRL_EDGE |
			STMPE610_INT_CTRL_ENABLE)
}

func (d *STMPE610) ReadReg8(addr uint8) uint8 {
	var txBuf []byte = []byte{0x80 + addr, 0x00}
	var rxBuf []byte = []byte{0x00, 0x00}
	d.spi.Tx(txBuf, rxBuf)
	//err := d.spi.Tx(txBuf, rxBuf)
	//check("ReadReg8()", err)
	return rxBuf[1]
}

func (d *STMPE610) WriteReg8(addr uint8, value uint8) {
	var buf []byte = []byte{addr, value}
	d.spi.Tx(buf, nil)
	//err := d.spi.Tx(buf, nil)
	//check("WriteReg8()", err)
}

func (d *STMPE610) ReadReg16(addr uint8) uint16 {
	var txBuf []byte = []byte{0x80 + addr, 0x81 + addr, 0x00}
	var rxBuf []byte = []byte{0x00, 0x00, 0x00}
	d.spi.Tx(txBuf, rxBuf)
	//err := d.spi.Tx(txBuf, rxBuf)
	//check("ReadReg16()", err)
	return (uint16(rxBuf[1]) << 8) | uint16(rxBuf[2])
}

func (d *STMPE610) WriteReg16(addr uint8, value uint16) {
    // Nicht implementiert
}

func (d *STMPE610) ReadData() (x, y uint16 /*, z uint8*/) {
	var txBuf []byte = []byte{0xD7, 0xD7, 0xD7, 0x00}
	var rxBuf []byte = []byte{0x00, 0x00, 0x00, 0x00}
	d.spi.Tx(txBuf, rxBuf)
	//err := d.spi.Tx(txBuf, rxBuf)
	//check("ReadData()", err)
	x = (uint16(rxBuf[1]) << 4) | (uint16(rxBuf[2]) >> 4)
	y = (uint16(rxBuf[2]&0x0F) << 8) | uint16(rxBuf[3])
	//z = uint8(rxBuf[4])
	return
}

func (d *STMPE610) SetCallback(cbFunc func(any), cbData any) {
	go func() {
		for {
			if d.pin.WaitForEdge(-1) {
				cbFunc(cbData)
			} else {
				//log.Printf("WaitForEdge() returned 'false'\n")
			}
		}
	}()
}

// Interne Check-Funktion, welche bei gravierenden Fehlern das Programm
// beendet.
func check(fnc string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", fnc, err)
	}
}
