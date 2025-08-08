// Package enabling parsing of .json.gz files in parallel. using local files or fetching from the internet.
package myjson

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var client = http.DefaultClient


type AnyJSON map[string]any;


/*
A function that collects the processed outputs of the ActionFunc that parses the input.

In the ParseInParallel, this manager is created once, while the actions workers are multiple,
and each of those is reading multiple lines and feeding the process output into the input channel.
*/
type ManagerFunc[T any, R any] func(<-chan T) R;

// Return the reader and the length. throwing errors upwards
type informativeReader func(string) (io.ReadCloser, int64, error)

const (
	WORKER_COUNT   = 16  // adjust based on CPU cores
	CHANNEL_BUFFER = 32 // buffer size for work queue
	HTTP_TIMEOUT = 300 // seconds
	BUFFER_SIZE = 1024 * 1024 * 5 //5MB
)


/*
Takes in multiple files, then for each files reads it, seperating
by new line. Then for each line it casts action(line), and takes
it output into a buffered channel for manager, which combines all the output
from the actions into one output.

Paramaters:
	- files[]		An array of files, serving as the source. can be real/url.
	- manager 		The function that collects all the output from the functions, and used to give the final result.
	- sourceType	The type of the file source. either a route to a real file. or http. (takes "file"/"http") 
*/
func ParseInParallel[T any, R any](files []string, manager ManagerFunc[T,R], sourceType string) (R, error) {
	var wg sync.WaitGroup
	var workerWg sync.WaitGroup  // <--- NEW wait group for workers
	workChan := make(chan T, CHANNEL_BUFFER)

	// Progress tracking
	var processed int64
	start := time.Now()
	total := int64(len(files))

	var readingMethod informativeReader
	switch sourceType {
	case "http":
		readingMethod = fetchWithTimeout 
	case "file":
		readingMethod = openAndSize
	default:
		fmt.Fprintf(os.Stderr, "Invalid SourceType: %v\n", sourceType)
		readingMethod = func(file string) (io.ReadCloser, int64, error) {
			return nil, 0 , fmt.Errorf("No reader")
		}
	}

	jobs := make(chan string, CHANNEL_BUFFER)

	// File producer
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(jobs)
		for _, file := range files {
			jobs <- file
		}
		fmt.Println("Done file producer")
	}()

	// Workers
	for w := 0; w < WORKER_COUNT; w++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for file := range jobs {
				processFile[T](file, readingMethod, workChan)

				curr := atomic.AddInt64(&processed, 1)
				if curr%10 == 0 || curr == total {
					elapsed := time.Since(start)
					remaining := total - curr
					rate := float64(curr) / elapsed.Seconds()
					eta := time.Duration(float64(remaining)/rate) * time.Second
					log.Printf("Progress: %d/%d | ETA: %s\n", curr, total, eta.Truncate(time.Second))
				}
			}
			fmt.Println("Done worker")
		}()
	}

	// Close workChan after all workers are done
	go func() {
		workerWg.Wait()
		close(workChan)
	}()

	// Manager processes results from workChan
	result := manager(workChan)

	// Wait for file producer (jobs channel closer)
	wg.Wait()
	return result, nil
}

func processFile[T any](filename string, getReader informativeReader, out chan<- T) {
	reader, _, err := getReader(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open source: %v\n", err);
		return
	}
	err = ProcessNDJSONInParallel(reader, out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing NDJSON: %v\n", err)
	}
	reader.Close()
}

/*
  Casts the action function on all jsons in the gzip file in paralle.
  Using multiple workers and streams to parse process the data, and then
  running the action function itself, all in parallel.
  
  The user should make sure that the action function has no race conditions,
  and to capture its output, it should output into a channel within itself.
 */
func ProcessNDJSONInParallel[T any](originalReader io.Reader, out chan<- T) error {
	gz, err := gzip.NewReader(originalReader)
	if err != nil {
		return fmt.Errorf("gzip reader error: %v", err)
	}
	defer gz.Close()

	reader := bufio.NewReaderSize(gz, BUFFER_SIZE)
	var json = jsoniter.ConfigFastest

	var item T;
	for {
		line, err := reader.ReadSlice('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("read error: %T %v", err, err)
		}
		if len(line) > 0 {
			if err := json.Unmarshal(line, &item); err != nil {
				fmt.Printf("JSON unmarshal error: %v\n", err)
				continue
			}
			out <- item
		}
		if err == io.EOF {
			break
		}
	}
	return nil
}

type cancelReadCloser struct {
    io.ReadCloser
    cancelFunc context.CancelFunc
}

func (c cancelReadCloser) Close() error {
    c.cancelFunc()
    return c.ReadCloser.Close() 
}

func fetchWithTimeout(url string) (io.ReadCloser, int64, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(HTTP_TIMEOUT)*time.Second)
	//defer cancel()

	// Create HTTP request with the context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		cancel()
		return nil, 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		cancel()
		return nil, 0, err
	}
	//defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("bad status: %s\nBody:\n%s", resp.Status, string(body)) //Pass error up.
	}
	
	lengthStr := resp.Header.Get("Content-Length")
	length, err := strconv.ParseInt(lengthStr, 10, 64)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed Atoi of length: %v\n", lengthStr) //Pass error up.
	} 

	resp.Body = cancelReadCloser{
		ReadCloser: resp.Body,
		cancelFunc: cancel,
	}
	return resp.Body, length, nil
}

//Open a file and get its size. While throwing errors up.
func openAndSize(filename string) (io.ReadCloser, int64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, 0, err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, 0, err
	}
	return file, fileInfo.Size(), err
} 
