package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
)

func imageToEdgeData(src image.Image, boundaryValue float64) [][]float64 {
	imageData := make([][]float64, src.Bounds().Dx())
	for i := 0; i < len(imageData); i++ {
		imageData[i] = make([]float64, src.Bounds().Dy())
	}

	texturing.Convolve(src, func(x, y int, values []color.Color) {
		for i := 0; i < 9; i++ {
			if values[4] != values[i] {
				imageData[x][y] = 0
				return
			}
		}

		_, _, _, a := values[4].RGBA()
		if a&255 == 255 {
			imageData[x][y] = -boundaryValue
		} else {
			imageData[x][y] = boundaryValue
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

	return png.Decode(logoFile)
}

func heatPropegate(data [][]float64, iterations int, decay float64) {
	for i := 0; i < iterations; i++ {
		texturing.ConvolveArray[float64](data, func(x, y int, values []float64) {
			if data[x][y] == 0 {
				return
			}
			total := values[0] + values[1] + values[2] + values[3] + values[5] + values[6] + values[7] + values[8]
			data[x][y] = (total / 8) * decay
		})
	}
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
				dst.SetRGBA(x, y, color.RGBA{R: byte(val * 255 * delta), G: 0, B: 0, A: 255})
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

func main() {
	maxHeat := 100.
	logoFileName := "logo.png"
	img, err := loadImage(logoFileName)
	check(err)
	imgData := imageToEdgeData(img, maxHeat)
	heatPropegate(imgData, 250, 0.9999)
	check(debugPropegation(imgData, "debug.png"))

	top := sdf.RoundedCylinder(vector3.New(600., 11., 600.), 600, 1, 15)
	oreoLogoSDF := func(v vector3.Float64) float64 {
		pixel := v.RoundToInt().XZ()

		if pixel.X() >= len(imgData) || pixel.X() < 0 {
			return 1
		}

		if pixel.Y() >= len(imgData[0]) || pixel.Y() < 0 {
			return 1
		}

		return math.Max(imgData[pixel.X()][pixel.Y()], top(v))
	}

	waferSDF := sdf.Union(oreoLogoSDF, sdf.RoundedCylinder(vector3.New(640., -16., 640.), 300, 1, 20))

	cookieField := marching.Field{
		Domain: geometry.NewAABBFromPoints(
			vector3.New(0., -40., 0.),
			vector3.New(1300., 40., 1300.),
		),
		Float1Functions: map[string]sample.Vec3ToFloat{
			modeling.PositionAttribute: waferSDF,
		},
	}

	// Colors I color picked from images of oreos
	colorA := vector3.New(0x60/255., 0x53/255., 0x4a/255.)
	colorB := vector3.New(0x69/255., 0x59/255., 0x4b/255.)

	// High contrast colors to help debug the noise.
	// colorA = vector3.New(1., 0., 0.)
	// colorB = vector3.New(0., 1., 0.)

	colorDif := colorB.Sub(colorA)
	vertexColors := make([]vector3.Float64, 0)

	oreoCookieTop := cookieField.March(modeling.PositionAttribute, 0.25, 0).
		ModifyFloat3Attribute(
			modeling.PositionAttribute,
			func(i int, v vector3.Float64) vector3.Float64 {
				vertNoise := vector3.New(
					noise.Perlin1D(v.X()),
					noise.Perlin1D(v.Y()),
					noise.Perlin1D(v.Z()),
				).Scale(30)

				// Compute a vertex color using perlin noise
				colorTime := (noise.Perlin3D(v.DivByConstant(10)) * 0.5) + 0.5
				colorNoise := colorA.Add(colorDif.Scale(colorTime))
				vertexColors = append(vertexColors, colorNoise)

				return v.Add(vertNoise)
			},
		).
		SetFloat3Attribute(modeling.ColorAttribute, vertexColors).
		Transform(
			meshops.LaplacianSmoothTransformer{
				Attribute:       modeling.PositionAttribute,
				Iterations:      30,
				SmoothingFactor: .1,
			},
			meshops.SmoothNormalsTransformer{},
			meshops.VertexColorSpaceTransformer{},
			meshops.CenterAttribute3DTransformer{},
		)

	flipRotation := modeling.UnitQuaternionFromTheta(math.Pi, vector3.Right[float64]())
	oreoCookieBottom := oreoCookieTop.
		Transform(
			meshops.RotateAttribute3DTransformer{
				Attribute: modeling.PositionAttribute,
				Amount:    flipRotation,
			},
			meshops.RotateAttribute3DTransformer{
				Attribute: modeling.NormalAttribute,
				Amount:    flipRotation,
			},
		).
		Translate(vector3.New(0., -150., 0.))

	icingField := marching.Field{
		Domain: geometry.NewAABBFromPoints(
			vector3.New(0., -40., 0.),
			vector3.New(1300., 40., 1300.),
		), Float1Functions: map[string]sample.Vec3ToFloat{
			modeling.PositionAttribute: sdf.RoundedCylinder(vector3.New(640., 0., 640.), 290, 1, 50),
		},
	}

	icing := icingField.March(modeling.PositionAttribute, 0.25, 0).
		ModifyFloat3Attribute(
			modeling.PositionAttribute,
			func(i int, v vector3.Float64) vector3.Float64 {
				vertNoise := vector3.New(
					noise.Perlin1D(v.X()/10),
					noise.Perlin1D(v.Y()/10),
					noise.Perlin1D(v.Z()/10),
				).Scale(10)
				return v.Add(vertNoise)
			},
		).
		Transform(
			meshops.LaplacianSmoothTransformer{
				Attribute:       modeling.PositionAttribute,
				Iterations:      30,
				SmoothingFactor: .1,
			},
			meshops.SmoothNormalsTransformer{},
			meshops.CenterAttribute3DTransformer{},
		).
		Translate(vector3.New(0., -70., 0.))

	gltf.SaveBinary(
		"oreo.glb",
		gltf.PolyformScene{
			Models: []gltf.PolyformModel{
				{Mesh: oreoCookieTop},
				{Mesh: icing,
					Material: &gltf.PolyformMaterial{
						PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
							BaseColorFactor: color.RGBA{
								R: uint8(0xf8),
								G: uint8(0xf6),
								B: uint8(0xf7),
								A: 255,
							},
						},
					},
				},
				{Mesh: oreoCookieBottom},
			},
		},
	)
}
