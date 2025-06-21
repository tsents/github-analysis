package myjson

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	jsoniter "github.com/json-iterator/go"
)
var json = jsoniter.ConfigCompatibleWithStandardLibrary


type AnyJSON map[string]any;
type ActionFunc[T any] func([]byte) (T, error);
type boundedActionFunc func([]byte);
type ManagerFunc[T any] func(<-chan T);

const (
	workerCount   = 8  // adjust based on CPU cores
	channelBuffer = 16 // buffer size for work queue
	readWorkersCount = 4 // adjust based on usecase? there are upsides/downsides for higher/lower.
)

/*
 Parses a list of files in paralle. The action in parsing
 is the action "action", and additionaly a manager goroutine
 is create (once!) with a in channel from the action goroutines.
*/
func ParseInParallel[T any](files []string, action ActionFunc[T], manager ManagerFunc[T]) {
	var wg sync.WaitGroup
	workChan := make(chan T)

	// Start the manager goroutine
	wg.Add(1)
	go func() {
		manager(workChan)
		wg.Done()
	}()

	var fileWg sync.WaitGroup

	// Progress tracking
	var processed int64
	start := time.Now()
	total := int64(len(files))

	// Start worker pool
	fileChan := make(chan string)

	// Bound action to the work channel
	var boundedAction boundedActionFunc = func(line []byte) {
		retValue, err := action(line)
		if (err != nil) {
			return // Can add stuff later for error printing.
		}
		workChan<-retValue
	} 

	for i := 0; i < readWorkersCount; i++ {
		fileWg.Add(1)
		go func() {
			defer fileWg.Done()
			for filename := range fileChan {
				file, err := os.Open(filename)
				if err != nil {
					fmt.Printf("Failed to open file: %v", err)
					continue
				}

				err = ProcessNDJSONInParallel(file, boundedAction)
				file.Close() // close explicitly
				if err != nil {
					fmt.Printf("Error processing NDJSON: %v", err)
					continue
				}

				// Update progress
				curr := atomic.AddInt64(&processed, 1)
				if curr%10 == 0 || curr == total {
					elapsed := time.Since(start)
					remaining := total - curr
					rate := float64(curr) / elapsed.Seconds()
					eta := time.Duration(float64(remaining)/rate) * time.Second
					log.Printf("Progress: %d/%d | ETA: %s\n", curr, total, eta.Truncate(time.Second))
				}
			}
		}()
	}

	// Feed file names to workers
	for _, f := range files {
		fileChan <- f
	}
	close(fileChan)

	// Wait for file processing and manager
	fileWg.Wait()
	close(workChan)
	wg.Wait()
}

/*
  Casts the action function on all jsons in the gzip file in paralle.
  Using multiple workers and streams to parse process the data, and then
  running the action function itself, all in parallel.
  
  The user should make sure that the action function has no race conditions,
  and to capture its output, it should output into a channel within itself.
 */
func ProcessNDJSONInParallel(file *os.File, action boundedActionFunc) error {
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
				action(line)
			}
		}(i)
	}

	// Read lines from file and feed into channel
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("read error: %T %v", err, err)
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

