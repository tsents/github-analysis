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

type ParsedEntry[T any] struct {
	RawEntry *[]byte
	Data     T
}

type AnyJSON map[string]any;


/*
A function that collects the processed outputs of the ActionFunc that parses the input.

In the ParseInParallel, this manager is created once, while the actions workers are multiple,
and each of those is reading multiple lines and feeding the process output into the input channel.

Make sure that the type T is immutable! this is because the same structure is reused to remove
memory overhead. a fixed for non-immutable structures will release in feature development
*/
type ManagerFunc[T any, R any] func(<-chan ParsedEntry[T]) R;

/*
A function that for a given name, returns the corresponding reader and length, throwing errors
encountered during runtime back to the caller. For example for http request, the name is the requested 
file and site.
*/
type informativeReader func(string) (io.ReadCloser, int64, error)

const (
	WORKER_COUNT   = 16  // adjust based on CPU cores
	CHANNEL_BUFFER = 32 // buffer size for work queue
	HTTP_TIMEOUT = 300 // seconds
	BUFFER_SIZE = 1024 * 1024 * 5 //5MB
)


/*
For a given set of files, and sourceType (http/file) the functions reads the file and
parsed it as a NDJSON, based on the struct T to unmarshal.


Then all the structs are collected by a channel into the "manager" function, that is a
thread that is meant to collect all the structers given into one desired output.


NOTE that the struct T is reused by the threads that are outputed, and any use of 
non immutable fields (all MAP types for example, strings are ok) will lead to race conditions (!). this will be fixed in feature development.
To bypass it change filed type T to *T to make the struct store only the pointer, and automaticly allocate new.

Parameters:
	- files[]		An array of files, serving as the source. can be real/url.
	- manager 		The function that collects all the output from the functions, and used to give the final result.
	- sourceType	The type of the file source. either a route to a real file. or http. (takes "file"/"http") 
*/
func ParseInParallel[T any, R any](files []string, manager ManagerFunc[T,R], sourceType string) (R, error) {
	var wg sync.WaitGroup
	var workerWg sync.WaitGroup  // <--- NEW wait group for workers
	workChan := make(chan ParsedEntry[T], CHANNEL_BUFFER)

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

func processFile[T any](filename string, getReader informativeReader, out chan<- ParsedEntry[T]) {
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
Unmarsjels all jsons read from the reader, assuming format of compressed NDJSON.
all read structures are then piped into the out channel for the caller to use.


NOTE that even if the caller does not intened to use the output, an iteration over the 
channel is needed to make sure that it is cleaned.

Parameters:
	- originalReader	The reader that is used to open the gzip. original in the sense that it is
	later wrapped in a gzip reader to unmarshel it.
	- out				The channel to push out the unmarshaled objects into, the type of the json struct
	is inferred based on the type of T.
 */
func ProcessNDJSONInParallel[T any](originalReader io.Reader, out chan<- ParsedEntry[T]) error {
	gz, err := gzip.NewReader(originalReader)
	if err != nil {
		return fmt.Errorf("gzip reader error: %v", err)
	}
	defer gz.Close()

	reader := bufio.NewReaderSize(gz, BUFFER_SIZE)
	var json = jsoniter.ConfigFastest

	var item ParsedEntry[T];
	for {
		line, err := reader.ReadSlice('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("read error: %T %v", err, err)
		}
		if len(line) > 0 {
			item.RawEntry = &line;
			if err := json.Unmarshal(line, &item.Data); err != nil {
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
		cancel()
		return nil, 0, fmt.Errorf("bad status: %s\nBody:\n%s", resp.Status, string(body)) //Pass error up.
	}
	
	lengthStr := resp.Header.Get("Content-Length")
	length, err := strconv.ParseInt(lengthStr, 10, 64)
	if err != nil {
		cancel()
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
