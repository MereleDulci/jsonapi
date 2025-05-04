package jsonapi

import (
	"encoding"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

type includesCache struct {
	includes []reflect.Value
}

func (c *includesCache) add(v reflect.Value) {
	c.includes = append(c.includes, v)
}

func (c *includesCache) contains(v reflect.Value) bool {
	for _, i := range c.includes {
		if i == v {
			return true
		}
	}
	return false
}

func Marshal(in interface{}) ([]byte, error) {
	inVal := reflect.ValueOf(in)
	if inVal.Kind() == reflect.Ptr {
		inVal = inVal.Elem()
	}

	if inVal.Kind() != reflect.Slice {
		return MarshalOne(in)
	}

	return MarshalMany(in)
}

func MarshalMany(in interface{}) ([]byte, error) {
	inVal := reflect.ValueOf(in)
	if inVal.Kind() == reflect.Ptr {
		inVal = inVal.Elem()
	}

	if inVal.Kind() != reflect.Slice {
		return nil, errors.New("input must be a slice")
	}

	out := make([]map[string]interface{}, inVal.Len())

	allIncludes := make([]interface{}, 0)
	for i := 0; i < inVal.Len(); i++ {
		next, includes, err := marshalNode(inVal.Index(i).Interface(), &includesCache{})
		if err != nil {
			return nil, err
		}
		out[i] = next
		allIncludes = append(allIncludes, includes...)
	}

	result := map[string]interface{}{
		"data": out,
	}

	if len(allIncludes) > 0 {
		result["included"] = deduplicateIncluded(allIncludes)
	}
	return json.Marshal(result)
}

func MarshalOne(in interface{}) ([]byte, error) {
	doc, includes, err := marshalNode(in, &includesCache{})
	if err != nil {
		return nil, err
	}

	out := map[string]interface{}{
		"data": doc,
	}

	if len(includes) > 0 {
		out["included"] = deduplicateIncluded(includes)
	}

	return json.Marshal(out)
}

func MixInMeta(source []byte, meta map[string]interface{}) ([]byte, error) {
	raw := make(map[string]interface{})
	err := json.Unmarshal(source, &raw)
	if err != nil {
		return nil, err
	}

	raw["meta"] = meta

	return json.Marshal(raw)
}

func marshalNode(node interface{}, refcache *includesCache) (map[string]interface{}, []interface{}, error) {
	inType := reflect.TypeOf(node)
	inVal := reflect.ValueOf(node)

	if inVal.Kind() == reflect.Ptr {
		inVal = inVal.Elem()
		inType = inType.Elem()
	}

	resourceType, err := getResourceType(inVal, inType)
	if err != nil {
		return nil, nil, err
	}
	resourceId, err := getResourceID(inVal, inType)
	if err != nil {
		return nil, nil, err
	}
	resourceAttrs, err := getAttributes(inVal, inType)
	if err != nil {
		return nil, nil, err
	}
	resourceRelationships, includes, err := getRelationships(inVal, inType, refcache)
	if err != nil {
		return nil, nil, err
	}

	return map[string]interface{}{
		"type":          resourceType,
		"id":            resourceId,
		"attributes":    resourceAttrs,
		"relationships": resourceRelationships,
	}, includes, nil
}

func getResourceID(inVal reflect.Value, inType reflect.Type) (string, error) {

	for i, n := 0, inType.NumField(); i < n; i++ {
		field := inType.Field(i)
		tag := field.Tag.Get("jsonapi")
		if tag != "" {
			parts := strings.Split(tag, ",")
			if len(parts) > 1 && parts[0] == "primary" {
				idField := inVal.FieldByName(field.Name)

				if idField.IsValid() {
					f := inVal.FieldByName("ID")
					switch f.Kind() {
					case reflect.String:
						return f.String(), nil
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						return strconv.FormatInt(f.Int(), 10), nil
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						return strconv.FormatUint(f.Uint(), 10), nil
					case reflect.Float32, reflect.Float64:
						return strconv.FormatFloat(f.Float(), 'f', -1, 64), nil
					default:
						textable, ok := f.Interface().(encoding.TextMarshaler)
						if ok {
							b, err := textable.MarshalText()
							if err != nil {
								return "", err
							}
							return string(b), nil
						}
						return "", errors.New("ID field must be a string, number or implement encoding.TextMarshaler")
					}
				}
			}
		}
	}

	return "", errors.New("no primary key found")
}

func getResourceType(inVal reflect.Value, inType reflect.Type) (string, error) {
	idField := inVal.FieldByName("ID")
	if idField.IsValid() {
		fType, _ := inType.FieldByName("ID")
		tag := fType.Tag.Get("jsonapi")
		if tag != "" {
			parts := strings.Split(tag, ",")
			if len(parts) > 1 && parts[0] == "primary" {
				return parts[1], nil
			}
		}
	}

	for i, n := 0, inType.NumField(); i < n; i++ {
		field := inType.Field(i)
		tag := field.Tag.Get("jsonapi")
		if tag != "" {
			parts := strings.Split(tag, ",")
			if len(parts) > 1 && parts[0] == "primary" {
				return parts[1], nil
			}
		}
	}

	return "", errors.New("no primary key found")
}

func getAttributes(inVal reflect.Value, inType reflect.Type) (map[string]interface{}, error) {
	attrs := map[string]interface{}{}

	for i, n := 0, inType.NumField(); i < n; i++ {
		field := inType.Field(i)
		encodedFieldName := getEncodedFieldName(field)

		if encodedFieldName == "-" {
			continue
		}

		jsonapiTag := field.Tag.Get("jsonapi")
		jsonTag := field.Tag.Get("json")
		if jsonapiTag != "" {
			parts := strings.Split(jsonapiTag, ",")
			if len(parts) == 0 {
				return nil, errors.New("invalid jsonapi tag")
			}
			if parts[0] != "attr" {
				continue
			}

			val := inVal.Field(i)
			if strings.Contains(jsonapiTag, ",omitempty") && isEmptyValue(val) {
				continue
			}

			attrs[encodedFieldName] = prepareAttributesNode(val)

		} else if jsonTag != "" {
			val := inVal.Field(i)

			if strings.Contains(jsonTag, ",omitempty") && isEmptyValue(val) {
				continue
			}

			attrs[encodedFieldName] = prepareAttributesNode(val)
		} else if field.IsExported() {
			//Exported field could represent an attribute or a relationship
			//Attribute values could have nested structs
			attrs[encodedFieldName] = prepareAttributesNode(inVal.Field(i))
		}
	}

	return attrs, nil
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Interface, reflect.Pointer:
		return v.IsZero()
	default:
		return false
	}
}

func getEncodedFieldName(field reflect.StructField) string {
	jsonapiTag := field.Tag.Get("jsonapi")
	if jsonapiTag != "" {
		parts := strings.Split(jsonapiTag, ",")
		if len(parts) > 1 && (parts[0] == "attr" || parts[0] == "relation") && parts[1] != "" {
			return parts[1]
		}
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if len(parts) == 1 && parts[0] == "-" { //Compatibility with json "-" conventions
			return "-"
		}
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}

	return toCamelCase(field.Name)
}

func prepareAttributesNode(field reflect.Value) interface{} {

	switch field.Kind() {
	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			return field.Interface().(time.Time).Format(time.RFC3339)
		}
		_, isMarshaler := field.Interface().(json.Marshaler)
		if isMarshaler {
			return field.Interface()
		}

		embed := map[string]interface{}{}

		for i, n := 0, field.NumField(); i < n; i++ {
			encodedFieldName := getEncodedFieldName(field.Type().Field(i))
			if encodedFieldName == "-" {
				continue
			}
			embed[encodedFieldName] = prepareAttributesNode(field.Field(i))
		}
		return embed
	case reflect.Pointer:
		if field.Elem().Kind() != reflect.Struct {
			return field.Interface()
		}

		if field.Elem().Type() == reflect.TypeOf(time.Time{}) {
			return field.Interface().(*time.Time).Format(time.RFC3339)
		}
		_, isMarshaler := field.Elem().Interface().(json.Marshaler)
		if isMarshaler {
			return field.Interface()
		}

		embed := map[string]interface{}{}
		for i, n := 0, field.Elem().NumField(); i < n; i++ {
			encodedFieldName := getEncodedFieldName(field.Elem().Type().Field(i))
			if encodedFieldName == "-" {
				continue
			}
			embed[encodedFieldName] = prepareAttributesNode(field.Elem().Field(i))
		}
		return embed
	case reflect.Slice:
		embed := make([]interface{}, field.Len())
		for i := 0; i < field.Len(); i++ {
			embed[i] = prepareAttributesNode(field.Index(i))
		}
		return embed
	default:
		return field.Interface()
	}
}

func getRelationships(inVal reflect.Value, inType reflect.Type, refcache *includesCache) (map[string]interface{}, []interface{}, error) {
	seen := make([]string, 0)
	rels := map[string]interface{}{}
	includes := make([]interface{}, 0)

	for i, n := 0, inType.NumField(); i < n; i++ {
		field := inType.Field(i)
		tag := field.Tag.Get("jsonapi")
		if tag != "" {
			parts := strings.Split(tag, ",")
			if len(parts) > 0 && parts[0] == "relation" {
				inner, include, err := prepareRelationshipNode(inVal.Field(i), refcache)
				if err != nil {
					return nil, nil, err
				}

				var relationshipName string
				if len(parts) > 1 {
					relationshipName = parts[1]
				} else {
					relationshipName = toCamelCase(field.Name)
				}

				if slices.Contains(seen, relationshipName) {
					return nil, nil, errors.New("relationship name already used: " + relationshipName)
				}
				seen = append(seen, relationshipName)

				rels[relationshipName] = map[string]interface{}{
					"data": inner,
				}

				includes = append(includes, include...)
			}
		}
	}

	return rels, includes, nil
}

func prepareRelationshipNode(topFieldValue reflect.Value, refcache *includesCache) (interface{}, []interface{}, error) {
	switch topFieldValue.Kind() {
	case reflect.Pointer:
		return prepareRelationshipNode(topFieldValue.Elem(), refcache)
	case reflect.Struct:
		refType, err := getResourceType(topFieldValue, topFieldValue.Type())
		if err != nil {
			return nil, nil, err
		}
		refId, err := getResourceID(topFieldValue, topFieldValue.Type())
		if err != nil {
			return nil, nil, err
		}

		relation := map[string]interface{}{
			"type": refType,
			"id":   refId,
		}

		//Breaks out of infinite recursion if there's a closed references loop in the provided structure
		if refcache.contains(topFieldValue) {
			return relation, nil, nil
		}
		refcache.add(topFieldValue)

		includeNode, internalIncludes, err := marshalNode(topFieldValue.Interface(), refcache)
		if err != nil {
			return nil, nil, err
		}

		return relation, append(internalIncludes, includeNode), nil

	case reflect.Slice:
		embed := make([]interface{}, topFieldValue.Len())
		includes := make([]interface{}, 0)
		for i := 0; i < topFieldValue.Len(); i++ {
			next, include, err := prepareRelationshipNode(topFieldValue.Index(i), refcache)
			if err != nil {
				return nil, nil, err
			}
			embed[i] = next
			includes = append(includes, include...)
		}
		return embed, includes, nil
	default:
		return nil, nil, nil
	}
}

func deduplicateIncluded(includes []interface{}) []interface{} {
	unique := map[string]interface{}{}
	out := make([]interface{}, 0)

	for _, include := range includes {
		doc := include.(map[string]interface{})
		resourceType := doc["type"]
		resourceId := doc["id"]
		if resourceId == "" {
			continue
		}

		key := resourceType.(string) + resourceId.(string)

		if _, ok := unique[key]; ok {
			attributesLeft := doc["attributes"].(map[string]interface{})
			attributesRight := unique[key].(map[string]interface{})["attributes"].(map[string]interface{})
			mergedAttributes := shallowMerge(attributesLeft, attributesRight, isAttributeZero)

			relationshipsLeft := doc["relationships"].(map[string]interface{})
			relationshipsRight := unique[key].(map[string]interface{})["relationships"].(map[string]interface{})
			mergedRelationships := shallowMerge(relationshipsLeft, relationshipsRight, isRelationshipZero)

			unique[key] = map[string]interface{}{
				"type":          resourceType,
				"id":            resourceId,
				"attributes":    mergedAttributes,
				"relationships": mergedRelationships,
			}

		} else {
			unique[key] = doc
		}
	}

	for _, v := range unique {
		out = append(out, v)
	}

	return out
}

type zeroPredicate func(i interface{}) bool

func shallowMerge(a, b map[string]interface{}, isZero zeroPredicate) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range a {
		if !isZero(v) {
			out[k] = v
		}
	}
	for k, v := range b {
		if !isZero(v) {
			out[k] = v
		}
	}
	return out
}

func isAttributeZero(attr interface{}) bool {
	if attr == nil {
		return true
	}
	return reflect.ValueOf(attr).IsZero()
}

func isRelationshipZero(rel interface{}) bool {
	data, ok := rel.(map[string]interface{})["data"]
	if !ok {
		return true
	}

	switch data.(type) {
	case map[string]interface{}:
		return len(data.(map[string]interface{})) == 0
	case []interface{}:
		return len(data.([]interface{})) == 0
	default:
		return true
	}
}

func toCamelCase(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}
