package helpers

import (
	"fmt"
	"reflect"
)

func AnytiseData(input any) any {
	if input != nil {
		inputVal := reflect.ValueOf(input)
		if inputVal.Kind() == reflect.Map {
			result := make(map[string]any)
			for _, key := range inputVal.MapKeys() {
				result[key.String()] = inputVal.MapIndex(key).Interface()
			}
			return result
		} else if inputVal.Kind() == reflect.Slice {
			result := make([]any, inputVal.Len())
			for i := 0; i < inputVal.Len(); i++ {
				result[i] = inputVal.Index(i).Interface()
			}
			return result
		} else {
			fmt.Println("Unsupported map type")
		}
	}
	return nil
}
