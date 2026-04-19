package adatft

import (
	"testing"
)

var (
	touch     *Touch
	v         float64
	distPlane *DistortedPlane
)

func TestTouchScreen(t *testing.T) {
	touch = OpenTouch(Rotate000)
	i := 0
	for event := range touch.EventQ {
		t.Logf("Event received: %v\n", event)
		i += 1
		if i >= 100 {
			break
		}
	}
}

func TestMap(t *testing.T) {
	distPlane = &DistortedPlane{}
	distPlane.ReadConfig(Rotate000)
	rawPos := TouchRawPos{RawX: 500, RawY: 500}
	pos, _ := distPlane.Transform(rawPos)
	t.Logf("got (%f, %f)\n", pos.X, pos.Y)
	//v = Map(0.0, 0.0, 1.0, 1.0, 10.0)
	//t.Logf("got %f, expected %f\n", v, 1.0)
	//v = Map(0.5, 0.0, 1.0, 1.0, 10.0)
	//t.Logf("got %f, expected %f\n", v, 5.5)
	//v = Map(1.0, 0.0, 1.0, 1.0, 10.0)
	//t.Logf("got %f, expected %f\n", v, 10.0)
}
