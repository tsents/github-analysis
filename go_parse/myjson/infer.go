package myjson

import (
	"sort"
	"bytes"
	"fmt"
	"os"
	"reflect"
)

/*
This is very costly. because although they are not many diffrent core types,
they ARE diffrent. there are many more close types that are counted as "diffrent".
sadly we might have to manualy create the types we can have.
*/
func InferManeger(in <-chan string) {
	uniqueEntries := make(map[string]struct{})
	for entry := range in {
		uniqueEntries[entry] = struct{}{}
	}
	for entry := range uniqueEntries {
	    file, err := os.OpenFile("types", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encountered %v\n", err);
		}
		defer file.Close()

		// Write string to file
		_, err = fmt.Fprintf(file, fmt.Sprintf("%v\n", entry))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encountered %v\n", err);
		}	
	}
}


// InferFlattenedTypes takes JSON data and returns a flattened map of paths to types.
func InferFlattenedTypes(data []byte) (string, error) {
    var root any
    if err := json.Unmarshal(data, &root); err != nil {
        return "", fmt.Errorf("unmarshal error: %w", err)
    }

    flatMap := make(map[string]string)
    flatten("", root, flatMap)

    // Sort keys and marshal deterministically
    keys := make([]string, 0, len(flatMap))
    for k := range flatMap {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    buffer := &bytes.Buffer{}
    buffer.WriteByte('{')
    for i, k := range keys {
        if i > 0 {
            buffer.WriteByte(',')
        }
        keyJSON, _ := json.Marshal(k)
        valJSON, _ := json.Marshal(flatMap[k])
        buffer.Write(keyJSON)
        buffer.WriteByte(':')
        buffer.Write(valJSON)
    }
    buffer.WriteByte('}')

    return buffer.String(), nil
}


func flatten(path string, v any, out map[string]string) {
	switch val := v.(type) {
	case map[string]any:
		for k, v2 := range val {
			fullKey := k
			if path != "" {
				fullKey = path + "." + k
			}
			flatten(fullKey, v2, out)
		}
	case []any:
		for i, item := range val {
			indexedPath := fmt.Sprintf("%s[%d]", path, i)
			flatten(indexedPath, item, out)
		}
	default:
		if path != "" {
			out[path] = typeOf(val)
		}
	}
}

func typeOf(v any) string {
	switch v.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "bool"
	case nil:
		return "null"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	default:
		return reflect.TypeOf(v).String()
	}
}

