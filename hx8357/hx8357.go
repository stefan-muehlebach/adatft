package hx8357

import (
	"log"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"time"
	//"periph.io/x/host/v3/sysfs"
)

// Konstanten für den Display-Chip HX8357.
// Dies sind die Codes aller Befehle, welche der ILI-Chip unterstützt.
// Mehr Infos zu den einzelnen Befehlen findet man in den Dokumentationen.
const (
	NOP      = 0x00 // No operation / dummy command
	SWRESET  = 0x01 // Software reset
	SLPIN    = 0x10 // Sleep In
	SLPOUT   = 0x11 // Sleep Out
	PTLON    = 0x12 // Partial mode on
	NORON    = 0x13 // Normal display mode on
	INVOFF   = 0x20 // Display inversion off
	INVON    = 0x21 // Display inversion on
	ALLPOFF  = 0x22 // All pixel off
	ALLPON   = 0x23 // All pixel on
	GAMSET   = 0x26 // Set Gamma curve
	DISPOFF  = 0x28 // Display off
	DISPON   = 0x29 // Display on
	CASET    = 0x2A // Set starting column address
	PASET    = 0x2B // Set starting row address
	RAMWR    = 0x2C // Memory write
	PLTAR    = 0x30 // Partial area
	VSCRDEF  = 0x33 // Vertical scrolling definition
	TEOFF    = 0x34 // Tearing effect line off
	TEON     = 0x35 // Tearing effect line on
	MADCTL   = 0x36 // Memory access control
	VSCRSADD = 0x37 // Vertical scrolling start address
	IDMOFF   = 0x38 // Idle mode off
	IDMON    = 0x39 // Idle mode on
	COLMOD   = 0x3A // Interface pixel format
	RAMWRCON = 0x3C // Memory write (continue)
	TESL     = 0x44 // Set tear effect scan lines
	WRDISBV  = 0x51 // Write display brightness
	WRCTRLD  = 0x53 // Write CTRL display
	WRCABC   = 0x55 // Write adaptive brightness control
	WRCABCMB = 0x5E // Write adapt. bright. control minimum brightness

	SETOSC       = 0xB0 // Set internal oscillator
	SETPOWER     = 0xB1 // Set power control
	SETDISPLAY   = 0xB2 // Set display control
	SETRGB       = 0xB3 // Set RGB interface
	SETCYC       = 0xB4 // Set display cycle
	SETBGP       = 0xB5 // Set BGP voltage
	SETVCOM      = 0xB6 // Set VCOM voltage
	SETOTP       = 0xB7 // Set OTP
	SETEXTC      = 0xB9 // Enter extension command
	SETDGC       = 0xC1 // Set digital gamma correction
	SETSTBA      = 0xC0 // Set source option
	SETDDB       = 0xC4 // Set DDB
	SETCABC      = 0xC9 // Set CABC
	SETPANEL     = 0xCC // Set panel characteristics
	SETGAMMA     = 0xE0 // Set gamma
	SETIMAGEI    = 0xE9 // Set image type
	SETMESSI     = 0xEA // Set command type
	SETCOLOR     = 0xEB // Set color
	SETREADINDEX = 0xFE // Set SPI read command address

	/*
		HX8357_FRMCTR1 = 0xB1
		HX8357_FRMCTR2 = 0xB2
		HX8357_FRMCTR3 = 0xB3
		HX8357_INVCTR  = 0xB4
		HX8357_DFUNCTR = 0xB6

		HX8357_PWCTR1  = 0xC0
		HX8357_PWCTR2  = 0xC1
		HX8357_PWCTR3  = 0xC2
		HX8357_PWCTR4  = 0xC3
		HX8357_PWCTR5  = 0xC4
		HX8357_VMCTR1  = 0xC5
		HX8357_VMCTR2  = 0xC7
		HX8357_PWCTRLA = 0xCB
		HX8357_PWCTRLB = 0xCF

		HX8357_RDID1 = 0xDA
		HX8357_RDID2 = 0xDB
		HX8357_RDID3 = 0xDC
		HX8357_RDID4 = 0xDD

		HX8357_GMCTRP1    = 0xE0
		HX8357_GMCTRN1    = 0xE1
		HX8357_GAMMACTRL1 = 0xE2 // Anschliessend ein Array von 16 Byte-Werten
		HX8357_GAMMACTRL2 = 0xE3 // Anschliessend ein Array von 64 Byte-Werten
		HX8357_DRVTICTRLA = 0xE8
		HX8357_DRVTICTRLB = 0xEA
		HX8357_PWOSEQCTR  = 0xED

		HX8357_GAMMA_3G = 0xF2
		HX8357_PMPRTCTR = 0xF7
	*/

	LONG_SIDE  = 480
	SHORT_SIDE = 320

	SPI_BLOCK_SIZE = 4096

)

// Die Variablen SpiDevFile und DatCmdPin enthalten die Verbindungsparameter
// zum Display-Chip über den Main-Kanal des SPI-Buses.
var (
	SpiDevFile = "/dev/spidev0.0"
	DatCmdPin  = "GPIO25"

    gammaPar = []uint8{
        // Positive polarity
        0x02, 0x0a, 0x11, 0x1d, 0x23, 0x35, 0x41, 0x4b,
        0x4b, 0x42, 0x3a, 0x27, 0x1b, 0x08, 0x09, 0x03,
        // Negative polarity
        0x02, 0x0a, 0x11, 0x1d, 0x23, 0x35, 0x41, 0x4b,
        0x4b, 0x42, 0x3a, 0x27, 0x1b, 0x08, 0x09, 0x03,
        // Some control bytes
        0x00, 0x01,
    }

    gammaLUT = []uint8{
        0x01,
        // Lookup values for red
        0x00, 0x08, 0x10, 0x18, 0x20, 0x28, 0x30, 0x38,
        0x40, 0x48, 0x50, 0x58, 0x60, 0x68, 0x70, 0x7a,
        0x80, 0x88, 0x90, 0x98, 0xa0, 0xa8, 0xb0, 0xb8,
        0xc0, 0xc8, 0xd0, 0xd8, 0xe0, 0xe8, 0xf0, 0xf8, 0xfc,
        // Lookup values for green
        0x00, 0x08, 0x10, 0x18, 0x20, 0x28, 0x30, 0x38,
        0x40, 0x48, 0x50, 0x58, 0x60, 0x68, 0x70, 0x78,
        0x80, 0x88, 0x90, 0x98, 0xa0, 0xa8, 0xb0, 0xb8,
        0xc0, 0xc8, 0xd0, 0xd8, 0xe0, 0xe8, 0xf0, 0xf8, 0xfc,
        // Lookup values for blue
        0x00, 0x08, 0x10, 0x18, 0x20, 0x28, 0x30, 0x38,
        0x40, 0x48, 0x50, 0x58, 0x60, 0x68, 0x70, 0x7a,
        0x80, 0x88, 0x90, 0x98, 0xa0, 0xa8, 0xb0, 0xb8,
        0xc0, 0xc8, 0xd0, 0xd8, 0xe0, 0xe8, 0xf0, 0xf8, 0xfc,
    }

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

	//spi, _ := sysfs.NewSPI(0, 0)
	//log.Printf("MaxTxSize(): %d", spi.MaxTxSize())
	//spi.Close()

	return d
}

// Schliesst die Verbindung zum HX8357.
func (d *HX8357) Close() {
	// d.spi.Close()
	// check("Close(): error in spi.Close()", err)
}

// Führt die Initialisierung des Chips durch. initParams ist ein Slice
// von Hardware-spezifischen Einstellungen. Beim HX8357 sind dies:
//
//	{ initMinimal, madctlParam }
func (d *HX8357) Init(initParams []any) {
	//    var initMinimal bool
	var madctlParam uint8
	var colmodParam uint8

	//    initMinimal = initParams[0].(bool)
	madctlParam = initParams[1].(uint8)
	colmodParam = initParams[2].(uint8)

	d.Cmd(DISPOFF) // Display Off
	time.Sleep(125 * time.Millisecond)

	d.Cmd(COLMOD) // Pixel format
	d.Data8(colmodParam)

	d.Cmd(SWRESET) // Reset the chip at the beginning
	time.Sleep(128 * time.Millisecond)

	d.Cmd(MADCTL) // Memory Access Control
	d.Data8(madctlParam)

	d.Cmd(WRCTRLD) // Write Control Display
	d.Data8(0x2c)  // Backlight Control Block: ON, Display Dimming: ON,
	// Backlight Control: ON

    d.Cmd(SETEXTC)
    d.DataArray([]byte{0xFF, 0x83, 0x57})

    d.Cmd(SETPANEL)
    d.Data8(0x00)

    d.Cmd(SETGAMMA)
    d.DataArray(gammaPar)

    //d.Cmd(SETDGC)
    //d.DataArray(gammaLUT)

    d.Cmd(SETEXTC)
    d.DataArray([]byte{0x01, 0x01, 0x01})

	d.Cmd(SLPOUT) // Exit Sleep
	time.Sleep(125 * time.Millisecond)

	d.Cmd(DISPON) // Display On
	time.Sleep(125 * time.Millisecond)
}

// Sende den Befehl in 'cmd' zum HX8357.
func (d *HX8357) Cmd(cmd uint8) {
	d.pin.Out(gpio.Low)
	d.spi.Tx([]byte{cmd}, nil)
	// err := d.spi.Tx([]byte{cmd}, nil)
	// check("Cmd()", err)
}

// Sende die Daten in 'value' (1 Byte) als Datenpaket zum HX8357.
func (d *HX8357) Data8(value uint8) {
	d.pin.Out(gpio.High)
	d.spi.Tx([]byte{value}, nil)
	// err := d.spi.Tx([]byte{value}, nil)
	// check("Data8()", err)
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
	// err := d.spi.Tx(txBuf, nil)
	// check("Data32()", err)
}

// Sendet die Daten aus dem Slice 'buf' als Daten zum HX8357. Dies ist bloss
// eine Hilfsfunktion, damit das Senden von Daten aus einem Slice einfacher
// aufzurufen ist und die ganzen Konvertierungen nicht im Hauptprogramm
// sichtbar sind.
func (d *HX8357) DataArray(buf []byte) {
	var countRemain int = len(buf)
	var sendSize, startIdx int

	d.pin.Out(gpio.High)
	if len(buf) <= SPI_BLOCK_SIZE {
		d.spi.Tx(buf, nil)
		// err := d.spi.Tx(buf, nil)
		// check("DataArray()", err)
	} else {
		startIdx = 0
		for countRemain > 0 {
			if countRemain > SPI_BLOCK_SIZE {
				sendSize = SPI_BLOCK_SIZE
			} else {
				sendSize = countRemain
			}
			d.spi.Tx(buf[startIdx:startIdx+sendSize], nil)
			// err := d.spi.Tx(buf[startIdx:startIdx+sendSize], nil)
			// check("DataArray()", err)
			countRemain -= sendSize
			startIdx += sendSize
		}
	}
}
