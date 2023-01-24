package trees_test

import (
	"math/rand"
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/modeling/trees"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestOctreeSingleTri(t *testing.T) {
	// ARRANGE ================================================================
	mesh := modeling.NewMesh(
		[]int{0, 1, 2},
		map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 1., 0.),
			},
		},
		nil,
		nil,
		nil,
	)
	tree := trees.FromMesh(mesh)

	// ACT ====================================================================
	_, p := tree.ClosestPoint(vector3.Zero[float64]())

	// ASSERT =================================================================
	assert.Equal(t, p, vector3.Zero[float64]())
}

func TestOctreeTwoTris(t *testing.T) {
	// ARRANGE ================================================================
	mesh := modeling.NewMesh(
		[]int{0, 1, 2, 0, 2, 3},
		map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 1., 0.),
				vector3.New(1., 0., 0.),
			},
		},
		nil,
		nil,
		nil,
	)
	tree := trees.FromMesh(mesh)

	// ACT ====================================================================
	_, p := tree.ClosestPoint(vector3.Zero[float64]())

	// ASSERT =================================================================
	assert.Equal(t, p, vector3.Zero[float64]())
}

func TestOctreeSphere(t *testing.T) {
	// ARRANGE ================================================================
	mesh := primitives.UVSphere(1, 100, 100)
	tree := trees.FromMesh(mesh)

	testPointCount := 1000
	testPoints := make([]vector3.Float64, testPointCount)
	for i := 0; i < testPointCount; i++ {
		testPoints[i] = vector3.New(
			-1+(rand.Float64()*2),
			-1+(rand.Float64()*2),
			-1+(rand.Float64()*2),
		).Normalized()
	}

	// ACT / ASSERT ===========================================================
	for i := 0; i < testPointCount; i++ {
		_, p := tree.ClosestPoint(testPoints[i])
		assert.InDelta(t, testPoints[i].X(), p.X(), 0.05)
		assert.InDelta(t, testPoints[i].Y(), p.Y(), 0.05)
		assert.InDelta(t, testPoints[i].Z(), p.Z(), 0.05)

	}
}

func TestOctreeLineSphere(t *testing.T) {
	// ARRANGE ================================================================
	mesh := modeling.NewLineStripMesh(
		map[string][]vector3.Float64{
			modeling.PositionAttribute: repeat.CirclePoints(100, 1),
		},
		nil,
		nil,
		nil,
	)
	tree := trees.FromMesh(mesh)

	testPointCount := 100
	testPoints := make([]vector3.Float64, testPointCount)
	for i := 0; i < testPointCount; i++ {
		testPoints[i] = vector3.New(
			-1+(rand.Float64()*2),
			0,
			-1+(rand.Float64()*2),
		).Normalized()
	}

	// ACT / ASSERT ===========================================================
	for i := 0; i < testPointCount; i++ {
		_, p := tree.ClosestPoint(testPoints[i].MultByConstant(5))
		assert.InDelta(t, testPoints[i].X(), p.X(), 0.05)
		assert.InDelta(t, testPoints[i].Y(), p.Y(), 0.05)
		assert.InDelta(t, testPoints[i].Z(), p.Z(), 0.05)
	}
}

var result vector3.Float64

func BenchmarkOctreeLineSphere(b *testing.B) {
	var r vector3.Float64

	mesh := modeling.NewLineStripMesh(
		map[string][]vector3.Float64{
			modeling.PositionAttribute: repeat.CirclePoints(10000, 1),
		},
		nil,
		nil,
		nil,
	)
	tree := trees.FromMesh(mesh)

	for n := 0; n < b.N; n++ {
		// always record the result of Fib to prevent
		// the compiler eliminating the function call.
		_, r = tree.ClosestPoint(vector3.New(1., 0., 1.))
	}
	// always store the result to a package level variable
	// so the compiler cannot eliminate the Benchmark itself.
	result = r
}
