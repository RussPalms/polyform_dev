package texturing

import (
	"image"
	"image/color"

	"github.com/EliCDavis/vector"
)

func Convolve(src image.Image, f func(x, y int, values []color.Color)) {
	for xIndex := 0; xIndex < src.Bounds().Dx(); xIndex++ {
		xLeft := xIndex - 1
		if xIndex == 0 {
			xLeft = xIndex + 1
		}
		xMid := xIndex
		xRight := xIndex + 1
		if xIndex == src.Bounds().Dx()-1 {
			xRight = xIndex - 1
		}

		for yIndex := 0; yIndex < src.Bounds().Dy(); yIndex++ {
			yBot := yIndex - 1
			if yIndex == 0 {
				yBot = yIndex + 1
			}
			yMid := yIndex
			yTop := yIndex + 1
			if yIndex == src.Bounds().Dx()-1 {
				yTop = yIndex - 1
			}

			f(xIndex, yIndex, []color.Color{
				src.At(xLeft, yTop), src.At(xMid, yTop), src.At(xRight, yTop),
				src.At(xLeft, yMid), src.At(xMid, yMid), src.At(xRight, yMid),
				src.At(xLeft, yBot), src.At(xMid, yBot), src.At(xRight, yBot),
			})
		}
	}
}

func averageColorComponents(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	r8 := r >> 8
	g8 := g >> 8
	b8 := b >> 8

	return (float64(r8+g8+b8) / 3.) / 255.
}

func ToNormal(src image.Image) image.Image {
	dst := image.NewRGBA(src.Bounds())
	scale := 1.
	Convolve(src, func(x, y int, vals []color.Color) {
		// float s[9] contains above samples
		n := vector.NewVector3(0, 0, 1)
		s0 := averageColorComponents(vals[0])
		s1 := averageColorComponents(vals[1])
		s2 := averageColorComponents(vals[2])
		s3 := averageColorComponents(vals[3])
		s5 := averageColorComponents(vals[5])
		s6 := averageColorComponents(vals[6])
		s7 := averageColorComponents(vals[7])
		s8 := averageColorComponents(vals[8])

		n = n.SetX(scale * -(s2 - s0 + 2*(s5-s3) + s8 - s6))
		n = n.SetY(scale * -(s6 - s0 + 2*(s7-s1) + s8 - s2))
		n = n.Normalized()

		dst.Set(x, y, color.RGBA{
			R: uint8((1 + n.X()) * 255),
			G: uint8((1 + n.Y()) * 255),
			B: uint8((1 + n.Z()) * 255),
			A: 255,
		})
	})
	return dst
}
