package myjson

import (
	"fmt"
	"testing"
)


func ExampleParseInParallel() {
	var manager ManagerFunc[BaseEvent, struct{}] = func(in <-chan BaseEvent) struct{}{
		for v := range in {
			fmt.Println(v)
		}	
		return struct{}{}
	}
	fmt.Println(ParseInParallel([]string{"https://data.gharchive.org/2025-05-15-15.json.gz"}, manager, "http"))
}

func TestParseInParallel(t *testing.T) {
	ExampleParseInParallel()
}
