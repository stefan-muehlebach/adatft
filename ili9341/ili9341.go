package ili9341

import (
    "log"
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

// Konstanten fuer den Display-Chip ILI9341.
// Dies sind die Codes aller Befehle, welche der ILI-Chip unterstuetzt. Mehr
// Infos zu den einzelnen Befehlen findet man in den Dokumentationen.
const (
    ILI9341_NOP        = 0x00		// No operation / dummy command
    ILI9341_SWRESET    = 0x01		// Software reset
    ILI9341_RDDID      = 0x04		// Read display identification info
    ILI9341_RDDST      = 0x09
    ILI9341_RDMODE     = 0x0A
    ILI9341_RDMADCTL   = 0x0B
    ILI9341_RDPIXFMT   = 0x0C
    ILI9341_RDIMGFMT   = 0x0D
    ILI9341_RDSELFDIAG = 0x0F

    ILI9341_SLPIN      = 0x10
    ILI9341_SLPOUT     = 0x11
    ILI9341_PTLON      = 0x12
    ILI9341_NORON      = 0x13

    ILI9341_INVOFF     = 0x20
    ILI9341_INVON      = 0x21
    ILI9341_GAMMASET   = 0x26
    ILI9341_DISPOFF    = 0x28
    ILI9341_DISPON     = 0x29
    ILI9341_CASET      = 0x2A
    ILI9341_PASET      = 0x2B
    ILI9341_RAMWR      = 0x2C
    ILI9341_LUTSET     = 0x2D
    ILI9341_RAMRD      = 0x2E

    ILI9341_PTLAR      = 0x30
    ILI9341_MADCTL     = 0x36
    ILI9341_VSCRSADD   = 0x37
    ILI9341_PIXFMT     = 0x3A

    ILI9341_WRDISBV    = 0x51
    ILI9341_WRCTRLD    = 0x53
    ILI9341_WRCABC     = 0x55

    ILI9341_FRMCTR1    = 0xB1
    ILI9341_FRMCTR2    = 0xB2
    ILI9341_FRMCTR3    = 0xB3
    ILI9341_INVCTR     = 0xB4
    ILI9341_DFUNCTR    = 0xB6

    ILI9341_PWCTR1     = 0xC0
    ILI9341_PWCTR2     = 0xC1
    ILI9341_PWCTR3     = 0xC2
    ILI9341_PWCTR4     = 0xC3
    ILI9341_PWCTR5     = 0xC4
    ILI9341_VMCTR1     = 0xC5
    ILI9341_VMCTR2     = 0xC7
    ILI9341_PWCTRLA    = 0xCB
    ILI9341_PWCTRLB    = 0xCF

    ILI9341_RDID1      = 0xDA
    ILI9341_RDID2      = 0xDB
    ILI9341_RDID3      = 0xDC
    ILI9341_RDID4      = 0xDD

    ILI9341_GMCTRP1    = 0xE0
    ILI9341_GMCTRN1    = 0xE1
    ILI9341_GAMMACTRL1 = 0xE2  // Anschliessend ein Array von 16 Byte-Werten
    ILI9341_GAMMACTRL2 = 0xE3  // Anschliessend ein Array von 64 Byte-Werten
    ILI9341_DRVTICTRLA = 0xE8
    ILI9341_DRVTICTRLB = 0xEA
    ILI9341_PWOSEQCTR  = 0xED

    ILI9341_GAMMA_3G   = 0xF2
    ILI9341_PMPRTCTR   = 0xF7

    ILI9341_SIDE_A     = 320
    ILI9341_SIDE_B     = 240

    ILI9341_WIDTH      = 320
    ILI9341_HEIGHT     = 240

    SPI_BLOCK_SIZE     = 4096
)

// Es ist moeglich, verschiedene Libraries fuer die SPI-Anbindung des
// ILI-Chips zu verwenden. Dieses Interface beschreibt alle Funktionen, welche
// von einer SPI-Anbindung implementiert werden muessen.
type ILIInterface interface {
    // Schliesst die Verbindung zum ILI-Chip und gibt alle Ressourcen in
    // Zusammenhang mit dieser Verbindung frei.
    Close()

    // Sendet einen Befehl (Command) zum Chip. Das ist in der Regel ein
    // 8 Bit Wert.
    Cmd(cmd uint8)

    // Sendet 8 Bit als Daten zum Chip. In den meisten Faellen ist dies ein
    // Argument eines Befehls, der vorgaengig via Cmd gesendet wird.
    Data8(val uint8)

    // Analog Data8, jedoch mit 32 Bit Daten.
    Data32(val uint32)

    // Der gesamte Slice buf wird gesendet.
    DataArray(buf []byte)
}

// Die Variablen SpiDevFile und DatCmdPin enthalten die Verbindungsparameter
// zum Display-Chip über den Main-Kanal des SPI-Buses.
var (
    SpiDevFile   = "/dev/spidev0.0"
    DatCmdPin    = "GPIO25"
)

// Dies ist der Datentyp, welche fuer die Verbindung zum ILI9341 via SPI
// steht. Im Wesentlichen handelt es sich dabei um den Filedescriptor auf
// das Device-File fuer die SPI-Verbindung und den Pin, welcher fuer die
// Command/Data-Leitung verwendet wird.
type ILI9341 struct {
    spi spi.Conn
    pin gpio.PinIO
}

// Damit wird die Verbindung zum ILI9341 geoeffnet und der Pin fur die Wahl
// zwischen Befehlen, resp. Daten definiert. Die Initialisierung des Chips
// wird in einer separaten Funktion (Init()) durchgefuehrt!
func OpenILI9341(speedHz physic.Frequency) (*ILI9341) {
    var err error
    var d *ILI9341
    var p spi.PortCloser

    d = &ILI9341{}
    p, err = spireg.Open(SpiDevFile)
    check("OpenILI9341(): error on spireg.Open()", err)

    d.spi, err = p.Connect(speedHz * physic.Hertz, spi.Mode0, 8)
    check("OpenILI9341(): error on port.Connect()", err)

    d.pin = gpioreg.ByName(DatCmdPin)
    if d.pin == nil {
        log.Fatal("OpenILI9341(): gpio io pin not found")
    }

    return d
}

// Schliesst die Verbindung zum ILI9341.
func (d *ILI9341) Close() {
//    err := d.spi.Close()
//    check("Close(): error in spi.Close()", err)
}

// Sende den Befehl in 'cmd' zum ILI9341.
func (d *ILI9341) Cmd(cmd uint8) {
    d.pin.Out(gpio.Low)
    d.spi.Tx([]byte{cmd}, nil)
//    err := d.spi.Tx([]byte{cmd}, nil)
//    check("Cmd()", err)
}

// Sende die Daten in 'value' (1 Byte) als Datenpaket zum ILI9341.
func (d *ILI9341) Data8(value uint8) {
    d.pin.Out(gpio.High)
    d.spi.Tx([]byte{value}, nil)
//    err := d.spi.Tx([]byte{value}, nil)
//    check("Data8()", err)
}

// Sende die Daten in 'value' (4 Bytes) als Datenpaket zum ILI9341.
func (d *ILI9341) Data32(value uint32) {
    var txBuf []byte = []byte{
        byte(value>>24),
        byte(value>>16),
        byte(value>>8),
        byte(value),
    }
    d.pin.Out(gpio.High)
    d.spi.Tx(txBuf, nil)
//    err := d.spi.Tx(txBuf, nil)
//    check("Data32()", err)
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
        d.spi.Tx(buf, nil)
//        err := d.spi.Tx(buf, nil)
//        check("DataArray()", err)
    } else {
        startIdx = 0
        for countRemain > 0 {
            if countRemain > SPI_BLOCK_SIZE {
                sendSize = SPI_BLOCK_SIZE
            } else {
                sendSize = countRemain
            }
            d.spi.Tx(buf[startIdx:startIdx+sendSize], nil)
//            err := d.spi.Tx(buf[startIdx:startIdx+sendSize], nil)
//            check("DataArray()", err)
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

