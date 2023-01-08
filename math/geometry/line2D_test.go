package geometry_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector"
)

func TestIntersect(t *testing.T) {
	l1 := geometry.NewLine2D(vector.NewVector2(0, 0), vector.NewVector2(1, 1))
	l2 := geometry.NewLine2D(vector.NewVector2(0, 1), vector.NewVector2(1, 0))
	if l1.Intersects(l2) == false {
		t.Error("Lines should have interesected")
	}
}

func TestDoesntIntersect(t *testing.T) {
	l1 := geometry.NewLine2D(vector.NewVector2(0, 0), vector.NewVector2(0, .4))
	l2 := geometry.NewLine2D(vector.NewVector2(0, 1), vector.NewVector2(1, 0))
	if l1.Intersects(l2) == true {
		t.Error("Lines should have not interesected")
	}
}
