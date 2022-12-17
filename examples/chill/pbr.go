package main

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/EliCDavis/mesh"
)

type PBRTextures struct {
	name     string
	color    image.Image
	normal   image.Image
	specular image.Image
}

func writeTex(name string, tex image.Image) error {
	if tex == nil {
		return nil
	}
	texOut, err := os.Create(name)
	if err != nil {
		return err
	}
	defer texOut.Close()
	return png.Encode(texOut, tex)
}

func (pbrt PBRTextures) Material() mesh.Material {
	colorPath := (pbrt.ColorPath())
	normalPath := (pbrt.NormalPath())
	specularPath := (pbrt.SpecularPath())
	return mesh.Material{
		Name:               pbrt.name,
		AmbientColor:       color.White,
		DiffuseColor:       color.White,
		ColorTextureURI:    &colorPath,
		NormalTextureURI:   &normalPath,
		SpecularTextureURI: &specularPath,
	}
}

func (pbrt PBRTextures) ColorPath() string {
	return pbrt.name + "_color.png"
}

func (pbrt PBRTextures) NormalPath() string {
	return pbrt.name + "_normal.png"
}

func (pbrt PBRTextures) SpecularPath() string {
	return pbrt.name + "_specular.png"
}

func (pbrt PBRTextures) Save() error {
	err := writeTex(pbrt.ColorPath(), pbrt.color)
	if err != nil {
		return err
	}

	err = writeTex(pbrt.NormalPath(), pbrt.normal)
	if err != nil {
		return err
	}

	return writeTex(pbrt.SpecularPath(), pbrt.specular)
}
