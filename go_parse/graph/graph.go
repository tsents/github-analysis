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
	"encoding/binary"
)

/*
Implements uniform implementations of grpah reads and writes.
*/

type Integer interface {
	~uint | ~int | ~int8 | ~int16 | ~int32 | ~int64 |
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

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

/*
Read graph of (raw test) format. Node Neighbor Neighbor ... 
Each node is seperated by newline and each Neighbor by space.
*/
func ReadNeighborGraph[T Integer](filename string) Graph[T, struct{}] {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encountered %v\n", err);
	}
	defer file.Close()

	userProjects := make(Graph[T, struct{}])
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
        userID, err := parseInteger[T](tokens[0])
        if err != nil {
            continue // skip malformed user ID
        }

        if _, ok := userProjects[userID]; !ok {
            userProjects[userID] = make(map[T]struct{})
        }

        for _, t := range tokens[1:] {
            projID, err := parseInteger[T](t)
            if err == nil {
                userProjects[userID][projID] = struct{}{}
            }
        }
		if (processed % logEvery == 0) {
			elapsed := time.Since(start)
			remaining := totalCount - processed
			rate := float64(processed) / elapsed.Seconds()
			eta := time.Duration(float64(remaining)/rate) * time.Second
			log.Printf("ReadNeighborGraph Progress: %d/%d | ETA: %s\n", processed, totalCount, eta.Truncate(time.Second))
		}	
	}
    return userProjects
}

/*
Function for generic parsing of integer
*/
func parseInteger[T Integer](s string) (T, error) {
	var zero T
	// Use the bit size based on the type
	var bitSize int
	switch any(zero).(type) {
	case int8:
		bitSize = 8
	case int16:
		bitSize = 16
	case int32:
		bitSize = 32
	case int64:
		bitSize = 64
	default:
		bitSize = 0 // Default to int
	}

	i, err := strconv.ParseInt(s, 10, bitSize)
	if err != nil {
		return zero, err
	}
	return T(i), nil
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


/*
Save grpah in format node | deg | neighbor | neighbor ... | node | deg | negihbor ...
*/
func WriteNeighborGraphBinary[T Integer](filename string, graph map[T]map[T]struct{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	processed := 0
	totalCount := len(graph)
	start := time.Now()

	for node, neighbors := range graph {
		degree := T(len(neighbors))

		// Write node
		if err := binary.Write(file, binary.LittleEndian, node); err != nil {
			return err
		}

		// Write degree
		if err := binary.Write(file, binary.LittleEndian, degree); err != nil {
			return err
		}

		// Write neighbors
		for neighbor := range neighbors {
			if err := binary.Write(file, binary.LittleEndian, neighbor); err != nil {
				return err
			}
		}

		processed++
		if (processed % logEvery == 0) {
			elapsed := time.Since(start)
			remaining := totalCount - processed
			rate := float64(processed) / elapsed.Seconds()
			eta := time.Duration(float64(remaining)/rate) * time.Second
			log.Printf("WriteNeighborGraphBinary Progress: %d/%d | ETA: %s\n", processed, totalCount, eta.Truncate(time.Second))
		}	
	}

	return nil
}

//TEST
func ReadNeighborGraphBinary[T Integer](filename string) (map[T]map[T]struct{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	graph := make(map[T]map[T]struct{})
	var node, degree T

	for {
		// Read node
		if err := binary.Read(file, binary.LittleEndian, &node); err != nil {
			break // likely EOF
		}

		// Read degree
		if err := binary.Read(file, binary.LittleEndian, &degree); err != nil {
			return nil, err
		}

		neighbors := make(map[T]struct{}, degree)

		// Read neighbors
		var neighbor T
		for i := T(0); i < degree; i++ {
			if err := binary.Read(file, binary.LittleEndian, &neighbor); err != nil {
				return nil, err
			}
			neighbors[neighbor] = struct{}{}
		}

		graph[node] = neighbors
	}

	return graph, nil
}
