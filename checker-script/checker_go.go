package checker

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run check_ndjson.go <file.json.gz>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	isNDJSON, err := checkNDJSONFormat(file)
	if err != nil {
		log.Fatalf("Error checking NDJSON format: %v", err)
	}

	if isNDJSON {
		fmt.Println("✅ This file appears to be in NDJSON format (newline-delimited JSON).")
	} else {
		fmt.Println("❌ This file does NOT appear to be in NDJSON format.")
	}
}

func checkNDJSONFormat(file *os.File) (bool, error) {
	gz, err := gzip.NewReader(file)
	if err != nil {
		return false, fmt.Errorf("gzip reader error: %v", err)
	}
	defer gz.Close()

	reader := bufio.NewReader(gz)
	lineCount := 0
	validCount := 0

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return false, fmt.Errorf("read error: %v", err)
		}
		lineCount++

		// Ignore empty lines
		if len(line) == 0 || string(line) == "\n" {
			continue
		}

		var js map[string]interface{}
		if json.Unmarshal(line, &js) == nil {
			validCount++
		} else {
			log.Printf("Invalid JSON at line %d", lineCount)
		}

		if lineCount >= 1000 || err == io.EOF {
			break
		}
	}

	return validCount > 0 && float64(validCount)/float64(lineCount) > 0.95, nil
}

