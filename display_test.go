package adatft

import (
	"image"
	"image/draw"
	"image/png"
	"log"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/stefan-muehlebach/gg"
	"github.com/stefan-muehlebach/gg/colors"
	draw2 "golang.org/x/image/draw"
	"periph.io/x/conn/v3/physic"
)

const (
	randSeed    = 12_345_678
	imageFile01 = "testbild01.png"
	imageFile02 = "testbild02.png"
)

var (
	disp            *Display
	pixBuf          *ILIImage
	fWidth, fHeight float64
	tempBild, testBild01, testBild02,
	workImage *image.RGBA
	rectFull, rectHalve, rectHalve02,
	rectHalve03, rectQuart, rectCust, rect image.Rectangle
	srcPoint                          image.Point
	gc                                *gg.Context
	gcImage                           *image.RGBA
	err                               error
	plane                             *DistortedPlane
	touchData                         TouchRawPos
	touchPos                          TouchPos
	backColor, fillColor, borderColor colors.Color
	borderWidth                       float64
	spiSpeed                          int64
)

func init() {
	//log.Printf("%d, %v", len(os.Args), os.Args)
	spiSpeed, err = strconv.ParseInt(os.Args[len(os.Args)-1], 10, 32)
	if err == nil {
		SPISpeedHz = physic.Frequency(spiSpeed)
	}

	disp = OpenDisplay(Rotate090)
	fWidth, fHeight = float64(Width), float64(Height)

	rectFull = image.Rect(0, 0, Width, Height)
	rectHalve = image.Rect(Width/4, Height/4, 3*Width/4, 3*Height/4)
	rectQuart = image.Rect(3*Width/8, 3*Height/8, 5*Width/8, 5*Height/8)
	rectHalve02 = image.Rect(0, Height/4, Width, 3*Height/4)
	rectHalve03 = image.Rect(Width/4, 0, 3*Width/4, Height)
	rectCust = image.Rect(0, 0, Width/3, Height/3)

	pixBuf = NewILIImage(rectFull)

	fh, err := os.Open(imageFile01)
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()
	img, err := png.Decode(fh)
	if err != nil {
		log.Fatal(err)
	}
	testBild01 = image.NewRGBA(rectFull)
	draw.Draw(testBild01, rectFull, img, image.Point{}, draw.Src)

	workImage = image.NewRGBA(rectFull)

	srcPoint = image.Pt(0, 0)

	plane = &DistortedPlane{}
	plane.ReadConfig(Rotate090)

	gc = gg.NewContext(Width, Height)
	gcImage = gc.Image().(*image.RGBA)

	backColor = colors.LightGreen
	fillColor = colors.CadetBlue
	borderColor = colors.WhiteSmoke
	borderWidth = 5.0

	rand.Seed(randSeed)
}

func TestSendFullImage(t *testing.T) {
	pixBuf.Clear()
	pixBuf.Convert(testBild01)
	disp.sendImage(pixBuf)
}
func TestSendHalveImage(t *testing.T) {
	pixBuf.Clear()
	pixBuf.Convert(testBild01)
	disp.sendImage(pixBuf.SubImage(rectHalve).(*ILIImage))
}
func TestSendQuartImage(t *testing.T) {
	pixBuf.Clear()
	pixBuf.Convert(testBild01)
	disp.sendImage(pixBuf.SubImage(rectQuart).(*ILIImage))
}

func BenchmarkSendFullImage(b *testing.B) {
	pixBuf.Clear()
	pixBuf.Convert(testBild01)
	for b.Loop() {
		disp.sendImage(pixBuf)
	}
}
func BenchmarkSendHalveImage(b *testing.B) {
	pixBuf.Clear()
	pixBuf.Convert(testBild01)
	for b.Loop() {
		disp.sendImage(pixBuf.SubImage(rectHalve).(*ILIImage))
	}
}
func BenchmarkSendQuartImage(b *testing.B) {
	pixBuf.Clear()
	pixBuf.Convert(testBild01)
	for b.Loop() {
		disp.sendImage(pixBuf.SubImage(rectQuart).(*ILIImage))
	}
}

// Test der synchronisierten Draw-Funktionen
func TestDrawSyncFull(t *testing.T) {
	gc.SetFillColor(colors.Black)
	gc.Clear()
	gc.DrawImage(testBild01.SubImage(rectFull), 0, 0)
	disp.DrawSync(gc.Image())
}
func BenchmarkDrawSyncFull(b *testing.B) {
	gc.SetFillColor(colors.Black)
	gc.Clear()
	gc.DrawImage(testBild01.SubImage(rectFull), 0, 0)
	for b.Loop() {
		disp.DrawSync(gc.Image())
	}
}
func TestDrawSyncRand(t *testing.T) {
	rand.Seed(randSeed)
	gc.SetFillColor(colors.Black)
	gc.Clear()
	for i := 0; i < 5; i++ {
		dx := rand.Intn(240) - 120
		dy := rand.Intn(180) - 90
		rect := rectQuart.Add(image.Point{dx, dy})
		draw.Draw(gcImage, rect, testBild01, rect.Min, draw.Src)
	}
	disp.DrawSync(gc.Image())
}

// Test der asynchronen Draw-Funktionen.
func TestDrawAsyncFull(t *testing.T) {
	gc.SetFillColor(colors.Black)
	gc.Clear()
	gc.DrawImage(testBild01.SubImage(rectFull), 0, 0)
	disp.Draw(gc.Image())
}
func BenchmarkDrawAsyncFull(b *testing.B) {
	gc.SetFillColor(colors.Black)
	gc.Clear()
	gc.DrawImage(testBild01.SubImage(rectFull), 0, 0)
	for b.Loop() {
		disp.Draw(gc.Image())
	}
}
func TestDrawAsyncRand(t *testing.T) {
	rand.Seed(randSeed)
	gc.SetFillColor(colors.Black)
	gc.Clear()
	for i := 0; i < 5; i++ {
		dx := rand.Intn(240) - 120
		dy := rand.Intn(180) - 90
		rect := rectQuart.Add(image.Point{dx, dy})
		draw.Draw(gcImage, rect, testBild01, rect.Min, draw.Src)
	}
	disp.Draw(gc.Image())
}

// Test des Ermittelns der Bilddifferenzen.
func TestDiff(t *testing.T) {
	img := NewILIImage(image.Rect(0, 0, Width, Height))
	img.Clear()
	pixBuf.Clear()

	// Sollte keine Unterschiede bemerken, resp. ein leeres Rechteck liefern.
	rect := pixBuf.Diff(img)
	t.Logf("no changes; diff rect: %v", rect)
	if !rect.Empty() {
		t.Errorf("no changes; want (0,0)-(0,0), got %v", rect)
	}

	// Unterschied von einem Pixel bei (100,100)
	img.Set(160, 120, colors.Navy)
	rect = pixBuf.Diff(img)
	t.Logf("one pixel changed; diff rect: %v", rect)
	if rect.Size() != image.Pt(1, 1) {
		t.Errorf("one pixel changed; want (160,120)-(161,121), got %v", rect)
	}

	// Unterschied durch zwei Pixel und dadurch Rechteck erwartet.
	img.Clear()
	img.Set(80, 60, colors.Navy)
	img.Set(239, 179, colors.Navy)
	rect = pixBuf.Diff(img)
	t.Logf("second pixel changed; diff rect: %v", rect)
	if rect.Size() != image.Pt(160, 120) {
		t.Errorf("second pixel changed; want (80,60)-(240,180), got %v", rect)
	}

	// Unterschied durch zwei Pixel und dadurch Rechteck erwartet.
	img.Clear()
	img.Set(239, 60, colors.Navy)
	img.Set(80, 179, colors.Navy)
	rect = pixBuf.Diff(img)
	t.Logf("second pixel changed; diff rect: %v", rect)
	if rect.Size() != image.Pt(160, 120) {
		t.Errorf("second pixel changed; want (80,60)-(240,180), got %v", rect)
	}

	// Unterschiede liegen ganz an den Raendern: ganzes Bild sollte neu
	// gezeichnet werden
	img.Clear()
	img.Set(Width/2, 0, colors.Navy)
	img.Set(Width-1, Height/2, colors.Navy)
	img.Set(Width/2, Height-1, colors.Navy)
	img.Set(0, Height/2, colors.Navy)
	rect = pixBuf.Diff(img)
	t.Logf("edge pixel changed; diff rect: %v", rect)
	if rect.Size() != image.Pt(Width, Height) {
		t.Errorf("edge pixel changed; want %v, %v", img.Rect.Size(), rect)
	}
}

// Misst die Zeit, welche benoetigt wird um festzustellen, welche Teile eines
// Bildes sich veraendert haben.
// Zuerst fuer den Fall, dass sich gar nichts aendert, also das gesamte Bild
// durchsucht werden muss.
func BenchmarkDiffFull(b *testing.B) {
	img := NewILIImage(rectFull)
	img.Clear()
	pixBuf.Clear()
	for b.Loop() {
		rect = pixBuf.Diff(img)
	}
}

// Und danach fuer 20 zufaellig gesetzte Punkte.
func BenchmarkDiffRand(b *testing.B) {
	rand.Seed(randSeed)
	img := NewILIImage(rectFull)
	pixBuf.Clear()
	for b.Loop() {
		img.Clear()
		x0, y0 := rand.Intn(Width/2), rand.Intn(Height/2)
		x1, y1 := Width/2+rand.Intn(Width/2), Height/2+rand.Intn(Height/2)
		img.Set(x0, y0, colors.White)
		img.Set(x1, y1, colors.White)
		rect = pixBuf.Diff(img)
	}
}

// Misst die Zeit für die Konvertierung eines Bildes im image.RGBA-Format
// ins TFT-spezifische 666-/565-Format. Es gibt dazu vier Funktionen, welche
// vier verschiedene Ausschnitte des Bildes konvertieren: Full, Halve, Quart
// und Cust (siehe auch die Variablen dstRectXXX in der Funktion init()).
func BenchmarkConvertFull(b *testing.B) {
	for b.Loop() {
		pixBuf.Convert(testBild01)
	}
}
func BenchmarkConvertRand(b *testing.B) {
	rand.Seed(randSeed)
	for b.Loop() {
		x0, y0 := rand.Intn(Width/2), rand.Intn(Height/2)
		x1, y1 := Width/2+rand.Intn(Width/2), Height/2+rand.Intn(Height/2)
		rect := image.Rect(x0, y0, x1, y1)
		img := testBild01.SubImage(rect).(*image.RGBA)
		pixBuf.Convert(img)
	}
}

// Benchmark der Konvertierung von Touchscreen-Koordinaten nach Bildschirm-
// Koordinaten. TO DO: ev. sollte die Erzeugung der Touchscreen-Koordinaten
// aus der Zeitmessung entfernt werden.
func BenchmarkTransformPoint(b *testing.B) {
	rand.Seed(randSeed)
	x, y := uint16(rand.Intn(2<<16)), uint16(rand.Intn(2<<16))
	touchData = TouchRawPos{x, y, 0}
	for b.Loop() {
		touchPos, _ = plane.Transform(touchData)
	}
}

// Misst die Zeit für die Darstellung eines Bildes (resp. eines Teils davon)
// auf dem TFT. Es gibt dazu vier Funktionen, welche vier verschiedene
// Ausschnitte des Bildes darstellen: Full, Halve, Quart und Cust (siehe auch
// die Variablen dstRectXXX in der Funktion init()).
func BenchmarkSendFull(b *testing.B) {
	pixBuf.Convert(testBild01)
	for b.Loop() {
		disp.sendImage(pixBuf)
	}
}
func BenchmarkSendRand(b *testing.B) {
	rand.Seed(randSeed)
	pixBuf.Convert(testBild01)
	for b.Loop() {
		x0, y0 := rand.Intn(Width), rand.Intn(Height)
		x1, y1 := rand.Intn(Width), rand.Intn(Height)
		rect := image.Rect(x0, y0, x1, y1)
		img := pixBuf.SubImage(rect).(*ILIImage)
		disp.sendImage(img)
	}
}

// Misst schliesslich die Zeit, die fuer den gesamten Ablauf (Konvertierung,
// Differenz bilden und zum Display senden) verwendet wird.
func BenchmarkDrawFull(b *testing.B) {
	img := NewILIImage(image.Rect(0, 0, Width, Height))
	pixBuf.Clear()
	for b.Loop() {
		img.Convert(testBild01)
		rect := pixBuf.Diff(img)
		disp.sendImage(img.SubImage(rect).(*ILIImage))
	}
}
func BenchmarkDrawRand(b *testing.B) {
	rand.Seed(randSeed)
	imgA := NewILIImage(image.Rect(0, 0, Width, Height))
	imgB := NewILIImage(image.Rect(0, 0, Width, Height))
	imgB.Convert(testBild01)
	for b.Loop() {
		for j := 0; j < 2; j++ {
			x, y := rand.Intn(Width), rand.Intn(Height)
			testBild01.Set(x, y, colors.YellowGreen)
		}
		imgA.Convert(testBild01)
		rect := imgA.Diff(imgB)
		disp.sendImage(imgA.SubImage(rect).(*ILIImage))
		imgA, imgB = imgB, imgA
	}
}

// Zeichnet eine Anzahl zufälliger Rechtecke und aktualisiert nach jedem
// Rechteck den gesamten Bildschirm.
func BenchmarkDrawRectangles(b *testing.B) {
	var img *image.RGBA

	img = gc.Image().(*image.RGBA)

	rand.Seed(randSeed)
	gc.SetFillColor(backColor)
	gc.Clear()
	disp.Draw(gc.Image())
	gc.SetFillColor(fillColor)
	gc.SetStrokeColor(borderColor)
	gc.SetStrokeWidth(borderWidth)
	for b.Loop() {
		x, y, w, h := fWidth/2*rand.Float64(), fHeight/2*rand.Float64(),
			fWidth*rand.Float64(), fHeight*rand.Float64()
		gc.DrawRectangle(x, y, w, h)
		gc.FillStroke()
		disp.Draw(img)
	}
}

// Zeichnet eine Anzahl zufälliger Rechtecke und aktualisiert nach jedem
// Rechteck den gesamten Bildschirm.
func BenchmarkDrawCircles(b *testing.B) {
	var img *image.RGBA

	img = gc.Image().(*image.RGBA)

	rand.Seed(randSeed)
	gc.SetFillColor(backColor)
	gc.Clear()
	disp.Draw(gc.Image())
	gc.SetFillColor(fillColor)
	gc.SetStrokeColor(borderColor)
	gc.SetStrokeWidth(borderWidth)
	for b.Loop() {
		x, y, r := fWidth*rand.Float64(), fHeight*rand.Float64(),
			fHeight/2*rand.Float64()
		gc.DrawCircle(x, y, r)
		gc.FillStroke()
		disp.Draw(img)
	}
}

func BenchmarkDrawImageGG(b *testing.B) {
	gc.Clear()
	disp.DrawSync(gc.Image())
	for b.Loop() {
		gc.DrawImage(testBild01, 0, 0)
	}
	disp.DrawSync(gc.Image())
}

func BenchmarkDrawImageGo(b *testing.B) {
	out := gc.Image().(*image.RGBA)
	gc.Clear()
	disp.DrawSync(gc.Image())
	for b.Loop() {
		draw.Draw(out, out.Bounds(), testBild01, image.Point{0, 0}, draw.Src)
	}
	disp.DrawSync(gc.Image())
}

func BenchmarkDrawImageGo2(b *testing.B) {
	out := gc.Image().(*image.RGBA)
	gc.Clear()
	disp.DrawSync(gc.Image())
	for b.Loop() {
		draw2.Draw(out, out.Bounds(), testBild01, image.Point{0, 0}, draw2.Src)
	}
	disp.DrawSync(gc.Image())
}

func BenchmarkCopyImageGo2(b *testing.B) {
	out := gc.Image().(*image.RGBA)
	gc.Clear()
	disp.DrawSync(gc.Image())
	for b.Loop() {
		draw2.Copy(out, image.Point{0, 0}, testBild01, testBild01.Bounds(), draw2.Src, nil)
	}
	disp.DrawSync(gc.Image())
}
