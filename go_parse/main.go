package main

import (
	"flag"
	"fmt"
	"os"
	"parser/myjson"

	"net/http"
	_ "net/http/pprof"
)


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


func infer(files []string) {
	myjson.ParseInParallel(files, myjson.InferFlattenedTypes, myjson.InferManeger)
}

func collabGraph(files []string, output string) {
	manager := myjson.BoundGraphManager(output)
	myjson.ParseInParallel(files, myjson.CollabGraphAction, manager)
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
		case "infer":
			infer(files)
		case "testParse":
			testParse(files)
		case "collabGraph":
			collabGraph(files, *output)
	}
}

