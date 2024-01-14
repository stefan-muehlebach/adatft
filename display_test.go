package adatft

import (
	"time"
    "image"
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
    pixBuf *ILIImage
    fWidth, fHeight float64
    testBild, workImage *image.RGBA
    RectFull, RectHalve, RectQuart, RectCust image.Rectangle
    srcPoint image.Point
    rect image.Rectangle
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
    disp = OpenDisplay(Rotate000)
    fWidth, fHeight = float64(Width), float64(Height)

    pixBuf = NewILIImage(image.Rect(0, 0, Width, Height))

    fh, err := os.Open(imageFile)
    if err != nil {
        log.Fatal(err)
    }
    defer fh.Close()
    tmp, err := png.Decode(fh)
    if err != nil {
        log.Fatal(err)
    }
    testBild  = tmp.(*image.RGBA)
    workImage = image.NewRGBA(image.Rect(0, 0, Width, Height))

    RectFull  = image.Rect(  0,  0, Width, Height)
    RectHalve = image.Rect(Width/4, Height/4, 3*Width/4, 3*Height/4)
    RectQuart = image.Rect(3*Width/8, 3*Height/8, 5*Width/8, 5*Height/8)
    RectCust  = image.Rect(0, 0, Width/3, Height/3)

    srcPoint = image.Pt(0, 0)
    rect = image.Rectangle{}

    plane = &DistortedPlane{}
    plane.ReadConfig()

    gc = gg.NewContext(Width, Height)
    backColor   = colornames.LightGreen
    fillColor   = colornames.CadetBlue
    borderColor = colornames.WhiteSmoke
    borderWidth = 5.0
}

// Sync'ed Draw-Funktionen.
//
func TestDrawSyncFull(t *testing.T) {
    gc.SetFillColor(colornames.Navy)
    gc.Clear()
    disp.DrawSync(gc.Image())
    gc.DrawImage(testBild.SubImage(RectFull), 0, 0)
    disp.DrawSync(gc.Image())
}
func TestDrawSyncHalve(t *testing.T) {
    gc.SetFillColor(colornames.Navy)
    gc.Clear()
    disp.DrawSync(gc.Image())
    gc.DrawImage(testBild.SubImage(RectHalve), 0, 0)
    disp.DrawSync(gc.Image())
}
func TestDrawSyncQuart(t *testing.T) {
    gc.SetFillColor(colornames.Navy)
    gc.Clear()
    disp.DrawSync(gc.Image())
    gc.DrawImage(testBild.SubImage(RectQuart), 0, 0)
    disp.DrawSync(gc.Image())
}
func TestDrawSyncCust(t *testing.T) {
    gc.SetFillColor(colornames.Navy)
    gc.Clear()
    disp.DrawSync(gc.Image())

    rect := RectQuart
    draw.Draw(workImage, rect, testBild, rect.Min, draw.Src)
    disp.DrawSync(workImage)

    rect = RectQuart.Add(image.Point{100, 100})
    draw.Draw(workImage, rect, testBild, rect.Min, draw.Src)
    disp.DrawSync(workImage)

    disp.DrawSync(workImage)
}

// Async'ed Draw-Funktionen.
//
func TestDrawAsyncFull(t *testing.T) {
    gc.SetFillColor(colornames.Navy)
    gc.Clear()
    disp.Draw(gc.Image())
    gc.DrawImage(testBild.SubImage(RectFull), 0, 0)
    disp.Draw(gc.Image())
    time.Sleep(time.Second)
}
func TestDrawAsyncHalve(t *testing.T) {
    gc.SetFillColor(colornames.Navy)
    gc.Clear()
    disp.Draw(gc.Image())
    gc.DrawImage(testBild.SubImage(RectHalve), 0, 0)
    disp.Draw(gc.Image())
    time.Sleep(time.Second)
}
func TestDrawAsyncQuart(t *testing.T) {
    gc.SetFillColor(colornames.Navy)
    gc.Clear()
    disp.Draw(gc.Image())
    gc.DrawImage(testBild.SubImage(RectQuart), 0, 0)
    disp.Draw(gc.Image())
    time.Sleep(time.Second)
}
func TestDrawAsyncCust(t *testing.T) {
    gc.SetFillColor(color.Black)
    gc.Clear()
    disp.Draw(gc.Image())

    rect := RectQuart
    draw.Draw(workImage, rect, testBild, rect.Min, draw.Src)
    disp.Draw(workImage)

    rect = RectQuart.Add(image.Point{100, 100})
    draw.Draw(workImage, rect, testBild, rect.Min, draw.Src)
    disp.Draw(workImage)
    disp.Draw(workImage)
    time.Sleep(time.Second)
}

func TestImageDiff(t *testing.T) {
    imageA := NewILIImage(image.Rect(0, 0, Width, Height))
    imageB := NewILIImage(image.Rect(0, 0, Width, Height))

    imageA.Convert(testBild)
    imageB.Convert(testBild)

    rect := imageA.Diff(imageB)
    log.Printf("difference rectangle: %v", rect)

    imageB.Set(100, 100, colornames.Navy)
    rect = imageA.Diff(imageB)
    log.Printf("difference rectangle: %v", rect)

    imageB.Set(250, 100, colornames.Navy)
    rect = imageA.Diff(imageB)
    log.Printf("difference rectangle: %v", rect)

    imageB.Convert(testBild)
    imageB.Set(100, 100, colornames.Navy)
    imageB.Set(100, 210, colornames.Navy)
    rect = imageA.Diff(imageB)
    log.Printf("difference rectangle: %v", rect)


}
func BenchmarkDiffFull(b *testing.B) {
    imageA := NewILIImage(image.Rect(0, 0, Width, Height))
    imageB := NewILIImage(image.Rect(0, 0, Width, Height))

    imageA.Convert(testBild)
    imageB.Convert(testBild)

    b.ResetTimer()
    for i:=0; i<b.N; i++ {
        rect = imageA.Diff(imageB)
    }
}
func BenchmarkDiffHalve(b *testing.B) {
    imageA := NewILIImage(image.Rect(0, 0, Width, Height))
    imageB := NewILIImage(image.Rect(0, 0, Width, Height))

    imageA.Convert(testBild)
    imageB.Convert(testBild)

    imageB.Set(Width/4, Height/4, colornames.Navy)
    imageB.Set(3*Width/4, 3*Height/4, colornames.Navy)

    b.ResetTimer()
    for i:=0; i<b.N; i++ {
        rect = imageA.Diff(imageB)
    }
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
    img := testBild.SubImage(RectFull).(*image.RGBA)
    for i := 0; i < b.N; i++ {
        pixBuf.Convert(img)
    }
}
func BenchmarkConvertHalve(b *testing.B) {
    img := testBild.SubImage(RectHalve).(*image.RGBA)
    for i := 0; i < b.N; i++ {
        pixBuf.Convert(img)
    }
}
func BenchmarkConvertQuart(b *testing.B) {
    img := testBild.SubImage(RectQuart).(*image.RGBA)
    for i := 0; i < b.N; i++ {
        pixBuf.Convert(img)
    }
}
func BenchmarkConvertCust(b *testing.B) {
    img := testBild.SubImage(RectCust).(*image.RGBA)
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
        disp.DrawSync(testBild.SubImage(RectFull))
    }
}
func BenchmarkDrawHalve(b *testing.B) {
    gc.SetFillColor(color.Black)
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        disp.DrawSync(testBild.SubImage(RectHalve))
    }
}
func BenchmarkDrawQuart(b *testing.B) {
    gc.SetFillColor(color.Black)
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        disp.DrawSync(testBild.SubImage(RectQuart))
    }
}
func BenchmarkDrawCust(b *testing.B) {
    gc.SetFillColor(color.Black)
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        disp.DrawSync(testBild.SubImage(RectCust))
    }
}

// Zeichnet eine Anzahl zufälliger Rechtecke und aktualisiert nach jedem
// Rechteck den gesamten Bildschirm.
//
func BenchmarkDrawRectangles(b *testing.B) {
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
        x, y, w, h := fWidth/2*rand.Float64(), fHeight/2*rand.Float64(),
                fWidth*rand.Float64(), fHeight*rand.Float64()
        gc.DrawRectangle(x, y, w, h)
        gc.FillStroke()
        disp.DrawSync(img)
    }
}

// Zeichnet eine Anzahl zufälliger Rechtecke und aktualisiert nach jedem
// Rechteck den gesamten Bildschirm.
//
func BenchmarkDrawCircles(b *testing.B) {
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
        x, y, r := fWidth*rand.Float64(), fHeight*rand.Float64(),
                fHeight/2*rand.Float64()
        gc.DrawCircle(x, y, r)
        gc.FillStroke()
        disp.DrawSync(img)
    }
}

func BenchmarkDrawImageGG(b *testing.B) {
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i< b.N; i++ {
        gc.DrawImage(testBild, 0, 0)
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
        draw.Draw(out, out.Bounds(), testBild, image.Point{0, 0}, draw.Src)
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
        draw.Copy(out, image.Point{0, 0}, testBild, testBild.Bounds(), draw.Src, nil)
    }
    b.StopTimer()
    disp.DrawSync(gc.Image())
}

