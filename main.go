package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"stream-parser/graph"
	"stream-parser/myjson"
	collabgraph "stream-parser/myjson/collab_graph"
)


func collabGraph(files []string, inputType string) graph.Graph[uint32, struct{}]{
	manager := collabgraph.CollabGraphManeger
	collabGraph, err := myjson.ParseInParallel(files, collabgraph.CollabGraphAction, manager, inputType)
	if (err != nil) {
		fmt.Fprintf(os.Stderr, "Error encounted collabGraph: %s\n", err)
		return nil
	}
	return collabGraph
}

func weightedCollabGraph(files []string, inputType string) graph.Graph[uint32, uint32]{
	manager := collabgraph.WeightedCollabGraphManeger
	collabGraph, err := myjson.ParseInParallel(files, collabgraph.CollabGraphAction, manager, inputType)
	if (err != nil) {
		fmt.Fprintf(os.Stderr, "Error encounted collabGraph: %s\n", err)
		return nil
	}
	return collabGraph
}

func main() {
    // Define the action flag (-a or --action)
    action := flag.String("a", "", "action to perform")
    flag.StringVar(action, "action", "", "action to perform")


    output := flag.String("o", "", "action to perform")
    flag.StringVar(output, "output", "", "action to perform")


    inputType := flag.String("t", "file", "the type of input (http/file)\ndefault is file")
    flag.StringVar(inputType, "type", "file", "the type of input (http/file)\ndefualt is file")

    // Parse flags
    flag.Parse()

    // After parsing, the remaining args are the file names
    files := flag.Args()	

	go func() {
        fmt.Println("pprof listening at http://localhost:6060/debug/pprof/")
        if err := http.ListenAndServe("localhost:6060", nil); err != nil {
            fmt.Fprintf(os.Stderr, "pprof server error: %v", err)
        }
    }()

	if (*output == "") {
		*output = os.DevNull
		fmt.Println("WARNING: no output file given, output will be directed to /dev/null")
	}

	switch *action {
		case "collabGraph":
			outputGraph := collabGraph(files, *inputType)
			graph.NeighborOutputGraph(*output, outputGraph)
		case "weightedCollabGraph":
			outputGraph := weightedCollabGraph(files, *inputType)
			graph.EdgeListOutputGraph(*output, outputGraph)
		default:
			fmt.Println("Action not found")
			return
	}
}

