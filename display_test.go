package adatft

import (
    "image"
    "image/color"
    "image/png"
    "log"
    "math/rand"
    "os"
    "testing"
    . "github.com/stefan-muehlebach/adatft/ili9341"
    "github.com/stefan-muehlebach/gg"
    "golang.org/x/image/draw"
)

const (
    testImage = "testbild.png"
    speedHz   = 50_000_000
)

var (
    disp *Display
    pixBuf *Buffer
    img *image.RGBA
    gc *gg.Context
    err error
    plane *DistortedPlane
    touchData TouchData
    touchPos  TouchPos
    fillColor, strokeColor color.Color
)

func init() {
    Init()
    disp = &Display{}
    disp.dspi = OpenILI9341(speedHz)
    disp.InitDisplay(Rotate000)

    pixBuf = NewBuffer()

    fh, err := os.Open(testImage)
    if err != nil {
        log.Fatal(err)
    }
    defer fh.Close()
    tmp, err := png.Decode(fh)
    if err != nil {
        log.Fatal(err)
    }
    img = tmp.(*image.RGBA)

    plane = &DistortedPlane{}
    plane.ReadConfig()

    gc = gg.NewContext(Width, Height)
    fillColor   = color.RGBA{125, 52, 26, 160}
    strokeColor = color.White
}

func BenchmarkConvertImage(b *testing.B) {
    for i := 0; i < b.N; i++ {
        pixBuf.Convert(img)
    }
} 

func BenchmarkSendBuffer(b *testing.B) {
    for i := 0; i < b.N; i++ {
        disp.DrawBuffer(pixBuf)
    }
}

func BenchmarkTransformPoint(b *testing.B) {
    for i := 0; i< b.N; i++ {
        x, y := uint16(rand.Intn(2 << 16)), uint16(rand.Intn(2 << 16))
        touchData = TouchData{x, y}
        touchPos, _ = plane.Transform(touchData)
    }
}

func BenchmarkDrawRectangle(b *testing.B) {
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i< b.N; i++ {
        x, y, w, h := 160.0*rand.Float64(), 120.0*rand.Float64(),
                160.0*rand.Float64(), 120.0*rand.Float64()
        gc.DrawRectangle(x, y, w, h)
        gc.SetLineWidth(2.0)
        gc.SetFillColor(fillColor)
        gc.SetStrokeColor(strokeColor)
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
        gc.SetLineWidth(2.0)
        gc.SetFillColor(fillColor)
        gc.SetStrokeColor(strokeColor)
        gc.FillStroke()
        gc.ResetClip()
    }
    b.StopTimer()
    disp.DrawSync(gc.Image())
}

func BenchmarkDrawCircle(b *testing.B) {
    gc.Clear()
    disp.DrawSync(gc.Image())
    b.ResetTimer()
    for i := 0; i< b.N; i++ {
        x, y, r := 320.0*rand.Float64(), 240.0*rand.Float64(),
                20.0+40.0*rand.Float64()
        gc.DrawCircle(x, y, r)
        gc.SetLineWidth(2.0)
        gc.SetFillColor(fillColor)
        gc.SetStrokeColor(strokeColor)
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
        gc.SetLineWidth(2.0)
        gc.SetFillColor(fillColor)
        gc.SetStrokeColor(strokeColor)
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
        gc.DrawImage(img, 0, 0)
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
        draw.Draw(out, out.Bounds(), img, image.Point{0, 0}, draw.Src)
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
        draw.Copy(out, image.Point{0, 0}, img, img.Bounds(), draw.Src, nil)
    }
    b.StopTimer()
    disp.DrawSync(gc.Image())
}



