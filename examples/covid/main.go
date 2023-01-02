package main

import (
	"image/color"
	"math"
	"math/rand"
	"path"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector"
)

func tendrilField(start, direction vector.Vector3, radius, length float64, plumbs int) sample.Vec3ToFloat {

	endPoint := start.Add(direction.MultByConstant(length))
	fields := []sample.Vec3ToFloat{
		marching.Line(start, endPoint, radius, 1),
	}

	angleIncrement := (math.Pi * 2) / float64(plumbs)
	perpendicular := direction.Perpendicular().Normalized()
	for i := 0; i < plumbs; i++ {
		rot := modeling.UnitQuaternionFromTheta(float64(i)*angleIncrement, direction)

		plumbRadius := radius * (.7 + (rand.Float64() * .2))

		plumbSite := rot.
			Rotate(perpendicular).
			MultByConstant(0.5).
			Add(start.Add(direction.MultByConstant(length * 1.2)))
		fields = append(fields, marching.Sphere(plumbSite, plumbRadius, 1))
	}

	return sample.SumVec3ToFloat(fields...)
}

func virusField(center vector.Vector3, virusWidth float64) sample.Vec3ToFloat {
	fields := []sample.Vec3ToFloat{
		marching.Sphere(center, virusWidth, 1),
	}

	tendrilCount := 20 + int(math.Round(20*rand.Float64()))
	tendrilRadius := .5 //.2 + (.4 * rand.Float64())
	tendrilLength := .7 + (.4 * rand.Float64())
	tendrilSites := repeat.FibonacciSpherePoints(tendrilCount, virusWidth)
	for _, site := range tendrilSites {
		direction := site.Normalized()
		start := center.Add(site)
		numberOfPlumbs := 3 + int(math.Round(2*rand.Float64()))
		fields = append(fields, tendrilField(start, direction, tendrilRadius, tendrilLength, numberOfPlumbs))
	}

	irregularSites := repeat.FibonacciSpherePoints(tendrilCount/2, virusWidth*.5)
	for _, site := range irregularSites {
		start := center.Add(site)
		fields = append(fields, marching.Sphere(start, tendrilRadius*2.5, 1))
	}

	return sample.SumVec3ToFloat(fields...)
}

func main() {
	resolution := 100
	cubesPerUnit := 10.
	canvas := marching.NewMarchingCanvas(resolution, resolution, resolution, cubesPerUnit)

	center := vector.Vector3One().MultByConstant((float64(resolution) / cubesPerUnit) / 2.)

	virusRadius := 2.
	canvas.AddFieldParallel(virusField(center, virusRadius))

	virusColor := coloring.NewColorStack(
		coloring.NewColorStackEntry(4, 1, 1, color.RGBA{199, 195, 195, 255}),
		coloring.NewColorStackEntry(1, 1, 1, color.RGBA{230, 50, 50, 255}),
		coloring.NewColorStackEntry(.1, 1, 1, color.RGBA{255, 112, 236, 255}),
	)

	textureMap := "covid.png"
	err := virusColor.Debug(path.Join("tmp/covid/", textureMap), 300, 100)
	if err != nil {
		panic(err)
	}

	uvs := make([]vector.Vector2, 0)
	mesh := canvas.March(.2).
		CenterFloat3Attribute(modeling.PositionAttribute).
		ScanFloat3Attribute(modeling.PositionAttribute, func(v vector.Vector3) {
			x := (v.Length() - 1.5) / 2.1
			uvs = append(uvs, vector.NewVector2(x, 0.5))
		}).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs).
		SmoothLaplacian(10, .1).
		CalculateSmoothNormals().
		SetMaterial(modeling.Material{
			Name:            "COVID",
			ColorTextureURI: &textureMap,
		})

	err = obj.Save("tmp/covid/covid.obj", mesh)
	if err != nil {
		panic(err)
	}
}
