package gc9503cv

import (
	"errors"
)

//-----------------------------------------------------------------------------

type RotationType byte

const (
	Rot000 RotationType = iota
	Rot090
	Rot180
	Rot270
)

func (r RotationType) String() string {
	switch r {
	case Rot000:
		return "Rot000"
	case Rot090:
		return "Rot090"
	case Rot180:
		return "Rot180"
	case Rot270:
		return "Rot270"
	default:
		return ""
	}
}

func (r *RotationType) Set(val string) error {
	switch val {
	case "Rot000", "rot000", "000", "0":
		*r = Rot000
	case "Rot090", "rot090", "090", "90":
		*r = Rot090
	case "Rot180", "rot180", "180":
		*r = Rot180
	case "Rot270", "rot270", "270":
		*r = Rot270
	default:
		return errors.New("Wrong rotation type")
	}
	return nil
}
