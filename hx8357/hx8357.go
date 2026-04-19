package hx8357

import (
	"log"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
)

// Konstanten für den Display-Chip HX8357.
// Dies sind die Codes aller Befehle, welche der ILI-Chip unterstützt.
// Mehr Infos zu den einzelnen Befehlen findet man in den Dokumentationen.
const (
	NOP          = 0x00 // No operation / dummy command
	SWRESET      = 0x01 // Software reset
	RDDIDIF      = 0x04
	RDNUMPE      = 0x05
	RDRED        = 0x06
	RDGREEN      = 0x07
	RDBLUE       = 0x08
	RDDPM        = 0x0A
	RDDMADCTL    = 0x0B
	RDDCOLMOD    = 0x0C
	RDDIM        = 0x0D
	RDDSM        = 0x0E
	RDDSDR       = 0x0F
	SLPIN        = 0x10 // Sleep In
	SLPOUT       = 0x11 // Sleep Out
	PTLON        = 0x12 // Partial mode on
	NORON        = 0x13 // Normal display mode on
	INVOFF       = 0x20 // Display inversion off
	INVON        = 0x21 // Display inversion on
	ALLPOFF      = 0x22 // All pixel off
	ALLPON       = 0x23 // All pixel on
	GAMSET       = 0x26 // Set Gamma curve
	DISPOFF      = 0x28 // Display off
	DISPON       = 0x29 // Display on
	CASET        = 0x2A // Set starting column address
	PASET        = 0x2B // Set starting row address
	RAMWR        = 0x2C // Memory write
	RAMRD        = 0x2E
	PLTAR        = 0x30 // Partial area
	VSCRDEF      = 0x33 // Vertical scrolling definition
	TEOFF        = 0x34 // Tearing effect line off
	TEON         = 0x35 // Tearing effect line on
	MADCTL       = 0x36 // Memory access control
	MAD_MY       = 0x80
	MAD_MX       = 0x40
	MAD_MV       = 0x20
	MAD_ML       = 0x10
	MAD_RGB      = 0x00
	MAD_BGR      = 0x08
	MAD_SS       = 0x04
	VSCRSADD     = 0x37 // Vertical scrolling start address
	IDMOFF       = 0x38 // Idle mode off
	IDMON        = 0x39 // Idle mode on
	COLMOD       = 0x3A // Interface pixel format
	RAMWRCON     = 0x3C // Memory write (continue)
	RAMRDCOM     = 0x3E
	TEARLINE     = 0x44 // Set tear effect scan lines
	GETSCAN      = 0x45
	WRDISBV      = 0x51 // Write display brightness
	RDDISBV      = 0x52
	WRCTRLD      = 0x53 // Write CTRL display
	RDCTRLD      = 0x54
	WRCABC       = 0x55 // Write adaptive brightness control
	RDCABS       = 0x56
	WRCABCMB     = 0x5E // Write adapt. bright. control minimum brightness
	RDCABCMB     = 0x5F
	RDABCSDR     = 0x68
	RDBWLB       = 0x70
	SETOSC       = 0xB0 // Set internal oscillator
	SETPOWER     = 0xB1 // Set power control
	SETDISPLAY   = 0xB2 // Set display control
	SETRGB       = 0xB3 // Set RGB interface
	SETCYC       = 0xB4 // Set display cycle
	SETBGP       = 0xB5 // Set BGP voltage
	SETVCOM      = 0xB6 // Set VCOM voltage
	SETOTP       = 0xB7 // Set OTP
	SETEXTC      = 0xB9 // Enter extension command
	SETSTBA      = 0xC0 // Set source option
	SETDGC       = 0xC1 // Set digital gamma correction
	SETID        = 0xC3
	SETDDB       = 0xC4 // Set DDB
	SETCABC      = 0xC9 // Set CABC
	SETPANEL     = 0xCC // Set panel characteristics
	SETGAMMA     = 0xE0 // Set gamma
	SETIMAGEI    = 0xE9 // Set image type
	SETMESSI     = 0xEA // Set command type
	SETCOLOR     = 0xEB // Set color
	SETREADINDEX = 0xFE // Set SPI read command address
	GETSPIREAD   = 0xFF

	LONG_SIDE  = 480
	SHORT_SIDE = 320

	SPI_BLOCK_SIZE = 4096
)

// Die Variablen SpiDevFile und DatCmdPin enthalten die Verbindungsparameter
// zum Display-Chip über den Main-Kanal des SPI-Buses.
var (
	SpiDevFile = "/dev/spidev0.0"
	DatCmdPin  = "GPIO25"
)

// Dies ist der Datentyp, welche für die Verbindung zum HX8357 via SPI
// steht. Im Wesentlichen handelt es sich dabei um den Filedescriptor auf
// das Device-File für die SPI-Verbindung und den Pin, welcher für die
// Command/Data-Leitung verwendet wird.
type HX8357 struct {
	spi spi.Conn
	pin gpio.PinIO
}

// Damit wird die Verbindung zum HX8357 geöffnet. Die Initialisierung des
// Chips wird in einer separaten Funktion (Init()) durchgeführt!
func Open(speedHz physic.Frequency) *HX8357 {
	var err error
	var d *HX8357
	var p spi.PortCloser

	d = &HX8357{}
	if p, err = spireg.Open(SpiDevFile); err != nil {
		log.Fatalf("OpenHX8357(): error on spireg.Open(): %v", err)
	}
	if d.spi, err = p.Connect(speedHz*physic.Hertz, spi.Mode0, 8); err != nil {
		log.Fatalf("OpenHX8357(): error on port.Connect(): %v", err)
	}
	if d.pin = gpioreg.ByName(DatCmdPin); d.pin == nil {
		log.Fatal("OpenHX8357(): gpio io pin not found")
	}

	return d
}

// Schliesst die Verbindung zum HX8357.
func (d *HX8357) Close() {
}

type InitCommand struct {
	Cmd byte
	Data []byte
	WaitMs int
}

func (d *HX8357) Init(rotation byte) (w, h int) {
	madctlParam := byte(MAD_RGB)

	switch rotation {
	case 0:
		madctlParam |= MAD_MX | MAD_MY
		w, h = SHORT_SIDE, LONG_SIDE
	case 1:
		madctlParam |= MAD_MV | MAD_MY
		w, h = LONG_SIDE, SHORT_SIDE
	case 2:
		w, h = SHORT_SIDE, LONG_SIDE
	case 3:
		madctlParam |= MAD_MV | MAD_MX
		w, h = LONG_SIDE, SHORT_SIDE
	}
	
	cmdList := []InitCommand{
		{DISPOFF, []byte{}, 125},
		{SWRESET, []byte{}, 128},
		{SETEXTC, []byte{0xFF, 0x83, 0x57}, 300},
		{SETRGB, []byte{0x80, 0x00, 0x06, 0x06}, 0},
		{SETVCOM, []byte{0x25}, 0},
		{SETOSC, []byte{0x68}, 0},
		{SETPANEL, []byte{0x05}, 0},
		{SETPOWER, []byte{0x00, 0x15, 0x1C, 0x1C, 0x83, 0xAA}, 0},
		{SETSTBA, []byte{0x50, 0x50, 0x01, 0x3C, 0x1E, 0x08}, 0},
		{SETCYC, []byte{0x02, 0x40, 0x00, 0x2A, 0x2A, 0x0D, 0x78}, 0},
		{SETGAMMA, []byte{0x02, 0x0A, 0x11, 0x1d, 0x23, 0x35, 0x41, 0x4b,
			0x4b, 0x42, 0x3A, 0x27, 0x1B, 0x08, 0x09, 0x03, 0x02, 0x0A,
			0x11, 0x1d, 0x23, 0x35, 0x41, 0x4b, 0x4b, 0x42, 0x3A, 0x27,
			0x1B, 0x08, 0x09, 0x03, 0x00, 0x01}, 0},
		{COLMOD, []byte{0x55}, 0},
		{MADCTL, []byte{madctlParam}, 0},
		{TEON, []byte{0x00}, 0},
		{TEARLINE, []byte{0x00, 0x02}, 0},
		{SLPOUT, []byte{}, 150},
		{DISPON, []byte{}, 150},
	}

	for _, initCmd := range cmdList {
		d.Cmd(initCmd.Cmd)
		if len(initCmd.Data) > 0 {
			d.DataArray(initCmd.Data)
		}
		if initCmd.WaitMs > 0 {
			time.Sleep(time.Duration(initCmd.WaitMs) * time.Millisecond)
		}
	}

	return w, h
}

// Sende den Befehl in 'cmd' zum HX8357.
func (d *HX8357) Cmd(cmd uint8) {
	d.pin.Out(gpio.Low)
	d.spi.Tx([]byte{cmd}, nil)
}

// Sende die Daten in 'value' (1 Byte) als Datenpaket zum HX8357.
func (d *HX8357) Data8(value uint8) {
	d.pin.Out(gpio.High)
	d.spi.Tx([]byte{value}, nil)
}

// Sende die Daten in 'value' (4 Bytes) als Datenpaket zum HX8357.
func (d *HX8357) Data32(value uint32) {
	var txBuf []byte = []byte{
		byte(value >> 24),
		byte(value >> 16),
		byte(value >> 8),
		byte(value),
	}
	d.pin.Out(gpio.High)
	d.spi.Tx(txBuf, nil)
}

// Sendet die Daten aus dem Slice 'buf' als Daten zum HX8357. Dies ist bloss
// eine Hilfsfunktion, damit das Senden von Daten aus einem Slice einfacher
// aufzurufen ist und die ganzen Konvertierungen nicht im Hauptprogramm
// sichtbar sind.
func (d *HX8357) DataArray(buf []byte) {
	var countRemain int = len(buf)
	var sendSize, startIdx int

	d.pin.Out(gpio.High)
	if countRemain <= SPI_BLOCK_SIZE {
		d.spi.Tx(buf, nil)
	} else {
		startIdx = 0
		for countRemain > 0 {
			if countRemain > SPI_BLOCK_SIZE {
				sendSize = SPI_BLOCK_SIZE
			} else {
				sendSize = countRemain
			}
			d.spi.Tx(buf[startIdx:startIdx+sendSize], nil)
			countRemain -= sendSize
			startIdx += sendSize
		}
	}
}
