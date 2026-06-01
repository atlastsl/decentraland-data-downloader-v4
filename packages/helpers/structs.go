package helpers

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
)

func ConvertMapToStruct(m map[string]any, target interface{}) error {
	config := &mapstructure.DecoderConfig{
		ErrorUnused: false,
		Result:      target,
	}
	decoder, e1 := mapstructure.NewDecoder(config)
	if e1 != nil {
		return e1
	}
	e2 := decoder.Decode(m)
	if e2 != nil {
		return e2
	}
	return nil
}

func ConvertAnyToStruct(a any, target interface{}) error {
	config := &mapstructure.DecoderConfig{
		ErrorUnused: false,
		Result:      target,
	}
	decoder, e1 := mapstructure.NewDecoder(config)
	if e1 != nil {
		return e1
	}
	e2 := decoder.Decode(a)
	if e2 != nil {
		return e2
	}
	return nil
}

func ConvertStructToMap(s any, target *map[string]any) error {
	jsonStr, err := json.Marshal(s)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonStr, target)
	if err != nil {
		return err
	}
	return nil
}
