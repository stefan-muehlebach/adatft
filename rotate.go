package adatft

import (
    "errors"
//    ili "github.com/stefan-muehlebach/adatft/ili9341"
    ili "github.com/stefan-muehlebach/adatft/hx8357"
)

// Rotationsmöglichkeiten des Displays. Es gibt (logischerweise) 4
// Möglichkeiten das Display zu rotieren. Dies hat Auswirkungen auf die
// Initialisierung des Displays, auf die globalen Variablen Width und Height
// und auf die Konfigurationsdateien, in welchen die Daten für die
// Transformation von Touch-Koordinaten auf Display-Koordianten abgelegt
// sind, etc.
type RotationType int

const (
    Rotate000 RotationType = iota
    Rotate090
    Rotate180
    Rotate270
)

func (rot RotationType) String() string {
    switch rot {
    case Rotate000:
        return "Rotate000"
    case Rotate090:
        return "Rotate090"
    case Rotate180:
        return "Rotate180"
    case Rotate270:
        return "Rotate270"
    default:
        return "(unknown rotation)"
    }
}

func (rot *RotationType) Set(s string) error {
    switch s {
    case "Rotate000":
        *rot = Rotate000
    case "Rotate090":
        *rot = Rotate090
    case "Rotate180":
        *rot = Rotate180
    case "Rotate270":
        *rot = Rotate270
    default:
        return errors.New("Unknown rotation: " + s)
    }
    return nil
}

// In RotationData sind nun alle von der Rotation abhängigen Einstellungen
// abgelegt. Es ist ein interner Datentyp, der wohl verwendet, aber nicht
// verändert werden kann.
type RotationData struct {
    iliParam      uint8
    calibDataFile string
    width, height int
}

var (
    rotDat = []RotationData{
        RotationData{0xe0, "Rotate000.json", ili.LONG_SIDE, ili.SHORT_SIDE},
        RotationData{0x80, "Rotate090.json", ili.SHORT_SIDE, ili.LONG_SIDE},
        RotationData{0x20, "Rotate180.json", ili.LONG_SIDE, ili.SHORT_SIDE},
        RotationData{0x40, "Rotate270.json", ili.SHORT_SIDE, ili.LONG_SIDE},
    }
)

