package hx8357

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

// Konstanten für den Display-Chip HX8357.
// Dies sind die Codes aller Befehle, welche der ILI-Chip unterstützt.
// Mehr Infos zu den einzelnen Befehlen findet man in den Dokumentationen.
const (
    HX8357_NOP        = 0x00 // No operation / dummy command
    HX8357_SWRESET    = 0x01 // Software reset
    HX8357_RDDID      = 0x04 // Read display identification info
    HX8357_RDDST      = 0x09
    HX8357_RDMODE     = 0x0A
    HX8357_RDMADCTL   = 0x0B
    HX8357_RDPIXFMT   = 0x0C
    HX8357_RDIMGFMT   = 0x0D
    HX8357_RDSELFDIAG = 0x0F

    HX8357_SLPIN  = 0x10
    HX8357_SLPOUT = 0x11
    HX8357_PTLON  = 0x12
    HX8357_NORON  = 0x13

    HX8357_INVOFF   = 0x20
    HX8357_INVON    = 0x21
    HX8357_GAMMASET = 0x26
    HX8357_DISPOFF  = 0x28
    HX8357_DISPON   = 0x29
    CASET    = 0x2A
    PASET    = 0x2B
    RAMWR    = 0x2C
    HX8357_LUTSET   = 0x2D
    HX8357_RAMRD    = 0x2E

    HX8357_PTLAR    = 0x30
    HX8357_MADCTL   = 0x36
    HX8357_VSCRSADD = 0x37
    HX8357_PIXFMT   = 0x3A

    HX8357_WRDISBV = 0x51
    HX8357_WRCTRLD = 0x53
    HX8357_WRCABC  = 0x55

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

    LONG_SIDE       = 480
    SHORT_SIDE      = 320

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
func Open(speedHz physic.Frequency) (*HX8357) {
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
    // d.spi.Close()
    // check("Close(): error in spi.Close()", err)
}

// Führt die Initialisierung des Chips durch. initParams ist ein Slice
// von Hardware-spezifischen Einstellungen. Beim HX8357 sind dies:
//     { initMinimal, madctlParam }
func (d *HX8357) Init(initParams []any) {
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

    d.Cmd(HX8357_DISPOFF) // Display On
    time.Sleep(125 * time.Millisecond)

    d.Cmd(HX8357_SWRESET) // Reset the chip at the beginning
    time.Sleep(128 * time.Millisecond)

    if !initMinimal {
        d.Cmd(0xEF)
        d.DataArray([]byte{0x03, 0x80, 0x02})

        d.Cmd(HX8357_PWCTRLB)
        d.DataArray([]byte{0x00, 0xc1, 0x30})

        d.Cmd(HX8357_PWOSEQCTR)
        d.DataArray([]byte{0x64, 0x03, 0x12, 0x81})

        d.Cmd(HX8357_DRVTICTRLA)
        d.DataArray([]byte{0x85, 0x00, 0x78})

        d.Cmd(HX8357_PWCTRLA)
        d.DataArray([]byte{0x39, 0x2c, 0x00, 0x34, 0x02})

        d.Cmd(HX8357_PMPRTCTR)
        d.Data8(0x20)

        d.Cmd(HX8357_DRVTICTRLB)
        d.DataArray([]byte{0x00, 0x00})

        d.Cmd(HX8357_PWCTR1)
        d.Data8(0x23)

        d.Cmd(HX8357_PWCTR2)
        d.Data8(0x10)

        d.Cmd(HX8357_VMCTR1)
        d.DataArray([]byte{0x3e, 0x28})

        d.Cmd(HX8357_VMCTR2)
        d.Data8(0x86)
    }

    d.Cmd(HX8357_MADCTL) // Memory Access Control
    d.Data8(madctlParam)

    if !initMinimal {
        d.Cmd(HX8357_VSCRSADD)
        d.Data8(0x00)
    }

    d.Cmd(HX8357_PIXFMT)
    d.Data8(0x66) // Fuer das 666-Format
    //d.Data8(0x55)        // Fuer das 565-Format

    if !initMinimal {
        //d.Cmd(HX8357_WRDISBV)
        //d.Data8(0x00)

        //d.Cmd(HX8357_WRCTRLD)
        //d.Data8(0x2c)

        d.Cmd(HX8357_FRMCTR1)
        d.DataArray([]byte{0x00, 0x18})

        d.Cmd(HX8357_DFUNCTR)
        d.DataArray([]byte{0x08, 0x82, 0x27})
    }

    d.Cmd(HX8357_GAMMA_3G) // Disable 3G (Gamma)
    d.Data8(0x00)

    d.Cmd(HX8357_GAMMASET) // Set gamma correction to custom
    d.Data8(0x01)           // curve 1

    d.Cmd(HX8357_GMCTRP1) // Positive Gamma Correction values
    d.DataArray(posGamma)

    d.Cmd(HX8357_GMCTRN1) // Negative Gamma Correction values
    d.DataArray(negGamma)

    d.Cmd(HX8357_SLPOUT) // Exit Sleep
    time.Sleep(125 * time.Millisecond)

    d.Cmd(HX8357_DISPON) // Display On
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
