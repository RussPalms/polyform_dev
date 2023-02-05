package materials

import (
	"math"
	"math/rand"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/vector/vector3"
)

type Dielectric struct {
	indexOfRefraction float64
}

func NewDielectric(indexOfRefraction float64) Dielectric {
	return Dielectric{indexOfRefraction}
}

func (d Dielectric) Scatter(in geometry.Ray, rec *rendering.HitRecord, attenuation *vector3.Float64, scattered *geometry.Ray) bool {
	refractionRatio := d.indexOfRefraction
	if rec.FrontFace {
		refractionRatio = (1.0 / d.indexOfRefraction)
	}

	unitDirection := in.Direction().Normalized()
	cosTheta := math.Min(unitDirection.Scale(-1).Dot(rec.Normal), 1.0)
	sinTheta := math.Sqrt(1.0 - (cosTheta * cosTheta))

	cannotRefract := refractionRatio*sinTheta > 1.0
	var direction vector3.Float64

	if cannotRefract || Reflectance(cosTheta, refractionRatio) > rand.Float64() {
		direction = unitDirection.Reflect(rec.Normal)
	} else {
		direction = unitDirection.Refract(rec.Normal, refractionRatio)
	}

	*scattered = geometry.NewRay(rec.Point, direction)
	*attenuation = vector3.New(1., 1., 1.)
	return true
}
