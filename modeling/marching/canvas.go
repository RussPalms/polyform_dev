package marching

import (
	"fmt"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

func interpolateVerts(v1, v2 vector.Vector3, v1v, v2v, isoLevel float64) vector.Vector3 {
	t := (isoLevel - v1v) / (v2v - v1v)

	return v2.Sub(v1).MultByConstant(t).Add(v1)
}

func LookupOrAdd(data *workingData, vert vector.Vector3) int {
	distritized := modeling.Vector3ToInt(vert, 4)

	if foundIndex, ok := data.vertLookup[distritized.String()]; ok {
		return foundIndex
	}

	index := len(data.verts)
	data.vertLookup[distritized.String()] = index
	data.verts = append(data.verts, vert)
	data.uvs = append(data.uvs, vert.XZ())
	return index
}

type MarchingCanvas struct {
	data         []float64
	width        int
	height       int
	depth        int
	cubesPerUnit float64
}

func NewMarchingCanvas(width, height, depth int, cubesPerUnit float64) *MarchingCanvas {

	if width < 0 {
		panic(fmt.Errorf("invalid marching cube width: %d", width))
	}

	if height < 0 {
		panic(fmt.Errorf("invalid marching cube height: %d", height))
	}

	if depth < 0 {
		panic(fmt.Errorf("invalid marching cube depth: %d", depth))
	}

	return &MarchingCanvas{
		data:         make([]float64, width*height*depth),
		width:        width,
		height:       height,
		depth:        depth,
		cubesPerUnit: cubesPerUnit,
	}
}

func (d MarchingCanvas) index(x, y, z int) int {
	return (z * d.width * d.height) + (y * d.width) + x
}

func (d MarchingCanvas) SetValue(x, y, z int, val float64) {
	d.data[d.index(x, y, z)] = val
}

func (d MarchingCanvas) AddValue(x, y, z int, val float64) {
	d.data[d.index(x, y, z)] += val
}

func (d MarchingCanvas) AddField(field sample.Vec3ToFloat) {

	for x := 0; x < d.width; x++ {
		for y := 0; y < d.height; y++ {
			for z := 0; z < d.depth; z++ {
				pos := vector.NewVector3(float64(x), float64(y), float64(z)).
					MultByConstant(d.cubesPerUnit)
				d.AddValue(x, y, z, field(pos))
			}
		}
	}

}

func (d MarchingCanvas) Sphere(pos vector.Vector3, radius float64) {

	xResolution := 100
	yResolution := 100
	zResolution := 100

	for x := -xResolution / 2; x < xResolution/2; x++ {
		for y := -yResolution / 2; y < yResolution/2; y++ {
			for z := -zResolution / 2; z < zResolution/2; z++ {
				if x == 0 && y == 0 && z == 0 {
					continue
				}

				cubePosition := vector.
					NewVector3(float64(x), float64(y), float64(z)).
					Normalized().
					MultByConstant(radius).
					Add(pos).
					DivByConstant(d.cubesPerUnit) // Convert to cube space

				cubeFloor := cubePosition.Floor()
				cubeDir := cubePosition.Sub(cubeFloor)
				len := cubeDir.Length()
				len += .1

				d.AddValue(int(cubeFloor.X()), int(cubeFloor.Y()), int(cubeFloor.Z()), len)
			}
		}
	}
}

type workingData struct {
	verts      []vector.Vector3
	uvs        []vector.Vector2
	vertLookup map[string]int
}

func (d MarchingCanvas) March(cutoff float64) modeling.Mesh {
	tris := make([]int, 0)
	marchingWorkingData := &workingData{
		verts:      make([]vector.Vector3, 0),
		uvs:        make([]vector.Vector2, 0),
		vertLookup: make(map[string]int),
	}

	// Vector3[] jitter = new Vector3[] {
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	//     (new Vector3(Random.value, Random.value, Random.value) / 10f) -(Vector3.one * 0.1f),
	// };

	for x := 0; x < d.width-1; x++ {
		for y := 0; y < d.height-1; y++ {
			for z := 0; z < d.depth-1; z++ {
				cubeCorners := []float64{
					d.data[d.index(x, y, z)],
					d.data[d.index(x+1, y, z)],
					d.data[d.index(x+1, y, z+1)],
					d.data[d.index(x, y, z+1)],
					d.data[d.index(x, y+1, z)],
					d.data[d.index(x+1, y+1, z)],
					d.data[d.index(x+1, y+1, z+1)],
					d.data[d.index(x, y+1, z+1)],
				}

				cubeCornersExistence := []bool{
					cubeCorners[0] < cutoff,
					cubeCorners[1] < cutoff,
					cubeCorners[2] < cutoff,
					cubeCorners[3] < cutoff,
					cubeCorners[4] < cutoff,
					cubeCorners[5] < cutoff,
					cubeCorners[6] < cutoff,
					cubeCorners[7] < cutoff,
				}

				xf := float64(x)
				yf := float64(y)
				zf := float64(z)

				cubeCornerPositions := []vector.Vector3{
					vector.NewVector3(xf, yf, zf),
					vector.NewVector3(xf+1, yf, zf),
					vector.NewVector3(xf+1, yf, zf+1),
					vector.NewVector3(xf, yf, zf+1),
					vector.NewVector3(xf, yf+1, zf),
					vector.NewVector3(xf+1, yf+1, zf),
					vector.NewVector3(xf+1, yf+1, zf+1),
					vector.NewVector3(xf, yf+1, zf+1),
				}

				var lookupIndex = 0
				if cubeCornersExistence[0] {
					lookupIndex |= 1
				}
				if cubeCornersExistence[1] {
					lookupIndex |= 2
				}
				if cubeCornersExistence[2] {
					lookupIndex |= 4
				}
				if cubeCornersExistence[3] {
					lookupIndex |= 8
				}
				if cubeCornersExistence[4] {
					lookupIndex |= 16
				}
				if cubeCornersExistence[5] {
					lookupIndex |= 32
				}
				if cubeCornersExistence[6] {
					lookupIndex |= 64
				}
				if cubeCornersExistence[7] {
					lookupIndex |= 128
				}

				for i := 0; triangulation[lookupIndex][i] != -1; i += 3 {
					// Get indices of corner points A and B for each of the three edges
					// of the cube that need to be joined to form the triangle.
					a0 := cornerIndexAFromEdge[triangulation[lookupIndex][i]]
					b0 := cornerIndexBFromEdge[triangulation[lookupIndex][i]]

					a1 := cornerIndexAFromEdge[triangulation[lookupIndex][i+1]]
					b1 := cornerIndexBFromEdge[triangulation[lookupIndex][i+1]]

					a2 := cornerIndexAFromEdge[triangulation[lookupIndex][i+2]]
					b2 := cornerIndexBFromEdge[triangulation[lookupIndex][i+2]]

					v1 := interpolateVerts(cubeCornerPositions[a0], cubeCornerPositions[b0], cubeCorners[a0], cubeCorners[b0], cutoff /*jitter*/)
					v2 := interpolateVerts(cubeCornerPositions[a1], cubeCornerPositions[b1], cubeCorners[a1], cubeCorners[b1], cutoff /*jitter*/)
					v3 := interpolateVerts(cubeCornerPositions[a2], cubeCornerPositions[b2], cubeCorners[a2], cubeCorners[b2], cutoff /*jitter*/)

					tris = append(
						tris,
						LookupOrAdd(marchingWorkingData, v1),
						LookupOrAdd(marchingWorkingData, v3),
						LookupOrAdd(marchingWorkingData, v2),
					)
				}
			}
		}
	}

	return modeling.NewMesh(
		tris,
		map[string][]vector.Vector3{
			modeling.PositionAttribute: marchingWorkingData.verts,
		},
		map[string][]vector.Vector2{
			modeling.TexCoordAttribute: marchingWorkingData.uvs,
		},
		nil,
		nil,
	)
}
