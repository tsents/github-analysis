package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"parser/data_analysis"
	"parser/graph"
	"parser/myjson"
	collabgraph "parser/myjson/collab_graph"
)

// Used to test empty manager and action
func testParse(files []string) {
	action := func(data []byte) (any, error) {
		value, err := myjson.UnmarshalPayload(data)
		if (err != nil) {
			fmt.Fprintf(os.Stderr, "Error encountered %v\n", err);
			return "", err
		}
		return value, nil
	}

	emtpyManeger := func(in <-chan any) {
		for _ = range in {}
	}
	myjson.ParseInParallel(files, action, emtpyManeger)
}



func collabGraph(files []string, output string) {
	manager := collabgraph.BoundGraphManager(output)
	myjson.ParseInParallel(files, collabgraph.CollabGraphAction, manager)
}

//DONT USE THIS
func userGraphFromCollab(file string) {
	graph := data_analysis.ReadCollabGraphToUserGraph(file)	
	_ = graph
}

//Not very usefull, but exsits
func repoGraphFromCollab(file string, output string) {
	repoGraph := data_analysis.ReadCollabToRepoGraph(file)
	err := graph.WriteNeighborGraphBinary(output, repoGraph)
	if (err != nil) {
		fmt.Fprintf(os.Stderr, "Error encounted in WriteGraphToFile %v\n", err)
	}
}

//DONT USE THIS. user graph is bad
func userGraphFromRepo(file string, output string) {
	graph, err := data_analysis.ConvertRepoFileToUserGraph(file)
	if (err != nil) {
		fmt.Fprintf(os.Stderr, "Error encounted in ConvertRepoFileToUserGraph %v\n", err)
	}
	_ = graph
}

func main() {
    // Define the action flag (-a or --action)
    action := flag.String("a", "", "action to perform")
    flag.StringVar(action, "action", "", "action to perform")


    output := flag.String("o", "", "action to perform")
    flag.StringVar(output, "output", "", "action to perform")

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
	switch *action {
		case "testParse":
			testParse(files)
		case "collabGraph":
			collabGraph(files, *output)
		case "repoGraphFromCollab":
			repoGraphFromCollab(files[0], *output)
		default:
			fmt.Println("Action not found")
	}
}

