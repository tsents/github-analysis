package myjson
/*
 Implements the collab graph action and maneger.
*/

import "fmt"


//Simillar to other defined, but slimmed to just id's, and no payloads.
type slimEvent struct {
    Actor     slimActor    `json:"actor"`
    Repo      slimRepo     `json:"repo"`
}


type slimActor struct {
    ID           int    `json:"id"`
}

type slimRepo struct {
    ID   int    `json:"id"`
}

func CollabGraphManeger(in <-chan slimEvent) {
	graph := make(map[int]map[int]struct{})
	for entry := range in {
		if graph[entry.Actor.ID] == nil {
			graph[entry.Actor.ID] = make(map[int]struct{})
		}
		graph[entry.Actor.ID][entry.Repo.ID] = struct{}{}
	}
	fmt.Println(graph)
}


// InferFlattenedTypes takes JSON data and returns a flattened map of paths to types.
func CollabGraphAction(data []byte) (slimEvent, error) {
	var slimEventCatch slimEvent;
	if err := json.Unmarshal(data, &slimEventCatch); err != nil {
		fmt.Errorf("SlimEvent Unmarshal error: %w", err)
		return slimEventCatch, err
	}
	return slimEventCatch, nil
}
