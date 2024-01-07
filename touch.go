package adatft

import (
    "fmt"
    "log"
    "time"
    stmpe "github.com/stefan-muehlebach/adatft/stmpe610"
)

const (
    tchSpeedHz     = 1_000_000
    eventQueueSize = 30
)

// Dies sind alle Pen- oder Touch-Events, auf welche man sich abonnieren kann.
// Vom SMTPE610 gibt es nur 3 Events: Press, Move und Release. Die anderen
// Events (wie Tap, DoubleTap, Enter oder Leave) sind virtuelle Events und
// werden im Package 'adagui' durch den Screen-Typ erzeugt.
//
type PenEventType uint8

const (
    PenPress PenEventType = iota
    PenDrag
    PenRelease
    numEvents
    sampleTime = 5 * time.Millisecond
)

func (pet PenEventType) String() (string) {
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

// Dieser Typ steht fuer das SPI Interface zum STMPE - dem Touchscreen.
//
type Touch struct {
    tspi TouchInterface
    EventQ PenEventChannelType
    DistortedPlane
    isOpen bool
}

// Dieser Typ enthält die rohen, unkalibrierten Display-Daten
//
type TouchRawPos struct {
    RawX, RawY uint16
    //RawZ uint8
}

func (td TouchRawPos) String() (string) {
    return fmt.Sprintf("(%4d, %4d)", td.RawX, td.RawY)
    //return fmt.Sprintf("(%4d, %4d, %3d)", td.RawX, td.RawY, td.RawZ)
}

// Während in diesem Typ die kalibrierten Postitionsdaten abgelegt werden.
//
type TouchPos struct {
    X, Y float64
    //Z uint8
}

func (tp TouchPos) String() (string) {
    return fmt.Sprintf("(%5.1f, %5.1f)", tp.X, tp.Y)
    //return fmt.Sprintf("(%5.1f, %5.1f, %08b)", tp.X, tp.Y, tp.Z)
}

// Jedes Ereignis des Touchscreens wird durch eine Variable des Typs
// 'Event' repraesentiert.
//
type PenEvent struct {
    Type       PenEventType
    TouchRawPos
    TouchPos
    Time       time.Time
    FifoSize   uint8
}

func (p1 TouchPos) Near(p2 TouchPos) (bool) {
    var dx, dy float64
    dx = p1.X-p2.X
    dy = p1.Y-p2.Y
    return dx*dx+dy*dy <= 50
}

// Dies ist der Funktionstyp für den PenEvent-Handler - also jene Funktion,
// welche beim Eintreffen eines Interrupts vom STMPE610 aufgerufen werden
// soll.
//
type PenEventHandlerType func(event PenEvent)

type PenEventChannelType chan PenEvent

//-----------------------------------------------------------------------------
//
// Funktionen
//
func OpenTouch() (*Touch) {
    var tch *Touch
    var devId uint16
    var revNr uint8

    tch = &Touch{}
    if isRaspberry {
        tch.tspi = stmpe.Open(tchSpeedHz)
    } else {
        tch.tspi = stmpe.OpenDummy(tchSpeedHz)
    }

    revNr = tch.tspi.ReadReg8(stmpe.STMPE610_ID_VER)
    devId = tch.tspi.ReadReg16(stmpe.STMPE610_CHIP_ID)
    if (devId != 0x0811) || (revNr != 0x03) {
        log.Fatalf("device id and/or revision numbers are not as expected: got (0x%04x, 0x%02x) should be (0x0811, 0x03)\n", devId, revNr)
    }

    // Initialisiere die Queue für applikatorische Events und setze den
    // Interrupt-Handler für Touch-Events.
    tch.EventQ = make(chan PenEvent, eventQueueSize)
    tch.tspi.SetCallback(eventDispatcher, tch)

    tch.tspi.Init(nil)
    tch.isOpen = true

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
//
func (tch *Touch) enqueueEvent(event PenEvent) {
    //defer func() {
    //    if x := recover(); x != nil {
    //        log.Printf("runtime panic: %v", x)
    //    }
    //}()
    select {
        case tch.EventQ <- event:
        default:
            log.Printf("Sending not possible: event queue full!\n")
    }
}

// Diese Funktion wird von 'aussen' aufgerufen und gibt das nächste Pen-Event
// zurück. Es ist eine Alternative zum Lesen aus der öffentlichen Event-Queue.
func (tch *Touch) WaitForEvent() (PenEvent) {
    return <-tch.EventQ
}

// Diese Hilfsfunktion dient der einfacheren Erstellung eines Event-Objektes.
func (tch *Touch) newPenEvent(typ PenEventType, rawPos TouchRawPos) (ev PenEvent) {
    ev.Type = typ
    ev.TouchRawPos = rawPos
    ev.TouchPos, _ = tch.Transform(rawPos)
    ev.Time = time.Now()
    ev.FifoSize = tch.tspi.ReadReg8(stmpe.STMPE610_FIFO_SIZE)
    return
}

func (tch *Touch) readRawPos() (td TouchRawPos) {
    td.RawX, td.RawY = tch.tspi.ReadData()
    //td.RawX, td.RawY, td.RawZ = tch.tspi.ReadData()
    return
}

// In diesen globalen Variablen werden Daten verwaltet, die vom
// Callback-Handler (siehe unten) benötigt werden.
var (
    posRaw TouchRawPos
    penUp  bool = true
)

// Diese Funktion ist der Callback-Handler, welcher beim Eintreten eines
// Interrupts vom Touchscreen aufgerufen wird. Effizienz ist der Schlüssel
// dieser Funktion, aber auch das korrekte Handling der darunterliegenden
// Hardware, sprich Verwalten des Interrupt-Systems.
func eventDispatcher(arg any) {
    var tch *Touch
    var evTyp PenEventType
    var ev PenEvent

    tch = arg.(*Touch)

    // intStatus enthält pro Interrupt den aktuellen Status (active,
    // not active) während in intEnable pro Interrupt festgehalten ist, ob
    // dieser Interrupt überhaupt eingeschaltet ist.
    intStatus := tch.tspi.ReadReg8(stmpe.STMPE610_INT_STA)
    intEnable := tch.tspi.ReadReg8(stmpe.STMPE610_INT_EN)

    if (intStatus & (stmpe.STMPE610_INT_TOUCH_DET |
            stmpe.STMPE610_INT_FIFO_TH)) == 0 {
        return
    }

    // Schalte alle (!) Interrupts aus.
    tch.tspi.WriteReg8(stmpe.STMPE610_INT_EN, 0x00)

    switch {
    case (intStatus & stmpe.STMPE610_INT_TOUCH_DET) != 0:
         if (tch.tspi.ReadReg8(stmpe.STMPE610_TSC_CTRL) & 0x80) == 0 {
             if !penUp {
                 ev = tch.newPenEvent(PenRelease, posRaw)
                 tch.enqueueEvent(ev)
                 penUp = true
             }
         }
         tch.tspi.WriteReg8(stmpe.STMPE610_INT_STA, stmpe.STMPE610_INT_TOUCH_DET)

    case (intStatus & stmpe.STMPE610_INT_FIFO_TH) != 0:
         for tch.tspi.ReadReg8(stmpe.STMPE610_FIFO_SIZE) > 0 {
             time.Sleep(sampleTime)      // NEU!!! ACHTUNG!!!
             posRaw = tch.readRawPos()
             evTyp = PenDrag
             if penUp {
                 evTyp = PenPress
                 penUp = false
             }
             ev = tch.newPenEvent(evTyp, posRaw)
             tch.enqueueEvent(ev)
         }
         tch.tspi.WriteReg8(stmpe.STMPE610_INT_STA, stmpe.STMPE610_INT_FIFO_TH)
         tch.tspi.WriteReg8(stmpe.STMPE610_FIFO_STA, 0x01)
         tch.tspi.WriteReg8(stmpe.STMPE610_FIFO_STA, 0x00)
    }

    // Schalte die Interrupts wieder ein.
    tch.tspi.WriteReg8(stmpe.STMPE610_INT_EN, intEnable)
}

