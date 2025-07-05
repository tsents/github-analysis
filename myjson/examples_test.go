package myjson

import (
	"fmt"
	"testing"
)

func ExampleParseInParallel() {
	var action ActionFunc[string] = func(line []byte) (string, error){
		return string(line), nil
	}
	var manager ManagerFunc[string, string] = func(in <-chan string) string{
		var combined string = ""; 
		for entry := range in {
			combined += string(entry[0])
		} 
		return combined
	}
	fmt.Println(ParseInParallel([]string{"https://data.gharchive.org/2025-05-15-15.json.gz"}, action, manager, "http"))
}

func TestParseInParallel(t *testing.T) {
	ExampleParseInParallel()
}
