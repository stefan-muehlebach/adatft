package adatft

import (
    "fmt"
    "log"
    "time"
    . "github.com/stefan-muehlebach/adatft/stmpe610"
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
    tspi *STMPE610
    EventQ PenEventChannelType
    DistortedPlane
    isOpen bool
}

// Dieser Typ enthält die rohen, unkalibrierten Display-Daten
//
type TouchData struct {
    RawX, RawY uint16
    //RawZ uint8
}

func (td TouchData) String() (string) {
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
    TouchData
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
    tch.tspi = OpenSTMPE610(tchSpeedHz)

    revNr = tch.tspi.ReadReg8(STMPE610_ID_VER)
    devId = tch.tspi.ReadReg16(STMPE610_CHIP_ID)
    if (devId != 0x0811) || (revNr != 0x03) {
        log.Fatalf("device id and/or revision numbers are not as expected: got (0x%04x, 0x%02x) should be (0x0811, 0x03)\n", devId, revNr)
    }
    tch.InitTouch()
    tch.isOpen = true

    return tch
}

func (tch *Touch) Close() {
    tch.isOpen = false
    close(tch.EventQ)
    tch.tspi.Close()
}

func (tch *Touch) InitTouch() {
    tch.EventQ = make(chan PenEvent, eventQueueSize)

    // Set the one and only function which will be called on an interrupt
    // from the STMPE610. We can only define one single function.
    tch.tspi.SetCallback(eventDispatcher, tch)

    // Konfiguration des Touchscreens. Diese Einstellungen wurden (wie auch
    // für das Display) aus vielen Code-Vorlagen und Dokumentationen aus
    // dem Internet zusammenorchestriert - geschmückt mit vielen Stunden
    // 'try and error'. Verbesserungen und Vorschläge sind jederzeit herzlich
    // willkommen.
    //
    // System Register (STMPE610_SYS_XXX)
    //
    tch.tspi.WriteReg8(STMPE610_SYS_CTRL1,
            STMPE610_SYS_CTRL1_RESET)
    time.Sleep(10 * time.Millisecond)
    for i := uint8(0); i < 65; i++ {
        tch.tspi.ReadReg8(i)
    }
    tch.tspi.WriteReg8(STMPE610_SYS_CTRL2,
            0x00)

    // Touchscreen Register (STMPE610_TSC_XXX)
    //
    // Touchscreen Controller Control
    // - set the window tracking feature to 8 pixels
    // - acquire X, Y data
    // - enable touch screen control
    //
    tch.tspi.WriteReg8(STMPE610_TSC_CTRL,
            STMPE610_TSC_CTRL_WTRK8 |
            STMPE610_TSC_CTRL_XY |
            STMPE610_TSC_CTRL_EN)

    // Analog Digital Converter Register (STMPE610_ADC_XXX)
    // (Wozu braucht es diese?)
    //
    tch.tspi.WriteReg8(STMPE610_ADC_CTRL1,
            STMPE610_ADC_CTRL1_12BIT |
            STMPE610_ADC_CTRL1_96CLK)		// Ada
            //STMPE610_ADC_CTRL1_36CLK)

    tch.tspi.WriteReg8(STMPE610_ADC_CTRL2,
            STMPE610_ADC_CTRL2_6_5MHZ)

    tch.tspi.WriteReg8(STMPE610_ADC_CAPT,
            STMPE610_ADC_CAPT_ALL)

    // Touchscreen Controller Configuration
    // - average 8 samples
    // - set a touch detect delay of 1ms
    // - set a settling time of 5ms
    //
    tch.tspi.WriteReg8(STMPE610_TSC_CFG,
            STMPE610_TSC_CFG_8SAMPLE |
            STMPE610_TSC_CFG_DELAY_1MS |
            STMPE610_TSC_CFG_SETTLE_5MS)
            // STMPE610_TSC_CFG_8SAMPLE |
            // STMPE610_TSC_CFG_DELAY_500US |
            // STMPE610_TSC_CFG_SETTLE_500US)

    // Don't collect any Z data since we cannot relay on this feature!
    //tch.tspi.WriteReg8(STMPE610_TSC_FRACTION_Z,
    //        STMPE610_TSC_FRACT_Z_2_6)

    // FIFO Register (STMPE610_FIFO_XXX)
    //
    tch.tspi.WriteReg8(STMPE610_FIFO_TH, 1)
    tch.tspi.WriteReg8(STMPE610_FIFO_STA,
            STMPE610_FIFO_STA_RESET)
    tch.tspi.WriteReg8(STMPE610_FIFO_STA, 0x00)

    // Interrupt Register (STMPE610_INT_XXX)
    //
    // Wir abonnieren uns auf zwei Events: das Drücken, respl. Loslassen
    // des Bildschirms (beide Ereignisse generieren das gleiche Event) sowie
    // das Erreichen eines bestimmten Schwellwertes bei der FIFO-Queue
    tch.tspi.WriteReg8(STMPE610_INT_EN,
            STMPE610_INT_TOUCH_DET |
            STMPE610_INT_FIFO_TH)

    tch.tspi.WriteReg8(STMPE610_TSC_I_DRIVE,
            STMPE610_TSC_I_DRIVE_50MA)

    //tch.tspi.WriteReg8(STMPE610_TSC_SHIELD,
    //        STMPE610_TSC_GROUND_X_P |
    //        STMPE610_TSC_GROUND_X_N |
    //        STMPE610_TSC_GROUND_Y_P |
    //        STMPE610_TSC_GROUND_Y_N)

    // Reset all interupts to begin with
    tch.tspi.WriteReg8(STMPE610_INT_STA, 0xFF)

    // Mit diesem Register schliesslich, wird das Interrupt-System aktiviert.
    tch.tspi.WriteReg8(STMPE610_INT_CTRL,
            STMPE610_INT_CTRL_EDGE |
            STMPE610_INT_CTRL_ENABLE)
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

// Dies ist eine Hilfsfunktion, mit der einfach ein PenEvent erzeugt werden
// kann.
func (tch *Touch) newPenEvent(typ PenEventType, pos TouchData) (ev PenEvent) {
    ev.Type = typ
    ev.TouchPos, _ = tch.Transform(pos)
    //ev.X, ev.Y, _ = tch.Transform(pos.RawX, pos.RawY)
    //ev.Z = pos.RawZ
    ev.TouchData = pos
    ev.Time = time.Now()
    ev.FifoSize = tch.tspi.ReadReg8(STMPE610_FIFO_SIZE)
    return
}

func (tch *Touch) readPosition() (td TouchData) {
    td.RawX, td.RawY = tch.tspi.ReadData()
    //td.RawX, td.RawY, td.RawZ = tch.tspi.ReadData()
    return
}

// In diesen globalen Variablen werden Daten verwaltet, die vom
// Callback-Handler (siehe unten) benötigt werden.
var (
    pos   TouchData
    penUp bool = true
)

// Diese Funktion ist der Callback-Handler, welcher beim Eintreten eines
// Interrupts vom Touchscreen aufgerufen wird. Effizienz ist der Schlüssel
// dieser Funktion, aber auch das korrekte Handling der darunterliegenden
// Hardware, sprich Verwalten des Interrupt-Systems.
func eventDispatcher(arg any) {
    var tch *Touch
    var ev PenEvent
    var evTyp PenEventType

    tch = arg.(*Touch)
    intStatus := tch.tspi.ReadReg8(STMPE610_INT_STA)
    intEnable := tch.tspi.ReadReg8(STMPE610_INT_EN)

    if (intStatus & 0x03) == 0 {
        return
    }

    tch.tspi.WriteReg8(STMPE610_INT_EN, 0x00)

    switch {
    case (intStatus & STMPE610_INT_TOUCH_DET) != 0:
         if (tch.tspi.ReadReg8(STMPE610_TSC_CTRL) & 0x80) == 0 {
             if !penUp {
                 ev = tch.newPenEvent(PenRelease, pos)
                 tch.enqueueEvent(ev)
                 penUp = true
             }
         }
         tch.tspi.WriteReg8(STMPE610_INT_STA, STMPE610_INT_TOUCH_DET)

    case (intStatus & STMPE610_INT_FIFO_TH) != 0:
         for tch.tspi.ReadReg8(STMPE610_FIFO_SIZE) > 0 {
             time.Sleep(sampleTime)      // NEU!!! ACHTUNG!!!
             pos = tch.readPosition()
             evTyp = PenDrag
             if penUp {
                 evTyp = PenPress
                 penUp = false
             }
             ev = tch.newPenEvent(evTyp, pos)
             tch.enqueueEvent(ev)
         }
         tch.tspi.WriteReg8(STMPE610_INT_STA, STMPE610_INT_FIFO_TH)
         tch.tspi.WriteReg8(STMPE610_FIFO_STA, 0x01)
         tch.tspi.WriteReg8(STMPE610_FIFO_STA, 0x00)
    }

    tch.tspi.WriteReg8(STMPE610_INT_EN, intEnable)
}

