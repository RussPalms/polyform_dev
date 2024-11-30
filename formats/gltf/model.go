package gltf

import (
	"image/color"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/animation"
)

type PolyformScene struct {
	Models []PolyformModel
	Lights []KHR_LightsPunctual
}

// PolyformModel is a utility structure for reading/writing to GLTF format within
// polyform, and not an actual concept found within the GLTF format.
type PolyformModel struct {
	Name     string
	Mesh     modeling.Mesh
	Material *PolyformMaterial

	Skeleton   *animation.Skeleton
	Animations []animation.Sequence
}

type PolyformMaterial struct {
	Name                 string
	Extras               map[string]any
	AlphaMode            *MaterialAlphaMode
	AlphaCutoff          *float64
	PbrMetallicRoughness *PolyformPbrMetallicRoughness
	Extensions           []MaterialExtension
	NormalTexture        *PolyformNormal
	EmissiveFactor       color.Color
}

type PolyformPbrMetallicRoughness struct {
	BaseColorFactor          color.Color
	BaseColorTexture         *PolyformTexture
	MetallicFactor           *float64
	RoughnessFactor          *float64
	MetallicRoughnessTexture *PolyformTexture
}

type PolyformNormal struct {
	PolyformTexture
	Scale *float64
}

type PolyformTexture struct {
	URI     string
	Sampler *Sampler
}

func (pm *PolyformMaterial) equal(other *PolyformMaterial) bool {
	if pm == other {
		return true
	}

	if pm.Name != other.Name {
		return false
	}
	if !pm.PbrMetallicRoughness.equal(other.PbrMetallicRoughness) {
		return false
	}
	if !colorsEqual(pm.EmissiveFactor, other.EmissiveFactor) {
		return false
	}

	if (pm.AlphaMode == nil) != (other.AlphaMode == nil) {
		return false
	} else if pm.AlphaMode != nil && other.AlphaMode != nil && *pm.AlphaMode != *other.AlphaMode {
		return false
	}

	if !float64PtrsEqual(pm.AlphaCutoff, other.AlphaCutoff) {
		return false
	}
	if len(pm.Extensions) != len(other.Extensions) {
		return false
	}
	for i, ext := range pm.Extensions {
		if ext != other.Extensions[i] {
			return false
		}
	}
	return true
}

func (pt *PolyformTexture) equal(other *PolyformTexture) bool {
	if pt == other {
		return true
	}

	if pt.URI != other.URI {
		return false
	}

	if pt.Sampler == other.Sampler {
		return true
	} else if pt.Sampler == nil || other.Sampler == nil {
		return false
	}

	if pt.Sampler.MagFilter != other.Sampler.MagFilter ||
		pt.Sampler.MinFilter != other.Sampler.MinFilter ||
		pt.Sampler.WrapS != other.Sampler.WrapS ||
		pt.Sampler.WrapT != other.Sampler.WrapT {
		return false
	}

	return true
}

func (pt *PolyformNormal) equal(other *PolyformNormal) bool {
	if !pt.PolyformTexture.equal(&other.PolyformTexture) {
		return false
	}
	return float64PtrsEqual(pt.Scale, other.Scale)
}

func (pmr *PolyformPbrMetallicRoughness) equal(other *PolyformPbrMetallicRoughness) bool {
	if pmr == other {
		return true
	}

	if !float64PtrsEqual(pmr.MetallicFactor, other.MetallicFactor) ||
		!float64PtrsEqual(pmr.RoughnessFactor, other.RoughnessFactor) ||
		!colorsEqual(pmr.BaseColorFactor, other.BaseColorFactor) {
		return false
	}

	if !pmr.BaseColorTexture.equal(other.BaseColorTexture) ||
		!pmr.MetallicRoughnessTexture.equal(other.MetallicRoughnessTexture) {
		return false
	}

	return true
}

// Helper functions for comparing nullable values
func float64PtrsEqual(a, b *float64) bool {
	if a == b {
		return true
	} else if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func colorsEqual(a, b color.Color) bool {
	if a == b {
		return true
	} else if a == nil || b == nil {
		return false
	}
	// Since color.Color is an interface, we can only check for basic RGBA equality
	r1, g1, b1, a1 := a.(color.Color).RGBA()
	r2, g2, b2, a2 := b.(color.Color).RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
