package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/drawing/texturing/normals"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type StemMesh struct {
	nodes.StructData[gltf.PolyformModel]

	StemResolution nodes.NodeOutput[float64]
	TopDip         nodes.NodeOutput[float64]
}

func (sm *StemMesh) Mesh() nodes.NodeOutput[gltf.PolyformModel] {
	return nodes.StructNodeOutput[gltf.PolyformModel]{
		Definition: sm,
	}
}

func (sm StemMesh) Process() (gltf.PolyformModel, error) {
	stemCanvas := marching.NewMarchingCanvas(sm.StemResolution.Data())

	sides := 6

	fields := make([]marching.Field, 0)
	angleInc := (math.Pi * 2.) / float64(sides)

	topPoint := 0.2

	fields = append(fields, marching.Line(
		vector3.New(0., 0.05, 0.),
		vector3.New(0., topPoint*.95, 0.),
		0.02,
		1,
	))

	for i := 0; i < sides; i++ {
		rot := quaternion.FromTheta(angleInc*float64(i), vector3.Up[float64]())

		rotatedPoints := rot.RotateArray([]vector3.Float64{
			vector3.New(.15, 0.08, -.025+(rand.Float64()*.05)),
			vector3.New(.05, 0.05, 0.),
			vector3.New(.03, topPoint, 0.),
		})

		fields = append(
			fields,
			marching.VarryingThicknessLine(
				[]sdf.LinePoint{
					{
						Point:  rotatedPoints[0],
						Radius: 0.01 + (rand.Float64() * 0.005),
					},
					{
						Point:  rotatedPoints[1],
						Radius: 0.02 + (rand.Float64() * 0.02),
					},
					{
						Point:  rotatedPoints[2],
						Radius: 0.02 + (rand.Float64() * 0.01),
					},
				},
				1,
			),
		)
	}
	stemCanvas.AddFieldParallel(marching.CombineFields(fields...))

	mesh := stemCanvas.
		MarchParallel(0).
		Transform(
			meshops.LaplacianSmoothTransformer{
				Iterations:      20,
				SmoothingFactor: 0.1,
			},
			meshops.TranslateAttribute3DTransformer{
				Amount: vector3.New(0., 1-sm.TopDip.Data()+0.055, 0.),
			},
			meshops.SmoothNormalsTransformer{},
		)

	pumpkinVerts := mesh.Float3Attribute(modeling.PositionAttribute)
	newUVs := make([]vector2.Float64, pumpkinVerts.Len())
	center := vector3.New(0., 0.5, 0.)
	up := vector3.Up[float64]()
	for i := 0; i < pumpkinVerts.Len(); i++ {
		vert := pumpkinVerts.At(i)

		xzTheta := math.Atan2(vert.Z(), vert.X())
		xzTheta = math.Abs(xzTheta) // Avoid the UV seam

		dir := vert.Sub(center)
		angle := math.Acos(dir.Dot(up) / (dir.Length() * up.Length()))

		newUVs[i] = vector2.New(xzTheta/(math.Pi*2), angle)
	}

	return gltf.PolyformModel{
		Name: "Stem",
		Mesh: mesh.SetFloat2Attribute(modeling.TexCoordAttribute, newUVs),
		Material: &gltf.PolyformMaterial{
			PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
				BaseColorTexture: &gltf.PolyformTexture{
					URI: "Texturing/stem.png",
				},
				MetallicRoughnessTexture: &gltf.PolyformTexture{
					URI: "Texturing/stem-roughness.png",
				},
			},
			NormalTexture: &gltf.PolyformNormal{
				PolyformTexture: gltf.PolyformTexture{
					URI: "Texturing/stem-normal.png",
				},
			},
		},
	}, nil
}

type StemNormalImage struct {
	nodes.StructData[generator.Artifact]

	NumberOfLines nodes.NodeOutput[int]
}

func (sr *StemNormalImage) Image() nodes.NodeOutput[generator.Artifact] {
	return nodes.StructNodeOutput[generator.Artifact]{
		Definition: sr,
	}
}

func (sni StemNormalImage) Process() (generator.Artifact, error) {
	dim := 1024
	img := image.NewRGBA(image.Rect(0, 0, dim, dim))
	// normals.Fill(img)

	n := noise.NewTilingNoise(dim, 1/64., 5)

	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			val := n.Noise(x, y)
			// p := n.Noise(vector2.New(xDim*10, yDim*10), 100)
			p := (val * 128) + 128

			img.Set(x, y, color.RGBA{
				R: byte(p), // byte(len * 255),
				G: byte(p),
				B: byte(p),
				A: 255,
			})
		}
	}

	img = texturing.ToNormal(img)

	numLines := sni.NumberOfLines.Data()

	spacing := float64(dim) / float64(numLines)
	halfSpacing := float64(spacing) / 2.

	segments := 8
	yInc := float64(dim) / float64(segments)
	halfYInc := yInc / 2.

	for i := 0; i < numLines; i++ {
		dir := normals.Subtractive
		if rand.Float64() > 0.5 {
			dir = normals.Additive
		}

		startX := (float64(i) * spacing) + (spacing / 2)
		width := 10 + (rand.Float64() * 20)

		start := vector2.New(startX, 0)
		for seg := 0; seg < segments-1; seg++ {
			end := vector2.New(
				startX-(halfSpacing/2)+(rand.Float64()*halfSpacing),
				start.Y()+halfYInc+(yInc*rand.Float64()),
			)
			normals.Line{
				Start:           start,
				End:             end,
				Width:           width,
				NormalDirection: dir,
			}.Round(img)
			start = end
		}

		normals.Line{
			Start:           start,
			End:             vector2.New(startX, float64(dim)),
			Width:           width,
			NormalDirection: dir,
		}.Round(img)

	}

	return &generator.ImageArtifact{Image: img}, nil
}

type StemRoughness struct {
	nodes.StructData[generator.Artifact]

	Dimensions nodes.NodeOutput[int]
	Roughness  nodes.NodeOutput[float64]
}

func (sr *StemRoughness) Image() nodes.NodeOutput[generator.Artifact] {
	return nodes.StructNodeOutput[generator.Artifact]{
		Definition: sr,
	}
}

func (sr StemRoughness) Process() (generator.Artifact, error) {
	dim := sr.Dimensions.Data()
	stemRoughnessImage := image.NewRGBA(image.Rect(0, 0, dim, dim))

	for x := 0; x < dim; x++ {
		for y := 0; y < dim; y++ {
			stemRoughnessImage.Set(x, y, color.RGBA{
				R: 0, // byte(len * 255),
				G: byte(255 * sr.Roughness.Data()),
				B: 0,
				A: 255,
			})
		}
	}

	return &generator.ImageArtifact{Image: stemRoughnessImage}, nil
}
