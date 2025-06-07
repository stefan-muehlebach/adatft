package adatft

import (
	"testing"
)

var (
	touch *Touch
)

func TestTouchScreen(t *testing.T) {
	touch = OpenTouch(Rotate000)
	for event := range touch.EventQ {
		t.Logf("Event received: %v\n", event)
	}
}
