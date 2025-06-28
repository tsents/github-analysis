package collabgraph

/*
 Implements the collab graph action and maneger.
 And implement a reader for the collab graph from the outputfile.
*/

import (
	"fmt"
	"os"
	"stream-parser/myjson"
)

//Simillar to other defined, but slimmed to just id's, and no payloads.
type slimEvent struct {
    Actor     slimActor    `json:"actor"`
    Repo      slimRepo     `json:"repo"`
}


type slimActor struct {
    ID           uint32    `json:"id"`
}

type slimRepo struct {
    ID   uint32    `json:"id"`
}


const logEvery = 1000000


func BoundGraphManager(outputFile string) myjson.ManagerFunc[slimEvent] {
	return func(in <-chan slimEvent) {
		collabGraphManeger(in, outputFile)
	}
}

func collabGraphManeger(in <-chan slimEvent, outputFile string) {
	graph := make(map[uint32]map[uint32]struct{})
	for entry := range in {
		if graph[entry.Actor.ID] == nil {
			graph[entry.Actor.ID] = make(map[uint32]struct{})
		}
		graph[entry.Actor.ID][entry.Repo.ID] = struct{}{}
	}

	file, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encountered in opening %v: %v\n", outputFile, err);
	}
	defer file.Close()

	for user := range graph {
		_, err = fmt.Fprintf(file, fmt.Sprintf("%v", user))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encountered in Fprintf, at collabGraphManeger %v\n", err);
		}	
		for repo := range graph[user] {
			// Write string to file
			_, err = fmt.Fprintf(file, fmt.Sprintf(" %v", repo))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encountered in Fprintf, at collabGraphManeger %v\n", err);
			}	
		}
		_, err = fmt.Fprintln(file, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encountered in Fprintf, at collabGraphManeger %v\n", err);
		}	
	}
}


// InferFlattenedTypes takes JSON data and returns a flattened map of paths to types.
func CollabGraphAction(data []byte) (slimEvent, error) {
	var slimEventCatch slimEvent;
	if err := myjson.Myjson.Unmarshal(data, &slimEventCatch); err != nil {
		fmt.Println("SlimEvent Unmarshal error: ", err)
		return slimEventCatch, err
	}
	return slimEventCatch, nil
}

