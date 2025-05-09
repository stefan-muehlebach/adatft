package adatft

import (
	"encoding/json"
	"errors"
	"math"
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
	d := &CalibData{}
	data, err := os.ReadFile(fileName)
	if err != nil {
		adalog.Fatal(err)
	}
	err = json.Unmarshal(data, d)
	if err != nil {
		adalog.Fatal(err)
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
	RawPosList             [NumRefPoints]TouchRawPos
	PosList                [NumRefPoints]TouchPos
	m1, m2, n1, n2, o1, o2 float64
	ax, ay                 float64
}

// Schreibt die aktuelle Konfiguration in das Default-File.
//func (d *DistortedPlane) WriteConfig() {
//	fileName := filepath.Join(confDir, calibDataFile)
//	d.WriteConfigFile(fileName)
//}
//func (d *DistortedPlane) WriteConfig(dataFile string) {
//	fileName := filepath.Join(confDir, dataFile)
//	d.WriteConfigFile(fileName)
//}

// Schreibt die aktuelle Konfiguration in das angegebene File. Der Pfad kann
// absolut oder relativ angegeben werden. Als Dateiformat wird JSON verwendet.
func (d *DistortedPlane) WriteConfigFile(fileName string) {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		adalog.Fatal(err)
	}
	err = os.WriteFile(fileName, data, 0644)
	if err != nil {
		adalog.Fatal(err)
	}
}

// Liest die Konfiguration aus dem Default-File.
func (d *DistortedPlane) ReadConfig(rot RotationType) {
	fileName := filepath.Join(confDir, calibDataFile)
	d.ReadConfigFile(fileName, rot)

	/*
		off := 0
		calibData := ReadCalibData()
		d.PosList = calibData.PosList
		switch rot {
		case Rotate000:
			off = 0
		case Rotate090:
			off = 1
			d.PosList[1].X, d.PosList[3].Y = d.PosList[3].Y, d.PosList[1].X
			d.PosList[2].X, d.PosList[2].Y = d.PosList[2].Y, d.PosList[2].X
		case Rotate180:
			off = 2
		case Rotate270:
			off = 3
			d.PosList[1].X, d.PosList[3].Y = d.PosList[3].Y, d.PosList[1].X
			d.PosList[2].X, d.PosList[2].Y = d.PosList[2].Y, d.PosList[2].X
		}
		for i := range NumRefPoints {
			d.RawPosList[(int(i)+off)%int(NumRefPoints)] = calibData.RawPosList[i]
		}

		d.update()
	*/
}

// Liest die Konfiguration aus dem angegebenen File. Der Pfad kann absolut
// oder relativ angegeben werden. Als Dateiformat wird JSON verwendet.
func (d *DistortedPlane) ReadConfigFile(fileName string, rot RotationType) {
	off := 0
	calibData := ReadCalibDataFile(fileName)
	d.PosList = calibData.PosList
	switch rot {
	case Rotate000:
		off = 0
	case Rotate090:
		off = 1
		d.PosList[1].X, d.PosList[3].Y = d.PosList[3].Y, d.PosList[1].X
		d.PosList[2].X, d.PosList[2].Y = d.PosList[2].Y, d.PosList[2].X
	case Rotate180:
		off = 2
	case Rotate270:
		off = 3
		d.PosList[1].X, d.PosList[3].Y = d.PosList[3].Y, d.PosList[1].X
		d.PosList[2].X, d.PosList[2].Y = d.PosList[2].Y, d.PosList[2].X
	}
	for i := range NumRefPoints {
		d.RawPosList[(int(i)+off)%int(NumRefPoints)] = calibData.RawPosList[i]
	}

	d.update()
}

func (d *DistortedPlane) SetRefPoint(id RefPointType, rawPos TouchRawPos,
	pos TouchPos) {
	d.RawPosList[id] = rawPos
	d.PosList[id] = pos
	d.update()
}

func (d *DistortedPlane) SetRefPoints(rawPosList []TouchRawPos,
	posList []TouchPos) {
	for id := RefTopLeft; id < NumRefPoints; id++ {
		d.RawPosList[id] = rawPosList[id]
		d.PosList[id] = posList[id]
	}
}

// Transformiert die Touchscreen-Koordinaten in rawPos zu Bildschirm-
// Koordinaten.
func (d *DistortedPlane) Transform(rawPos TouchRawPos) (pos TouchPos,
	err error) {
	var p1, p2, bx, cx, by, cy float64
	var tx, ty float64

	p1 = float64(rawPos.RawX) - float64(d.RawPosList[0].RawX)
	p2 = float64(rawPos.RawY) - float64(d.RawPosList[0].RawY)

	bx = p1*d.o2 - d.m1*d.n2 - p2*d.o1 + d.n1*d.m2
	cx = p1*d.n2 - p2*d.n1
	tx = (-bx - math.Sqrt(bx*bx-4*d.ax*cx)) / (2 * d.ax)

	by = p1*d.o2 - d.n1*d.m2 - p2*d.o1 + d.m1*d.n2
	cy = p1*d.m2 - p2*d.m1
	ty = (-by + math.Sqrt(by*by-4*d.ay*cy)) / (2 * d.ay)

	pos.X = (1-tx)*d.PosList[0].X + tx*d.PosList[2].X
	pos.Y = (1-ty)*d.PosList[0].Y + ty*d.PosList[2].Y

	if pos.X < 0.0 || pos.X >= float64(Width) {
		pos.X = max(pos.X, 0.0)
		pos.X = min(pos.X, float64(Width)-1.0)
		err = errors.New("coordinate outside reasonable range")
	}
	if pos.Y < 0.0 || pos.Y >= float64(Height) {
		pos.Y = max(pos.Y, 0.0)
		pos.Y = min(pos.Y, float64(Height)-1.0)
		err = errors.New("coordinate outside reasonable range")
	}
	return pos, err
}

func (d *DistortedPlane) update() {
	d.m1 = float64(d.RawPosList[1].RawX) - float64(d.RawPosList[0].RawX)
	d.m2 = float64(d.RawPosList[1].RawY) - float64(d.RawPosList[0].RawY)
	d.n1 = float64(d.RawPosList[3].RawX) - float64(d.RawPosList[0].RawX)
	d.n2 = float64(d.RawPosList[3].RawY) - float64(d.RawPosList[0].RawY)
	d.o1 = float64(d.RawPosList[2].RawX) - float64(d.RawPosList[3].RawX) - d.m1
	d.o2 = float64(d.RawPosList[2].RawY) - float64(d.RawPosList[3].RawY) - d.m2
	d.ax = d.m2*d.o1 - d.m1*d.o2
	d.ay = d.n2*d.o1 - d.n1*d.o2
}
