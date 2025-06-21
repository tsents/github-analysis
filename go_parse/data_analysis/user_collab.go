package data_analysis

import (
	"log"
	"time"
	"parser/graph"
)


const logEvery = 1000000


/*
This contains a script that takes the collab graph int the format of User -> Repos
into a user collab graph of User -> User based on shared repos.

in terms of implementation we can do an unweighted version based on just if there
is any interaction. or a weighted based on the number of shared repos.

there are more options such as based on the number of interactions within the repo as well
but this is not implemented yet and might not be as usefull (for example, what if this is a very
active repo? will you now after reading all, normalize to the number of interaction?)
*/

type userGraph = graph.Graph[int, struct{}] // a graph of user <-> user
type collabGraph = graph.Graph[int, struct{}] // a graph of user -> repos
type repoGraph = graph.Graph[int, struct{}] // A graph of repo -> users


func ReadCollabGraphToUserGraph(filename string) userGraph {	
	var collabGraph userGraph = graph.ReadNeighborGraph(filename)
	var repoGraph repoGraph = ConvertCollabToRepoGraph(collabGraph)
	return ConvertRepoToUserGraph(repoGraph)
}

func ReadCollabToRepoGraph(filename string) repoGraph {
	var collabGraph userGraph = graph.ReadNeighborGraph(filename)
	return ConvertCollabToRepoGraph(collabGraph)
}

func ConvertCollabToRepoGraph(graph collabGraph) repoGraph {
	repoGraph := make(repoGraph)
	
	start := time.Now()
	processed := 0
	totalCount := len(graph)

	for user := range graph {
		for repo := range graph[user] {
			if _, ok := repoGraph[repo]; !ok { // if not repo in repoGraph
				repoGraph[repo] = make(map[int]struct{})
			}
			repoGraph[repo][user] = struct{}{}
		}

		processed++

		if (processed % logEvery == 0) {
			elapsed := time.Since(start)
			remaining := totalCount - processed
			rate := float64(processed) / elapsed.Seconds()
			eta := time.Duration(float64(remaining)/rate) * time.Second
			log.Printf("ConvertCollabToRepoGraph Progress: %d/%d | ETA: %s\n", processed, totalCount, eta.Truncate(time.Second))
		}	

	}
	return repoGraph
}

func ConvertRepoToUserGraph(graph repoGraph) userGraph {
	myUserGraph := make(userGraph)
	for repo := range graph {
		for user := range graph[repo] {
			if _, ok := myUserGraph[user]; !ok { // if not repo in repoGraph
				myUserGraph[user] = make(map[int]struct{})
			}

			for otherUser := range graph[repo] {
				if (user > otherUser) { // Avoid multiple writes
					myUserGraph[user][otherUser] = struct{}{}
				}
			}
		}
	} 
	return myUserGraph
}
