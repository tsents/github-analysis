package collabgraph

/*
 Implements the collab graph action and maneger.
 And implement a reader for the collab graph from the outputfile.
*/

import (
	"fmt"
	"stream-parser/graph"
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



func CollabGraphManeger(in <-chan slimEvent) graph.Graph[uint32, struct{}] {
	graph := make(graph.Graph[uint32, struct{}])
	for entry := range in {
		if graph[entry.Actor.ID] == nil {
			graph[entry.Actor.ID] = make(map[uint32]struct{})
		}
		graph[entry.Actor.ID][entry.Repo.ID] = struct{}{}
	}
	return graph
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

