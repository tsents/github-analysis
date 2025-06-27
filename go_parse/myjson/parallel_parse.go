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
	"net/http"
	"context"

	jsoniter "github.com/json-iterator/go"
)

//Define the json engine for this package, and all subpackages that will inherit it.
var Myjson = jsoniter.ConfigCompatibleWithStandardLibrary
var client = http.DefaultClient


type AnyJSON map[string]any;

/*
A function that is done on each json in the file (parses the bytes[] to json).

This function has generic [T] that indecates its output into a manager function that
collects outputs of all files and lines into one nice output.
*/
type ActionFunc[T any] func([]byte) (T, error);

//An actionFunc that its output is binded into a channel of type T.
type bindedActionFunc func([]byte);

/*
A function that collects the processed outputs of the ActionFunc that parses the input.

In the ParseInParallel, this manager is created once, while the actions workers are multiple,
and each of those is reading multiple lines and feeding the process output into the input channel.
*/
type ManagerFunc[T any] func(<-chan T);

const (
	workerCount   = 16  // adjust based on CPU cores
	channelBuffer = 16 // buffer size for work queue
	readWorkersCount = 2 // adjust based on usecase? there are upsides/downsides for higher/lower.
	httpTimeout = 300 // seconds
)

/*
 Parses a list of files in paralle. The action in parsing
 is the action "action", and additionaly a manager goroutine
 is create (once!) with a in channel from the action
 goroutines to allow communication (of type T).
 The file format is expected to be NDJSON (new line delimited json),
 that is compressed using gz. (thus filename.json.jz).

Paramaters:
	- files[]		An array of files, serving as the source. can be real/url.
	- action		The action to preform on each json entry in the file.
	- manager 		The function that collects all the output from the functions, and used to give the final result.
	- sourceType	The type of the file source. either a route to a real file. or http. (takes "file"/"http") 
*/
func ParseInParallel[T any](files []string, action ActionFunc[T], manager ManagerFunc[T], sourceType string) {
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
	var bindedAction bindedActionFunc = func(line []byte) {
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
				processFile(filename, bindedAction, sourceType)

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

func processFile(filename string, bindedAction bindedActionFunc, sourceType string) {
	var reader io.ReadCloser
	var err error
	switch sourceType {
		case "file":
			reader, err = os.Open(filename)
		case "http":
			reader, err = fetchWithTimeout(filename, httpTimeout)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open source: %v\n", err);
		return
	}
	defer reader.Close()
	err = ProcessNDJSONInParallel(reader, bindedAction)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing NDJSON: %v\n", err)
	}
}

/*
  Casts the action function on all jsons in the gzip file in paralle.
  Using multiple workers and streams to parse process the data, and then
  running the action function itself, all in parallel.
  
  The user should make sure that the action function has no race conditions,
  and to capture its output, it should output into a channel within itself.
 */
func ProcessNDJSONInParallel(originalReader io.ReadCloser, action bindedActionFunc) error {
	gz, err := gzip.NewReader(originalReader)
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

func fetchWithTimeout(url string, timeoutSeconds int) (io.ReadCloser, error) {
	// Create a context with timeout
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)

	// Create HTTP request with the context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("bad status: %s\nBody:\n%s", resp.Status, string(body)) //Pass error up.
	}

	// Return the response body as io.Reader
	// The caller is responsible for closing resp.Body
	return resp.Body, nil
}

