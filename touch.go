package adatft

import (
	"fmt"
	"log"
	"time"

	hw "github.com/stefan-muehlebach/adatft/stmpe610"
)

// Mit diesen Konstanten wird einerseits die Frequenz fuer den SPI-Bus
// definiert und die Groesse der Queue fuer die Touch-Events festgelegt.
const (
	tchSpeedHz     = 500_000
	eventQueueSize = 30
	sampleTime     = 7 * time.Millisecond
)

// Dies sind alle Pen- oder Touch-Events, auf welche man sich abonnieren kann.
// Aktuell sind dies 3 Events: Press, Drag und Release, wobei die Reihenfolge
// der Events immer die folgende ist:
//
//	Press -> Drag -> Drag -> ... -> Release
type PenEventType uint8

const (
	// PenPress wird erzeugt, wenn der Touch-Screen gedrueckt wird.
	PenPress PenEventType = iota
	// PenDrag wird erzeugt, wenn sich die Position des Drucks auf dem
	// Touchscreen veraendert.
	PenDrag
	// PenRelease wird erzeugt, wenn der Druck auf dem Touch-Screen nicht
	// mehr gemessen werden kann.
	PenRelease
)

func (pet PenEventType) String() string {
	switch pet {
	case PenPress:
		return "PenPress"
	case PenDrag:
		return "PenDrag"
	case PenRelease:
		return "PenRelease"
	}
	return "(unknown event)"
}

// Dieser Typ enthält die rohen, unkalibrierten Display-Daten
type TouchRawPos struct {
	RawX, RawY uint16
	RawZ       uint8
}

func (td TouchRawPos) String() string {
	//return fmt.Sprintf("(%4d, %4d)", td.RawX, td.RawY)
	return fmt.Sprintf("(%4d, %4d, %3d)", td.RawX, td.RawY, td.RawZ)
}

// Während in diesem Typ die kalibrierten Postitionsdaten abgelegt werden.
type TouchPos struct {
	X, Y float64
	Z    uint8
}

func (tp TouchPos) String() string {
	//return fmt.Sprintf("(%5.1f, %5.1f)", tp.X, tp.Y)
	return fmt.Sprintf("(%5.1f, %5.1f, %08b)", tp.X, tp.Y, tp.Z)
}

// Jedes Ereignis des Touchscreens wird durch eine Variable des Typs
// 'Event' repraesentiert.
type PenEvent struct {
	Type PenEventType
	TouchRawPos
	TouchPos
	Time time.Time
	//FifoSize uint8
}

func (p1 TouchPos) Near(p2 TouchPos) bool {
	var dx, dy float64
	dx = p1.X - p2.X
	dy = p1.Y - p2.Y
	return dx*dx+dy*dy <= 50
}

// Dies ist der Funktionstyp für den PenEvent-Handler - also jene Funktion,
// welche beim Eintreffen eines Interrupts vom STMPE610 aufgerufen werden
// soll.
type PenEventHandlerType func(event PenEvent)

type PenEventChannelType chan PenEvent

// Dieser Typ steht fuer das SPI Interface zum STMPE - dem Touchscreen.
type Touch struct {
	tspi   TouchInterface
	EventQ PenEventChannelType
	plane  DistortedPlane
	isOpen bool
}

// Funktionen
func OpenTouch(rot RotationType) *Touch {
	var tch *Touch
	var devId uint16
	var revNr uint8

	tch = &Touch{}
	if isRaspberry {
		tch.tspi = hw.Open(tchSpeedHz)
	} else {
		tch.tspi = hw.OpenDummy(tchSpeedHz)
	}

	revNr = tch.tspi.ReadReg8(hw.ID_VER)
	devId = tch.tspi.ReadReg16(hw.CHIP_ID)
	if (devId != 0x0811) || (revNr != 0x03) {
		adalog.Fatalf("device id and/or revision numbers are not as expected: got (0x%04x, 0x%02x) should be (0x0811, 0x03)\n", devId, revNr)
	}

	// Initialisiere die Queue für applikatorische Events und setze den
	// Interrupt-Handler für Touch-Events.
	tch.EventQ = make(chan PenEvent, eventQueueSize)
	tch.tspi.SetCallback(eventDispatcher, tch)

	tch.tspi.Init(nil)
	tch.isOpen = true

	tch.plane.ReadConfig(rot)

	return tch
}

func (tch *Touch) Close() {
	tch.isOpen = false
	close(tch.EventQ)
	tch.tspi.Close()
}

// Mit dieser Funktion wird ein neues Pen-Event in die zentrale Event-Queue
// gestellt (welche dann von der Applikation ausgelesen werden muss).
// Diese Operation darf nicht blockierend ausgeführt werden, andernfalls
// würde der Event-Handler blockiert - was in meinen Augen gravierender ist.
//
// Mit dem auskommentierten Code kann für Testzwecke dafür gesorgt werden,
// dass bei einem Fehler ein Runtime-Panic ausgelöst wird.
func (tch *Touch) enqueueEvent(event PenEvent) {
	defer func() {
		if x := recover(); x != nil {
			log.Printf("Runtime panic: %v\n", x)
		}
	}()
	select {
	case tch.EventQ <- event:
	default:
		log.Printf("Sending not possible: event queue full!\n")
	}
}

// Diese Funktion wird von 'aussen' aufgerufen und gibt das nächste Pen-Event
// zurück. Es ist eine Alternative zum Lesen aus der öffentlichen Event-Queue.
func (tch *Touch) WaitForEvent() PenEvent {
	return <-tch.EventQ
}

// Diese Hilfsfunktion dient der einfacheren Erstellung eines Event-Objektes.
func (tch *Touch) newPenEvent(typ PenEventType, rawPos TouchRawPos) (ev PenEvent) {
	ev.Type = typ
	ev.TouchRawPos = rawPos
	ev.TouchPos, _ = tch.plane.Transform(rawPos)
	ev.Time = time.Now()
	//ev.FifoSize = tch.tspi.ReadReg8(hw.FIFO_SIZE)
	return
}

func (t *Touch) readRawPos() (td TouchRawPos) {
	//td.RawX, td.RawY = tch.tspi.ReadData()
	td.RawX, td.RawY, td.RawZ = t.ReadData()
	return
}

func (t *Touch) BufferLen() uint8 {
	return t.tspi.ReadReg8(hw.FIFO_SIZE)
}

func (t *Touch) ReadData() (x, y uint16, z uint8) {
	cnt := t.BufferLen()
	for cnt > 0 {
		x, y, z = t.tspi.ReadData()
		cnt--
	}
	t.tspi.WriteReg8(hw.FIFO_STA, hw.FIFO_STA_RESET)
	t.tspi.WriteReg8(hw.FIFO_STA, 0)
	return
}

// In diesen globalen Variablen werden Daten verwaltet, die vom
// Callback-Handler (siehe unten) benötigt werden.
var (
	posRaw TouchRawPos
	penUp  bool = true
)

/*
func eventDispatcher(arg any) {
	var ev PenEvent

	t := arg.(*Touch)

	if t.BufferLen() == 0 {
		for i := 0; i < 10; i++ {
			time.Sleep(5 * time.Millisecond)
			if t.BufferLen() > 0 {
				posRaw = t.readRawPos()
				ev = t.newPenEvent(PenPress, posRaw)
				t.enqueueEvent(ev)
				break
			}
		}
		t.tspi.WriteReg8(hw.INT_STA, 0xFF)
		return
	}

	posRaw = t.readRawPos()
	ev = t.newPenEvent(PenRelease, posRaw)
	t.enqueueEvent(ev)
	t.tspi.WriteReg8(hw.INT_STA, 0xFF)
}
*/

// Diese Funktion ist der Callback-Handler, welcher beim Eintreten eines
// Interrupts vom Touchscreen aufgerufen wird. Effizienz ist der Schlüssel
// dieser Funktion, aber auch das korrekte Handling der darunterliegenden
// Hardware, sprich Verwalten des Interrupt-Systems.
func eventDispatcher(arg any) {
	var t *Touch
	var evTyp PenEventType
	var ev PenEvent

	t = arg.(*Touch)

	intEnable := t.tspi.ReadReg8(hw.INT_EN)
	// log.Printf("ISR called (INT_EN: %08b)\n", intEnable)
	for {
		time.Sleep(sampleTime) // NEU!!! ACHTUNG!!!
		intStatus := t.tspi.ReadReg8(hw.INT_STA)
		// log.Printf("  INT_STA: %08b\n", intStatus)
		if (intStatus & intEnable) == 0 {
			break
		}

		if (intStatus & hw.INT_TOUCH_DET) != 0 {
			// log.Printf("    INT_TOUCH_DET\n")
			if (t.tspi.ReadReg8(hw.TSC_CTRL) & 0x80) == 0 {
				if !penUp {
					// log.Printf("      Pen up\n")
					ev = t.newPenEvent(PenRelease, posRaw)
					t.enqueueEvent(ev)
					penUp = true
				}
			} else {
				// log.Printf("      Pen down\n")
			}
			t.tspi.WriteReg8(hw.INT_STA, hw.INT_TOUCH_DET)
		}

		if (intStatus & hw.INT_FIFO_TH) != 0 {
			// log.Printf("    INT_FIFO_TH\n")
			for t.tspi.ReadReg8(hw.FIFO_SIZE) > 0 {
				// log.Printf("      FIFO_SIZE > 0\n")
				posRaw = t.readRawPos()
				evTyp = PenDrag
				if penUp {
					evTyp = PenPress
					penUp = false
				}
				ev = t.newPenEvent(evTyp, posRaw)
				t.enqueueEvent(ev)
			}
			t.tspi.WriteReg8(hw.INT_STA, hw.INT_FIFO_TH)
		}
	}
}
