package myjson

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"sync"

	jsoniter "github.com/json-iterator/go"
)
var json = jsoniter.ConfigCompatibleWithStandardLibrary


type AnyJSON map[string]any;
type ActionFunc func([]byte);

const (
	workerCount   = 8  // adjust based on CPU cores
	channelBuffer = 16 // buffer size for work queue
)

/*
  Casts the action function on all jsons in the gzip file in paralle.
  Using multiple workers and streams to parse process the data, and then
  running the action function itself, all in parallel.
  
  The user should make sure that the action function has no race conditions,
  and to capture its output, it should output into a channel within itself.
 */
func ProcessNDJSONInParallel(file *os.File, action ActionFunc) error {
	gz, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("gzip reader error: %v", err)
	}
	defer gz.Close()

	reader := bufio.NewReader(gz)

	lines := make(chan []byte, channelBuffer)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for line := range lines {
				processLine(workerID, line, action)
			}
		}(i)
	}

	// Read lines from file and feed into channel
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("read error: %v", err)
		}
		if len(line) > 0 {
			lines <- line
		}
		if err == io.EOF {
			break
		}
	}
	close(lines)
	wg.Wait()
	return nil
}

func processLine(workerID int, line []byte, action ActionFunc) {
	action(line)
}

