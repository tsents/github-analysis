package collabgraph

/*
 Implements the collab graph action and maneger.
 And implement a reader for the collab graph from the outputfile.
*/

import (
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


/*
Takes in parsed slimEvents, that hold any user <-> repo interaction.
then generates a 2-partite graph of form graph[user] = repoSet.
returns the unweighted graph of format Graph[uint32, void]
*/
func CollabGraphManeger(in <-chan myjson.ParsedEntry[slimEvent]) graph.Graph[uint32, struct{}] {
	graph := make(graph.Graph[uint32, struct{}])
	for entry := range in {
		entryData := entry.Data
		if graph[entryData.Actor.ID] == nil {
			graph[entryData.Actor.ID] = make(map[uint32]struct{})
		}
		graph[entryData.Actor.ID][entryData.Repo.ID] = struct{}{}
	}
	return graph
}

// A weighted version of the CollabGraphManeger, counting how many time a
// user <-> repo interaction was held. for further information refer to CollabGraphManeger.
func WeightedCollabGraphManeger(in <-chan myjson.ParsedEntry[slimEvent]) graph.Graph[uint32, uint32] {
	graph := make(graph.Graph[uint32, uint32])
	for entry := range in {
		entryData := entry.Data
		if graph[entryData.Actor.ID] == nil {
			graph[entryData.Actor.ID] = make(map[uint32]uint32)
		}
		graph[entryData.Actor.ID][entryData.Repo.ID] += 1
	}
	return graph
}

