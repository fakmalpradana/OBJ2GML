package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// XML namespaces and schema declarations
const (
	xmlHeader = `<?xml version="1.0" encoding="UTF-8"?>
<!-- OBJ to CityGML LOD2 Converter Output -->
<!-- copyrights 2025 © Fairuz Akmal Pradana | fakmalpradana@gmail.com  -->
`
)

// CityGML structures based on the provided schema
type CityModel struct {
	XMLName        xml.Name `xml:"core:CityModel"`
	GML            string   `xml:"xmlns:gml,attr"`
	Core           string   `xml:"xmlns:core,attr"`
	Bldg           string   `xml:"xmlns:bldg,attr"`
	App            string   `xml:"xmlns:app,attr"`
	Gen            string   `xml:"xmlns:gen,attr"`
	Grp            string   `xml:"xmlns:grp,attr"`
	XLink          string   `xml:"xmlns:xlink,attr"`
	XSI            string   `xml:"xmlns:xsi,attr"`
	SchemaLocation string   `xml:"xsi:schemaLocation,attr"`

	BoundedBy        BoundedBy          `xml:"gml:boundedBy"`
	CityObjectMember []CityObjectMember `xml:"core:cityObjectMember"`
}

type BoundedBy struct {
	Envelope Envelope `xml:"gml:Envelope"`
}

type Envelope struct {
	SrsName      string `xml:"srsName,attr"`
	SrsDimension string `xml:"srsDimension,attr,omitempty"`
	LowerCorner  string `xml:"gml:lowerCorner"`
	UpperCorner  string `xml:"gml:upperCorner"`
}

type CityObjectMember struct {
	Building Building `xml:"bldg:Building"`
}

type Building struct {
	ID                 string                    `xml:"gml:id,attr"`
	Function           string                    `xml:"bldg:function,omitempty"`
	YearOfConstruction string                    `xml:"bldg:yearOfConstruction,omitempty"`
	RoofType           string                    `xml:"bldg:roofType,omitempty"`
	MeasuredHeight     MeasuredHeight            `xml:"bldg:measuredHeight,omitempty"`
	Lod2Solid          *Lod2Solid                `xml:"bldg:lod2Solid,omitempty"`
	BoundedBy          []BoundarySurfaceProperty `xml:"bldg:boundedBy,omitempty"`
}

type MeasuredHeight struct {
	Value string `xml:",chardata"`
	UOM   string `xml:"uom,attr"`
}

type Lod2Solid struct {
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

// New structures for LOD2 boundary surfaces
type BoundarySurfaceProperty struct {
	RoofSurface   *RoofSurface   `xml:"bldg:RoofSurface,omitempty"`
	WallSurface   *WallSurface   `xml:"bldg:WallSurface,omitempty"`
	GroundSurface *GroundSurface `xml:"bldg:GroundSurface,omitempty"`
}

type RoofSurface struct {
	ID               string               `xml:"gml:id,attr"`
	Lod2MultiSurface MultiSurfaceProperty `xml:"bldg:lod2MultiSurface"`
}

type WallSurface struct {
	ID               string               `xml:"gml:id,attr"`
	Lod2MultiSurface MultiSurfaceProperty `xml:"bldg:lod2MultiSurface"`
}

type GroundSurface struct {
	ID               string               `xml:"gml:id,attr"`
	Lod2MultiSurface MultiSurfaceProperty `xml:"bldg:lod2MultiSurface"`
}

type MultiSurfaceProperty struct {
	MultiSurface MultiSurface `xml:"gml:MultiSurface"`
}

type MultiSurface struct {
	ID            string          `xml:"gml:id,attr,omitempty"`
	SurfaceMember []SurfaceMember `xml:"gml:surfaceMember"`
}

// OBJ file structures
type OBJVertex struct {
	X, Y, Z float64
}

type OBJFace struct {
	VertexIndices []int
	Material      string
}

// MTL material structure
type MTLMaterial struct {
	Name string
	Kd   [3]float64 // Diffuse color
}

// Vector3D represents a 3D vector
type Vector3D struct {
	X, Y, Z float64
}

// Main function
func main() {
	// Parse command-line arguments
	inputDir := flag.String("input", "", "Directory containing OBJ files")
	outputDir := flag.String("output", "", "Directory for output CityGML files")
	epsgCode := flag.String("epsg", "32748", "EPSG code for the coordinate reference system")
	flag.Parse()

	if *inputDir == "" || *outputDir == "" {
		fmt.Println("Usage: obj2citygml -input <input_directory> -output <output_directory> [-epsg <epsg_code>]")
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

		err := convertOBJToCityGML(objFile, outputFile, fileNameWithoutExt, *epsgCode)
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

// Parse MTL file to extract materials
func parseMTLFile(filePath string) (map[string]MTLMaterial, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	materials := make(map[string]MTLMaterial)
	var currentMaterial string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "newmtl":
			if len(fields) > 1 {
				currentMaterial = fields[1]
				materials[currentMaterial] = MTLMaterial{Name: currentMaterial}
			}
		case "Kd":
			if len(fields) > 3 && currentMaterial != "" {
				r, _ := strconv.ParseFloat(fields[1], 64)
				g, _ := strconv.ParseFloat(fields[2], 64)
				b, _ := strconv.ParseFloat(fields[3], 64)
				mat := materials[currentMaterial]
				mat.Kd = [3]float64{r, g, b}
				materials[currentMaterial] = mat
			}
		}
	}

	return materials, scanner.Err()
}

// Enhanced OBJ file parser that captures material assignments
func parseOBJFile(filePath string) ([]OBJVertex, []OBJFace, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, "", err
	}
	defer file.Close()

	var vertices []OBJVertex
	var faces []OBJFace
	var mtlLib string
	currentMaterial := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "v":
			if len(fields) >= 4 {
				x, _ := strconv.ParseFloat(fields[1], 64)
				y, _ := strconv.ParseFloat(fields[2], 64)
				z, _ := strconv.ParseFloat(fields[3], 64)
				vertices = append(vertices, OBJVertex{x, y, z})
			}
		case "mtllib":
			if len(fields) > 1 {
				mtlLib = fields[1]
			}
		case "usemtl":
			if len(fields) > 1 {
				currentMaterial = fields[1]
			}
		case "f":
			if len(fields) >= 4 {
				var indices []int
				for _, f := range fields[1:] {
					parts := strings.Split(f, "/")
					index, _ := strconv.Atoi(parts[0])
					indices = append(indices, index-1) // OBJ indices are 1-based
				}
				faces = append(faces, OBJFace{indices, currentMaterial})
			}
		}
	}

	return vertices, faces, mtlLib, scanner.Err()
}

// Determine if a face is a roof, wall, or ground surface based on its normal and material
func classifySurface(face OBJFace, vertices []OBJVertex, material string) string {
	if strings.Contains(material, "Roof") {
		return "Roof"
	}
	if strings.Contains(material, "Wall") {
		return "Wall"
	}
	if strings.Contains(material, "Ground") {
		return "Ground"
	}

	// If material name doesn't give us a clue, use the face normal
	// Calculate face normal
	if len(face.VertexIndices) >= 3 {
		v1 := vertices[face.VertexIndices[0]]
		v2 := vertices[face.VertexIndices[1]]
		v3 := vertices[face.VertexIndices[2]]

		// Calculate two edges
		edge1 := Vector3D{v2.X - v1.X, v2.Y - v1.Y, v2.Z - v1.Z}
		edge2 := Vector3D{v3.X - v1.X, v3.Y - v1.Y, v3.Z - v1.Z}

		// Calculate cross product to get normal
		normal := Vector3D{
			edge1.Y*edge2.Z - edge1.Z*edge2.Y,
			edge1.Z*edge2.X - edge1.X*edge2.Z,
			edge1.X*edge2.Y - edge1.Y*edge2.X,
		}

		// Normalize
		length := math.Sqrt(normal.X*normal.X + normal.Y*normal.Y + normal.Z*normal.Z)
		if length > 0 {
			normal.X /= length
			normal.Y /= length
			normal.Z /= length
		}

		// Check if normal is pointing upward (roof), horizontal (wall), or downward (ground)
		if normal.Z > 0.7 {
			return "Roof"
		} else if normal.Z < -0.7 {
			return "Ground"
		} else {
			return "Wall"
		}
	}

	// Default to Wall if we can't determine
	return "Wall"
}

// Convert OBJ file to CityGML
func convertOBJToCityGML(objFile, outputFile, buildingID, epsgCode string) error {
	// Parse OBJ file
	vertices, faces, mtlLib, err := parseOBJFile(objFile)
	if err != nil {
		return fmt.Errorf("error parsing OBJ file: %v", err)
	}

	// Parse MTL file if available
	var materials map[string]MTLMaterial
	if mtlLib != "" {
		mtlFile := filepath.Join(filepath.Dir(objFile), mtlLib)
		materials, err = parseMTLFile(mtlFile)
		if err != nil {
			fmt.Printf("Warning: Could not parse MTL file: %v\n", err)
		}
	}

	// Create CityGML model
	model := CreateCityGMLModel(vertices, faces, materials, buildingID, epsgCode)

	// Write to file
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer file.Close()

	// Write XML header
	file.WriteString(xmlHeader)

	// Marshal and write the model
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	if err := encoder.Encode(model); err != nil {
		return fmt.Errorf("error encoding CityGML: %v", err)
	}

	return nil
}

// Create CityGML model from OBJ data
func CreateCityGMLModel(vertices []OBJVertex, faces []OBJFace, materials map[string]MTLMaterial, buildingID, epsgCode string) CityModel {
	// Calculate bounding box
	minX, minY, minZ := math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
	maxX, maxY, maxZ := -math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64

	for _, v := range vertices {
		minX = math.Min(minX, v.X)
		minY = math.Min(minY, v.Y)
		minZ = math.Min(minZ, v.Z)
		maxX = math.Max(maxX, v.X)
		maxY = math.Max(maxY, v.Y)
		maxZ = math.Max(maxZ, v.Z)
	}

	// Group faces by their surface type
	roofFaces := []OBJFace{}
	wallFaces := []OBJFace{}
	groundFaces := []OBJFace{}

	for _, face := range faces {
		surfaceType := classifySurface(face, vertices, face.Material)
		switch surfaceType {
		case "Roof":
			roofFaces = append(roofFaces, face)
		case "Wall":
			wallFaces = append(wallFaces, face)
		case "Ground":
			groundFaces = append(groundFaces, face)
		}
	}

	// Create CityGML model
	model := CityModel{
		GML:            "http://www.opengis.net/gml",
		Core:           "http://www.opengis.net/citygml/2.0",
		Bldg:           "http://www.opengis.net/citygml/building/2.0",
		App:            "http://www.opengis.net/citygml/appearance/2.0",
		Gen:            "http://www.opengis.net/citygml/generics/2.0",
		Grp:            "http://www.opengis.net/citygml/cityobjectgroup/2.0",
		XLink:          "http://www.w3.org/1999/xlink",
		XSI:            "http://www.w3.org/2001/XMLSchema-instance",
		SchemaLocation: "http://www.opengis.net/citygml/building/2.0 http://schemas.opengis.net/citygml/building/2.0/building.xsd",

		BoundedBy: BoundedBy{
			Envelope: Envelope{
				SrsName:      fmt.Sprintf("urn:ogc:def:crs:EPSG::%s", epsgCode),
				SrsDimension: "3",
				LowerCorner:  fmt.Sprintf("%f %f %f", minX, minY, minZ),
				UpperCorner:  fmt.Sprintf("%f %f %f", maxX, maxY, maxZ),
			},
		},
	}

	// Create building
	building := Building{
		ID:             fmt.Sprintf("Building_%s", buildingID),
		MeasuredHeight: MeasuredHeight{Value: fmt.Sprintf("%f", maxZ-minZ), UOM: "m"},
	}

	// Create boundary surfaces
	boundedBy := []BoundarySurfaceProperty{}

	// Create roof surface
	// Create roof surface
	if len(roofFaces) > 0 {
		roofSurface := createRoofSurface(buildingID, vertices, roofFaces)
		boundedBy = append(boundedBy, BoundarySurfaceProperty{RoofSurface: &roofSurface})
	}

	// Create wall surface
	if len(wallFaces) > 0 {
		wallSurface := createWallSurface(buildingID, vertices, wallFaces)
		boundedBy = append(boundedBy, BoundarySurfaceProperty{WallSurface: &wallSurface})
	}

	// Create ground surface
	if len(groundFaces) > 0 {
		groundSurface := createGroundSurface(buildingID, vertices, groundFaces)
		boundedBy = append(boundedBy, BoundarySurfaceProperty{GroundSurface: &groundSurface})
	}

	// Add boundary surfaces to building
	building.BoundedBy = boundedBy

	// Create solid geometry for LOD2
	building.Lod2Solid = createLod2Solid(buildingID, vertices, faces)

	// Add building to city model
	model.CityObjectMember = []CityObjectMember{{Building: building}}

	return model
}

// Create LOD2 Solid geometry
func createLod2Solid(buildingID string, vertices []OBJVertex, faces []OBJFace) *Lod2Solid {
	surfaceMembers := []SurfaceMember{}

	for i, face := range faces {
		// Create polygon
		polygon := Polygon{
			ID: fmt.Sprintf("Solid_%s_p%d", buildingID, i),
			Exterior: PolygonExterior{
				LinearRing: LinearRing{
					PosList: createPosListFromFace(face, vertices),
				},
			},
		}

		// Add to surface members
		surfaceMembers = append(surfaceMembers, SurfaceMember{Polygon: polygon})
	}

	return &Lod2Solid{
		Solid: Solid{
			ID: fmt.Sprintf("Solid_%s", buildingID),
			Exterior: Exterior{
				CompositeSurface: CompositeSurface{
					SurfaceMember: surfaceMembers,
				},
			},
		},
	}
}

// Create a boundary surface of the specified type
// Replace createBoundarySurface with three type-specific functions
func createRoofSurface(buildingID string, vertices []OBJVertex, faces []OBJFace) RoofSurface {
	surfaceMembers := createSurfaceMembers("RoofSurface", buildingID, vertices, faces)

	multiSurface := MultiSurfaceProperty{
		MultiSurface: MultiSurface{
			ID:            fmt.Sprintf("RoofSurface_%s_ms", buildingID),
			SurfaceMember: surfaceMembers,
		},
	}

	return RoofSurface{
		ID:               fmt.Sprintf("RoofSurface_%s", buildingID),
		Lod2MultiSurface: multiSurface,
	}
}

func createWallSurface(buildingID string, vertices []OBJVertex, faces []OBJFace) WallSurface {
	surfaceMembers := createSurfaceMembers("WallSurface", buildingID, vertices, faces)

	multiSurface := MultiSurfaceProperty{
		MultiSurface: MultiSurface{
			ID:            fmt.Sprintf("WallSurface_%s_ms", buildingID),
			SurfaceMember: surfaceMembers,
		},
	}

	return WallSurface{
		ID:               fmt.Sprintf("WallSurface_%s", buildingID),
		Lod2MultiSurface: multiSurface,
	}
}

func createGroundSurface(buildingID string, vertices []OBJVertex, faces []OBJFace) GroundSurface {
	surfaceMembers := createSurfaceMembers("GroundSurface", buildingID, vertices, faces)

	multiSurface := MultiSurfaceProperty{
		MultiSurface: MultiSurface{
			ID:            fmt.Sprintf("GroundSurface_%s_ms", buildingID),
			SurfaceMember: surfaceMembers,
		},
	}

	return GroundSurface{
		ID:               fmt.Sprintf("GroundSurface_%s", buildingID),
		Lod2MultiSurface: multiSurface,
	}
}

// Helper function to create surface members
func createSurfaceMembers(surfaceType, buildingID string, vertices []OBJVertex, faces []OBJFace) []SurfaceMember {
	surfaceMembers := []SurfaceMember{}

	for i, face := range faces {
		polygon := Polygon{
			ID: fmt.Sprintf("%s_%s_p%d", surfaceType, buildingID, i),
			Exterior: PolygonExterior{
				LinearRing: LinearRing{
					PosList: createPosListFromFace(face, vertices),
				},
			},
		}

		surfaceMembers = append(surfaceMembers, SurfaceMember{Polygon: polygon})
	}

	return surfaceMembers
}

// Create a position list string from a face
func createPosListFromFace(face OBJFace, vertices []OBJVertex) string {
	var posList strings.Builder

	for _, idx := range face.VertexIndices {
		if idx < len(vertices) {
			v := vertices[idx]
			posList.WriteString(fmt.Sprintf("%f %f %f ", v.X, v.Y, v.Z))
		}
	}

	// Close the polygon by repeating the first vertex
	if len(face.VertexIndices) > 0 && face.VertexIndices[0] < len(vertices) {
		v := vertices[face.VertexIndices[0]]
		posList.WriteString(fmt.Sprintf("%f %f %f", v.X, v.Y, v.Z))
	}

	return posList.String()
}
