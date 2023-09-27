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
        return "Top Left"
    case RefTopRight:
        return "Top Right"
    case RefBottomRight:
        return "Bottom Right"
    case RefBottomLeft:
        return "Bottom Left"
    }
    return "(unknow reference point)"
}

