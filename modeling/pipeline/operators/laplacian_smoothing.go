package operators

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/pipeline"
	"github.com/EliCDavis/vector"
)

func SmoothLaplacian(iterations int, smoothingFactor float64) pipeline.Command {
	return pipeline.NewCommand(
		pipeline.PermissionForResources(
			pipeline.RequireMeshPrimitive(),
			pipeline.RequireMeshFloat3Attribute(modeling.PositionAttribute),
		),
		pipeline.PermissionForResources(
			pipeline.RequireMeshFloat3Attribute(modeling.PositionAttribute),
		),
		func(v *pipeline.View) {
			lut := v.VertexNeighborTable()

			vertices := make([]vector.Vector3, v.AttributeLength())
			v.ScanFloat3AttributeParallel(modeling.PositionAttribute, func(i int, v vector.Vector3) {
				vertices[i] = v
			})

			for i := 0; i < iterations; i++ {
				for vi, vertex := range vertices {
					vs := vector.Vector3Zero()

					for vn := range lut.Lookup(vi) {
						vs = vs.Add(vertices[vn])
					}

					vertices[vi] = vertex.Add(
						vs.
							DivByConstant(float64(lut.Count(vi))).
							Sub(vertex).
							MultByConstant(smoothingFactor))
				}
			}

			v.SetFloat3Attribute(modeling.PositionAttribute, vertices)
		},
	)
}
