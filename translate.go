package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

func main() {
	// Configuration parameters
	inputDir := "export/3D_UJI_MONEV_LOD1.obj" // Replace with actual path
	translationX := 700621.357389              // Replace with desired X translation
	translationY := 9311966.06841              // Replace with desired Y translation
	translationZ := 0.0                        // Z translation (default: 0)

	// Create output directory name by appending "_translated" to input directory
	dirName := filepath.Base(inputDir)
	parentDir := filepath.Dir(inputDir)
	outputDir := filepath.Join(parentDir, dirName+"_translated")

	// Create output directory if it doesn't exist
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Find all OBJ files in the input directory
	files, err := filepath.Glob(filepath.Join(inputDir, "*.obj"))
	if err != nil {
		fmt.Printf("Error finding OBJ files: %v\n", err)
		return
	}

	totalFiles := len(files)
	fmt.Printf("Found %d OBJ files to process\n", totalFiles)

	// Use a wait group to track completion of goroutines
	var wg sync.WaitGroup

	// Channel to collect results
	results := make(chan bool, totalFiles)
	errorFiles := make(chan string, totalFiles)

	// Process files concurrently with worker pool
	maxWorkers := 4 // Adjust based on your CPU cores
	semaphore := make(chan struct{}, maxWorkers)

	for _, file := range files {
		wg.Add(1)

		go func(filePath string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			fileName := filepath.Base(filePath)
			outputFile := filepath.Join(outputDir, fileName)

			err := translateOBJFile(filePath, outputFile, translationX, translationY, translationZ)
			if err != nil {
				errorFiles <- fileName
			} else {
				results <- true
			}
		}(file)
	}

	// Close channels when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
		close(errorFiles)
	}()

	// Count successful translations
	successCount := 0
	for range results {
		successCount++
	}

	// Collect error files
	var failedFiles []string
	for fileName := range errorFiles {
		failedFiles = append(failedFiles, fileName)
	}

	// Print summary
	fmt.Printf("Successfully translated %d from %d obj files\n", successCount, totalFiles)

	if len(failedFiles) > 0 {
		fmt.Printf("Failed to translate %d files: %v\n", len(failedFiles), failedFiles)
	}
}

// translateOBJFile reads an OBJ file, translates its vertices, and writes to output
func translateOBJFile(inputPath, outputPath string, tx, ty, tz float64) error {
	// Open input file
	inFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer inFile.Close()

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	scanner := bufio.NewScanner(inFile)
	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	// Increase scanner buffer size for large files
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	// Process file line by line
	for scanner.Scan() {
		line := scanner.Text()

		// Check if the line defines a vertex
		if len(line) > 2 && line[0] == 'v' && line[1] == ' ' {
			// Parse vertex coordinates
			parts := strings.Fields(line)
			if len(parts) >= 4 { // "v x y z" format
				x, err1 := strconv.ParseFloat(parts[1], 64)
				y, err2 := strconv.ParseFloat(parts[2], 64)
				z, err3 := strconv.ParseFloat(parts[3], 64)

				if err1 == nil && err2 == nil && err3 == nil {
					// Apply translation
					x += tx
					y += ty
					z += tz

					// Write translated vertex efficiently
					fmt.Fprintf(writer, "v %g %g %g", x, y, z)

					// Add any additional vertex data (color, etc.)
					for i := 4; i < len(parts); i++ {
						fmt.Fprintf(writer, " %s", parts[i])
					}
					fmt.Fprintln(writer)
					continue
				}
			}
		}

		// Write unchanged line
		fmt.Fprintln(writer, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input file: %v", err)
	}

	return nil
}
