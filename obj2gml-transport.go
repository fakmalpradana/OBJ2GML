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
<!-- OBJ to CityGML Converter Output - Complete Model Preservation -->
<!-- copyrights 2025 © Fairuz Akmal Pradana | fakmalpradana@gmail.com  -->
`
)

// CityGML structures adapted for Transportation 2.0 (transportation.xsd)
type CityModel struct {
	XMLName        xml.Name                  `xml:"core:CityModel"`
	GML            string                    `xml:"xmlns:gml,attr"`
	Core           string                    `xml:"xmlns:core,attr"`
	Tran           string                    `xml:"xmlns:tran,attr"`
	App            string                    `xml:"xmlns:app,attr,omitempty"`
	Gen            string                    `xml:"xmlns:gen,attr,omitempty"`
	Grp            string                    `xml:"xmlns:grp,attr,omitempty"`
	XLink          string                    `xml:"xmlns:xlink,attr,omitempty"`
	XSI            string                    `xml:"xmlns:xsi,attr,omitempty"`
	SchemaLocation string                    `xml:"xsi:schemaLocation,attr,omitempty"`

	BoundedBy        *BoundedBy                 `xml:"gml:boundedBy,omitempty"`
	CityObjectMember []CityObjectMemberTran     `xml:"core:cityObjectMember"`
}

// BoundedBy / Envelope (reuse pattern from your original)
type BoundedBy struct {
	Envelope Envelope `xml:"gml:Envelope"`
}

type Envelope struct {
	SrsName      string `xml:"srsName,attr,omitempty"`
	SrsDimension string `xml:"srsDimension,attr,omitempty"`
	LowerCorner  string `xml:"gml:lowerCorner,omitempty"`
	UpperCorner  string `xml:"gml:upperCorner,omitempty"`
}

// CityObjectMember for transportation (holds one of many transportation types)
type CityObjectMemberTran struct {
	TransportationComplex *TransportationComplex `xml:"tran:TransportationComplex,omitempty"`
	Road                  *Road                  `xml:"tran:Road,omitempty"`
	Railway               *Railway               `xml:"tran:Railway,omitempty"`
	Track                 *Track                 `xml:"tran:Track,omitempty"`
	Square                *Square                `xml:"tran:Square,omitempty"`
}

// Generic TransportationComplex (base type in XSD)
type TransportationComplex struct {
	ID       string  `xml:"gml:id,attr,omitempty"`
	Class    string  `xml:"tran:class,omitempty"`
	Function string  `xml:"tran:function,omitempty"`
	Usage    string  `xml:"tran:usage,omitempty"`

	// LOD geometries (optional)
	Lod0Network    *Lod0Network     `xml:"tran:lod0Network,omitempty"`
	Lod1MultiSurf  *LodMultiSurface `xml:"tran:lod1MultiSurface,omitempty"`
	Lod2MultiSurf  *LodMultiSurface `xml:"tran:lod2MultiSurface,omitempty"`
	Lod3MultiSurf  *LodMultiSurface `xml:"tran:lod3MultiSurface,omitempty"`
	Lod4MultiSurf  *LodMultiSurface `xml:"tran:lod4MultiSurface,omitempty"`
}

type Road struct {
    GmlID                  string                `xml:"gml:id,attr"`
    Name                   string                `xml:"gml:name,omitempty"`
    Class                  string                `xml:"tran:class,omitempty"`
    Function               string                `xml:"tran:function,omitempty"`
    Usage                  string                `xml:"tran:usage,omitempty"`

    Lod0Network            *Lod0Network          `xml:"tran:lod0Network,omitempty"`
    Lod1MultiSurface       *LodMultiSurface      `xml:"tran:lod1MultiSurface,omitempty"`
    Lod2MultiSurface       *LodMultiSurface      `xml:"tran:lod2MultiSurface,omitempty"`

    // Bagian penting untuk representasi jalan
    TrafficAreas           []TrafficAreaProperty          `xml:"tran:trafficArea,omitempty"`
    AuxiliaryTrafficAreas  []AuxiliaryTrafficAreaProperty `xml:"tran:auxiliaryTrafficArea,omitempty"`
}

// TrafficArea (jalur utama jalan)
type TrafficAreaProperty struct {
    TrafficArea TrafficArea `xml:"tran:TrafficArea"`
}

type TrafficArea struct {
    GmlID            string                `xml:"gml:id,attr"`
    Class            string                `xml:"tran:class,omitempty"`
    Function         string                `xml:"tran:function,omitempty"`
    Usage            string                `xml:"tran:usage,omitempty"`
    Lod2MultiSurface *MultiSurfaceProperty `xml:"tran:lod2MultiSurface,omitempty"`
}

type MultiSurfaceProperty struct {
    MultiSurface MultiSurface `xml:"gml:MultiSurface"`
}

// AuxiliaryTrafficArea (contoh: trotoar, median)
type AuxiliaryTrafficAreaProperty struct {
    AuxiliaryTrafficArea AuxiliaryTrafficArea `xml:"tran:AuxiliaryTrafficArea"`
}

type AuxiliaryTrafficArea struct {
    GmlID            string                `xml:"gml:id,attr"`
    Class            string                `xml:"tran:class,omitempty"`
    Function         string                `xml:"tran:function,omitempty"`
    Usage            string                `xml:"tran:usage,omitempty"`
    Lod2MultiSurface *MultiSurfaceProperty `xml:"tran:lod2MultiSurface,omitempty"`
}

type Railway struct {
    GmlID       string `xml:"gml:id,attr,omitempty"`
    Name     string `xml:"gml:name,omitempty"`
    Class    string `xml:"tran:class,omitempty"`
    Function string `xml:"tran:function,omitempty"`
    Usage    string `xml:"tran:usage,omitempty"`

    Lod0Network   *Lod0Network     `xml:"tran:lod0Network,omitempty"`
    Lod1MultiSurf *LodMultiSurface `xml:"tran:lod1MultiSurface,omitempty"`
}

type Track struct {
    GmlID       string `xml:"gml:id,attr,omitempty"`
    Name     string `xml:"gml:name,omitempty"`
    Class    string `xml:"tran:class,omitempty"`
    Function string `xml:"tran:function,omitempty"`
    Usage    string `xml:"tran:usage,omitempty"`

    Lod1MultiSurf *LodMultiSurface `xml:"tran:lod1MultiSurface,omitempty"`
}

type Square struct {
    GmlID       string `xml:"gml:id,attr,omitempty"`
    Name     string `xml:"gml:name,omitempty"`
    Class    string `xml:"tran:class,omitempty"`
    Function string `xml:"tran:function,omitempty"`
    Usage    string `xml:"tran:usage,omitempty"`

    Lod1MultiSurf *LodMultiSurface `xml:"tran:lod1MultiSurface,omitempty"`
}

// LOD geometry wrappers
type Lod0Network struct {
	// In many CityGML files, lod0Network references a geometric complex or xlink.
	// Keep xlink:href to reference existing network elements, or expand to full GeometricComplex.
	Href string `xml:"xlink:href,attr,omitempty"`
}

type LodMultiSurface struct {
	MultiSurface MultiSurface `xml:"gml:MultiSurface"`
}

// Geometry: MultiSurface -> surfaceMember -> Polygon (same GML pattern)
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
	// Note: you can add <gml:interior> if needed
}

type PolygonExterior struct {
	LinearRing LinearRing `xml:"gml:LinearRing"`
}

type LinearRing struct {
	PosList string `xml:"gml:posList"`
}

// OBJ file structures (unchanged)
type OBJVertex struct {
	X, Y, Z float64
}

type OBJFace []int

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

	    // deteksi otomatis dari group
		transportType, err := detectTransportTypeFromOBJ(objFile)
		if err != nil {
			fmt.Printf("Error detecting transport type for %s: %v\n", baseFileName, err)
			continue
		}

		err = convertOBJToCityGML(objFile, outputFile, fileNameWithoutExt, *epsgCode, transportType)
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

func detectTransportTypeFromOBJ(path string) (string, error) {
    file, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "g ") {
            groupName := strings.ToLower(line[2:]) // ambil setelah "g "

            switch {
            case strings.Contains(groupName, "krl"),
                strings.Contains(groupName, "mrt"),
                strings.Contains(groupName, "lrt"),
                strings.Contains(groupName, "rail"):
                return "railway", nil

            case strings.Contains(groupName, "jl"),
                strings.Contains(groupName, "jalan"),
                strings.Contains(groupName, "jln"),
                strings.Contains(groupName, "road"):
                return "road", nil
            }
        }
    }

    if err := scanner.Err(); err != nil {
        return "", err
    }

    // default
    return "", fmt.Errorf("no valid transport type found in OBJ: %s", path)
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

// Convert OBJ file to CityGML (Road)
func convertOBJToCityGML(inputPath, outputPath, objectID, epsgCode, transportType string) error {
	var member CityObjectMemberTran

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
		GML:            "http://www.opengis.net/gml",
		Core:           "http://www.opengis.net/citygml/2.0",
		Tran:           "http://www.opengis.net/citygml/transportation/2.0",
		App:            "http://www.opengis.net/citygml/appearance/2.0",
		Gen:            "http://www.opengis.net/citygml/generics/2.0",
		Grp:            "http://www.opengis.net/citygml/cityobjectgroup/2.0",
		XLink:          "http://www.w3.org/1999/xlink",
		XSI:            "http://www.w3.org/2001/XMLSchema-instance",
		SchemaLocation: "http://www.opengis.net/citygml/2.0 http://schemas.opengis.net/citygml/2.0/cityGMLBase.xsd " +
			"http://www.opengis.net/citygml/transportation/2.0 http://schemas.opengis.net/citygml/transportation/2.0/transportation.xsd",
		BoundedBy: &BoundedBy{
			Envelope: Envelope{
				SrsName:      fmt.Sprintf("http://www.opengis.net/def/crs/EPSG/0/%s", epsgCode),
				SrsDimension: "3",
				LowerCorner:  fmt.Sprintf("%f %f %f", minX, minY, minZ),
				UpperCorner:  fmt.Sprintf("%f %f %f", maxX, maxY, maxZ),
			},
		},
	}

	switch strings.ToLower(transportType) {
	case "railway":
		railway := Railway{
			GmlID:       objectID,
			Class:    "railway",
			Function: "transport",
			Usage:    "public",
			Lod1MultiSurf: &LodMultiSurface{
				MultiSurface: MultiSurface{ID: fmt.Sprintf("%s-ms", objectID)},
			},
		}

		// isi polygons dari faces
		for i, face := range faces {
			face = ensureConsistentWindingOrder(vertices, face)

			polygonID := fmt.Sprintf("%s-polygon-%d", objectID, i)

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
			posList := strings.TrimSpace(posListBuilder.String())

			surfaceMember := SurfaceMember{
				Polygon: Polygon{
					ID: polygonID,
					Exterior: PolygonExterior{
						LinearRing: LinearRing{
							PosList: posList,
						},
					},
				},
			}

			railway.Lod1MultiSurf.MultiSurface.SurfaceMember =
				append(railway.Lod1MultiSurf.MultiSurface.SurfaceMember, surfaceMember)
		}

		member = CityObjectMemberTran{Railway: &railway}

	default: // fallback ke Road
		road := Road{
			GmlID:    objectID,
			Class:    "road",
			Function: "transport",
			Usage:    "public",
			Lod1MultiSurface: &LodMultiSurface{
				MultiSurface: MultiSurface{ID: fmt.Sprintf("%s-ms", objectID)},
			},
		}

		for i, face := range faces {
			face = ensureConsistentWindingOrder(vertices, face)

			polygonID := fmt.Sprintf("%s-polygon-%d", objectID, i)

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
			posList := strings.TrimSpace(posListBuilder.String())

			surfaceMember := SurfaceMember{
				Polygon: Polygon{
					ID: polygonID,
					Exterior: PolygonExterior{
						LinearRing: LinearRing{
							PosList: posList,
						},
					},
				},
			}
			

			road.Lod1MultiSurface.MultiSurface.SurfaceMember =
				append(road.Lod1MultiSurface.MultiSurface.SurfaceMember, surfaceMember)
		}

		member = CityObjectMemberTran{Road: &road}
	}

	// 🚀 Tambahkan transportasi (Road/Railway) ke CityModel
	cityModel.CityObjectMember = append(
		cityModel.CityObjectMember,
		member,
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
