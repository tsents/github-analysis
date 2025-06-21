package myjson
/*
 Implements the collab graph action and maneger.
*/

import "fmt"


func CollabGraphManeger(in <-chan SlimEvent) {
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
func CollabGraphAction(data []byte) (SlimEvent, error) {
    var slimEventCatch SlimEvent
	err := slimEventCatch.UnmarshalJSON(data)
    return slimEventCatch, err
}
