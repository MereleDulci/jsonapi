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
			err = fmt.Errorf("[jsonapi.UnmarshalPatchesSlice] recovered from: %w", r.(error))
		}
	}()

	modelVal := reflect.New(model.Elem())
	if modelVal.Kind() != reflect.Ptr && modelVal.Elem().Kind() != reflect.Struct {
		return nil, errors.New("invalid model type")
	}

	for i, patch := range patches {
		switch patch.Op {
		case "replace", "test":
			if patch.Path == "" {
				return nil, errors.New("invalid patch operation - empty path")
			}

			patchPathParts := strings.Split(strings.TrimPrefix(patch.Path, "/"), "/")

			fieldVal, jsonapiType, err := digIn(modelVal, patchPathParts)
			if err != nil {
				return nil, fmt.Errorf("failed to dig into patch path: %w", err)
			}

			if fieldVal.IsValid() {
				if jsonapiType == "relation" {
					fieldPrimitiveType := fieldVal.Type()
					fieldPrimitiveVal := fieldVal

					if fieldVal.Kind() == reflect.Slice {
						fieldPrimitiveType = fieldPrimitiveType.Elem()
						if fieldPrimitiveType.Kind() == reflect.Ptr {
							//fieldPrimitiveType is []T, need to step into T
							fieldPrimitiveType = fieldPrimitiveType.Elem()
						}
						fieldPrimitiveVal = reflect.New(fieldPrimitiveType).Elem()

						//Should iterate over all elements of the slice one by one
						list, ok := patch.Value.([]interface{})
						if !ok {
							return nil, errors.New("invalid patch operation - cannot target slice with non-slice value with replace")
						}
						idFieldType, idFieldVal, resourceName := getIdFieldVal(fieldPrimitiveType, fieldPrimitiveVal)
						for i, item := range list {
							if err := unmarshalID(idFieldType, idFieldVal, item, resourceName); err != nil {
								return nil, fmt.Errorf("failed to unmarshal referenced ID: %w", err)
							}
							list[i] = idFieldVal.Interface()
						}
						patches[i].Value = list
					} else {
						idFieldType, idFieldVal, resourceName := getIdFieldVal(fieldPrimitiveType, fieldPrimitiveVal)
						if err := unmarshalID(idFieldType, idFieldVal, patch.Value, resourceName); err != nil {
							return nil, fmt.Errorf("failed to unmarshal referenced ID: %w", err)
						}
						patches[i].Value = idFieldVal.Interface()
					}
				} else {
					unmarshalSingleAttribute(fieldVal, patch.Value)
					patches[i].Value = fieldVal.Interface()
				}
			}
		case "add":
			//Applicable to lists only, so need to validate the targets
			patchPathParts := strings.Split(strings.TrimPrefix(patch.Path, "/"), "/")
			fieldVal, jsonapiType, err := digIn(modelVal, patchPathParts)
			if err != nil {
				return nil, fmt.Errorf("failed to dig into patch path: %w", err)
			}

			if fieldVal.IsValid() {
				fieldType := fieldVal.Type()
				isPtr := fieldType.Kind() == reflect.Ptr

				if isPtr {
					fieldType = fieldType.Elem()
				}

				if fieldType.Kind() == reflect.Slice {
					fieldPrimitiveType := fieldVal.Type().Elem()
					if isPtr {
						//fieldPrimitiveType is []T, need to step into T
						fieldPrimitiveType = fieldPrimitiveType.Elem()
					}
					fieldPrimitiveVal := reflect.New(fieldPrimitiveType).Elem()

					if jsonapiType == "relation" {
						//In case of relation we can only receive id value, but the target type is unknown
						//fieldPrimitiveVal is now relation value which is a nil struct that should have a primary field somewhere
						//Should be similar to modelType and modelVal in unmarshalOne here
						idFieldType, idFieldVal, resourceName := getIdFieldVal(fieldPrimitiveType, fieldPrimitiveVal)
						if err := unmarshalID(idFieldType, idFieldVal, patch.Value, resourceName); err != nil {
							return nil, fmt.Errorf("failed to unmarshal referenced ID: %w", err)
						}
						patches[i].Value = idFieldVal.Interface()
					} else {
						unmarshalSingleAttribute(fieldPrimitiveVal, patch.Value)
						patches[i].Value = fieldPrimitiveVal.Interface()
					}
				} else {
					return nil, errors.New("invalid patch operation - target field is not a slice")
				}
			} else {
				return nil, fmt.Errorf("invalid patch operation - cannot target %s", patch.Path)
			}
		case "":
			return nil, errors.New("invalid patch operation - empty op")
		default:
			continue //Stepping over other options for now for compatibility
		}
	}

	return patches, nil
}

func digIn(modelVal reflect.Value, pathParts []string) (reflect.Value, string, error) {
	var fieldVal reflect.Value
	var jsonapiType string
	for i := 0; i < modelVal.Elem().NumField(); i++ {
		field := modelVal.Elem().Type().Field(i)
		jsonapiType = getJsonapiFieldType(field)
		if jsonapiType == "" {
			jsonapiType = "attr" //Defaults to attr if not set or only json tag is provided
		}

		//Only handle explicit and implicit attributes
		//Primary cannot be patched, reference types are managed on the application level
		attrName := getAttributeName(field)
		if attrName == pathParts[0] {
			fieldVal = modelVal.Elem().Field(i)
			break
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

	return fieldVal, jsonapiType, nil
}

func getIdFieldVal(modelType reflect.Type, modelVal reflect.Value) (reflect.StructField, reflect.Value, string) {
	if modelType.Kind() == reflect.Ptr { //Unwrap potential pointer
		modelType = modelType.Elem()
		modelVal = reflect.New(modelType).Elem()
	}

	for i := 0; i < modelType.NumField(); i++ {
		fieldType := modelType.Field(i)
		fieldVal := modelVal.Field(i)

		tag := fieldType.Tag.Get("jsonapi")
		if tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] == "primary" {
				return fieldType, fieldVal, parts[1]
			}
		}
	}

	//No primary field found on relationship, has to be a configuration mistake so panic here is appropriate
	panic(errors.New("no primary field found on relationship field type"))
}
