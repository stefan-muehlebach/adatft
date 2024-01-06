package adatft

type RefPointType uint8

const (
    RefTopLeft RefPointType = iota
    RefTopRight
    RefBottomRight
    RefBottomLeft
    NumRefPoints
)

func (pt RefPointType) String() (string) {
    switch pt {
    case RefTopLeft:
        return "TopLeft"
    case RefTopRight:
        return "TopRight"
    case RefBottomRight:
        return "BottomRight"
    case RefBottomLeft:
        return "BottomLeft"
    }
    return "(unknow reference point)"
}

