package extrude

import (
	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/vector"
)

type LinePoint struct {
	Point  vector.Vector3
	Up     vector.Vector3
	Width  float64
	Height float64
}

func directionsOfLinePoints(points []LinePoint) []vector.Vector3 {
	pointVec := make([]vector.Vector3, len(points))
	for i, point := range points {
		pointVec[i] = point.Point
	}
	return directionOfPoints(pointVec)
}

func Line(linePoints []LinePoint) mesh.Mesh {
	if len(linePoints) < 2 {
		panic("extruding a line requires 2 or more points")
	}

	vertices := make([]vector.Vector3, 0)
	normals := make([]vector.Vector3, 0)
	uvs := [][]vector.Vector2{make([]vector.Vector2, 0)}
	directions := directionsOfLinePoints(linePoints)
	for i, p := range linePoints {

		low := p.Point.Add(p.Up.MultByConstant(p.Height))
		outDir := directions[i].Cross(p.Up).MultByConstant(p.Width)

		rightPoint := low.Add(outDir)
		leftPoint := low.Sub(outDir)

		rightNormal := p.Up
		leftNormal := p.Up

		if p.Width != 0 {
			rightNormal = rightPoint.Sub(p.Point).Normalized().Cross(directions[i]).MultByConstant(-1)
			leftNormal = leftPoint.Sub(p.Point).Normalized().Cross(directions[i]).MultByConstant(-1)
		}

		vertices = append(
			vertices,
			p.Point,
			rightPoint,
			leftPoint,
		)

		normals = append(
			normals,
			p.Up,
			rightNormal,
			leftNormal,
		)
	}

	tris := make([]int, 0)
	for i := 1; i < len(linePoints); i++ {

		front := i * 3
		back := (i - 1) * 3

		frontMiddle := front
		frontRight := front + 1
		frontLeft := front + 2

		backMiddle := back
		backRight := back + 1
		backLeft := back + 2

		tris = append(
			tris,

			// Right Side
			frontMiddle, backMiddle, backRight,
			frontMiddle, backRight, frontRight,

			// Left Side
			frontMiddle, frontLeft, backMiddle,
			frontLeft, backLeft, backMiddle,
		)
	}

	return mesh.NewMesh(
		tris,
		vertices,
		normals,
		uvs,
	)
}
