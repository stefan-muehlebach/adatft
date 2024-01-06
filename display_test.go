package adatft

import (
    "image"
    //"image/color"
    "image/png"
    "log"
    "math/rand"
    "os"
    "testing"
    "github.com/stefan-muehlebach/gg"
    "github.com/stefan-muehlebach/gg/color"
    "github.com/stefan-muehlebach/gg/colornames"
    "golang.org/x/image/draw"
)

const (
    imageFile = "testbild.png"
    speedHz   = 50_000_000
)

var (
    disp *Display
    pixBuf *Buffer
    testImage *image.RGBA
    dstRectFull, dstRectHalve, dstRectQuart, dstRectCust image.Rectangle
    srcPoint image.Point
    gc *gg.Context
    err error
    plane *DistortedPlane
    touchData TouchRawPos
    touchPos  TouchPos
    backColor, fillColor, borderColor color.Color
    borderWidth float64
)

func init() {
    Init()
    disp = OpenDisplay(Rotate090)

    pixBuf = NewBuffer(Width, Height)

    fh, err := os.Open(imageFile)
    if err != nil {
        log.Fatal(err)
    }
    defer fh.Close()
    tmp, err := png.Decode(fh)
    if err != nil {
        log.Fatal(err)
    }
    testImage = tmp.(*image.RGBA)

    dstRectFull  = image.Rect(  0,  0, 320, 240)
    dstRectHalve = image.Rect( 80, 60, 240, 180)
    dstRectQuart = image.Rect(120, 90, 200, 150)
    dstRectCust  = image.Rect(120, 80, 300, 200)

    srcPoint = image.Pt(0, 0)

    plane = &DistortedPlane{}
    plane.ReadConfig()

    gc = gg.NewContext(Width, Height)
    backColor   = colornames.LightGreen
    fillColor   = colornames.CadetBlue
    borderColor = colornames.WhiteSmoke
    borderWidth = 5.0
}

// Benchmark der Konvertierung von Touchscreen-Koordinaten nach Bildschirm-
// Koordinaten. TO DO: ev. sollte die Erzeugung der Touchscreen-Koordinaten
// aus der Zeitmessung entfernt werden.
//
func BenchmarkTransformPoint(b *testing.B) {
    x, y := uint16(rand.Intn(2 << 16)), uint16(rand.Intn(2 << 16))
    b.ResetTimer()
    for i := 0; i< b.N; i++ {
        touchData = TouchRawPos{x, y}
        touchPos, _ = plane.Transform(touchData)
    }
}

// Misst die Zeit für die Konvertierung eines Bildes im image.RGBA-Format
// ins TFT-spezifische 666-/565-Format. Es gibt dazu vier Funktionen, welche
// vier verschiedene Ausschnitte des Bildes konvertieren: Full, Halve, Quart
// und Cust (siehe auch die Variablen dstRectXXX in der Funktion init()).
//
func BenchmarkConvertFull(b *testing.B) {
    img := testImage.SubImage(dstRectFull).(*image.RGBA)
    for i := 0; i < b.N; i++ {
        pixBuf.Convert(img)
    }
}
func BenchmarkConvertHalve(b *testing.B) {
    img := testImage.SubImage(dstRectHalve).(*image.RGBA)
    for i := 0; i < b.N; i++ {
        pixBuf.Convert(img)
    }
}
func BenchmarkConvertQuart(b *testing.B) {
    img := testImage.SubImage(dstRectQuart).(*image.RGBA)
    for i := 0; i < b.N; i++ {
        pixBuf.Convert(img)
    }
}
func BenchmarkConvertCust(b *testing.B) {
    img := testImage.SubImage(dstRectCust).(*image.RGBA)
    for i := 0; i < b.N; i++ {
        pixBuf.Convert(img)
    }
}

// Misst die Zeit für die Darstellung eines Bildes (resp. eines Teils davon)
// auf dem TFT. Es gibt dazu vier Funktionen, welche vier verschiedene
// Ausschnitte des Bildes darstellen: Full, Halve, Quart und Cust (siehe auch
// die Variablen dstRectXXX in der Funktion init()).
//
func BenchmarkDrawFull(b *testing.B) {
    gc.SetFillColor(color.Black)
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        disp.DrawSync(testImage.SubImage(dstRectFull))
    }
}
func BenchmarkDrawHalve(b *testing.B) {
    gc.SetFillColor(color.Black)
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        disp.DrawSync(testImage.SubImage(dstRectHalve))
    }
}
func BenchmarkDrawQuart(b *testing.B) {
    gc.SetFillColor(color.Black)
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        disp.DrawSync(testImage.SubImage(dstRectQuart))
    }
}
func BenchmarkDrawCust(b *testing.B) {
    gc.SetFillColor(color.Black)
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        disp.DrawSync(testImage.SubImage(dstRectCust))
    }
}

// Misst die Zeit für das Zeichnen von zufälligen Rechtecken. Damit wird im
// Wesentlichen die Performance der Zeichenoperationen in gg gemessen
// TO DO: Ev. sollten diese Benchmarks in das Package gg verschoben werden.
//
func BenchmarkDrawRectangle(b *testing.B) {
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i< b.N; i++ {
        x, y, w, h := 160.0*rand.Float64(), 120.0*rand.Float64(),
                160.0*rand.Float64(), 120.0*rand.Float64()
        gc.DrawRectangle(x, y, w, h)
        gc.SetStrokeWidth(2.0)
        gc.SetFillColor(fillColor)
        gc.SetStrokeColor(borderColor)
        gc.FillStroke()
    }
    b.StopTimer()
    disp.DrawSync(gc.Image())
}

func BenchmarkDrawRectangleClipped(b *testing.B) {
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i< b.N; i++ {
        x, y, w, h := 160.0*rand.Float64(), 120.0*rand.Float64(),
                160.0*rand.Float64(), 120.0*rand.Float64()
        gc.DrawRectangle(x, y, w, h)
        gc.ClipPreserve()
        gc.SetStrokeWidth(2.0)
        gc.SetFillColor(fillColor)
        gc.SetStrokeColor(borderColor)
        gc.FillStroke()
        gc.ResetClip()
    }
    b.StopTimer()
    disp.DrawSync(gc.Image())
}

// Zeichnet eine Anzahl zufälliger Rechtecke und aktualisiert nach jedem
// Rechteck den gesamten Bildschirm.
//
func BenchmarkDrawRectanglesFull(b *testing.B) {
    var img *image.RGBA

    img = gc.Image().(*image.RGBA)

    rand.Seed(123_456)
    gc.SetFillColor(backColor)
    gc.Clear()
    disp.DrawSync(gc.Image())
    gc.SetFillColor(fillColor)
    gc.SetStrokeColor(borderColor)
    gc.SetStrokeWidth(borderWidth)
    b.ResetTimer()
    for i := 0; i< b.N; i++ {
        x, y, w, h := 160.0*rand.Float64(), 120.0*rand.Float64(),
                160.0*rand.Float64(), 120.0*rand.Float64()
        gc.DrawRectangle(x, y, w, h)
        gc.FillStroke()
        disp.DrawSync(img)
    }
}

// Zeichnet eine Anzahl zufälliger Rechtecke und aktualisiert nach jedem
// Rechteck nur den Bereich, der sich verändert hat.
//
func BenchmarkDrawRectanglesSubImage(b *testing.B) {
    var img *image.RGBA

    img = gc.Image().(*image.RGBA)

    rand.Seed(123_456)
    gc.SetFillColor(backColor)
    gc.Clear()
    disp.DrawSync(gc.Image())
    gc.SetFillColor(fillColor)
    gc.SetStrokeColor(borderColor)
    gc.SetStrokeWidth(borderWidth)
    b.ResetTimer()
    for i := 0; i< b.N; i++ {
        x, y, w, h := 160.0*rand.Float64(), 120.0*rand.Float64(),
                160.0*rand.Float64(), 120.0*rand.Float64()
        rect := image.Rect(int(x), int(y), int(x+w), int(y+h)).Inset(-1)
        gc.DrawRectangle(x, y, w, h)
        gc.FillStroke()
        // disp.DrawSync(img)
        disp.DrawSync(img.SubImage(rect))
    }
}

func BenchmarkDrawCircle(b *testing.B) {
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i< b.N; i++ {
        x, y, r := 320.0*rand.Float64(), 240.0*rand.Float64(),
                20.0+40.0*rand.Float64()
        gc.DrawCircle(x, y, r)
        gc.SetStrokeWidth(2.0)
        gc.SetFillColor(fillColor)
        gc.SetStrokeColor(borderColor)
        gc.FillStroke()
    }
    b.StopTimer()
    disp.DrawSync(gc.Image())
}

func BenchmarkDrawCircleClipped(b *testing.B) {
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i< b.N; i++ {
        x, y, r := 320.0*rand.Float64(), 240.0*rand.Float64(),
                20.0+40.0*rand.Float64()
        gc.DrawCircle(x, y, r)
        gc.ClipPreserve()
        gc.SetStrokeWidth(2.0)
        gc.SetFillColor(fillColor)
        gc.SetStrokeColor(borderColor)
        gc.FillStroke()
        gc.ResetClip()
    }
    b.StopTimer()
    disp.DrawSync(gc.Image())
}

func BenchmarkDrawImageGG(b *testing.B) {
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i< b.N; i++ {
        gc.DrawImage(testImage, 0, 0)
    }
    b.StopTimer()
    disp.DrawSync(gc.Image())
}

func BenchmarkDrawImageGo(b *testing.B) {
    out := gc.Image().(*image.RGBA)
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i< b.N; i++ {
        draw.Draw(out, out.Bounds(), testImage, image.Point{0, 0}, draw.Src)
    }
    b.StopTimer()
    disp.DrawSync(gc.Image())
}

func BenchmarkCopyImageGo(b *testing.B) {
    out := gc.Image().(*image.RGBA)
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i< b.N; i++ {
        draw.Copy(out, image.Point{0, 0}, testImage, testImage.Bounds(), draw.Src, nil)
    }
    b.StopTimer()
    disp.DrawSync(gc.Image())
}

