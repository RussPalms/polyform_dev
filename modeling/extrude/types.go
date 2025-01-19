package extrude

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[ScrewNode](factory)
	refutil.RegisterType[CircleNode](factory)
	generator.RegisterTypes(factory)
}
