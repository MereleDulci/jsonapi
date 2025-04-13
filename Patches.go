package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type PatchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func UnmarshalPatches(data []byte, model reflect.Type) (patches []PatchOp, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from: %w", r.(error))
		}
	}()

	patches = make([]PatchOp, 0)
	err = json.Unmarshal(data, &patches)
	if err != nil {
		return nil, err
	}

	modelVal := reflect.New(model.Elem())
	if modelVal.Kind() != reflect.Ptr && modelVal.Elem().Kind() != reflect.Struct {
		return nil, errors.New("invalid model type")
	}

	for i, patch := range patches {
		if patch.Path == "" {
			return nil, errors.New("invalid patch operation")
		}

		if patch.Op == "" {
			return nil, errors.New("invalid patch operation")
		}

		var fieldVal reflect.Value
		for i := 0; i < modelVal.Elem().NumField(); i++ {
			field := modelVal.Elem().Type().Field(i)
			attrName := getAttributeName(field)

			if attrName == strings.TrimPrefix(patch.Path, "/") {
				fieldVal = modelVal.Elem().Field(i)
				break
			}
		}

		unmarshalSingleAttribute(fieldVal, patch.Value)
		patches[i].Value = fieldVal.Interface()
	}

	return patches, nil
}
