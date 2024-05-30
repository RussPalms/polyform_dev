package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/formats/potree"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/urfave/cli/v2"
)

type builldModelJob struct {
	ByteSize   uint64
	ByteOffset uint64
	NumPoints  uint32
	PlyStart   int64
}

func buildModelWorker(
	ctx *cli.Context,
	jobs <-chan *builldModelJob,
	metadata *potree.Metadata,
	largestPointCount int,
	plyFilename string,
	plyHeaderOffset int,
	out chan<- int,
) {
	octreeFile, err := openOctreeFile(ctx)
	if err != nil {
		panic(err)
	}
	defer octreeFile.Close()

	pointsProcessed := 0
	potreeBuf := make([]byte, largestPointCount*metadata.BytesPerPoint())
	positionBuf := make([]vector3.Float64, largestPointCount)
	colorBuf := make([]vector3.Float64, largestPointCount)
	plyBuf := make([]byte, largestPointCount*27)

	plyFile, err := os.OpenFile(plyFilename, os.O_WRONLY, 0)
	if err != nil {
		panic(err)
	}
	defer plyFile.Close()

	for job := range jobs {
		count := int(job.NumPoints)
		if count == 0 {
			continue
		}

		_, err := octreeFile.Seek(int64(job.ByteOffset), 0)
		if err != nil {
			panic(err)
		}

		_, err = io.ReadFull(octreeFile, potreeBuf[:job.ByteSize])
		if err != nil {
			panic(err)
		}

		potree.LoadNodePositionDataIntoArray(metadata, potreeBuf[:job.ByteSize], positionBuf[:job.NumPoints])
		potree.LoadNodeColorDataIntoArray(metadata, potreeBuf[:job.ByteSize], colorBuf[:job.NumPoints])
		endien := binary.LittleEndian

		offset := 0
		for i := 0; i < count; i++ {
			p := positionBuf[i]
			endien.PutUint64(plyBuf[offset:], math.Float64bits(p.X()))
			endien.PutUint64(plyBuf[offset+8:], math.Float64bits(p.Y()))
			endien.PutUint64(plyBuf[offset+16:], math.Float64bits(p.Z()))

			c := colorBuf[i].Scale(255)
			plyBuf[offset+24] = byte(c.X())
			plyBuf[offset+25] = byte(c.Y())
			plyBuf[offset+26] = byte(c.Z())

			offset += 27
		}

		_, err = plyFile.Seek(int64(plyHeaderOffset)+(job.PlyStart*27), 0)
		if err != nil {
			panic(err)
		}
		_, err = plyFile.Write(plyBuf[:count*27])
		if err != nil {
			panic(err)
		}

		pointsProcessed += count
	}

	out <- pointsProcessed
}

func buildModelWithChildren(ctx *cli.Context, root *potree.OctreeNode, metadata *potree.Metadata) error {
	plyFilename := ctx.String("out")
	plyFile, err := os.Create(plyFilename)
	if err != nil {
		return err
	}
	defer plyFile.Close()

	largestPointCount := 0
	root.Walk(func(o *potree.OctreeNode) {
		if o.NumPoints > uint32(largestPointCount) {
			largestPointCount = int(o.NumPoints)
		}
	})

	header := ply.Header{
		Format: ply.BinaryLittleEndian,
		Elements: []ply.Element{
			{
				Name:  ply.VertexElementName,
				Count: int64(root.PointCount()),
				Properties: []ply.Property{
					&ply.ScalarProperty{PropertyName: "x", Type: ply.Double},
					&ply.ScalarProperty{PropertyName: "y", Type: ply.Double},
					&ply.ScalarProperty{PropertyName: "z", Type: ply.Double},
					&ply.ScalarProperty{PropertyName: "red", Type: ply.UChar},
					&ply.ScalarProperty{PropertyName: "green", Type: ply.UChar},
					&ply.ScalarProperty{PropertyName: "blue", Type: ply.UChar},
				},
			},
		},
	}

	buf := &bytes.Buffer{}
	err = header.Write(buf)
	if err != nil {
		return err
	}
	headerOffset := buf.Len()
	_, err = plyFile.Write(buf.Bytes())
	if err != nil {
		return err
	}

	err = plyFile.Truncate(int64(headerOffset) + int64(27*root.PointCount()))
	if err != nil {
		return err
	}

	workerCount := runtime.NumCPU() * 2
	jobs := make(chan *builldModelJob, workerCount)
	meshes := make(chan int, workerCount)

	for i := 0; i < workerCount; i++ {
		go buildModelWorker(ctx, jobs, metadata, largestPointCount, plyFilename, headerOffset, meshes)
	}

	var plyStart int64
	root.Walk(func(o *potree.OctreeNode) {
		jobs <- &builldModelJob{
			ByteSize:   o.ByteSize,
			ByteOffset: o.ByteOffset,
			NumPoints:  o.NumPoints,
			PlyStart:   plyStart,
		}
		plyStart += int64(o.NumPoints)
	})
	close(jobs)

	for i := 0; i < workerCount; i++ {
		<-meshes
	}

	return nil
}

func buildModel(ctx *cli.Context, node *potree.OctreeNode, metadata *potree.Metadata) (*modeling.Mesh, error) {
	octreeFile, err := openOctreeFile(ctx)
	if err != nil {
		return nil, err
	}
	defer octreeFile.Close()

	_, err = octreeFile.Seek(int64(node.ByteOffset), 0)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, node.ByteSize)
	_, err = io.ReadFull(octreeFile, buf)
	if err != nil {
		return nil, err
	}

	mesh := potree.LoadNode(node, metadata, buf)
	return &mesh, nil
}

func findNode(name string, node *potree.OctreeNode) (*potree.OctreeNode, error) {
	if name == node.Name {
		return node, nil
	}

	for _, c := range node.Children {
		if strings.Index(name, c.Name) == 0 {
			return findNode(name, c)
		}
		log.Println(c.Name)
	}

	return nil, fmt.Errorf("%s can't find child node with name %s", node.Name, name)
}

var ExtractPointcloudCommand = &cli.Command{
	Name: "extract-pointcloud",
	Flags: []cli.Flag{
		metadataFlag,
		hierarchyFlag,
		octreeFlag,
		&cli.StringFlag{
			Name:  "node",
			Value: "r",
			Usage: "Name of node to extract point data from",
		},
		&cli.BoolFlag{
			Name:  "include-children",
			Value: false,
			Usage: "Whether or not to include children data",
		},
		&cli.StringFlag{
			Name:  "out",
			Value: "out.ply",
			Usage: "Name of ply file to write pointcloud data too",
		},
	},
	Action: func(ctx *cli.Context) error {
		metadata, hierarchy, err := loadHierarchy(ctx)
		if err != nil {
			return err
		}

		startNode, err := findNode(ctx.String("node"), hierarchy)
		if err != nil {
			return err
		}

		if ctx.Bool("include-children") {
			start := time.Now()
			err = buildModelWithChildren(ctx, startNode, metadata)
			log.Printf("PLY written in %s", time.Since(start))
			return err
		}

		mesh, err := buildModel(ctx, startNode, metadata)
		if err != nil {
			return err
		}
		fmt.Fprintf(ctx.App.Writer, "Writing pointcloud with %d points to %s", mesh.Indices().Len(), ctx.String("out"))
		return ply.SaveBinary(ctx.String("out"), *mesh)
	},
}
