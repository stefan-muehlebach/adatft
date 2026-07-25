package adatft

import (
	//"flag"
	"image"
	"image/draw"
	"image/png"
	"log"
	"math/rand/v2"
	"os"
	"testing"

	. "github.com/stefan-muehlebach/adatft/iliimg"
	"github.com/stefan-muehlebach/gg"
	"github.com/stefan-muehlebach/gg/colors"
	draw2 "golang.org/x/image/draw"
)

const (
	randSeed1   = 12_345_678
	randSeed2   = 87_654_321
	imageFile01 = "testbild01.png"
	imageFile02 = "testbild02.png"
)

var (
	disp                                        *Display
	touch                                       *Touch
	pointer                                     *Pointer
	pixBuf                                      *ILIImage
	width, height                               int
	fWidth, fHeight                             float64
	tempBild, testBild01, testBild02, workImage *image.RGBA
	rectFull, rectCust, rect                    image.Rectangle
	srcPoint                                    image.Point
	gc                                          *gg.Context
	gcImage                                     *image.RGBA
	err                                         error
	plane                                       *DistortedPlane
	touchData                                   TouchRawPos
	touchPos                                    TouchPos
	backColor, fillColor, borderColor           colors.RGBA
	borderWidth                                 float64
	spiSpeed                                    int64
	pcg                                         *rand.PCG
	rnd                                         *rand.Rand
)

// func TestMain(m *testing.M) {
func init() {
	//flag.Var(&dspSpeed, "speed", "Freq (Hz) of SPI bus")
	//flag.Parse()

	disp = OpenDisplay(Rotate000)
	bounds := disp.Bounds()
	width, height = bounds.Dx(), bounds.Dy()
	fWidth, fHeight = float64(width), float64(height)

	rectFull = image.Rect(0, 0, width, height)
	rectCust = image.Rect(0, 0, width/3, height/3)

	pixBuf = NewILIImage(rectFull)
	pixBuf.SetLSBFirst()

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

	gc = gg.NewContext(width, height)
	gcImage = gc.Image().(*image.RGBA)

	backColor = colors.LightGreen
	fillColor = colors.CadetBlue
	borderColor = colors.WhiteSmoke
	borderWidth = 5.0

	pcg = rand.NewPCG(randSeed1, randSeed2)
	rnd = rand.New(pcg)

	//m.Run()

	//disp.Close()
}

func TestSendImageFull(t *testing.T) {
	pixBuf.Clear()
	pixBuf.Convert(testBild01)
	disp.dspi.Send(pixBuf)
}

// Misst die Zeit für die Darstellung eines Bildes (resp. eines Teils davon)
// auf dem TFT. Es gibt dazu vier Funktionen, welche vier verschiedene
// Ausschnitte des Bildes darstellen: Full, Halve, Quart und Cust (siehe auch
// die Variablen dstRectXXX in der Funktion init()).
func BenchmarkSendImageFull(b *testing.B) {
	pixBuf.Clear()
	pixBuf.Convert(testBild01)
	for b.Loop() {
		disp.dspi.Send(pixBuf)
	}
}
func BenchmarkSendImageRand(b *testing.B) {
	pcg.Seed(randSeed1, randSeed2)
	pixBuf.Convert(testBild01)
	for b.Loop() {
		x0, y0 := rnd.IntN(width), rnd.IntN(height)
		x1, y1 := rnd.IntN(width), rnd.IntN(height)
		rect := image.Rect(x0, y0, x1, y1)
		img := pixBuf.SubImage(rect).(*ILIImage)
		disp.dspi.Send(img)
	}
}

// Test der synchronisierten Draw-Funktionen
func TestDrawSyncFull(t *testing.T) {
	gc.SetFillColor(colors.Black)
	gc.Clear()
	//gc.DrawImage(testBild01.SubImage(rectFull), 0, 0)
	draw.Draw(gcImage, rectFull, testBild01, rectFull.Min, draw.Src)
	disp.DrawSync(gc.Image())
}
func TestDrawSyncRand(t *testing.T) {
	pcg.Seed(randSeed1, randSeed2)
	gc.SetFillColor(colors.Black)
	gc.Clear()
	for i := 0; i < 5; i++ {
		x0, y0 := rnd.IntN(width), rnd.IntN(height)
		x1, y1 := rnd.IntN(width), rnd.IntN(height)
		rect := image.Rect(x0, y0, x1, y1)
		draw.Draw(gcImage, rect, testBild01, rect.Min, draw.Src)
	}
	disp.DrawSync(gc.Image())
}
func BenchmarkDrawSyncFull(b *testing.B) {
	gc.SetFillColor(colors.Black)
	gc.Clear()
	//gc.DrawImage(testBild01.SubImage(rectFull), 0, 0)
	draw.Draw(gcImage, rectFull, testBild01, rectFull.Min, draw.Src)
	for b.Loop() {
		disp.DrawSync(gc.Image())
	}
}

// Test der asynchronen Draw-Funktionen.
func TestDrawFull(t *testing.T) {
	gc.SetFillColor(colors.Black)
	gc.Clear()
	draw.Draw(gcImage, rectFull, testBild01, rectFull.Min, draw.Src)
	//gc.DrawImage(testBild01, 0, 0)
	disp.Draw(gc.Image())
}
func BenchmarkDrawFull(b *testing.B) {
	gc.SetFillColor(colors.Black)
	gc.Clear()
	draw.Draw(gcImage, rectFull, testBild01, rectFull.Min, draw.Src)
	//gc.DrawImage(testBild01.SubImage(rectFull), 0, 0)
	for b.Loop() {
		disp.Draw(gc.Image())
	}
}
func TestDrawRand(t *testing.T) {
	pcg.Seed(randSeed1, randSeed2)
	gc.SetFillColor(colors.Black)
	gc.Clear()
	for i := 0; i < 5; i++ {
		x0, y0 := rnd.IntN(width), rnd.IntN(height)
		x1, y1 := rnd.IntN(width), rnd.IntN(height)
		rect := image.Rect(x0, y0, x1, y1)
		draw.Draw(gcImage, rect, testBild01, rect.Min, draw.Src)
	}
	disp.Draw(gc.Image())
}

// Test des Ermittelns der Bilddifferenzen.
func TestDiff(t *testing.T) {
	img := NewILIImage(image.Rect(0, 0, width, height))
	img.Clear()
	pixBuf.Clear()

	// Sollte keine Unterschiede bemerken, resp. ein leeres Rechteck liefern.
	rect := pixBuf.Diff(img)
	//t.Logf("no changes; diff rect: %v", rect)
	if !rect.Empty() {
		t.Errorf("no changes; want (0,0)-(0,0), got %v", rect)
	}

	// Unterschied von einem Pixel bei (100,100)
	img.Set(160, 120, colors.Navy)
	rect = pixBuf.Diff(img)
	//t.Logf("one pixel changed; diff rect: %v", rect)
	if rect.Size() != image.Pt(1, 1) {
		t.Errorf("one pixel changed; want (160,120)-(161,121), got %v", rect)
	}

	// Unterschied durch zwei Pixel und dadurch Rechteck erwartet.
	img.Clear()
	img.Set(80, 60, colors.Navy)
	img.Set(239, 179, colors.Navy)
	rect = pixBuf.Diff(img)
	//t.Logf("second pixel changed; diff rect: %v", rect)
	if rect.Size() != image.Pt(160, 120) {
		t.Errorf("second pixel changed; want (80,60)-(240,180), got %v", rect)
	}

	// Unterschied durch zwei Pixel und dadurch Rechteck erwartet.
	img.Clear()
	img.Set(239, 60, colors.Navy)
	img.Set(80, 179, colors.Navy)
	rect = pixBuf.Diff(img)
	//t.Logf("second pixel changed; diff rect: %v", rect)
	if rect.Size() != image.Pt(160, 120) {
		t.Errorf("second pixel changed; want (80,60)-(240,180), got %v", rect)
	}

	// Unterschiede liegen ganz an den Raendern: ganzes Bild sollte neu
	// gezeichnet werden
	img.Clear()
	img.Set(width/2, 0, colors.Navy)
	img.Set(width-1, height/2, colors.Navy)
	img.Set(width/2, height-1, colors.Navy)
	img.Set(0, height/2, colors.Navy)
	rect = pixBuf.Diff(img)
	//t.Logf("edge pixel changed; diff rect: %v", rect)
	if rect.Size() != image.Pt(width, height) {
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
	pcg.Seed(randSeed1, randSeed2)
	img := NewILIImage(rectFull)
	pixBuf.Clear()
	R, G, B := byte(0), byte(0), byte(0)
	for b.Loop() {
		img.Clear()
		color := colors.RGBA{R: R, G: G, B: B, A: 0xff}
		x0, y0 := rnd.IntN(width/2), rnd.IntN(height/2)
		x1, y1 := width/2+rnd.IntN(width/2), height/2+rnd.IntN(height/2)
		img.Set(x0, y0, color)
		img.Set(x1, y1, color)
		rect = pixBuf.Diff(img)
		if R < 255 {
			R++
		} else {
			R = 0
			if G < 255 {
				G++
			} else {
				G = 0
				if B < 255 {
					B++
				} else {
					B = 0
				}
			}
		}
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
	pcg.Seed(randSeed1, randSeed2)
	for b.Loop() {
		x0, y0 := rnd.IntN(width/2), rnd.IntN(height/2)
		x1, y1 := width/2+rnd.IntN(width/2), height/2+rnd.IntN(height/2)
		rect := image.Rect(x0, y0, x1, y1)
		img := testBild01.SubImage(rect).(*image.RGBA)
		pixBuf.Convert(img)
	}
}

// Benchmark der Konvertierung von Touchscreen-Koordinaten nach Bildschirm-
// Koordinaten. TO DO: ev. sollte die Erzeugung der Touchscreen-Koordinaten
// aus der Zeitmessung entfernt werden.
func BenchmarkTransform(b *testing.B) {
	pcg.Seed(randSeed1, randSeed2)
	x, y := uint16(rnd.IntN(2<<16)), uint16(rnd.IntN(2<<16))
	touchData = TouchRawPos{x, y, 0}
	for b.Loop() {
		touchPos, _ = plane.Transform(touchData)
	}
}

// Misst schliesslich die Zeit, die fuer den gesamten Ablauf (Konvertierung,
// Differenz bilden und zum Display senden) verwendet wird.
func BenchmarkAllFull(b *testing.B) {
	img := NewILIImage(image.Rect(0, 0, width, height))
	pixBuf.Clear()
	for b.Loop() {
		img.Convert(testBild01)
		rect := pixBuf.Diff(img)
		disp.dspi.Send(img.SubImage(rect).(*ILIImage))
	}
}
func BenchmarkAllRand(b *testing.B) {
	pcg.Seed(randSeed1, randSeed2)
	imgA := NewILIImage(image.Rect(0, 0, width, height))
	imgB := NewILIImage(image.Rect(0, 0, width, height))
	imgB.Convert(testBild01)
	for b.Loop() {
		for j := 0; j < 2; j++ {
			x, y := rnd.IntN(width), rnd.IntN(height)
			testBild01.Set(x, y, colors.YellowGreen)
		}
		imgA.Convert(testBild01)
		rect := imgA.Diff(imgB)
		disp.dspi.Send(imgA.SubImage(rect).(*ILIImage))
		imgA, imgB = imgB, imgA
	}
}

// Zeichnet eine Anzahl zufälliger Rechtecke und aktualisiert nach jedem
// Rechteck den gesamten Bildschirm.
func BenchmarkDrawRectangles(b *testing.B) {
	var img *image.RGBA

	img = gc.Image().(*image.RGBA)

	pcg.Seed(randSeed1, randSeed2)
	gc.SetFillColor(backColor)
	gc.Clear()
	disp.Draw(gc.Image())
	gc.SetFillColor(fillColor)
	gc.SetStrokeColor(borderColor)
	gc.SetStrokeWidth(borderWidth)
	for b.Loop() {
		x, y, w, h := fWidth/2*rnd.Float64(), fHeight/2*rnd.Float64(),
			fWidth*rnd.Float64(), fHeight*rnd.Float64()
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

	pcg.Seed(randSeed1, randSeed2)
	gc.SetFillColor(backColor)
	gc.Clear()
	disp.Draw(gc.Image())
	gc.SetFillColor(fillColor)
	gc.SetStrokeColor(borderColor)
	gc.SetStrokeWidth(borderWidth)
	for b.Loop() {
		x, y, r := fWidth*rnd.Float64(), fHeight*rnd.Float64(),
			fHeight/2*rnd.Float64()
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

func TestPointerDevice(t *testing.T) {
	pointer = OpenPointer("/dev/input/event0", disp.Bounds())
	i := 0
	for event := range touch.EventQ {
		t.Logf("Event received: %v\n", event)
		i += 1
		if i >= 100 {
			break
		}
	}
}
