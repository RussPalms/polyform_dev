package main

import (
	"fmt"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

type InfoNode = nodes.StructNode[string, InfoNodeData]

type InfoNodeData struct {
	Original nodes.NodeOutput[modeling.Mesh]
	Final    nodes.NodeOutput[modeling.Mesh]
}

func (in InfoNodeData) Process() (string, error) {
	original := in.Original.Value().AttributeLength()
	final := in.Final.Value().AttributeLength()

	return fmt.Sprintf("Points: %d / %d\nPruned: %d", final, original, original-final), nil
}
