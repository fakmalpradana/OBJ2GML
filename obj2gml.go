package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CityGML structures based on the provided schema
type CityGML struct {
	XMLName   xml.Name `xml:"CityModel"`
	GML       string   `xml:"xmlns:gml,attr"`
	CityGML   string   `xml:"xmlns:citygml,attr"`
	Bldg      string   `xml:"xmlns:bldg,attr"`
	XLink     string   `xml:"xmlns:xlink,attr"`
	XSI       string   `xml:"xmlns:xsi,attr"`
	SchemaLoc string   `xml:"xsi:schemaLocation,attr"`

	BoundedBy        BoundedBy          `xml:"boundedBy"`
	CityObjectMember []CityObjectMember `xml:"cityObjectMember"`
}

type BoundedBy struct {
	Envelope Envelope `xml:"Envelope"`
}

type Envelope struct {
	SrsName     string `xml:"srsName,attr"`
	LowerCorner string `xml:"lowerCorner"`
	UpperCorner string `xml:"upperCorner"`
}

type CityObjectMember struct {
	Building Building `xml:"bldg:Building"`
}

type Building struct {
	ID                 string    `xml:"gml:id,attr"`
	Function           string    `xml:"bldg:function,omitempty"`
	YearOfConstruction string    `xml:"bldg:yearOfConstruction,omitempty"`
	RoofType           string    `xml:"bldg:roofType,omitempty"`
	MeasuredHeight     string    `xml:"bldg:measuredHeight,omitempty"`
	Lod1Solid          Lod1Solid `xml:"bldg:lod1Solid"`
}

type Lod1Solid struct {
	Solid Solid `xml:"gml:Solid"`
}

type Solid struct {
	ID       string   `xml:"gml:id,attr"`
	Exterior Exterior `xml:"gml:exterior"`
}

type Exterior struct {
	CompositeSurface CompositeSurface `xml:"gml:CompositeSurface"`
}

type CompositeSurface struct {
	SurfaceMember []SurfaceMember `xml:"gml:surfaceMember"`
}

type SurfaceMember struct {
	Polygon Polygon `xml:"gml:Polygon"`
}

type Polygon struct {
	ID       string          `xml:"gml:id,attr"`
	Exterior PolygonExterior `xml:"gml:exterior"`
}

type PolygonExterior struct {
	LinearRing LinearRing `xml:"gml:LinearRing"`
}

type LinearRing struct {
	PosList string `xml:"gml:posList"`
}

// OBJ file structures
type OBJVertex struct {
	X, Y, Z float64
}

type OBJFace []int

// Main function
func main() {
	// Parse command-line arguments
	inputDir := flag.String("input", "", "Directory containing OBJ files")
	outputDir := flag.String("output", "", "Directory for output CityGML files")
	epsg := flag.String("epsg", "", "EPSG Code")
	flag.Parse()

	if *inputDir == "" || *outputDir == "" {
		fmt.Println("Usage: obj2citygml -input <input_directory> -output <output_directory>")
		return
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Find all OBJ files in the input directory
	objFiles, err := filepath.Glob(filepath.Join(*inputDir, "*.obj"))
	if err != nil {
		fmt.Printf("Error finding OBJ files: %v\n", err)
		return
	}

	fmt.Printf("Found %d OBJ files to process\n", len(objFiles))
	successCount := 0
	errorFiles := []string{}

	// Process each OBJ file
	for _, objFile := range objFiles {
		baseFileName := filepath.Base(objFile)
		fileNameWithoutExt := strings.TrimSuffix(baseFileName, filepath.Ext(baseFileName))
		outputFile := filepath.Join(*outputDir, fileNameWithoutExt+".gml")

		err := convertOBJToCityGML(*epsg, objFile, outputFile, fileNameWithoutExt)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", baseFileName, err)
			errorFiles = append(errorFiles, baseFileName)
		} else {
			successCount++
		}
	}

	// Print summary
	fmt.Printf("Successfully converted %d from %d OBJ files\n", successCount, len(objFiles))
	if len(errorFiles) > 0 {
		fmt.Printf("Failed to convert %d files: %v\n", len(errorFiles), errorFiles)
	}
}

// Convert OBJ file to CityGML
func convertOBJToCityGML(epsg, inputPath, outputPath, buildingID string) error {
	// Read and parse OBJ file
	vertices, faces, err := parseOBJFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to parse OBJ file: %v", err)
	}

	// Calculate bounding box
	minX, minY, minZ := float64(999999), float64(999999), float64(999999)
	maxX, maxY, maxZ := float64(-999999), float64(-999999), float64(-999999)

	for _, v := range vertices {
		if v.X < minX {
			minX = v.X
		}
		if v.Y < minY {
			minY = v.Y
		}
		if v.Z < minZ {
			minZ = v.Z
		}
		if v.X > maxX {
			maxX = v.X
		}
		if v.Y > maxY {
			maxY = v.Y
		}
		if v.Z > maxZ {
			maxZ = v.Z
		}
	}

	// Calculate height
	height := maxZ - minZ

	// Create CityGML structure
	cityGML := CityGML{
		GML:       "http://www.opengis.net/gml",
		CityGML:   "http://www.opengis.net/citygml/2.0",
		Bldg:      "http://www.opengis.net/citygml/building/2.0",
		XLink:     "http://www.w3.org/1999/xlink",
		XSI:       "http://www.w3.org/2001/XMLSchema-instance",
		SchemaLoc: "http://www.opengis.net/citygml/building/2.0 http://schemas.opengis.net/citygml/building/2.0/building.xsd",
		BoundedBy: BoundedBy{
			Envelope: Envelope{
				// SrsName:     "urn:ogc:def:crs:EPSG::32748",
				SrsName:     fmt.Sprintf("%s %s", "urn:ogc:def:crs:EPSG::", epsg),
				LowerCorner: fmt.Sprintf("%f %f %f", minX, minY, minZ),
				UpperCorner: fmt.Sprintf("%f %f %f", maxX, maxY, maxZ),
			},
		},
	}

	// Create building
	building := Building{
		ID:                 buildingID,
		YearOfConstruction: strconv.Itoa(time.Now().Year()),
		MeasuredHeight:     fmt.Sprintf("%f", height),
		RoofType:           "1000", // Default roof type
		Lod1Solid: Lod1Solid{
			Solid: Solid{
				ID: fmt.Sprintf("%s-solid", buildingID),
				Exterior: Exterior{
					CompositeSurface: CompositeSurface{},
				},
			},
		},
	}

	// Add faces to the building
	for i, face := range faces {
		polygonID := fmt.Sprintf("%s-polygon-%d", buildingID, i)

		// Create posList from face vertices
		var posListBuilder strings.Builder
		for _, vIdx := range face {
			if vIdx > 0 && vIdx <= len(vertices) {
				v := vertices[vIdx-1]
				posListBuilder.WriteString(fmt.Sprintf("%f %f %f ", v.X, v.Y, v.Z))
			}
		}

		// Add first vertex again to close the polygon
		if len(face) > 0 {
			vIdx := face[0]
			if vIdx > 0 && vIdx <= len(vertices) {
				v := vertices[vIdx-1]
				posListBuilder.WriteString(fmt.Sprintf("%f %f %f", v.X, v.Y, v.Z))
			}
		}

		surfaceMember := SurfaceMember{
			Polygon: Polygon{
				ID: polygonID,
				Exterior: PolygonExterior{
					LinearRing: LinearRing{
						PosList: posListBuilder.String(),
					},
				},
			},
		}

		building.Lod1Solid.Solid.Exterior.CompositeSurface.SurfaceMember = append(
			building.Lod1Solid.Solid.Exterior.CompositeSurface.SurfaceMember, surfaceMember)
	}

	// Add building to city model
	cityObjectMember := CityObjectMember{
		Building: building,
	}
	cityGML.CityObjectMember = append(cityGML.CityObjectMember, cityObjectMember)

	// Generate XML
	output, err := xml.MarshalIndent(cityGML, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to generate XML: %v", err)
	}

	// Add XML header
	xmlData := []byte(xml.Header + string(output))

	// Write to file
	if err := ioutil.WriteFile(outputPath, xmlData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	return nil
}

// Parse OBJ file
func parseOBJFile(filePath string) ([]OBJVertex, []OBJFace, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var vertices []OBJVertex
	var faces []OBJFace

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "v":
			// Parse vertex
			if len(fields) < 4 {
				continue
			}

			x, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				continue
			}

			y, err := strconv.ParseFloat(fields[2], 64)
			if err != nil {
				continue
			}

			z, err := strconv.ParseFloat(fields[3], 64)
			if err != nil {
				continue
			}

			vertices = append(vertices, OBJVertex{X: x, Y: y, Z: z})

		case "f":
			// Parse face
			if len(fields) < 4 {
				continue
			}

			var face OBJFace
			for i := 1; i < len(fields); i++ {
				// Handle different face formats (v, v/vt, v/vt/vn)
				vertexStr := strings.Split(fields[i], "/")[0]
				idx, err := strconv.Atoi(vertexStr)
				if err != nil {
					continue
				}
				face = append(face, idx)
			}

			if len(face) >= 3 {
				faces = append(faces, face)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return vertices, faces, nil
}
