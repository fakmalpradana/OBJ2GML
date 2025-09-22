package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// XML namespaces and schema declarations
const (
	xmlHeader = `<?xml version="1.0" encoding="UTF-8"?>
<!-- OBJ to CityGML WaterBody Converter Output -->
<!-- copyrights 2025 -->
`
)

// CityModel structures adapted for WaterBody 2.0 (waterBody.xsd)
type CityModel struct {
	XMLName        xml.Name                 `xml:"core:CityModel"`
	GML            string                   `xml:"xmlns:gml,attr"`
	Core           string                   `xml:"xmlns:core,attr"`
	Wtr            string                   `xml:"xmlns:wtr,attr,omitempty"`  // konsisten prefix
	App            string                   `xml:"xmlns:app,attr,omitempty"`
	Gen            string                   `xml:"xmlns:gen,attr,omitempty"`
	Grp            string                   `xml:"xmlns:grp,attr,omitempty"`
	XLink          string                   `xml:"xmlns:xlink,attr,omitempty"`
	XSI            string                   `xml:"xmlns:xsi,attr,omitempty"`
	SchemaLocation string                   `xml:"xsi:schemaLocation,attr,omitempty"`

	BoundedBy        *BoundedBy             `xml:"gml:boundedBy,omitempty"`
	CityObjectMember []CityObjectMemberWtr  `xml:"core:cityObjectMember"`
}

// CityObjectMember for waterbody objects
type CityObjectMemberWtr struct {
	WaterBody *WaterBody `xml:"wtr:WaterBody,omitempty"`
}

// Envelope / BoundedBy
type BoundedBy struct {
	Envelope Envelope `xml:"gml:Envelope"`
}

type Envelope struct {
	SrsName      string `xml:"srsName,attr,omitempty"`
	SrsDimension string `xml:"srsDimension,attr,omitempty"`
	LowerCorner  string `xml:"gml:lowerCorner,omitempty"`
	UpperCorner  string `xml:"gml:upperCorner,omitempty"`
}

// Main WaterBody object
type WaterBody struct {
	ID       string   `xml:"gml:id,attr,omitempty"`
	Name     string   `xml:"gml:name,omitempty"`
	Class    string   `xml:"wtr:class,omitempty"`
	Function []string `xml:"wtr:function,omitempty"`
	Usage    []string `xml:"wtr:usage,omitempty"`

	// LOD representations
	Lod0MultiSurface *LodMultiSurface `xml:"wtr:lod0MultiSurface,omitempty"`
	Lod1MultiSurface *LodMultiSurface `xml:"wtr:lod1MultiSurface,omitempty"`
	Lod1Solid        *SolidProperty   `xml:"wtr:lod1Solid,omitempty"`
	Lod2Solid        *SolidProperty   `xml:"wtr:lod2Solid,omitempty"`
	Lod3Solid        *SolidProperty   `xml:"wtr:lod3Solid,omitempty"`
	Lod4Solid        *SolidProperty   `xml:"wtr:lod4Solid,omitempty"`

	BoundedBy []WaterBoundary `xml:"wtr:boundedBy,omitempty"`
}

// Boundary surfaces
type WaterBoundary struct {
	WaterSurface       *WaterSurface       `xml:"wtr:WaterSurface,omitempty"`
	WaterGroundSurface *WaterGroundSurface `xml:"wtr:WaterGroundSurface,omitempty"`
}

// WaterSurface
type WaterSurface struct {
	ID    string `xml:"gml:id,attr,omitempty"`
	Class string `xml:"wtr:class,omitempty"`

	Lod2MultiSurface *LodMultiSurface `xml:"wtr:lod2MultiSurface,omitempty"`
	Lod3MultiSurface *LodMultiSurface `xml:"wtr:lod3MultiSurface,omitempty"`
	Lod4MultiSurface *LodMultiSurface `xml:"wtr:lod4MultiSurface,omitempty"`
}

// WaterGroundSurface
type WaterGroundSurface struct {
	ID    string `xml:"gml:id,attr,omitempty"`
	Class string `xml:"wtr:class,omitempty"`

	Lod2MultiSurface *LodMultiSurface `xml:"wtr:lod2MultiSurface,omitempty"`
	Lod3MultiSurface *LodMultiSurface `xml:"wtr:lod3MultiSurface,omitempty"`
	Lod4MultiSurface *LodMultiSurface `xml:"wtr:lod4MultiSurface,omitempty"`
}

// Shared LOD geometry structures (reuse from bridge code)
type SolidProperty struct {
	Solid Solid `xml:"gml:Solid"`
}

type Solid struct {
	ID       string        `xml:"gml:id,attr,omitempty"`
	Exterior SolidExterior `xml:"gml:exterior"`
}

type SolidExterior struct {
	Shell Shell `xml:"gml:Shell"`
}

type Shell struct {
	SurfaceMembers []SurfaceMember `xml:"gml:surfaceMember"`
}

type LodMultiSurface struct {
	MultiSurface MultiSurface `xml:"gml:MultiSurface"`
}

type MultiSurface struct {
	ID            string          `xml:"gml:id,attr,omitempty"`
	SurfaceMember []SurfaceMember `xml:"gml:surfaceMember"`
}

type SurfaceMember struct {
	Polygon Polygon `xml:"gml:Polygon"`
}

type Polygon struct {
	ID       string          `xml:"gml:id,attr,omitempty"`
	Exterior PolygonExterior `xml:"gml:exterior"`
}

type PolygonExterior struct {
	LinearRing LinearRing `xml:"gml:LinearRing"`
}

type LinearRing struct {
	PosList string `xml:"gml:posList"`
}

// OBJ structures (unchanged, optional kalau mau dipakai di converter)
type OBJVertex struct {
	X, Y, Z float64
}

type OBJFace []int

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

// Calculate normal vector for a triangle
func calculateNormal(v1, v2, v3 OBJVertex) Vector3D {
	// Calculate vectors from v1 to v2 and v1 to v3
	ux := v2.X - v1.X
	uy := v2.Y - v1.Y
	uz := v2.Z - v1.Z

	vx := v3.X - v1.X
	vy := v3.Y - v1.Y
	vz := v3.Z - v1.Z

	// Cross product
	nx := uy*vz - uz*vy
	ny := uz*vx - ux*vz
	nz := ux*vy - uy*vx

	// Normalize
	length := math.Sqrt(nx*nx + ny*ny + nz*nz)
	if length > 0 {
		nx /= length
		ny /= length
		nz /= length
	}

	return Vector3D{X: nx, Y: ny, Z: nz}
}

// // Ensure consistent winding order for face
// func ensureConsistentWindingOrder(vertices []OBJVertex, face OBJFace, up Vector3D) OBJFace {
// 	if len(face) < 3 {
// 		return face
// 	}

// 	// Get vertices for the face
// 	v1 := vertices[face[0]-1]
// 	v2 := vertices[face[1]-1]
// 	v3 := vertices[face[2]-1]

// 	// Calculate normal
// 	normal := calculateNormal(v1, v2, v3)

// 	// Dot product dengan arah "up"
// 	dot := normal.X*up.X + normal.Y*up.Y + normal.Z*up.Z

// 	// Kalau arahnya berlawanan, reverse
// 	if dot < 0 {
// 		for i, j := 0, len(face)-1; i < j; i, j = i+1, j-1 {
// 			face[i], face[j] = face[j], face[i]
// 		}
// 	}

// 	return face
// }

// Ensure consistent winding order for face
func ensureConsistentWindingOrder(vertices []OBJVertex, face OBJFace) OBJFace {
	if len(face) < 3 {
		return face
	}

	// Get vertices for the face
	v1 := vertices[face[0]-1]
	v2 := vertices[face[1]-1]
	v3 := vertices[face[2]-1]

	// Calculate normal
	normal := calculateNormal(v1, v2, v3)

	// If normal is pointing inward (negative Z), reverse the winding order
	// This is a simplification - in a real application, you'd need a more sophisticated check
	if normal.Z < 0 {
		// Reverse the face indices
		for i, j := 0, len(face)-1; i < j; i, j = i+1, j-1 {
			face[i], face[j] = face[j], face[i]
		}
	}

	return face
}

// Convert OBJ file to CityGML (WaterBody)
func convertOBJToCityGML(inputPath, outputPath, waterbodyID, epsgCode string) error {	
	// up := Vector3D{X: 0, Y: 0, Z: 1}

	
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

	// Create CityGML root
	cityModel := CityModel{
		GML:  "http://www.opengis.net/gml",
		Core: "http://www.opengis.net/citygml/2.0",
		Wtr:  "http://www.opengis.net/citygml/waterbody/2.0",
		App:  "http://www.opengis.net/citygml/appearance/2.0",
		Gen:  "http://www.opengis.net/citygml/generics/2.0",
		Grp:  "http://www.opengis.net/citygml/cityobjectgroup/2.0",
		XLink: "http://www.w3.org/1999/xlink",
		XSI:   "http://www.w3.org/2001/XMLSchema-instance",
		SchemaLocation: "http://www.opengis.net/citygml/2.0 http://schemas.opengis.net/citygml/2.0/cityGMLBase.xsd " +
			"http://www.opengis.net/citygml/waterbody/2.0 http://schemas.opengis.net/citygml/waterbody/2.0/waterBody.xsd",
		BoundedBy: &BoundedBy{
			Envelope: Envelope{
				SrsName:      fmt.Sprintf("http://www.opengis.net/def/crs/EPSG/0/%s", epsgCode),
				SrsDimension: "3",
				LowerCorner:  fmt.Sprintf("%f %f %f", minX, minY, minZ),
				UpperCorner:  fmt.Sprintf("%f %f %f", maxX, maxY, maxZ),
			},
		},
	}

	// Create WaterBody object 🚀
	waterbody := WaterBody{
		ID:       waterbodyID,
		Class:    "waterbody",
		Function: []string{"river"}, // slice, bisa lebih dari satu fungsi
		Lod1MultiSurface: &LodMultiSurface{
			MultiSurface: MultiSurface{ID: fmt.Sprintf("%s-ms", waterbodyID)},
		},
	}

	// Convert faces -> gml:surfaceMember
	for i, face := range faces {
		face = ensureConsistentWindingOrder(vertices, face)

		polygonID := fmt.Sprintf("%s-polygon-%d", waterbodyID, i)

		var posListBuilder strings.Builder
		for _, vIdx := range face {
			if vIdx > 0 && vIdx <= len(vertices) {
				v := vertices[vIdx-1]
				posListBuilder.WriteString(fmt.Sprintf("%f %f %f ", v.X, v.Y, v.Z))
			}
		}

		// close polygon ring
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

		waterbody.Lod1MultiSurface.MultiSurface.SurfaceMember =
			append(waterbody.Lod1MultiSurface.MultiSurface.SurfaceMember, surfaceMember)
	}

	// 🚀 Add WaterBody into CityModel
	cityModel.CityObjectMember = append(
		cityModel.CityObjectMember,
		CityObjectMemberWtr{WaterBody: &waterbody},
	)

	// Generate XML
	output, err := xml.MarshalIndent(cityModel, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to generate XML: %v", err)
	}

	// Add XML header
	xmlData := []byte(xmlHeader + string(output))

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
