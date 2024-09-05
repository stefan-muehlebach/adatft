package adatft

import (
	"fmt"
	"time"
)

// Dieser Typ dient der Zeitmessung.
type Stopwatch struct {
	t time.Time
	d time.Duration
	n int
}

func NewStopwatch() *Stopwatch {
	return &Stopwatch{}
}

// Mit Start wird eine neue Messung begonnen.
func (s *Stopwatch) Start() {
	s.t = time.Now()
}

// Stop beendet die Messung und aktualisiert die Variablen, welche die totale
// Messdauer als auch die Anzahl Messungen enthalten.
func (s *Stopwatch) Stop() {
	s.d += time.Since(s.t)
	s.n += 1
}

// Setzt die gemessene Dauer auf 0 und die Anzahl Messungen ebenfalls.
func (s *Stopwatch) Reset() {
	s.d = 0
	s.n = 0
}

// Retourniert die totale Messdauer.
func (s *Stopwatch) Total() time.Duration {
	return s.d
}

// Retourniert die Anzahl Messungen.
func (s *Stopwatch) Num() int {
	return s.n
}

// Berechnet die durchschnittliche Messdauer (also den Quotienten von
// Total() / Num()).
func (s *Stopwatch) Avg() time.Duration {
	if s.n == 0 {
		return 0
	}
	return s.d / time.Duration(s.n)
}

var (
	// Mit ConvWatch wird die Zeit gemessen, welche für das Konvertieren der
	// Bilder vom RGBA-Format in das 565-/666-Format verwendet wird.
	// Misst im Wesentlichen die aktive Zeit der Methode 'Convert'.
	ConvWatch = NewStopwatch()

	// Mit DispWatch wird die Zeit gemessen, welche für das Senden der Bilder
	// zum Display verwendet wird. Misst im Wesentlichen die aktive Zeit
	// der Methode 'sendImage'.
	DispWatch = NewStopwatch()

	// PaintWatch und AnimWatch koennen von der Applikation zum Messen der
	// Zeiten fuer das Zeichnen, resp. Animieren des Bildschirminhalts
	// verwendet werden. Das Package macht von diesen Zeitmessern keinen
	// Gebrauch.
	PaintWatch = NewStopwatch()
	AnimWatch  = NewStopwatch()
)

// Gibt eine Reihe von Messdaten aus, mit denen die Performance der Umgebung
// eingeschaetzt werden kann.
//
// Als Daumenregel gilt: wenn die applikatorische Zeit pro Frame
// (PaintWatch.Avg()) groesser ist als die Zeit, welche fuer die
// Darstellung benoetigt wird (DispWatch.Avg()), dann besteht Bedarf nach
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
