package main

import (
	"log"
	"fmt"
	"flag"
	"os"
	"parser/myjson"
	"sync"
	"time"
	"sync/atomic"
)

import _ "net/http/pprof"
import "net/http"


func testParse(files []string) {
	var wg sync.WaitGroup
	workChan := make(chan any)
	action := func(data []byte) {
		value, err := myjson.UnmarshalPayload(data)
		if (err != nil) {
			fmt.Fprintf(os.Stderr, "Error encountered %v\n", err);
			return;
		}
		workChan <- value
	}

	// Start the manager goroutine
	wg.Add(1)
	go func() {
		for _ = range workChan {
			continue
		}
		wg.Done()
	}()

	var fileWg sync.WaitGroup
	numWorkers := 2 // Adjust based on CPU cores or IO limits

	// Progress tracking
	var processed int64
	start := time.Now()
	total := int64(len(files))

	// Start worker pool
	fileChan := make(chan string)

	for i := 0; i < numWorkers; i++ {
		fileWg.Add(1)
		go func() {
			defer fileWg.Done()
			for filename := range fileChan {
				file, err := os.Open(filename)
				if err != nil {
					log.Printf("Failed to open file: %v", err)
					continue
				}

				err = myjson.ProcessNDJSONInParallel(file, action)
				file.Close() // close explicitly
				if err != nil {
					log.Printf("Error processing NDJSON: %v", err)
					continue
				}

				// Update progress
				curr := atomic.AddInt64(&processed, 1)
				if curr%10 == 0 || curr == total {
					elapsed := time.Since(start)
					remaining := total - curr
					rate := float64(curr) / elapsed.Seconds()
					eta := time.Duration(float64(remaining)/rate) * time.Second
					log.Printf("Progress: %d/%d | ETA: %s", curr, total, eta.Truncate(time.Second))
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


func infer(files []string) {
	var wg sync.WaitGroup
	workChan := make(chan string)
	action := myjson.InferFlattenedDecorator(workChan)

	// Start the manager goroutine
	wg.Add(1)
	go func() {
		myjson.InferManeger(workChan)
		wg.Done()
	}()

	var fileWg sync.WaitGroup
	numWorkers := 2 // Adjust based on CPU cores or IO limits

	// Progress tracking
	var processed int64
	start := time.Now()
	total := int64(len(files))

	// Start worker pool
	fileChan := make(chan string)

	for i := 0; i < numWorkers; i++ {
		fileWg.Add(1)
		go func() {
			defer fileWg.Done()
			for filename := range fileChan {
				file, err := os.Open(filename)
				if err != nil {
					log.Printf("Failed to open file: %v", err)
					continue
				}

				err = myjson.ProcessNDJSONInParallel(file, action)
				file.Close() // close explicitly
				if err != nil {
					log.Printf("Error processing NDJSON: %v", err)
					continue
				}

				// Update progress
				curr := atomic.AddInt64(&processed, 1)
				if curr%10 == 0 || curr == total {
					elapsed := time.Since(start)
					remaining := total - curr
					rate := float64(curr) / elapsed.Seconds()
					eta := time.Duration(float64(remaining)/rate) * time.Second
					log.Printf("Progress: %d/%d | ETA: %s", curr, total, eta.Truncate(time.Second))
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

func main() {
    // Define the action flag (-a or --action)
    action := flag.String("a", "", "action to perform")
    flag.StringVar(action, "action", "", "action to perform")

    // Parse flags
    flag.Parse()

    // After parsing, the remaining args are the file names
    files := flag.Args()	

	go func() {
        log.Println("pprof listening at http://localhost:6060/debug/pprof/")
        if err := http.ListenAndServe("localhost:6060", nil); err != nil {
            log.Fatalf("pprof server error: %v", err)
        }
    }()
	switch *action {
		case "infer":
			infer(files)
		case "testParse":
			testParse(files)
	}
}
