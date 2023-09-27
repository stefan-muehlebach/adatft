package adatft

import (
    "errors"
    "encoding/json"
    "log"
    "math"
    "os"
    "path/filepath"
    . "mju.net/adatft/ili9341"
)

var (
    calibDataFile string
)

// Der Touchscreen hat ein eigenes Koordinatensystem, welches mit den Pixel-
// Koordinaten des Bildschirms nicht zusammenpasst. Fuer die Konvertierung der
// Tchscreen-Koordinaten in Bildschirm-Koordinaten wird der Datentyp
// in diesem File verwendet.
//
// Fuer die Kalibrierung des Tchscreens braucht es 4 Referenzpunkte auf
// dem Display, welche moeglichst in die Ecken des Displays gesetzt werden
// muessen.
//

// Dieser Typ schliesslich wird gebraucht, um von den verzerrten Touchscreen-
// Koordinaten auf Display-Koordinaten des Bildschirms umzurechnen.
//
type DistortedPlane struct {
    DataList [NumRefPoints]TouchData
    PosList  [NumRefPoints]TouchPos
    m1, m2, n1, n2, o1, o2 float64
    ax, ay float64
}

//-----------------------------------------------------------------------------
//
func (d *DistortedPlane) WriteConfig() {
    fileName := filepath.Join(appConfDir, calibDataFile)
    d.WriteConfigFile(fileName)
}

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

func (d *DistortedPlane) ReadConfig() {
    fileName := filepath.Join(appConfDir, calibDataFile)
    d.ReadConfigFile(fileName)
}

func (d *DistortedPlane) ReadConfigFile(fileName string) {
    data, err := os.ReadFile(fileName)
    if err != nil {
        log.Fatal(err)
    }
    err = json.Unmarshal(data, d)
    if err != nil {
        log.Fatal(err)
    }
    d.update()
}

//-----------------------------------------------------------------------------
//
func (d *DistortedPlane) SetRefPoint(id RefPointType, touchData TouchData,
        touchPos TouchPos) {
    d.DataList[id] = touchData
    d.PosList[id]  = touchPos
    d.update()
}

func (d *DistortedPlane) SetRefPoints(DataList []TouchData, posList []TouchPos) {
    for id := RefTopLeft; id < NumRefPoints; id++ {
        d.DataList[id] = DataList[id]
        d.PosList[id] = posList[id]
    }
}

func (d *DistortedPlane) Transform(touchData TouchData) (touchPos TouchPos,
        err error) {
    var p1, p2, bx, cx, by, cy float64
    var tx, ty float64

    p1 = float64(touchData.RawX) - float64(d.DataList[0].RawX)
    p2 = float64(touchData.RawY) - float64(d.DataList[0].RawY)

    bx = p1*d.o2 - d.m1*d.n2 - p2*d.o1 + d.n1*d.m2
    cx = p1*d.n2 - p2*d.n1
    tx = (-bx - math.Sqrt(bx*bx - 4*d.ax*cx))/(2*d.ax)

    by = p1*d.o2 - d.n1*d.m2 - p2*d.o1 + d.m1*d.n2
    cy = p1*d.m2 - p2*d.m1
    ty = (-by + math.Sqrt(by*by - 4*d.ay*cy))/(2*d.ay)

    touchPos.X = (1-tx) * d.PosList[0].X + tx * d.PosList[2].X
    touchPos.Y = (1-ty) * d.PosList[0].Y + ty * d.PosList[2].Y

    if touchPos.X < 0.0 || touchPos.X >= ILI9341_WIDTH {
        touchPos.X = math.Max(touchPos.X, 0.0)
        touchPos.X = math.Min(touchPos.X, ILI9341_WIDTH-1.0)
        err = errors.New("coordinate outside reasonable range")
    }
    if touchPos.Y < 0.0 || touchPos.Y >= ILI9341_HEIGHT {
        touchPos.Y = math.Max(touchPos.Y, 0.0)
        touchPos.Y = math.Min(touchPos.Y, ILI9341_HEIGHT-1.0)
        err = errors.New("coordinate outside reasonable range")
    }
    return touchPos, err
}

func (d *DistortedPlane) update() {
    d.m1 = float64(d.DataList[1].RawX) - float64(d.DataList[0].RawX)
    d.m2 = float64(d.DataList[1].RawY) - float64(d.DataList[0].RawY)
    d.n1 = float64(d.DataList[3].RawX) - float64(d.DataList[0].RawX)
    d.n2 = float64(d.DataList[3].RawY) - float64(d.DataList[0].RawY)
    d.o1 = float64(d.DataList[2].RawX) - float64(d.DataList[3].RawX) - d.m1
    d.o2 = float64(d.DataList[2].RawY) - float64(d.DataList[3].RawY) - d.m2
    d.ax = d.m2*d.o1 - d.m1*d.o2
    d.ay = d.n2*d.o1 - d.n1*d.o2
}

