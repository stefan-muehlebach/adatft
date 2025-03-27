package ili9341

import (
    "log"
    "time"
    "periph.io/x/conn/v3/gpio"
    "periph.io/x/conn/v3/gpio/gpioreg"
    "periph.io/x/conn/v3/physic"
    "periph.io/x/conn/v3/spi"
    "periph.io/x/conn/v3/spi/spireg"
    //"periph.io/x/conn/gpio"
    //"periph.io/x/conn/gpio/gpioreg"
    //"periph.io/x/conn/physic"
    //"periph.io/x/conn/spi"
    //"periph.io/x/conn/spi/spireg"
)

// Konstanten für den Display-Chip ILI9341.
// Dies sind die Codes aller Befehle, welche der ILI-Chip unterstützt.
// Mehr Infos zu den einzelnen Befehlen findet man in den Dokumentationen.
const (
    NOP        = 0x00 // No operation / dummy command
    SWRESET    = 0x01 // Software reset
    RDDID      = 0x04 // Read display identification info
    RDDST      = 0x09
    RDMODE     = 0x0A
    RDMADCTL   = 0x0B
    RDPIXFMT   = 0x0C
    RDIMGFMT   = 0x0D
    RDSELFDIAG = 0x0F

    SLPIN  = 0x10
    SLPOUT = 0x11
    PTLON  = 0x12
    NORON  = 0x13

    INVOFF   = 0x20
    INVON    = 0x21
    GAMMASET = 0x26
    DISPOFF  = 0x28
    DISPON   = 0x29
    CASET    = 0x2A
    PASET    = 0x2B
    RAMWR    = 0x2C
    LUTSET   = 0x2D
    RAMRD    = 0x2E

    PTLAR    = 0x30
    MADCTL   = 0x36
    VSCRSADD = 0x37
    PIXFMT   = 0x3A

    WRDISBV = 0x51
    WRCTRLD = 0x53
    WRCABC  = 0x55

    FRMCTR1 = 0xB1
    FRMCTR2 = 0xB2
    FRMCTR3 = 0xB3
    INVCTR  = 0xB4
    DFUNCTR = 0xB6

    PWCTR1  = 0xC0
    PWCTR2  = 0xC1
    PWCTR3  = 0xC2
    PWCTR4  = 0xC3
    PWCTR5  = 0xC4
    VMCTR1  = 0xC5
    VMCTR2  = 0xC7
    PWCTRLA = 0xCB
    PWCTRLB = 0xCF

    RDID1 = 0xDA
    RDID2 = 0xDB
    RDID3 = 0xDC
    RDID4 = 0xDD

    GMCTRP1    = 0xE0
    GMCTRN1    = 0xE1
    GAMMACTRL1 = 0xE2 // Anschliessend ein Array von 16 Byte-Werten
    GAMMACTRL2 = 0xE3 // Anschliessend ein Array von 64 Byte-Werten
    DRVTICTRLA = 0xE8
    DRVTICTRLB = 0xEA
    PWOSEQCTR  = 0xED

    GAMMA_3G = 0xF2
    PMPRTCTR = 0xF7

    LONG_SIDE  = 320
    SHORT_SIDE = 240

    SPI_BLOCK_SIZE = 4096
)

// Die Variablen SpiDevFile und DatCmdPin enthalten die Verbindungsparameter
// zum Display-Chip über den Main-Kanal des SPI-Buses.
var (
    SpiDevFile = "/dev/spidev0.0"
    DatCmdPin  = "GPIO25"
)

// Dies ist der Datentyp, welche für die Verbindung zum ILI9341 via SPI
// steht. Im Wesentlichen handelt es sich dabei um den Filedescriptor auf
// das Device-File für die SPI-Verbindung und den Pin, welcher für die
// Command/Data-Leitung verwendet wird.
type ILI9341 struct {
    spi spi.Conn
    pin gpio.PinIO
}

// Damit wird die Verbindung zum ILI9341 geöffnet. Die Initialisierung des
// Chips wird in einer separaten Funktion (Init()) durchgeführt!
func Open(speedHz physic.Frequency) (*ILI9341) {
    var err error
    var d *ILI9341
    var p spi.PortCloser

    d = &ILI9341{}
    if p, err = spireg.Open(SpiDevFile); err != nil {
        log.Fatalf("OpenILI9341(): error on spireg.Open(): %v", err)
    }
    if d.spi, err = p.Connect(speedHz*physic.Hertz, spi.Mode0, 8); err != nil {
        log.Fatalf("OpenILI9341(): error on port.Connect(): %v", err)
    }
    if d.pin = gpioreg.ByName(DatCmdPin); d.pin == nil {
        log.Fatal("OpenILI9341(): gpio io pin not found")
    }

    return d
}

// Schliesst die Verbindung zum ILI9341.
func (d *ILI9341) Close() {
    // d.spi.Close()
    // check("Close(): error in spi.Close()", err)
}

// Führt die Initialisierung des Chips durch. initParams ist ein Slice
// von Hardware-spezifischen Einstellungen. Beim ILI9341 sind dies:
//     { initMinimal, madctlParam }
func (d *ILI9341) Init(initParams []any) {
    var initMinimal bool
    var madctlParam uint8

    initMinimal = initParams[0].(bool)
    madctlParam = initParams[1].(uint8)

    var posGamma []uint8 = []uint8{
        0x0f, 0x31, 0x2b, 0x0c, 0x0e, 0x08, 0x4e,
        0xf1,
        0x37, 0x07, 0x10, 0x03, 0x0e, 0x09, 0x00,
    }
    var negGamma []uint8 = []uint8{
        0x00, 0x0e, 0x14, 0x03, 0x11, 0x07, 0x31,
        0xc1,
        0x48, 0x08, 0x0f, 0x0c, 0x31, 0x36, 0x0f,
    }
    /*
       var colorLut []uint8
       colorLut = make([]uint8, 128)
       for i := 0; i < 32; i++ {
           colorLut[i] = uint8(2*i)
       }
       for i := 0; i < 64; i++ {
           colorLut[i+32] = uint8(i)
       }
       for i := 0; i < 32; i++ {
           colorLut[i+96] = uint8(2*i)
       }
    */

    /*
       macroGamma = make([]uint8, 16)
       for i := 0; i < 16; i++ {
           macroGamma[i] = 0x00
       }
       microGamma = make([]uint8, 64)
       for i := 0; i < 64; i++ {
           microGamma[i] = 0x00
       }
    */

    d.Cmd(DISPOFF) // Display On
    time.Sleep(125 * time.Millisecond)

    d.Cmd(SWRESET) // Reset the chip at the beginning
    time.Sleep(128 * time.Millisecond)

    if !initMinimal {
        d.Cmd(0xEF)
        d.DataArray([]byte{0x03, 0x80, 0x02})

        d.Cmd(PWCTRLB)
        d.DataArray([]byte{0x00, 0xc1, 0x30})

        d.Cmd(PWOSEQCTR)
        d.DataArray([]byte{0x64, 0x03, 0x12, 0x81})

        d.Cmd(DRVTICTRLA)
        d.DataArray([]byte{0x85, 0x00, 0x78})

        d.Cmd(PWCTRLA)
        d.DataArray([]byte{0x39, 0x2c, 0x00, 0x34, 0x02})

        d.Cmd(PMPRTCTR)
        d.Data8(0x20)

        d.Cmd(DRVTICTRLB)
        d.DataArray([]byte{0x00, 0x00})

        d.Cmd(PWCTR1)
        d.Data8(0x23)

        d.Cmd(PWCTR2)
        d.Data8(0x10)

        d.Cmd(VMCTR1)
        d.DataArray([]byte{0x3e, 0x28})

        d.Cmd(VMCTR2)
        d.Data8(0x86)
    }

    d.Cmd(MADCTL) // Memory Access Control
    d.Data8(madctlParam)

    if !initMinimal {
        d.Cmd(VSCRSADD)
        d.Data8(0x00)
    }

    d.Cmd(PIXFMT)
    //d.Data8(0x66) // Fuer das 666-Format
    d.Data8(0x55)        // Fuer das 565-Format

    if !initMinimal {
        //d.Cmd(WRDISBV)
        //d.Data8(0x00)

        //d.Cmd(WRCTRLD)
        //d.Data8(0x2c)

        d.Cmd(FRMCTR1)
        d.DataArray([]byte{0x00, 0x18})

        d.Cmd(DFUNCTR)
        d.DataArray([]byte{0x08, 0x82, 0x27})
    }

    d.Cmd(GAMMA_3G) // Disable 3G (Gamma)
    d.Data8(0x00)

    d.Cmd(GAMMASET) // Set gamma correction to custom
    d.Data8(0x01)           // curve 1

    d.Cmd(GMCTRP1) // Positive Gamma Correction values
    d.DataArray(posGamma)

    d.Cmd(GMCTRN1) // Negative Gamma Correction values
    d.DataArray(negGamma)

    d.Cmd(SLPOUT) // Exit Sleep
    time.Sleep(125 * time.Millisecond)

    d.Cmd(DISPON) // Display On
    time.Sleep(125 * time.Millisecond)
}

// Sende den Befehl in 'cmd' zum ILI9341.
func (d *ILI9341) Cmd(cmd uint8) {
    d.pin.Out(gpio.Low)
    //d.spi.Tx([]byte{cmd}, nil)
    err := d.spi.Tx([]byte{cmd}, nil)
    check("Cmd()", err)
}

// Sende die Daten in 'value' (1 Byte) als Datenpaket zum ILI9341.
func (d *ILI9341) Data8(value uint8) {
    d.pin.Out(gpio.High)
    //d.spi.Tx([]byte{value}, nil)
    err := d.spi.Tx([]byte{value}, nil)
    check("Data8()", err)
}

// Sende die Daten in 'value' (4 Bytes) als Datenpaket zum ILI9341.
func (d *ILI9341) Data32(value uint32) {
    var txBuf []byte = []byte{
        byte(value >> 24),
        byte(value >> 16),
        byte(value >> 8),
        byte(value),
    }
    d.pin.Out(gpio.High)
    //d.spi.Tx(txBuf, nil)
    err := d.spi.Tx(txBuf, nil)
    check("Data32()", err)
}

// Sendet die Daten aus dem Slice 'buf' als Daten zum ILI9341. Dies ist bloss
// eine Hilfsfunktion, damit das Senden von Daten aus einem Slice einfacher
// aufzurufen ist und die ganzen Konvertierungen nicht im Hauptprogramm
// sichtbar sind.
func (d *ILI9341) DataArray(buf []byte) {
    var countRemain int = len(buf)
    var sendSize, startIdx int

    d.pin.Out(gpio.High)
    if len(buf) <= SPI_BLOCK_SIZE {
        //d.spi.Tx(buf, nil)
        err := d.spi.Tx(buf, nil)
        check("DataArray()", err)
    } else {
        startIdx = 0
        for countRemain > 0 {
            if countRemain > SPI_BLOCK_SIZE {
                sendSize = SPI_BLOCK_SIZE
            } else {
                sendSize = countRemain
            }
            //d.spi.Tx(buf[startIdx:startIdx+sendSize], nil)
            err := d.spi.Tx(buf[startIdx:startIdx+sendSize], nil)
            check("DataArray()", err)
            countRemain -= sendSize
            startIdx += sendSize
        }
    }
}

// Interne Check-Funktion, welche bei gravierenden Fehlern das Programm
// beendet.
func check(fnc string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", fnc, err)
	}
}

