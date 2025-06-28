package data_analysis

import (
	"log"
	"time"
	"stream-parser/graph"
	"os"
	"encoding/binary"
)


const logEvery = 1000000
const logEveryBytes = 1000


/*
This contains a script that takes the collab graph int the format of User -> Repos
into a user collab graph of User -> User based on shared repos.

in terms of implementation we can do an unweighted version based on just if there
is any interaction. or a weighted based on the number of shared repos.

there are more options such as based on the number of interactions within the repo as well
but this is not implemented yet and might not be as usefull (for example, what if this is a very
active repo? will you now after reading all, normalize to the number of interaction?)

IMPORTATNT - USER <-> USER graphs won't be used! this is becuase in fact
they contain high amount of cliques adn thus have VERY high degrees.
*/

type userGraph = graph.Graph[uint32, struct{}] // a graph of user <-> user
type collabGraph = graph.Graph[uint32, struct{}] // a graph of user -> repos
type repoGraph = graph.Graph[uint32, struct{}] // A graph of repo -> users


func ReadCollabGraphToUserGraph(filename string) userGraph {	
	var collabGraph userGraph = graph.ReadNeighborGraph[uint32](filename)
	var repoGraph repoGraph = ConvertCollabToRepoGraph(collabGraph)
	return ConvertRepoToUserGraph(repoGraph)
}

func ReadCollabToRepoGraph(filename string) repoGraph {
	var collabGraph userGraph = graph.ReadNeighborGraph[uint32](filename)
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
				repoGraph[repo] = make(map[uint32]struct{})
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
				myUserGraph[user] = make(map[uint32]struct{})
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
/*
This implementation is more optimized then just reading the whole
repo graph and then converting. becuase no saving of the repograph is needed.
*/
func ConvertRepoFileToUserGraph(filename string) (userGraph, error) {

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	graph := make(userGraph)
	var node, degree uint32

	var processed uint32 = 0;
	fileInfo, err := file.Stat()
	fileSize := fileInfo.Size() // int64
	start := time.Now()
	
	for {
		// Read node
		if err := binary.Read(file, binary.LittleEndian, &node); err != nil {
			break // likely EOF
		}

		// Read degree
		if err := binary.Read(file, binary.LittleEndian, &degree); err != nil {
			return nil, err
		}

		var neighbors []uint32 = make([]uint32, degree)

		// Read neighbors
		var neighbor uint32
		for i := uint32(0); i < degree; i++ {
			if err := binary.Read(file, binary.LittleEndian, &neighbor); err != nil {
				return nil, err
			}
			neighbors[i] = neighbor
		}
		
		for i := range neighbors {
			for j := range neighbors {
				if (neighbors[i] > neighbors[j]) {
					if _, ok := graph[neighbors[i]]; !ok {
						graph[neighbors[i]] = map[uint32]struct{}{}
					}
					graph[neighbors[i]][neighbors[j]] = struct{}{}
				}
			}
		}

		processed += (1 + 1 + degree) * 4 //Node | degree | neighbors * (uint32 in bytes [4])
		if (processed % logEveryBytes == 0) {
			elapsed := time.Since(start)
			remaining := uint32(fileSize) - processed
			rate := float64(processed) / elapsed.Seconds()
			eta := time.Duration(float64(remaining)/rate) * time.Second
			log.Printf("ConvertCollabToRepoGraph Progress: %d/%d | ETA: %s\n", processed, fileSize, eta.Truncate(time.Second))
		}	

	}

	return graph, nil
}
