package adatft

import (
    "fmt"
    "time"
)

var (
    // ConvTime enthält die kumulierte Zeit, welche für das Konvertieren der
    // Bilder vom RGBA-Format in das 565-/666-Format verwendet wird.
    // Misst im Wesentlichen die aktive Zeit der Methode 'Convert'.    
    ConvWatch = NewStopwatch()
    
    // DispTime enthält die kumulierte Zeit, welche für das Senden der Bilder
    // zum Display verwendet wird. Misst im Wesentlichen die aktive Zeit
    // der Methode 'sendImage'.
    DispWatch = NewStopwatch()

    // PaintTime kann von der Applikation verwendet werden, um die kumulierte
    // Zeit zu erfassen, die von der Applikation selber zum Zeichnen des
    // Bildschirms verwendet wird.
    PaintWatch = NewStopwatch()
    AnimWatch  = NewStopwatch()
)

// Gibt eine Reihe von Messdaten aus, mit denen die Performance der Umgebung
// eingeschaetzt werden kann. Drei Bereiche werden gemessen:
//   - Die applikatorische Zeit (NumPaint, PaintTime).
//   - Die Zeit, welche fuer die Konvertierung der Bilder ins ILI-spezifische
//     Format benoetigt wird (NumConf, ConvTime).
//   - Die Zeit, welche fuer die Darstellung der Bilder auf dem TFT benoetigt
//     wird (NumDisp, DispTime).
// Als Daumenregel gilt: wenn die applikatorische Zeit pro Frame
// (PaintTime / NumPaint) groesser ist als die Zeit, welche fuer die
// Darstellung benoetigt wird (DispTime / NumDisp), dann besteht Bedarf nach
// Optimierung.
func PrintStat() {
    fmt.Printf("total:\n")
    fmt.Printf("  %d frames\n", ConvWatch.Num())
    fmt.Printf("application painting:\n")
    fmt.Printf("  %v total\n", PaintWatch.Total())
    fmt.Printf("  %v / frame\n", PaintWatch.Avg())
    fmt.Printf("buffer conversion:\n")
    fmt.Printf("  %v total\n", ConvWatch.Total())
    fmt.Printf("  %v / frame\n", ConvWatch.Avg())
    fmt.Printf("sending to SPI:\n")
    fmt.Printf("  %v total\n", DispWatch.Total())
    fmt.Printf("  %v / frame\n", DispWatch.Avg())
}

type Stopwatch struct {
    t time.Time
    d time.Duration
    n int
}

func NewStopwatch() (*Stopwatch) {
    return &Stopwatch{}
}

func (s *Stopwatch) Start() {
    s.t = time.Now()
}

func (s *Stopwatch) Stop() {
    s.d += time.Since(s.t)
    s.n += 1
}

func (s *Stopwatch) Reset() {
    s.d = 0
    s.n = 0
}

func (s *Stopwatch) Total() (time.Duration) {
    return s.d
}

func (s *Stopwatch) Avg() (time.Duration) {
    return s.d / time.Duration(s.n)
}

func (s *Stopwatch) Num() (int) {
    return s.n
}

