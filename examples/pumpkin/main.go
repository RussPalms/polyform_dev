package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"time"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/math/colors"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/fogleman/gg"
)

func closestTimeOnMultiLineSegment(point vector3.Float64, multiLine []vector3.Float64, totalLength float64) float64 {
	if len(multiLine) < 2 {
		panic("line segment required 2 or more points")
	}

	minDist := math.MaxFloat64

	closestTime := 0.
	lengthTraversed := 0.
	for i := 1; i < len(multiLine); i++ {
		line := geometry.NewLine3D(multiLine[i-1], multiLine[i])
		lineLength := line.Length()
		dist := line.ClosestPointOnLine(point).Distance(point)
		if dist < minDist {
			minDist = dist
			closestTime = (lengthTraversed + (lineLength * line.ClosestTimeOnLine(point))) / totalLength
		}
		lengthTraversed += lineLength
	}

	return closestTime
}

func newPumpkinMesh(
	cubersPerUnit float64,
	maxWidth, topDip, distanceFromCenter, wedgeLineRadius float64,
	sides int,
	imageField [][]float64,
	useImageField bool,
) modeling.Mesh {
	canvas := marching.NewMarchingCanvas(cubersPerUnit)

	outerPoints := []vector3.Float64{
		vector3.New(0., .3, distanceFromCenter),
		vector3.New(0., .25, distanceFromCenter+(maxWidth*0.5)),
		vector3.New(0., 0.5, distanceFromCenter+maxWidth),
		vector3.New(0., .8, distanceFromCenter+(maxWidth*0.75)),
		vector3.New(0., 1-topDip, distanceFromCenter),
	}

	pointsBoundsLower, pointsBoundsHigher := vector3.Float64Array(outerPoints).Bounds()
	boundsCenter := pointsBoundsHigher.Midpoint(pointsBoundsLower)
	innerPoints := vector3.Float64Array(outerPoints).
		Add(boundsCenter.Scale(-1)).
		Scale(0.3).
		Add(boundsCenter)

	fields := make([]marching.Field, 0)
	angleInc := (math.Pi * 2.) / float64(sides)
	for i := 0; i < sides; i++ {
		rot := modeling.UnitQuaternionFromTheta(angleInc*float64(i), vector3.Up[float64]())

		outer := []sdf.LinePoint{
			{Point: rot.Rotate(outerPoints[0]), Radius: 0.33 * wedgeLineRadius},
			{Point: rot.Rotate(outerPoints[1]), Radius: 0.33 * wedgeLineRadius},
			{Point: rot.Rotate(outerPoints[2]), Radius: 1.00 * wedgeLineRadius},
			{Point: rot.Rotate(outerPoints[3]), Radius: 0.66 * wedgeLineRadius},
			{Point: rot.Rotate(outerPoints[4]), Radius: 0.33 * wedgeLineRadius},
		}

		inner := []sdf.LinePoint{
			{Point: rot.Rotate(innerPoints[0]), Radius: 0.33 * wedgeLineRadius},
			{Point: rot.Rotate(innerPoints[1]), Radius: 0.33 * wedgeLineRadius},
			{Point: rot.Rotate(innerPoints[2]), Radius: 1.00 * wedgeLineRadius},
			{Point: rot.Rotate(innerPoints[3]), Radius: 0.66 * wedgeLineRadius},
			{Point: rot.Rotate(innerPoints[4]), Radius: 0.33 * wedgeLineRadius},
		}
		fields = append(fields, marching.Subtract(marching.VarryingThicknessLine(outer, 1), marching.VarryingThicknessLine(inner, 2)))
	}

	allFields := marching.CombineFields(fields...)

	pumpkinField := allFields
	if useImageField {
		pumpkinField = marching.Subtract(
			allFields,
			marching.Field{
				Domain: allFields.Domain,
				Float1Functions: map[string]sample.Vec3ToFloat{
					modeling.PositionAttribute: func(f vector3.Float64) float64 {

						pixel := f.XY().
							Scale(float64(len(imageField)) * 2).
							RoundToInt().
							Sub(vector2.New(-len(imageField)/2, int(float64(len(imageField))*0.75)))

						if pixel.X() < 0 || pixel.X() >= len(imageField) {
							return 10
						}

						if pixel.Y() < 0 || pixel.Y() >= len(imageField) {
							return 10
						}

						if f.Z() < .2 {
							return 10
						}

						return -imageField[pixel.X()][len(imageField)-1-pixel.Y()]
					},
				},
			},
		)
	}

	addFieldStart := time.Now()
	canvas.AddField(pumpkinField)
	log.Printf("time to add field: %s", time.Since(addFieldStart))

	marchStart := time.Now()
	mesh := canvas.MarchParallel(0)
	log.Printf("time to march: %s", time.Since(marchStart))

	mesh = mesh.Transform(
		meshops.LaplacianSmoothTransformer{
			Iterations:      20,
			SmoothingFactor: 0.1,
		},
		meshops.SmoothNormalsTransformer{},
	)

	pumpkinVerts := mesh.Float3Attribute(modeling.PositionAttribute)
	newUVs := make([]vector2.Float64, pumpkinVerts.Len())
	totalLength := vector3.Array[float64](outerPoints).Distance()

	newColors := make([]vector3.Float64, pumpkinVerts.Len())
	for i := 0; i < pumpkinVerts.Len(); i++ {
		vert := pumpkinVerts.At(i)
		xzTheta := math.Atan2(vert.Z(), vert.X())
		if math.IsNaN(xzTheta) {
			// xzTheta += math.Pi*2
			continue
		}

		rot := modeling.UnitQuaternionFromTheta(xzTheta, vector3.Up[float64]())
		rotatedLine := rot.RotateArray(outerPoints)

		y := closestTimeOnMultiLineSegment(vert, rotatedLine, totalLength)
		if y < 0 || y > 1.001 {
			panic(y)
		}

		newUVs[i] = vector2.New(-xzTheta/(math.Pi*2), y)
		newUVs[i] = vector2.New(-xzTheta/(math.Pi*2), vert.Y())
		y = xzTheta / (math.Pi * 2)
		// secondHalf := 0.
		// if y > 0.5 {
		// 	secondHalf = (y - 0.5) * 2
		// 	y = 1
		// } else {
		// 	y *= 2
		// }
		newColors[i] = vector3.New(y, 0, 0.)
	}

	// METHOD 1 ===============================================================
	// Works okay, issues from the dip of the top of the pumpkin causing the
	// texture to reverse directions
	// pumpkinVerts := mesh.Float3Attribute(modeling.PositionAttribute)
	// newUVs := make([]vector2.Float64, pumpkinVerts.Len())

	// for i := 0; i < pumpkinVerts.Len(); i++ {
	// 	  vert := pumpkinVerts.At(i)
	// 	  xzPos := vert.XZ()
	// 	  xzTheta := math.Atan2(xzPos.Y(), xzPos.X())
	// 	  newUVs[i] = vector2.New(xzTheta/(math.Pi*2), vert.Y())
	// }

	mesh = mesh.SetFloat2Attribute(modeling.TexCoordAttribute, newUVs) //.SetFloat3Attribute(modeling.ColorAttribute, newColors)

	return mesh
}

func pumpkinStem(maxWidth, minWidth, length, tipOffset float64) marching.Field {
	return marching.VarryingThicknessLine([]sdf.LinePoint{
		{Point: vector3.New(0., 0., 0.), Radius: maxWidth},
		{Point: vector3.New(0., length*.8, 0.), Radius: minWidth},
		{Point: vector3.New(tipOffset, length, 0.), Radius: minWidth},
	}, 1)
}

func imageToEdgeData(src image.Image, fillValue float64) [][]float64 {
	imageData := make([][]float64, src.Bounds().Dx())
	for i := 0; i < len(imageData); i++ {
		imageData[i] = make([]float64, src.Bounds().Dy())
	}

	texturing.Convolve(src, func(x, y int, kernel []color.Color) {
		if texturing.SimpleEdgeTest(kernel) {
			imageData[x][y] = 0
			return
		}

		if colors.RedEqual(kernel[4], 255) {
			imageData[x][y] = -fillValue
		} else {
			imageData[x][y] = fillValue
		}
	})

	return imageData
}

func loadImage(imageName string) (image.Image, error) {
	logoFile, err := os.Open(imageName)
	if err != nil {
		return nil, err
	}
	defer logoFile.Close()

	img, _, err := image.Decode(logoFile)

	return img, err
}

func heatPropegate(data [][]float64, iterations int, decay float64) [][]float64 {
	tempData := make([][]float64, len(data))
	for r := 0; r < len(tempData); r++ {
		tempData[r] = make([]float64, len(data[r]))
	}

	for i := 0; i < iterations; i++ {
		toConvole := data
		toStore := tempData
		if i%2 == 1 {
			toConvole = tempData
			toStore = data
		}
		texturing.ConvolveArray[float64](toConvole, func(x, y int, kernel []float64) {
			if toConvole[x][y] == 0 {
				return
			}
			total := kernel[0] + kernel[1] + kernel[2] + kernel[3] + kernel[5] + kernel[6] + kernel[7] + kernel[8]
			toStore[x][y] = (total / 8) * decay
		})
	}

	if iterations%2 == 1 {
		return tempData
	}
	return data
}

func debugPropegation(data [][]float64, filename string) error {
	dst := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{X: len(data), Y: len(data[0])}})

	max := -math.MaxFloat64
	min := math.MaxFloat64
	for x := 0; x < len(data); x++ {
		row := data[x]
		for y := 0; y < len(row); y++ {
			max = math.Max(max, row[y])
			min = math.Min(min, row[y])
		}
	}

	delta := max - min

	for x := 0; x < len(data); x++ {
		row := data[x]
		for y := 0; y < len(row); y++ {
			val := row[y] / delta
			if val > 0 {
				dst.SetRGBA(x, y, color.RGBA{R: byte(val * 255), G: 0, B: 0, A: 255})
			} else {
				dst.SetRGBA(x, y, color.RGBA{R: 0, G: byte(val * -255), B: 0, A: 255})
			}
		}
	}

	imgFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer imgFile.Close()
	return png.Encode(imgFile, dst)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

//go:embed uv.jpg
var debugUVImageData []byte

func main() {

	maxHeat := 100.
	logoFileName := "face.jpg"
	img, err := loadImage(logoFileName)
	check(err)
	imgData := imageToEdgeData(img, maxHeat)
	imgData = heatPropegate(imgData, 250, 0.9999)
	check(debugPropegation(imgData, "debug.png"))

	app := generator.App{
		Name:        "Pumpkin",
		Version:     "0.0.1",
		Description: "Making a pumpkin for Haloween",
		Authors: []generator.Author{
			{
				Name: "Eli C Davis",
			},
		},
		WebScene: &room.WebScene{
			Fog: room.WebSceneFog{
				Near:  2,
				Far:   10,
				Color: coloring.WebColor{R: 0x13, G: 0x0b, B: 0x3c, A: 255},
			},
			Ground:     coloring.WebColor{R: 0x4f, G: 0x6d, B: 0x55, A: 255},
			Background: coloring.WebColor{R: 0x13, G: 0x0b, B: 0x3c, A: 255},
			Lighting:   coloring.WebColor{R: 0xff, G: 0xd8, B: 0x94, A: 255},
		},
		Generator: &generator.Generator{
			SubGenerators: map[string]*generator.Generator{
				"Texturing": {
					Parameters: &generator.GroupParameter{
						Parameters: []generator.Parameter{
							&generator.ColorParameter{
								Name:         "Base Color",
								DefaultValue: coloring.WebColor{R: 0xf9, G: 0x81, B: 0x1f, A: 255},
							},
							&generator.ColorParameter{
								Name:         "Line Color",
								DefaultValue: coloring.WebColor{R: 0, G: 0x81, B: 0x1f, A: 255},
							},
							&generator.IntParameter{
								Name:         "Lines",
								DefaultValue: 8,
							},
						},
					},
					Producers: map[string]generator.Producer{
						"pumpkin.png": func(c *generator.Context) (generator.Artifact, error) {
							const texDimension = 1024

							ctx := gg.NewContext(texDimension, texDimension)
							ctx.SetColor(c.Parameters.Color("Base Color"))

							ctx.DrawRectangle(0, 0, texDimension, texDimension)
							ctx.Fill()

							lines := c.Parameters.Int("Lines")

							ctx.SetColor(c.Parameters.Color("Line Color"))
							ctx.SetLineWidth(2)
							spacing := texDimension / (lines)
							for i := 0; i < lines; i++ {
								xDim := float64((spacing / 2) + (spacing * i))
								ctx.DrawLine(xDim, 0, xDim, texDimension)
								ctx.Stroke()
							}

							return generator.ImageArtifact{
								Image: ctx.Image(),
							}, nil
						},
					},
				},
			},
			Parameters: &generator.GroupParameter{
				Name: "Pumpkin",
				Parameters: []generator.Parameter{
					&generator.FloatParameter{
						Name:         "Cubes Per Unit",
						DefaultValue: 40,
					},

					&generator.IntParameter{
						Name:         "Wedges",
						DefaultValue: 10,
					},

					&generator.FloatParameter{
						Name:         "Wedge Spacing",
						DefaultValue: .1,
					},

					&generator.FloatParameter{
						Name:         "Wedge Radius",
						DefaultValue: .3,
					},

					&generator.FloatParameter{
						Name:         "Max Width",
						DefaultValue: .3,
					},

					&generator.FloatParameter{
						Name:         "Top Dip",
						DefaultValue: .2,
					},

					&generator.ColorParameter{
						Name:         "Light Color",
						DefaultValue: coloring.WebColor{R: 0xf4, G: 0xf5, B: 0xad, A: 255},
					},

					&generator.BoolParameter{
						Name:         "Carve",
						DefaultValue: true,
					},

					&generator.GroupParameter{
						Name: "Stem",
						Parameters: []generator.Parameter{
							&generator.ColorParameter{
								Name:         "Color",
								DefaultValue: coloring.WebColor{R: 0x6d, G: 0x52, B: 0x40, A: 255},
							},
							&generator.FloatParameter{
								Name:         "Base Width",
								DefaultValue: 0.07,
							},
							&generator.FloatParameter{
								Name:         "Tip Width",
								DefaultValue: 0.03,
							},
							&generator.FloatParameter{
								Name:         "Length",
								DefaultValue: 0.3,
							},
							&generator.FloatParameter{
								Name:         "Tip Offset",
								DefaultValue: 0.1,
							},
						},
					},
				},
			},
			Producers: map[string]generator.Producer{

				"uvMap.png": func(c *generator.Context) (generator.Artifact, error) {
					img, err := jpeg.Decode(bytes.NewReader(debugUVImageData))
					return &generator.ImageArtifact{Image: img}, err
				},
				"pumpkin.glb": func(c *generator.Context) (generator.Artifact, error) {

					pumpkinMesh := newPumpkinMesh(
						c.Parameters.Float64("Cubes Per Unit"),
						c.Parameters.Float64("Max Width"),
						c.Parameters.Float64("Top Dip"),
						c.Parameters.Float64("Wedge Spacing"),
						c.Parameters.Float64("Wedge Radius"),
						c.Parameters.Int("Wedges"),
						imgData,
						c.Parameters.Bool("Carve"),
					)

					stemParams := c.Parameters.Group("Stem")
					stemCanvas := marching.NewMarchingCanvas(c.Parameters.Float64("Cubes Per Unit"))
					stemCanvas.AddFieldParallel(pumpkinStem(
						stemParams.Float64("Base Width"),
						stemParams.Float64("Tip Width"),
						stemParams.Float64("Length"),
						stemParams.Float64("Tip Offset"),
					))
					stem := stemCanvas.
						MarchParallel(0).
						Transform(
							meshops.LaplacianSmoothTransformer{
								Iterations:      20,
								SmoothingFactor: 0.1,
							},
							meshops.TranslateAttribute3DTransformer{
								Amount: vector3.New(0., 1-c.Parameters.Float64("Top Dip"), 0.),
							},
						)

					// texturingParams := c.Parameters.Group("Texturing")

					return generator.GltfArtifact{
						Scene: gltf.PolyformScene{
							Models: []gltf.PolyformModel{
								{
									Name: "Pumpkin",
									Mesh: pumpkinMesh,
									Material: &gltf.PolyformMaterial{
										PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
											BaseColorTexture: &gltf.PolyformTexture{
												URI: "Texturing/pumpkin.png", //"uvMap.png",
												Sampler: &gltf.Sampler{
													WrapS: gltf.SamplerWrap_REPEAT,
													WrapT: gltf.SamplerWrap_REPEAT,
												},
											},
											// BaseColorFactor: texturingParams.Color("Base Color"),
											// MetallicFactor:  1,
											// RoughnessFactor: 0,
										},
										Extensions: []gltf.MaterialExtension{
											// gltf.PolyformMaterialsUnlit{},
										},
									},
								},
								{
									Name: "Stem",
									Mesh: stem,
									Material: &gltf.PolyformMaterial{
										PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
											BaseColorFactor: stemParams.Color("Color"),
										},
									},
								},
							},
							Lights: []gltf.KHR_LightsPunctual{
								{
									Type:     gltf.KHR_LightsPunctualType_Point,
									Position: vector3.New(0., 0.5, 0.),
									Color:    c.Parameters.Color("Light Color"),
								},
							},
						},
					}, nil
				},
			},
		},
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
