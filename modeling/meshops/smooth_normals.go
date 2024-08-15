package meshops

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type SmoothNormalsTransformer struct{}

func (smt SmoothNormalsTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	if err = RequireTopology(m, modeling.TriangleTopology); err != nil {
		return
	}

	if err = RequireV3Attribute(m, modeling.PositionAttribute); err != nil {
		return
	}

	return SmoothNormals(m), nil
}

func SmoothNormals(m modeling.Mesh) modeling.Mesh {
	check(RequireTopology(m, modeling.TriangleTopology))
	check(RequireV3Attribute(m, modeling.PositionAttribute))

	vertices := m.Float3Attribute(modeling.PositionAttribute)
	normals := make([]vector3.Float64, vertices.Len())
	for i := range normals {
		normals[i] = vector3.Zero[float64]()
	}

	tris := m.Indices()
	for triIndex := 0; triIndex < tris.Len(); triIndex += 3 {
		p1 := tris.At(triIndex)
		p2 := tris.At(triIndex + 1)
		p3 := tris.At(triIndex + 2)
		// normalize(cross(B-A, C-A))
		normalized := vertices.At(p2).Sub(vertices.At(p1)).Cross(vertices.At(p3).Sub(vertices.At(p1)))

		// This occurs whenever the given tri is actually just a line
		if math.IsNaN(normalized.X()) {
			continue
		}

		normals[p1] = normals[p1].Add(normalized)
		normals[p2] = normals[p2].Add(normalized)
		normals[p3] = normals[p3].Add(normalized)
	}

	zero := vector3.Zero[float64]()
	for i, n := range normals {
		if n == zero {
			continue
		}
		normals[i] = n.Normalized()
	}

	return m.SetFloat3Attribute(modeling.NormalAttribute, normals)
}

type SmoothNormalsImplicitWeldTransformer struct {
	Distance float64
}

func (smt SmoothNormalsImplicitWeldTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	if err = RequireTopology(m, modeling.TriangleTopology); err != nil {
		return
	}

	if err = RequireV3Attribute(m, modeling.PositionAttribute); err != nil {
		return
	}

	return SmoothNormalsImplicitWeld(m, smt.Distance), nil
}

func SmoothNormalsImplicitWeld(m modeling.Mesh, distance float64) modeling.Mesh {
	if distance < 0 {
		panic(fmt.Errorf("weld distance can not be negative, recieved: %f", distance))
	}

	check(RequireTopology(m, modeling.TriangleTopology))
	check(RequireV3Attribute(m, modeling.PositionAttribute))

	tree := m.ToPointCloud().OctTree()

	vertices := m.Float3Attribute(modeling.PositionAttribute)
	normals := make([]vector3.Float64, vertices.Len())
	for i := range normals {
		normals[i] = vector3.Zero[float64]()
	}

	tris := m.Indices()
	for triIndex := 0; triIndex < tris.Len(); triIndex += 3 {
		p1 := tris.At(triIndex)
		p2 := tris.At(triIndex + 1)
		p3 := tris.At(triIndex + 2)

		// normalize(cross(B-A, C-A))
		normalized := vertices.At(p2).Sub(vertices.At(p1)).Cross(vertices.At(p3).Sub(vertices.At(p1)))

		// This occurs whenever the given tri is actually just a line
		if math.IsNaN(normalized.X()) {
			continue
		}

		welds := tree.ElementsWithinRange(vertices.At(p1), distance)
		for _, weld := range welds {
			normals[weld] = normals[weld].Add(normalized)
		}

		welds = tree.ElementsWithinRange(vertices.At(p2), distance)
		for _, weld := range welds {
			normals[weld] = normals[weld].Add(normalized)
		}

		welds = tree.ElementsWithinRange(vertices.At(p3), distance)
		for _, weld := range welds {
			normals[weld] = normals[weld].Add(normalized)
		}
	}

	zero := vector3.Zero[float64]()
	for i, n := range normals {
		if n == zero {
			continue
		}
		normals[i] = n.Normalized()
	}

	return m.SetFloat3Attribute(modeling.NormalAttribute, normals)
}

type SmoothNormalsNode = nodes.StructNode[modeling.Mesh, SmoothNormalsNodeData]

type SmoothNormalsNodeData struct {
	Mesh nodes.NodeOutput[modeling.Mesh]
}

func (snn SmoothNormalsNodeData) Process() (modeling.Mesh, error) {
	if snn.Mesh == nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}
	return SmoothNormals(snn.Mesh.Value()), nil
}

type SmoothNormalsImplicitWeldNode = nodes.StructNode[modeling.Mesh, SmoothNormalsImplicitWeldNodeData]

type SmoothNormalsImplicitWeldNodeData struct {
	Mesh     nodes.NodeOutput[modeling.Mesh]
	Distance nodes.NodeOutput[float64]
}

func (snn SmoothNormalsImplicitWeldNodeData) Process() (modeling.Mesh, error) {
	if snn.Mesh == nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}
	distance := 0.
	if snn.Distance != nil {
		distance = snn.Distance.Value()
	}
	return SmoothNormalsImplicitWeld(snn.Mesh.Value(), distance), nil
}
