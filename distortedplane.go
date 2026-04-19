package adatft

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type RefPointType uint8

const (
	RefTopLeft RefPointType = iota
	RefTopRight
	RefBottomRight
	RefBottomLeft
	NumRefPoints
)

func (pt RefPointType) String() string {
	switch pt {
	case RefTopLeft:
		return "TopLeft"
	case RefTopRight:
		return "TopRight"
	case RefBottomRight:
		return "BottomRight"
	case RefBottomLeft:
		return "BottomLeft"
	}
	return "(unknow reference point)"
}

var (
	calibDataFile = "TouchCalib.json"
)

type CalibData struct {
	RawPosList [NumRefPoints]TouchRawPos
	PosList    [NumRefPoints]TouchPos
}

// Liest die Konfiguration aus dem Default-File.
func ReadCalibData() *CalibData {
	fileName := filepath.Join(confDir, calibDataFile)
	return ReadCalibDataFile(fileName)
}

// Liest die Konfiguration aus dem angegebenen File. Der Pfad kann absolut
// oder relativ angegeben werden. Als Dateiformat wird JSON verwendet.
func ReadCalibDataFile(fileName string) *CalibData {
	var data []byte
	var err error

	d := &CalibData{}
	if data, err = os.ReadFile(fileName); err != nil {
		log.Fatalf("Couldn't read %s: %v", fileName, err)
	}
	if err = json.Unmarshal(data, d); err != nil {
		log.Fatalf("Couldn't unmarshal calibration file: %v", err)
	}
	return d
}

// Der Touchscreen hat ein eigenes Koordinatensystem, welches mit den Pixel-
// Koordinaten des Bildschirms erst einmal nichts gemeinsam hat (eigener
// Ursprung, eigene Skalierung, etc.). Ausserdem kann das Touchscreen-
// Koordinatensystem gegenüber dem Bildschirm-Koord.system verzerrt sein, d.h.
// die jeweiligen Koordinaten-Achsen müssen nicht parallel zueinander sein.
//
// Für die Konvertierung der Touchscreen-Koordinaten in Bildschirm-Koordinaten
// wird der Datentyp DistortedPlane verwendet.
type DistortedPlane struct {
	Rot					   RotationType
	RawPosList             [NumRefPoints]TouchRawPos
	PosList                [NumRefPoints]TouchPos
//	m1, m2, n1, n2, o1, o2 float64
//	ax, ay                 float64
}

// Schreibt die aktuelle Konfiguration in das angegebene File. Der Pfad kann
// absolut oder relativ angegeben werden. Als Dateiformat wird JSON verwendet.
func (d *DistortedPlane) WriteConfigFile(fileName string) {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(fileName, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// Liest die Konfiguration aus dem Default-File.
func (d *DistortedPlane) ReadConfig(rot RotationType) {
	fileName := filepath.Join(confDir, calibDataFile)
	d.ReadConfigFile(fileName, rot)
}

// Liest die Konfiguration aus dem angegebenen File. Der Pfad kann absolut
// oder relativ angegeben werden. Als Dateiformat wird JSON verwendet.
func (d *DistortedPlane) ReadConfigFile(fileName string, rot RotationType) {
	off := int(rot)
	calibData := ReadCalibDataFile(fileName)
	d.Rot = rot
	d.PosList = calibData.PosList
	switch rot {
	case Rotate090, Rotate270:
		d.PosList[1].X, d.PosList[3].Y = d.PosList[3].Y, d.PosList[1].X
		d.PosList[2].X, d.PosList[2].Y = d.PosList[2].Y, d.PosList[2].X
	}
	for i := range NumRefPoints {
		d.RawPosList[i] = calibData.RawPosList[(int(i)+off)%int(NumRefPoints)]
	}

	//log.Printf("posList   : %+v\n", d.PosList)
	//log.Printf("rawPosList: %+v\n", d.RawPosList)
}

func (d *DistortedPlane) SetRefPoint(id RefPointType, rawPos TouchRawPos,
	pos TouchPos) {
	d.RawPosList[id] = rawPos
	d.PosList[id] = pos
}

func (d *DistortedPlane) SetRefPoints(rawPosList []TouchRawPos,
	posList []TouchPos) {
	for id := RefTopLeft; id < NumRefPoints; id++ {
		d.RawPosList[id] = rawPosList[id]
		d.PosList[id] = posList[id]
	}
}

func (d *DistortedPlane) Transform(rawPos TouchRawPos) (pos TouchPos, err error) {
	switch d.Rot {
	case Rotate000, Rotate180:
		pos.X = Map(float64(rawPos.RawX),
			float64(d.RawPosList[0].RawX), float64(d.RawPosList[1].RawX),
			d.PosList[0].X, d.PosList[1].X)
		pos.Y = Map(float64(rawPos.RawY),
			float64(d.RawPosList[0].RawY), float64(d.RawPosList[3].RawY),
			d.PosList[0].Y, d.PosList[3].Y)
	case Rotate090, Rotate270:
		pos.X = Map(float64(rawPos.RawY),
			float64(d.RawPosList[0].RawY), float64(d.RawPosList[1].RawY),
			d.PosList[0].X, d.PosList[1].X)
		pos.Y = Map(float64(rawPos.RawX),
			float64(d.RawPosList[0].RawX), float64(d.RawPosList[3].RawX),
			d.PosList[0].Y, d.PosList[3].Y)
	}
	pos.Z = rawPos.RawZ
	//log.Printf("(%d, %d) -> (%.0f, %.0f) (%d)\n",
	//	rawPos.RawX, rawPos.RawY, pos.X, pos.Y, pos.Z)
	return pos, nil
}

type Mappable interface {
	~int | ~int16 | ~float64
}

func Map[In, Out Mappable](valIn, lbIn, ubIn In, lbOut, ubOut Out) (valOut Out) {
	return lbOut + Out(valIn-lbIn)*(ubOut-lbOut)/Out(ubIn-lbIn)
}
