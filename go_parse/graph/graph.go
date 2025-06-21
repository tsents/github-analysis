package graph

import (
	"os"
	"fmt"
	"time"
	"log"
	"bufio"
	"io"
	"strings"
	"strconv"
)

/*
Implements uniform implementations of grpah reads and writes.
*/

type Graph[T comparable,U any] map[T]map[T]U


const logEvery = 100000

/*
Outputs a graph in the format of Vertex Neighbor Neighbor... seperated by newline.
*/
func NeighborOutputGraph[T comparable,U any](outputFile string, graph Graph[T,U]) {
	file, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encountered %v\n", err);
	}
	defer file.Close()

	for user := range graph {
		_, err = fmt.Fprintf(file, fmt.Sprintf("%v", user))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encountered %v\n", err);
		}	
		for repo := range graph[user] {
			// Write string to file
			_, err = fmt.Fprintf(file, fmt.Sprintf(" %v", repo))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encountered %v\n", err);
			}	
		}
		_, err = fmt.Fprintln(file, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encountered %v\n", err);
		}	
	}
}

func ReadNeighborGraph(filename string) Graph[int, struct{}] {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encountered %v\n", err);
	}
	defer file.Close()

	userProjects := make(Graph[int, struct{}])
    reader := bufio.NewReader(file)

	start := time.Now()
	processed := 0
	totalCount, _ := countLines(filename)	

    for {
        line, err := reader.ReadString('\n')
        if len(line) == 0 && err != nil {
            break // EOF or error
        }

		processed++

        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }

        tokens := strings.Fields(line)
        userID, err := strconv.Atoi(tokens[0])
        if err != nil {
            continue // skip malformed user ID
        }

        if _, ok := userProjects[userID]; !ok {
            userProjects[userID] = make(map[int]struct{})
        }

        for _, t := range tokens[1:] {
            projID, err := strconv.Atoi(t)
            if err == nil {
                userProjects[userID][projID] = struct{}{}
            }
        }
		if (processed % logEvery == 0) {
			elapsed := time.Since(start)
			remaining := totalCount - processed
			rate := float64(processed) / elapsed.Seconds()
			eta := time.Duration(float64(remaining)/rate) * time.Second
			log.Printf("Read & Prase Progress: %d/%d | ETA: %s\n", processed, totalCount, eta.Truncate(time.Second))
		}	
	}
    return userProjects
}


func countLines(filename string) (int, error) {
    file, err := os.Open(filename)
    if err != nil {
        return 0, err
    }
    defer file.Close()

    buf := make([]byte, 1024*1024) // 1MB buffer
    count := 0

    for {
        n, err := file.Read(buf)
        if n == 0 {
            break
        }

        for _, b := range buf[:n] {
            if b == '\n' {
                count++
            }
        }

        if err != nil {
            if err == io.EOF {
                break
            }
            return count, err
        }
    }
    return count, nil
}


