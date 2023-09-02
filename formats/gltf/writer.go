package gltf

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"math"

	"github.com/EliCDavis/bitlib"
	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/modeling/animation"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type Writer struct {
	buf          *bytes.Buffer
	bitW         *bitlib.Writer
	bytesWritten int

	accessors   []Accessor
	bufferViews []BufferView
	meshes      []Mesh
	nodes       []Node
	materials   []Material

	skins      []Skin
	animations []Animation

	textures     []Texture
	images       []Image
	samplers     []Sampler
	textureInfos []TextureInfo
	scene        []int
}

func NewWriter() *Writer {
	buf := &bytes.Buffer{}
	return &Writer{
		buf:         buf,
		bitW:        bitlib.NewWriter(buf, binary.LittleEndian),
		bufferViews: make([]BufferView, 0),
		accessors:   make([]Accessor, 0),
		nodes:       make([]Node, 0),
		meshes:      make([]Mesh, 0),
		materials:   make([]Material, 0),
		skins:       make([]Skin, 0),
		animations:  make([]Animation, 0),
	}
}

func (w Writer) WriteVector4AsFloat32(v vector4.Float64) {
	w.bitW.Float32(float32(v.X()))
	w.bitW.Float32(float32(v.Y()))
	w.bitW.Float32(float32(v.Z()))
	w.bitW.Float32(float32(v.W()))
}

func (w Writer) WriteVector4AsByte(v vector4.Float64) {
	w.bitW.Byte(uint8(v.X()))
	w.bitW.Byte(uint8(v.Y()))
	w.bitW.Byte(uint8(v.Z()))
	w.bitW.Byte(uint8(v.W()))
}

func (w Writer) WriteVector3AsFloat32(v vector3.Float64) {
	w.bitW.Float32(float32(v.X()))
	w.bitW.Float32(float32(v.Y()))
	w.bitW.Float32(float32(v.Z()))
}

func (w Writer) WriteVector3AsByte(v vector3.Float64) {
	w.bitW.Byte(uint8(v.X()))
	w.bitW.Byte(uint8(v.Y()))
	w.bitW.Byte(uint8(v.Z()))
}

func (w Writer) WriteVector2AsFloat32(v vector2.Float64) {
	w.bitW.Float32(float32(v.X()))
	w.bitW.Float32(float32(v.Y()))
}

func (w Writer) WriteVector2AsByte(v vector2.Float64) {
	w.bitW.Byte(uint8(v.X()))
	w.bitW.Byte(uint8(v.Y()))
}

func (w *Writer) WriteVector4(accessorComponentType AccessorComponentType, data iter.ArrayIterator[vector4.Float64]) {
	accessorType := AccessorType_VEC4

	min := vector4.Fill(math.MaxFloat64)
	max := vector4.Fill(-math.MaxFloat64)

	if accessorComponentType == AccessorComponentType_FLOAT {
		for i := 0; i < data.Len(); i++ {
			v := data.At(i)
			min = vector4.Min(min, v)
			max = vector4.Max(max, v)
			w.WriteVector4AsFloat32(v)
		}
	}

	if accessorComponentType == AccessorComponentType_UNSIGNED_BYTE {
		for i := 0; i < data.Len(); i++ {
			v := data.At(i)
			min = vector4.Min(min, v)
			max = vector4.Max(max, v)
			w.WriteVector4AsByte(v)
		}
	}

	minArr := []float64{min.X(), min.Y(), min.Z(), min.W()}
	maxArr := []float64{max.X(), max.Y(), max.Z(), max.W()}
	datasize := data.Len() * 4 * accessorComponentType.Size()

	w.accessors = append(w.accessors, Accessor{
		BufferView:    ptrI(len(w.bufferViews)),
		ComponentType: accessorComponentType,
		Type:          accessorType,
		Count:         data.Len(),
		Min:           minArr,
		Max:           maxArr,
	})

	w.bufferViews = append(w.bufferViews, BufferView{
		Buffer:     0,
		ByteOffset: w.bytesWritten,
		ByteLength: datasize,
		Target:     ARRAY_BUFFER,
	})

	w.bytesWritten += datasize
}

func (w *Writer) WriteVector3(accessorComponentType AccessorComponentType, data iter.ArrayIterator[vector3.Float64]) {
	accessorType := AccessorType_VEC3

	min := vector3.Fill(math.MaxFloat64)
	max := vector3.Fill(-math.MaxFloat64)

	if accessorComponentType == AccessorComponentType_FLOAT {
		for i := 0; i < data.Len(); i++ {
			v := data.At(i)
			min = vector3.Min(min, v)
			max = vector3.Max(max, v)
			w.WriteVector3AsFloat32(v)
		}
	}

	if accessorComponentType == AccessorComponentType_UNSIGNED_BYTE {
		for i := 0; i < data.Len(); i++ {
			v := data.At(i)
			min = vector3.Min(min, v)
			max = vector3.Max(max, v)
			w.WriteVector3AsByte(v)
		}
	}

	minArr := []float64{min.X(), min.Y(), min.Z()}
	maxArr := []float64{max.X(), max.Y(), max.Z()}
	datasize := data.Len() * 3 * accessorComponentType.Size()

	w.accessors = append(w.accessors, Accessor{
		BufferView:    ptrI(len(w.bufferViews)),
		ComponentType: accessorComponentType,
		Type:          accessorType,
		Count:         data.Len(),
		Min:           minArr,
		Max:           maxArr,
	})

	w.bufferViews = append(w.bufferViews, BufferView{
		Buffer:     0,
		ByteOffset: w.bytesWritten,
		ByteLength: datasize,
		Target:     ARRAY_BUFFER,
	})

	w.bytesWritten += datasize
}

func (w *Writer) WriteVector2(accessorComponentType AccessorComponentType, data iter.ArrayIterator[vector2.Float64]) {
	accessorType := AccessorType_VEC2

	min := vector2.Fill(math.MaxFloat64)
	max := vector2.Fill(-math.MaxFloat64)

	if accessorComponentType == AccessorComponentType_FLOAT {
		for i := 0; i < data.Len(); i++ {
			v := data.At(i)
			min = vector2.Min(min, v)
			max = vector2.Max(max, v)
			w.WriteVector2AsFloat32(v)
		}
	}

	if accessorComponentType == AccessorComponentType_UNSIGNED_BYTE {
		for i := 0; i < data.Len(); i++ {
			v := data.At(i)
			min = vector2.Min(min, v)
			max = vector2.Max(max, v)
			w.WriteVector2AsByte(v)
		}
	}

	minArr := []float64{min.X(), min.Y()}
	maxArr := []float64{max.X(), max.Y()}
	datasize := data.Len() * 2 * accessorComponentType.Size()

	w.accessors = append(w.accessors, Accessor{
		BufferView:    ptrI(len(w.bufferViews)),
		ComponentType: accessorComponentType,
		Type:          accessorType,
		Count:         data.Len(),
		Min:           minArr,
		Max:           maxArr,
	})

	w.bufferViews = append(w.bufferViews, BufferView{
		Buffer:     0,
		ByteOffset: w.bytesWritten,
		ByteLength: datasize,
		Target:     ARRAY_BUFFER,
	})

	w.bytesWritten += datasize
}

func (w *Writer) WriteIndices(indices iter.ArrayIterator[int], attributeSize int) {
	indiceSize := indices.Len()

	componentType := AccessorComponentType_UNSIGNED_INT

	if attributeSize > math.MaxUint16 {
		for i := 0; i < indices.Len(); i++ {
			w.bitW.UInt32(uint32(indices.At(i)))
		}
		indiceSize *= 4
	} else {
		for i := 0; i < indices.Len(); i++ {
			w.bitW.UInt16(uint16(indices.At(i)))
		}
		indiceSize *= 2
		componentType = AccessorComponentType_UNSIGNED_SHORT
	}

	w.accessors = append(w.accessors, Accessor{
		BufferView:    ptrI(len(w.bufferViews)),
		ComponentType: componentType,
		Type:          AccessorType_SCALAR,
		Count:         indices.Len(),
	})

	w.bufferViews = append(w.bufferViews, BufferView{
		Buffer:     0,
		ByteOffset: w.bytesWritten,
		ByteLength: indiceSize,
		Target:     ELEMENT_ARRAY_BUFFER,
	})

	w.bytesWritten += indiceSize
}

func (w *Writer) AddMaterial(mat PolyformMaterial) *int {
	var pbr *PbrMetallicRoughness

	pbr = &PbrMetallicRoughness{
		BaseColorFactor: &[4]float64{1, 1, 1, 1},
	}

	if mat.PbrMetallicRoughness != nil {
		polyPBR := mat.PbrMetallicRoughness

		pbr.MetallicFactor = polyPBR.MetallicFactor
		pbr.RoughnessFactor = polyPBR.RoughnessFactor

		if polyPBR.BaseColorFactor != nil {
			r, g, b, a := polyPBR.BaseColorFactor.RGBA()
			// polyPBR.BaseColorFactor.
			pbr.BaseColorFactor = &[4]float64{
				float64(r) / math.MaxUint16,
				float64(g) / math.MaxUint16,
				float64(b) / math.MaxUint16,
				float64(a) / math.MaxUint16,
			}
		}

		if polyPBR.BaseColorTexture != nil {
			pbr.BaseColorTexture = &TextureInfo{Index: len(w.textures)}

			w.textures = append(w.textures, Texture{
				Sampler: ptrI(len(w.samplers)),
				Source:  ptrI(len(w.images)),
			})

			w.images = append(w.images, Image{
				URI: polyPBR.BaseColorTexture.URI,
			})
			w.samplers = append(w.samplers, Sampler{})
		}

		if polyPBR.MetallicRoughnessTexture != nil {
			pbr.MetallicRoughnessTexture = &TextureInfo{Index: len(w.textures)}

			w.textures = append(w.textures, Texture{
				Sampler: ptrI(len(w.samplers)),
				Source:  ptrI(len(w.images)),
			})

			w.images = append(w.images, Image{
				URI: polyPBR.MetallicRoughnessTexture.URI,
			})
			w.samplers = append(w.samplers, Sampler{})
		}
	}

	w.materials = append(w.materials, Material{
		ChildOfRootProperty: ChildOfRootProperty{
			Name: mat.Name,
		},
		PbrMetallicRoughness: pbr,
	})

	return ptrI(len(w.materials) - 1)
}

func (w *Writer) AddSkin(skeleton animation.Skeleton) (*int, int) {
	skeletonNodes := flattenSkeletonToNodes(1, skeleton, w.buf)
	w.scene = append(w.scene, len(w.nodes))
	w.nodes = append(w.nodes, skeletonNodes...)

	jointIndices := make([]int, len(skeletonNodes))
	for i := 0; i < len(skeletonNodes); i++ {
		jointIndices[i] = i + 1 // +1 because we're offsetting from mesh node
	}

	w.accessors = append(w.accessors, Accessor{
		BufferView:    ptrI(len(w.bufferViews)),
		ComponentType: AccessorComponentType_FLOAT,
		Type:          AccessorType_MAT4,
		Count:         len(skeletonNodes),
		// Min:           []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		// Max:           []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	})

	inverseBindMAtrixLen := len(skeletonNodes) * 4 * 16

	w.bufferViews = append(w.bufferViews, BufferView{
		Buffer:     0,
		ByteOffset: w.bytesWritten,
		ByteLength: inverseBindMAtrixLen,
		// Target:     ARRAY_BUFFER,
	})
	w.bytesWritten += inverseBindMAtrixLen

	w.skins = []Skin{
		{
			Joints:              jointIndices,
			InverseBindMatrices: len(w.accessors) - 1,
		},
	}
	return ptrI(len(w.skins) - 1), w.scene[len(w.scene)-1]
}

func (w *Writer) AddAnimations(animations []animation.Sequence, skeleton animation.Skeleton, skeletonNode int) {
	for i, animation := range animations {

		min := vector3.New(math.MaxFloat64, math.MaxFloat64, math.MaxFloat64)
		max := vector3.New(-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64)

		for _, frame := range animation.Frames() {
			v := frame.Val()
			min = vector3.Min(min, v)
			max = vector3.Max(max, v)
			w.WriteVector3AsFloat32(v)
		}

		datasize := len(animation.Frames()) * 3 * 4

		animationDataBufferView := BufferView{
			Buffer:     0,
			ByteOffset: w.bytesWritten,
			ByteLength: datasize,
		}
		animationDataBufferViewIndex := len(w.bufferViews)

		animationDataAccessor := Accessor{
			BufferView:    ptrI(animationDataBufferViewIndex),
			ComponentType: AccessorComponentType_FLOAT,
			Type:          AccessorType_VEC3,
			Count:         len(animation.Frames()),
			Min:           []float64{min.X(), min.Y(), min.Z()},
			Max:           []float64{max.X(), max.Y(), max.Z()},
		}
		animationDataAccessorIndex := len(w.accessors)

		w.accessors = append(w.accessors, animationDataAccessor)
		w.bufferViews = append(w.bufferViews, animationDataBufferView)

		w.bytesWritten += datasize

		// Time Data ============================================================

		minTime := math.MaxFloat64
		maxTime := -math.MaxFloat64

		for _, frame := range animation.Frames() {
			minTime = math.Min(minTime, frame.Time())
			maxTime = math.Max(maxTime, frame.Time())
			w.bitW.Float32(float32(frame.Time()))
		}

		datasize = len(animation.Frames()) * 4

		timeBufferView := BufferView{
			Buffer:     0,
			ByteOffset: w.bytesWritten,
			ByteLength: datasize,
		}
		timeBufferViewIndex := len(w.bufferViews)

		timeAccessor := Accessor{
			BufferView:    ptrI(timeBufferViewIndex),
			ComponentType: AccessorComponentType_FLOAT,
			Type:          AccessorType_SCALAR,
			Count:         len(animation.Frames()),
			Min:           []float64{minTime},
			Max:           []float64{maxTime},
		}

		timeAccessorIndex := len(w.accessors)
		w.accessors = append(w.accessors, timeAccessor)
		w.bufferViews = append(w.bufferViews, timeBufferView)

		w.bytesWritten += datasize

		w.animations = append(w.animations, Animation{
			Samplers: []AnimationSampler{
				{
					Interpolation: AnimationSamplerInterpolation_LINEAR,
					Input:         timeAccessorIndex,
					Output:        animationDataAccessorIndex,
				},
			},
			Channels: []AnimationChannel{
				{
					Target: AnimationChannelTarget{
						Path: AnimationChannelTargetPath_TRANSLATION,
						Node: skeleton.Lookup(animation.Joint()) + skeletonNode,
					},
					Sampler: i,
				},
			},
		})
	}
}

func (w *Writer) AddMesh(model PolyformModel) {
	primitiveAttributes := make(map[string]int)

	for _, val := range model.Mesh.Float4Attributes() {
		primitiveAttributes[polyformToGLTFAttribute(val)] = len(w.accessors)
		w.WriteVector4(attributeType(val), model.Mesh.Float4Attribute(val))
	}

	for _, val := range model.Mesh.Float3Attributes() {
		primitiveAttributes[polyformToGLTFAttribute(val)] = len(w.accessors)
		w.WriteVector3(attributeType(val), model.Mesh.Float3Attribute(val))
	}

	for _, val := range model.Mesh.Float2Attributes() {
		primitiveAttributes[polyformToGLTFAttribute(val)] = len(w.accessors)
		w.WriteVector2(attributeType(val), model.Mesh.Float2Attribute(val))
	}

	indiceIndex := len(w.accessors)
	w.WriteIndices(model.Mesh.Indices(), model.Mesh.AttributeLength())

	meshIndex := len(w.meshes)

	nodeIndex := len(w.nodes)
	w.scene = append(w.scene, nodeIndex)
	w.nodes = append(w.nodes, Node{
		Mesh: &meshIndex,
	})

	var skinNode = 0
	// Skins nodes in the scene always have to come after the mesh itself, the
	// mesh node must act as the root node for the skinning to work
	if model.Skeleton != nil {
		var skinIndex *int
		skinIndex, skinNode = w.AddSkin(*model.Skeleton)
		w.nodes[nodeIndex].Skin = skinIndex
	}

	if len(model.Animations) > 0 {
		w.AddAnimations(model.Animations, *model.Skeleton, skinNode)
	}

	var materialIndex *int
	if model.Material != nil {
		materialIndex = w.AddMaterial(*model.Material)
	}

	w.meshes = append(w.meshes, Mesh{
		ChildOfRootProperty: ChildOfRootProperty{
			Name: model.Name,
		},
		Primitives: []Primitive{
			{
				Indices:    &indiceIndex,
				Attributes: primitiveAttributes,
				Material:   materialIndex,
			},
		},
	})
}

type BufferEmbeddingStrategy int

const (
	BufferEmbeddingStrategy_Base64Encode BufferEmbeddingStrategy = iota
	BufferEmbeddingStrategy_GLB
)

func (w Writer) ToGLTF(embeddingStrategy BufferEmbeddingStrategy) Gltf {
	buffer := Buffer{
		ByteLength: w.bytesWritten,
	}

	if embeddingStrategy == BufferEmbeddingStrategy_Base64Encode {
		buffer.URI = "data:application/octet-stream;base64," + base64.StdEncoding.EncodeToString(w.buf.Bytes())
	}

	return Gltf{
		Asset:       defaultAsset(),
		Buffers:     []Buffer{buffer},
		BufferViews: w.bufferViews,
		Accessors:   w.accessors,

		// Skins: skins,
		Scene: 0,
		Scenes: []Scene{
			{
				Nodes: w.scene,
			},
		},

		Skins:      w.skins,
		Animations: w.animations,

		Nodes:     w.nodes,
		Meshes:    w.meshes,
		Materials: w.materials,
		Textures:  w.textures,
		Images:    w.images,
		Samplers:  w.samplers,
	}
}

func (w Writer) WriteGLB(out io.Writer) error {
	jsonBytes, err := json.Marshal(w.ToGLTF(BufferEmbeddingStrategy_GLB))
	if err != nil {
		return err
	}
	jsonByteLen := len(jsonBytes)
	jsonPadding := (4 - (jsonByteLen % 4)) % 4
	jsonByteLen += jsonPadding

	binBytes := w.buf.Bytes()
	binByteLen := len(binBytes)
	binPadding := (4 - (binByteLen % 4)) % 4
	binByteLen += binPadding

	bitWriter := bitlib.NewWriter(out, binary.LittleEndian)

	// https://registry.khronos.org/glTF/specs/2.0/glTF-2.0.pdf
	// magic MUST be equal to equal 0x46546C67. It is ASCII string glTF and can
	// be used to identify data as Binary glTF
	bitWriter.UInt32(0x46546C67)

	// GLB version
	bitWriter.UInt32(2)

	// Length of entire document
	totalLen := jsonByteLen + binByteLen + 12 + 8
	if binByteLen > 0 {
		totalLen += 8
	}
	bitWriter.UInt32(uint32(totalLen))

	// JSON CHUNK =============================================================

	// Chunk Length
	bitWriter.UInt32(uint32(jsonByteLen))

	// JSON Chunk Identifier
	bitWriter.UInt32(0x4E4F534A)

	// JSON data
	bitWriter.ByteArray(jsonBytes)

	// Padding to make it align to a 4 byte boundary
	for i := 0; i < jsonPadding; i++ {
		bitWriter.Byte(0x20)
	}

	// OPTIONAL BIN CHUNK =====================================================

	// Don't write anything if the bin data is empty
	if binByteLen == 0 {
		return bitWriter.Error()
	}

	// Chunk Length
	bitWriter.UInt32(uint32(binByteLen))

	// BIN Chunk Identifier
	bitWriter.UInt32(0x004E4942)

	// Bin data
	bitWriter.ByteArray(binBytes)

	// Padding to make it align to a 4 byte boundary
	for i := 0; i < binPadding; i++ {
		bitWriter.Byte(0x00)
	}

	return bitWriter.Error()
}
