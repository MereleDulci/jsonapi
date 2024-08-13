package jsonapi

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

func UnmarshalManyAsType(payload []byte, model reflect.Type) ([]interface{}, error) {
	raw := map[string]interface{}{}
	err := json.Unmarshal(payload, &raw)
	if err != nil {
		return nil, err
	}

	included, ok := raw["included"].([]interface{})
	if !ok {
		included = make([]interface{}, 0)
	}

	data, ok := raw["data"].([]interface{})
	if !ok {
		return nil, errors.New("invalid data structure")
	}
	models := make([]interface{}, 0)

	for _, resource := range data {
		resourceData, ok := resource.(map[string]interface{})
		if !ok {
			return nil, errors.New("invalid data structure")
		}

		out := reflect.New(model.Elem()).Interface()

		err = unmarshalOne(resourceData, out, included)
		if err != nil {
			return nil, err
		}

		models = append(models, out)
	}

	return models, nil
}

func UnmarshalOneAsType(payload []byte, model reflect.Type) (interface{}, error) {
	raw := map[string]interface{}{}
	err := json.Unmarshal(payload, &raw)
	if err != nil {
		return nil, err
	}

	included, ok := raw["included"].([]interface{})
	if !ok {
		included = make([]interface{}, 0)
	}

	data, ok := raw["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid data structure")
	}
	modelVal := reflect.New(model.Elem())
	if modelVal.Kind() != reflect.Ptr && modelVal.Elem().Kind() != reflect.Struct {
		return nil, errors.New("invalid model type")
	}

	out := reflect.New(model.Elem()).Interface()
	err = unmarshalOne(data, out, included)
	if err != nil {
		return nil, err
	}
	return out, nil

}

func Unmarshal(data []byte, model interface{}) error {
	raw := map[string]interface{}{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	included, ok := raw["included"].([]interface{})
	if !ok {
		included = []interface{}{}
	}

	switch raw["data"].(type) {
	case map[string]interface{}:
		err = unmarshalOne(raw["data"].(map[string]interface{}), model, included)
		if err != nil {
			return err
		}
	case []interface{}:
		data, ok := raw["data"].([]interface{})
		if !ok {
			return errors.New("invalid data structure")
		}
		if reflect.TypeOf(model).Elem().Kind() != reflect.Slice {
			return errors.New("invalid model type")
		}

		if len(data) == 0 {
			return nil
		}
		acc := reflect.ValueOf(model).Elem()
		modelVal := reflect.ValueOf(model).Elem().Type().Elem()
		asPtr := modelVal.Kind() == reflect.Ptr
		if asPtr {
			modelVal = modelVal.Elem()
		}

		for _, resource := range data {
			resourceData, ok := resource.(map[string]interface{})
			if !ok {
				return errors.New("invalid data structure")
			}

			out := reflect.New(modelVal).Interface()

			err = unmarshalOne(resourceData, out, included)
			if err != nil {
				return err
			}

			if asPtr {
				acc.Set(reflect.Append(acc, reflect.ValueOf(out)))
			} else {
				acc.Set(reflect.Append(acc, reflect.ValueOf(out).Elem()))
			}
		}

	default:
		return errors.New("invalid data structure")
	}

	return nil
}

func unmarshalOne(data map[string]interface{}, model interface{}, included []interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from: %w", r.(error))
		}
	}()

	modelVal := reflect.ValueOf(model)
	modelType := reflect.TypeOf(model)

	if modelType.Kind() == reflect.Ptr {
		modelVal = modelVal.Elem()
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		return errors.New(fmt.Sprintf("model should be a struct to unmarshal single resource, got %s", modelType.Kind()))
	}

	resourceType := data["type"]
	resourceID := data["id"]
	resourceAttributes, attributesValid := data["attributes"].(map[string]interface{})

	resourceRelationships, relationshipsValid := data["relationships"].(map[string]interface{})

	for i := 0; i < modelType.NumField(); i++ {
		fieldType := modelType.Field(i)
		fieldVal := modelVal.Field(i)

		jsonapitag := fieldType.Tag.Get("jsonapi")
		if jsonapitag != "" && resourceID != nil {
			parts := strings.Split(jsonapitag, ",")
			if len(parts) > 1 {
				switch parts[0] {
				case "primary":
					//Check provided model type matches
					if parts[1] != resourceType {
						return errors.New("resource type does not match model type")
					}
					stringMarshallerType := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

					switch {
					case fieldVal.Type().Implements(stringMarshallerType):
						v := reflect.New(fieldVal.Type().Elem())
						v.MethodByName("UnmarshalText").
							Call([]reflect.Value{reflect.ValueOf([]byte(resourceID.(string)))})
						fieldVal.Set(v)
					case reflect.PointerTo(fieldVal.Type()).Implements(stringMarshallerType):
						v := reflect.New(fieldVal.Type())
						v.MethodByName("UnmarshalText").
							Call([]reflect.Value{reflect.ValueOf([]byte(resourceID.(string)))})
						fieldVal.Set(v.Elem())
					default:
						switch fieldVal.Kind() {
						case reflect.String:
							fieldVal.SetString(resourceID.(string))
						case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
							fieldVal.SetInt(int64(resourceID.(float64)))
						case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
							fieldVal.SetUint(uint64(resourceID.(float64)))
						default:

							return errors.New("ID field must be a string, number or implement encoding.TextUnmarshaler")
						}

					}
					continue //to process other fields
				}
			}
		}

		if attributesValid {
			unmarshalAttributes(fieldType, fieldVal, resourceAttributes)
		}
		if relationshipsValid {
			if err := unmarshalRelationships(fieldType, fieldVal, resourceRelationships, included); err != nil {
				return err
			}
		}

	}

	return nil
}

func unmarshalAttributes(fieldType reflect.StructField, fieldVal reflect.Value, resourceAttributes map[string]interface{}) {
	attributeName := getAttributeName(fieldType)

	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Errorf("unmarshal attribute %s: %w", attributeName, r.(error)))
		}
	}()
	if attribute, ok := resourceAttributes[attributeName]; ok {
		unmarshalSingleAttribute(fieldVal, attribute)
	}
}

func unmarshalSingleAttribute(fieldVal reflect.Value, attribute interface{}) {

	switch fieldVal.Kind() {
	case reflect.Struct:
		isHandled, err := unmarshalTime(fieldVal, attribute)
		if err != nil {
			panic(err)
		}
		if isHandled {
			return
		}
		isHandled, err = unmarshalUnmarshaler(fieldVal, attribute)
		if err != nil {
			panic(err)
		}
		if isHandled {
			return
		}

		var toFillIn = reflect.New(fieldVal.Type())

		for i := 0; i < toFillIn.Elem().NumField(); i++ {
			nextFieldType := toFillIn.Elem().Type().Field(i)
			nextFieldVal := toFillIn.Elem().Field(i)
			unmarshalAttributes(nextFieldType, nextFieldVal, attribute.(map[string]interface{}))
		}

		fieldVal.Set(toFillIn.Elem())
	case reflect.Pointer:

		isHandled, err := unmarshalTime(fieldVal, attribute)
		if err != nil {
			panic(err)
		}
		if isHandled {
			return
		}
		isHandled, err = unmarshalUnmarshaler(fieldVal, attribute)
		if err != nil {
			panic(err)
		}
		if isHandled {
			return
		}
		toFillIn := reflect.New(fieldVal.Type().Elem())

		if fieldVal.Type().Elem().Kind() == reflect.Struct {
			for i := 0; i < toFillIn.Elem().NumField(); i++ {
				nextFieldType := toFillIn.Elem().Type().Field(i)
				nextFieldVal := toFillIn.Elem().Field(i)
				unmarshalAttributes(nextFieldType, nextFieldVal, attribute.(map[string]interface{}))
			}
		} else {
			toFillIn.Elem().Set(castPrimitive(fieldVal.Type().Elem().Kind(), fieldVal.Type(), attribute))
		}

		fieldVal.Set(toFillIn)
	case reflect.Slice:
		dataSlice, ok := attribute.([]interface{})
		if !ok {
			//TODO form an error
			return
		}
		fieldValueKind := fieldVal.Type().Elem().Kind()
		fieldValueType := fieldVal.Type().Elem()

		reflection := reflect.MakeSlice(reflect.SliceOf(fieldValueType), 0, 0)
		reflectionValue := reflect.New(reflection.Type())
		reflectionValue.Elem().Set(reflection)

		slicePtr := reflect.ValueOf(reflectionValue.Interface())
		sliceValuePtr := slicePtr.Elem()

		for _, datapoint := range dataSlice {
			var primitiveVal reflect.Value

			if fieldValueKind == reflect.Ptr {
				primitiveVal = castPrimitivePointer(fieldValueType.Elem().Kind(), fieldValueType, datapoint)
			} else {
				primitiveVal = castPrimitive(fieldValueKind, fieldValueType, datapoint)
			}

			sliceValuePtr.Set(reflect.Append(sliceValuePtr, primitiveVal))
		}

		fieldVal.Set(reflect.ValueOf(sliceValuePtr.Interface()))
	default:
		fieldVal.Set(castPrimitive(fieldVal.Kind(), fieldVal.Type(), attribute))
	}

}

func unmarshalTime(fieldVal reflect.Value, attribute interface{}) (bool, error) {

	switch fieldVal.Type().Kind() {
	case reflect.Pointer:
		if fieldVal.Type().Elem() == reflect.TypeOf(time.Time{}) {
			if attribute == nil {
				fieldVal.Set(reflect.Zero(fieldVal.Type()))
				return true, nil
			}

			concreteValue, err := time.Parse(time.RFC3339, attribute.(string))
			if err != nil {
				return false, err
			}

			fieldVal.Set(reflect.ValueOf(&concreteValue))
			return true, nil
		}
	case reflect.Struct:
		if fieldVal.Type() == reflect.TypeOf(time.Time{}) {
			concreteValue, err := time.Parse(time.RFC3339, attribute.(string))
			if err != nil {
				return false, err
			}
			fieldVal.Set(reflect.ValueOf(concreteValue))
			return true, nil
		}
	}

	return false, nil
}

func unmarshalUnmarshaler(fieldVal reflect.Value, attribute interface{}) (bool, error) {
	switch fieldVal.Type().Kind() {
	case reflect.Ptr:
		if fieldVal.Type().Implements(reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()) {
			v := reflect.New(fieldVal.Type().Elem())
			recoded, err := json.Marshal(attribute)
			if err != nil {
				return false, err
			}
			v.MethodByName("UnmarshalJSON").
				Call([]reflect.Value{reflect.ValueOf(recoded)})
			fieldVal.Set(v)
			return true, nil
		}
	case reflect.Struct:
		if reflect.PointerTo(fieldVal.Type()).Implements(reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()) {
			v := reflect.New(fieldVal.Type())
			recoded, err := json.Marshal(attribute)
			if err != nil {
				return false, err
			}
			v.MethodByName("UnmarshalJSON").
				Call([]reflect.Value{reflect.ValueOf(recoded)})
			fieldVal.Set(v.Elem())
			return true, nil
		}
	}

	return false, nil
}

func castPrimitive(kind reflect.Kind, fieldType reflect.Type, attribute interface{}) reflect.Value {
	stringUnmarshallerType := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

	switch {
	case fieldType.Implements(stringUnmarshallerType):
		v := reflect.New(fieldType.Elem())
		v.MethodByName("UnmarshalText").
			Call([]reflect.Value{reflect.ValueOf([]byte(attribute.(string)))})
		return v.Elem()
	case reflect.PointerTo(fieldType).Implements(stringUnmarshallerType):
		v := reflect.New(fieldType)
		v.MethodByName("UnmarshalText").
			Call([]reflect.Value{reflect.ValueOf([]byte(attribute.(string)))})
		return v.Elem()
	default:
		switch kind {
		case reflect.Bool:
			return reflect.ValueOf(attribute.(bool))
		case reflect.Int8:
			return reflect.ValueOf(int8(attribute.(float64)))
		case reflect.Int:
			return reflect.ValueOf(int(attribute.(float64)))
		case reflect.Int16:
			return reflect.ValueOf(int16(attribute.(float64)))
		case reflect.Int32:
			return reflect.ValueOf(int32(attribute.(float64)))
		case reflect.Int64:
			return reflect.ValueOf(int64(attribute.(float64)))
		case reflect.Uint8:
			return reflect.ValueOf(uint8(attribute.(float64)))
		case reflect.Uint:
			return reflect.ValueOf(uint(attribute.(float64)))
		case reflect.Uint16:
			return reflect.ValueOf(uint16(attribute.(float64)))
		case reflect.Uint32:
			return reflect.ValueOf(uint32(attribute.(float64)))
		case reflect.Uint64:
			return reflect.ValueOf(uint64(attribute.(float64)))
		case reflect.Float32:
			return reflect.ValueOf(float32(attribute.(float64)))
		case reflect.Float64:
			return reflect.ValueOf(attribute.(float64))
		case reflect.String:
			return reflect.ValueOf(attribute.(string))
		default:
			return reflect.ValueOf(attribute)
		}
	}
}

func castPrimitivePointer(concreteKind reflect.Kind, concreteType reflect.Type, attribute interface{}) reflect.Value {
	stringUnmarshallerType := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	switch {
	case concreteType.Implements(stringUnmarshallerType):
		v := reflect.New(concreteType.Elem())
		v.MethodByName("UnmarshalText").
			Call([]reflect.Value{reflect.ValueOf([]byte(attribute.(string)))})
		return v
	case reflect.PointerTo(concreteType).Implements(stringUnmarshallerType):
		v := reflect.New(reflect.PointerTo(concreteType))
		v.MethodByName("UnmarshalText").
			Call([]reflect.Value{reflect.ValueOf([]byte(attribute.(string)))})
		return v
	default:
		switch concreteKind {
		case reflect.Bool:
			cVal := attribute.(bool)
			return reflect.ValueOf(&cVal)
		case reflect.Int8:
			cVal := int8(attribute.(float64))
			return reflect.ValueOf(&cVal)
		case reflect.Int:
			cVal := int(attribute.(float64))
			return reflect.ValueOf(&cVal)
		case reflect.Int16:
			cVal := int16(attribute.(float64))
			return reflect.ValueOf(&cVal)
		case reflect.Int32:
			cVal := int32(attribute.(float64))
			return reflect.ValueOf(&cVal)
		case reflect.Int64:
			cVal := int64(attribute.(float64))
			return reflect.ValueOf(&cVal)
		case reflect.Uint8:
			cVal := uint8(attribute.(float64))
			return reflect.ValueOf(&cVal)
		case reflect.Uint:
			cVal := uint(attribute.(float64))
			return reflect.ValueOf(&cVal)
		case reflect.Uint16:
			cVal := uint16(attribute.(float64))
			return reflect.ValueOf(&cVal)
		case reflect.Uint32:
			cVal := uint32(attribute.(float64))
			return reflect.ValueOf(&cVal)
		case reflect.Uint64:
			cVal := uint64(attribute.(float64))
			return reflect.ValueOf(&cVal)
		case reflect.Float32:
			cVal := float32(attribute.(float64))
			return reflect.ValueOf(&cVal)
		case reflect.Float64:
			cVal := attribute.(float64)
			return reflect.ValueOf(&cVal)
		case reflect.String:
			cVal := attribute.(string)
			return reflect.ValueOf(&cVal)
		default:
			return reflect.ValueOf(&attribute)
		}
	}
}

func unmarshalRelationships(fieldType reflect.StructField, fieldVal reflect.Value, resourceRelationships map[string]interface{}, included []interface{}) error {
	relationshipName := getAttributeName(fieldType)
	if relationship, ok := resourceRelationships[relationshipName]; ok {
		data, ok := relationship.(map[string]interface{})["data"] //normalised data of relationship containing type and id / list of ids
		if !ok {
			return errors.New("invalid relationship data structure")
		}
		err := unmarshalSingleRelationship(fieldVal, data, included)
		if err != nil {
			return err
		}
	}

	return nil
}

func unmarshalSingleRelationship(fieldVal reflect.Value, relationship interface{}, included []interface{}) error {
	//relationship here should be extended with attributes and references from corresponding included if available

	switch fieldVal.Kind() {
	case reflect.Struct:
		var toFillIn = reflect.New(fieldVal.Type())

		if err := unmarshalOne(resolveRelationshipData(relationship.(map[string]interface{}), included), toFillIn.Interface(), included); err != nil {
			return err
		}

		fieldVal.Set(toFillIn.Elem())
		return nil
	case reflect.Pointer:
		toFillIn := reflect.New(fieldVal.Type().Elem())

		if relationship == nil {
			//empty relationship is only legit for pointers
			//init pointer to nil
			fieldVal.Set(reflect.Zero(fieldVal.Type()))
			return nil
		}

		if err := unmarshalOne(resolveRelationshipData(relationship.(map[string]interface{}), included), toFillIn.Interface(), included); err != nil {
			return err
		}

		fieldVal.Set(toFillIn)

		return nil
	case reflect.Slice:
		dataSlice, ok := relationship.([]interface{})
		if !ok {
			return errors.New("invalid relationship data structure - expecting a slice of relationships")
		}

		reflection := reflect.MakeSlice(reflect.SliceOf(fieldVal.Type().Elem()), 0, 0)
		reflectionValue := reflect.New(reflection.Type())
		reflectionValue.Elem().Set(reflection)

		slicePtr := reflect.ValueOf(reflectionValue.Interface())
		sliceValuePtr := slicePtr.Elem()

		for _, datapoint := range dataSlice {
			toFillIn := reflect.New(fieldVal.Type().Elem())

			if fieldVal.Type().Elem().Kind() == reflect.Ptr {
				toFillIn = reflect.New(fieldVal.Type().Elem().Elem())
			}

			if err := unmarshalOne(resolveRelationshipData(datapoint.(map[string]interface{}), included), toFillIn.Interface(), included); err != nil {
				return err
			}

			var toAppend reflect.Value
			if fieldVal.Type().Elem().Kind() == reflect.Struct {
				toAppend = toFillIn.Elem()
			} else {
				toAppend = toFillIn
			}
			sliceValuePtr.Set(reflect.Append(sliceValuePtr, toAppend))
		}

		fieldVal.Set(reflect.ValueOf(sliceValuePtr.Interface()))
		return nil
	default:
		return errors.New("invalid relationship field type")
	}
}

func resolveRelationshipData(referenceData map[string]interface{}, included []interface{}) map[string]interface{} {
	referencedType := referenceData["type"]
	referencedId := referenceData["id"]

	for _, includedResource := range included {
		includedResourceData, ok := includedResource.(map[string]interface{})
		if !ok {
			continue
		}

		if includedResourceData["type"] == referencedType && includedResourceData["id"] == referencedId {
			return includedResourceData
		}
	}

	return referenceData
}

func getAttributeName(fieldType reflect.StructField) string {
	jsonapitag := fieldType.Tag.Get("jsonapi")
	if jsonapitag != "" {
		parts := strings.Split(jsonapitag, ",")
		if len(parts) > 1 {
			switch parts[0] {
			case "attr", "relation":
				return parts[1]
			}
		}
	}
	jsontag := fieldType.Tag.Get("json")
	if jsontag != "" {
		parts := strings.Split(jsontag, ",")
		return parts[0]
	}

	return toCamelCase(fieldType.Name)
}
