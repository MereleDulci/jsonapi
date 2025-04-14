package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
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

	return UnmarshalPatchesSlice(patches, model)
}

func UnmarshalPatchesSlice(patches []PatchOp, model reflect.Type) (out []PatchOp, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from: %w", r.(error))
		}
	}()

	modelVal := reflect.New(model.Elem())
	if modelVal.Kind() != reflect.Ptr && modelVal.Elem().Kind() != reflect.Struct {
		return nil, errors.New("invalid model type")
	}

	for i, patch := range patches {
		if patch.Op == "" {
			return nil, errors.New("invalid patch operation")
		}
		if !slices.Contains([]string{"replace", "test"}, patch.Op) {
			continue //Other operations will need special treatment
		}

		if patch.Path == "" {
			return nil, errors.New("invalid patch operation")
		}

		patchPathParts := strings.Split(strings.TrimPrefix(patch.Path, "/"), "/")

		fieldVal, err := digIn(modelVal, patchPathParts)
		if err != nil {
			return nil, fmt.Errorf("failed to dig into patch path: %w", err)
		}

		if fieldVal.IsValid() {
			unmarshalSingleAttribute(fieldVal, patch.Value)
			patches[i].Value = fieldVal.Interface()
		}
	}

	return patches, nil
}

func digIn(modelVal reflect.Value, pathParts []string) (reflect.Value, error) {
	var fieldVal reflect.Value
	for i := 0; i < modelVal.Elem().NumField(); i++ {
		field := modelVal.Elem().Type().Field(i)
		jsonapiType := getJsonapiFieldType(field)
		if jsonapiType == "attr" || jsonapiType == "" {
			//Only handle explicit and implicit attributes
			//Primary cannot be patched, reference types are managed on the application level
			attrName := getAttributeName(field)
			if attrName == pathParts[0] {
				fieldVal = modelVal.Elem().Field(i)
				break
			}
		}
	}

	if len(pathParts) > 1 {
		if fieldVal.Kind() == reflect.Ptr {
			return digIn(reflect.New(fieldVal.Type().Elem()), pathParts[1:])
		}
		if fieldVal.Kind() == reflect.Struct {
			return digIn(reflect.New(fieldVal.Type()), pathParts[1:])
		}
	}

	return fieldVal, nil
}
