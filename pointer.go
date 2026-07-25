package adatft

import (
	"image"
	"log"
	"time"
	"fmt"
	"os"
	"github.com/holoplot/go-evdev"
	"github.com/stefan-muehlebach/gg/geom"
)

// Dies sind alle Pen- oder Touch-Events, auf welche man sich abonnieren kann.
// Aktuell sind dies 3 Events: Press, Drag und Release, wobei die Reihenfolge
// der Events immer die folgende ist:
//
//	Press -> Drag -> Drag -> ... -> Release
type PointerEventType uint8

const (
	// PenPress wird erzeugt, wenn der Touch-Screen gedrueckt wird.
	PointerPress PointerEventType = iota
	// PenDrag wird erzeugt, wenn sich die Position des Drucks auf dem
	// Touchscreen veraendert.
	PointerDrag
	// PenRelease wird erzeugt, wenn der Druck auf dem Touch-Screen nicht
	// mehr gemessen werden kann.
	PointerRelease
	PointerMove
)

func (pet PointerEventType) String() string {
	switch pet {
	case PointerPress:
		return "PointerPress"
	case PointerDrag:
		return "PointerDrag"
	case PointerRelease:
		return "PointerRelease"
	case PointerMove:
		return "PointerMove"
	}
	return "(unknown event)"
}

// Während in diesem Typ die kalibrierten Postitionsdaten abgelegt werden.
type PointerPos image.Point

func (p PointerPos) String() string {
	//return fmt.Sprintf("(%5.1f, %5.1f)", tp.X, tp.Y)
	return fmt.Sprintf("(%5d, %5d)", p.X, p.Y)
}

func (p1 PointerPos) Near(p2 PointerPos) bool {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return dx*dx+dy*dy <= 50
}

// Dies ist der Funktionstyp für den PenEvent-Handler - also jene Funktion,
// welche beim Eintreffen eines Interrupts vom STMPE610 aufgerufen werden
// soll.
type PointerEventChannelType chan PointerEvent

// Dieser Typ steht fuer das SPI Interface zum STMPE - dem Touchscreen.
type Pointer struct {
	Pos                          PointerPos
	Cursor                       *Cursor
	Wheel                        int
	LeftBtn, MiddleBtn, RightBtn int32
	EventQ                       PointerEventChannelType
	dev                          *evdev.InputDevice
	rect                         image.Rectangle
	minWheel, maxWheel           int

	isOpen bool
}

var (
	devFile = "/dev/input/event4"
)

// Funktionen
func OpenPointer() *Pointer {
	var err error

	p := &Pointer{}
	p.dev, err = evdev.Open(devFile)
	if err != nil {
		log.Fatalf("evdev.Open failed: %v", err)
	}
	p.Cursor = OpenCursor("left_ptr.png", image.Point{6, 3})
	p.LeftBtn, p.MiddleBtn, p.RightBtn = 0, 0, 0
	p.EventQ = make(chan PointerEvent, eventQueueSize)
	p.isOpen = true
	go p.processEvents()
	return p
}

func (p *Pointer) Close() {
	p.isOpen = false
	close(p.EventQ)
	p.dev.Close()
}

func (p *Pointer) SetPosRange(rect geom.Rectangle) {
	p.rect = rect.Int()
	p.Pos = PointerPos(rect.Center().Int())
}

func (p *Pointer) SetWheelRange(ini, min, max int) {
	p.Wheel = ini
	p.minWheel = min
	p.maxWheel = max
}

// Diese Funktion wird von 'aussen' aufgerufen und gibt das nächste Pen-Event
// zurück. Es ist eine Alternative zum Lesen aus der öffentlichen Event-Queue.
func (p *Pointer) WaitForEvent() PointerEvent {
	return <-p.EventQ
}

// Mit dieser Funktion wird ein neues Pen-Event in die zentrale Event-Queue
// gestellt (welche dann von der Applikation ausgelesen werden muss).
// Diese Operation darf nicht blockierend ausgeführt werden, andernfalls
// würde der Event-Handler blockiert - was in meinen Augen gravierender ist.
//
// Mit dem auskommentierten Code kann für Testzwecke dafür gesorgt werden,
// dass bei einem Fehler ein Runtime-Panic ausgelöst wird.
func (p *Pointer) enqueueEvent(ev PointerEvent) {
	pev.Time = time.Now()
	defer func() {
		if x := recover(); x != nil {
			log.Printf("Runtime panic: %v\n", x)
		}
	}()
	select {
	case p.EventQ <- ev:
	default:
		log.Printf("Sending not possible: event queue full!\n")
	}
}

// Jedes Ereignis des Touchscreens wird durch eine Variable des Typs
// 'Event' repraesentiert.
type PointerEvent struct {
	Type              PointerEventType
	Time              time.Time
	Pos               PointerPos
	Wheel             int
	LeftBtn, MiddleBtn, RightBtn int32
}

var (
	pev PointerEvent
)

func (p *Pointer) processEvents() {
	for {
		e, err := p.dev.ReadOne()
		if err != nil {
			log.Fatalf("ReadOne() failed: %v", err)
		}
		switch e.Type {
		case evdev.EV_REL:
			switch e.Code {
			case evdev.REL_X:
				p.Pos.X += int(e.Value)
				if p.Pos.X < 0 {
					p.Pos.X = 0
				} else if p.Pos.X >= p.rect.Max.X {
					p.Pos.X = p.rect.Max.X - 1
				}
			case evdev.REL_Y:
				p.Pos.Y += int(e.Value)
				if p.Pos.Y < 0 {
					p.Pos.Y = 0
				} else if p.Pos.Y >= p.rect.Max.Y {
					p.Pos.Y = p.rect.Max.Y - 1
				}
			case evdev.REL_WHEEL:
				p.Wheel += int(e.Value)
				if p.Wheel < p.minWheel {
					p.Wheel = p.minWheel
				} else if p.Wheel >= p.maxWheel {
					p.Wheel = p.maxWheel - 1
				}
			default:
				continue
			}
			if p.LeftBtn != 0 || p.RightBtn != 0 {
				pev.Type = PointerDrag
			} else {
				pev.Type = PointerMove
			}
			pev.Pos = p.Pos
			pev.Wheel = p.Wheel

		case evdev.EV_KEY:
			switch e.Code {
			case evdev.BTN_LEFT, evdev.BTN_MIDDLE, evdev.BTN_RIGHT:
				if e.Value == 0 {
					pev.Type = PointerRelease
				} else {
					pev.Type = PointerPress
				}
				switch e.Code {
				case evdev.BTN_LEFT:
					p.LeftBtn = e.Value
				case evdev.BTN_MIDDLE:
					p.MiddleBtn = e.Value
				case evdev.BTN_RIGHT:
					p.RightBtn = e.Value
				}
				pev.LeftBtn = p.LeftBtn
				pev.MiddleBtn = p.MiddleBtn
				pev.RightBtn = p.RightBtn
			default:
				continue
			}

		default:
			continue
		}
		p.enqueueEvent(pev)
	}
}

//-----------------------------------------------------------------------------

type Cursor struct {
    img image.Image
    hotspot image.Point
}

func OpenCursor(fileName string, hotspot image.Point) *Cursor {
    c := &Cursor{}
    fh, err := os.Open(fileName)
    if err != nil {
        log.Fatalf("Couldn't open image file: %v", err)
    }
    defer fh.Close()
    c.img, _, err = image.Decode(fh)
    if err != nil {
        log.Fatalf("Couldn't decode image data: %v", err)
    }
    c.hotspot = hotspot
    return c
}   

