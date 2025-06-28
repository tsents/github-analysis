package myjson

import (
	"fmt"
	"testing"
)

func ExampleParseInParallel() {
	var action ActionFunc[string] = func(line []byte) (string, error){
		return string(line), nil
	}
	var manager ManagerFunc[string] = func(in <-chan string) {
		var combined string = ""; 
		for entry := range in {
			combined += string(entry[0])
		} 
		fmt.Println(combined)
	}
	ParseInParallel([]string{"https://data.gharchive.org/2025-05-15-15.json.gz"}, action, manager, "http")
}

func TestParseInParallel(t *testing.T) {
	ExampleParseInParallel()
}
