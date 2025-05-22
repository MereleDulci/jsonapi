package jsonapi

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMarshal_metadata(t *testing.T) {

	t.Run("should correctly extract the resource type ", func(t *testing.T) {
		input := struct {
			ID string `jsonapi:"primary,resource-name"`
		}{
			ID: "1",
		}

		plain, err := MarshalOne(input)
		if err != nil {
			t.Fatal(err)
		}

		ref, err := MarshalOne(&input)
		if err != nil {
			t.Fatal(err)
		}

		for _, v := range [][]byte{plain, ref} {
			check := map[string]interface{}{}
			if err := json.Unmarshal(v, &check); err != nil {
				t.Fatal(err)
			}

			if check["data"].(map[string]interface{})["type"] != "resource-name" {
				t.Fatal("unexpected resource type")
			}
		}
	})
}

func TestMarshal_single(t *testing.T) {

	t.Run("should correctly marshal a single resource with primitive types", func(t *testing.T) {
		input := struct {
			ID            string `jsonapi:"primary,resource-name"`
			ExportedField string
			private       string
			Muted         string `json:"-"`
		}{
			ID:            "1",
			ExportedField: "test",
			private:       "private",
			Muted:         "muted",
		}

		plain, err := MarshalOne(input)
		if err != nil {
			t.Fatal(err)
		}

		ref, err := MarshalOne(&input)
		if err != nil {
			t.Fatal(err)
		}

		for _, v := range [][]byte{plain, ref} {
			check := map[string]interface{}{}
			if err := json.Unmarshal(v, &check); err != nil {
				t.Fatal(err)
			}

			if check["data"].(map[string]interface{})["id"] != "1" {
				t.Fatal("unexpected id")
			}

			attrs, ok := check["data"].(map[string]interface{})["attributes"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected attributes type")
			}

			if attrs["exportedField"] != "test" {
				t.Fatal("unexpected attribute value")
			}

			if _, ok := attrs["private"]; ok {
				t.Fatal("private field should not be present")
			}

			if _, ok := attrs["muted"]; ok {
				t.Fatal("muted field should not be present")
			}
			if _, ok := attrs["-"]; ok {
				t.Fatal("muted field should not be present")
			}
			if _, ok := attrs[""]; ok {
				t.Fatal("muted field should not be present")
			}
		}
	})

	t.Run("should correctly marshal a single resource with nested structs", func(t *testing.T) {
		input := struct {
			ID             string `jsonapi:"primary,resource-name"`
			ExportedStruct struct {
				A     int
				Muted string `json:"-"`
			}
			ExportedRef *struct {
				A     int
				Muted string `json:"-"`
			}
			ExportedSlice       []string
			ExportedStructSlice []struct{ A int }
			ExportedRefSlice    []*struct{ A int }
		}{
			ID: "1",
			ExportedStruct: struct {
				A     int
				Muted string `json:"-"`
			}{A: 1, Muted: "muted"},
			ExportedRef: &struct {
				A     int
				Muted string `json:"-"`
			}{A: 1, Muted: "muted"},
			ExportedSlice:       []string{"a"},
			ExportedStructSlice: []struct{ A int }{{A: 1}},
			ExportedRefSlice:    []*struct{ A int }{{A: 1}},
		}

		plain, err := MarshalOne(input)
		if err != nil {
			t.Fatal(err)
		}

		ref, err := MarshalOne(&input)
		if err != nil {
			t.Fatal(err)
		}

		for _, v := range [][]byte{plain, ref} {
			check := map[string]interface{}{}
			if err := json.Unmarshal(v, &check); err != nil {
				t.Fatal(err)
			}

			attrs, ok := check["data"].(map[string]interface{})["attributes"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected attributes type")
			}

			if attrs["exportedStruct"].(map[string]interface{})["a"] != float64(1) {
				t.Fatal("unexpected attribute value on struct")
			}

			if _, ok := attrs["exportedStruct"].(map[string]interface{})["muted"]; ok {
				t.Fatal("muted field should not be present")
			}
			if _, ok := attrs["exportedStruct"].(map[string]interface{})["-"]; ok {
				t.Fatal("muted field should not be present")
			}
			if _, ok := attrs["exportedStruct"].(map[string]interface{})[""]; ok {
				t.Fatal("muted field should not be present")
			}

			if attrs["exportedRef"].(map[string]interface{})["a"] != float64(1) {
				t.Fatal("unexpected attribute value on ref")
			}

			if _, ok := attrs["exportedRef"].(map[string]interface{})["muted"]; ok {
				t.Fatal("muted field should not be present")
			}

			if _, ok := attrs["exportedRef"].(map[string]interface{})["-"]; ok {
				t.Fatal("muted field should not be present")
			}

			if _, ok := attrs["exportedRef"].(map[string]interface{})[""]; ok {
				t.Fatal("muted field should not be present")
			}

			if attrs["exportedSlice"].([]interface{})[0] != "a" {
				t.Fatal("unexpected attribute value on slice of primitives")
			}

			if attrs["exportedStructSlice"].([]interface{})[0].(map[string]interface{})["a"] != float64(1) {
				t.Fatal("unexpected attribute value on slice of structs")
			}

			if attrs["exportedRefSlice"].([]interface{})[0].(map[string]interface{})["a"] != float64(1) {
				t.Fatal("unexpected attribute value on slice of refs")
			}
		}
	})

	t.Run("should apply MarshalJSON method on nested structs if defined", func(t *testing.T) {
		input := struct {
			ID             string `jsonapi:"primary,resource-name"`
			ExportedStruct nestedWithMarshalInner
			ExportedRef    *nestedWithMarshalInner
		}{
			ID:             "1",
			ExportedStruct: nestedWithMarshalInner{A: 1},
			ExportedRef:    &nestedWithMarshalInner{A: 1},
		}

		plain, err := MarshalOne(input)
		if err != nil {
			t.Fatal(err)
		}

		ref, err := MarshalOne(&input)
		if err != nil {
			t.Fatal(err)
		}

		for _, v := range [][]byte{plain, ref} {
			check := map[string]interface{}{}
			if err := json.Unmarshal(v, &check); err != nil {
				t.Fatal(err)
			}

			attrs, ok := check["data"].(map[string]interface{})["attributes"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected attributes type")
			}

			if attrs["exportedStruct"].(map[string]interface{})["a"] == float64(1) {
				t.Fatal("expected Marshaller to rename the field")
			}
			if attrs["exportedStruct"].(map[string]interface{})["b"] != float64(1) {
				t.Fatal("unexpected attribute value on struct")
			}
			if attrs["exportedRef"].(map[string]interface{})["a"] == float64(1) {
				t.Fatal("expected Marshaller to rename the field")
			}
			if attrs["exportedRef"].(map[string]interface{})["b"] != float64(1) {
				t.Fatal("unexpected attribute value on struct")
			}
		}

	})

	t.Run("should correctly marshal non-primitive map values", func(t *testing.T) {
		type inner struct {
			A int `json:"a"`
		}

		input := struct {
			ID string           `jsonapi:"primary,resource-name"`
			M  map[string]inner `jsonapi:"attr,m"`
		}{
			ID: "1",
			M: map[string]inner{
				"key1": {A: 1},
			},
		}

		plain, err := MarshalOne(input)
		if err != nil {
			t.Fatal(err)
		}

		ref, err := MarshalOne(&input)
		if err != nil {
			t.Fatal(err)
		}

		for _, v := range [][]byte{plain, ref} {
			check := map[string]interface{}{}
			if err := json.Unmarshal(v, &check); err != nil {
				t.Fatal(err)
			}

			attrs, ok := check["data"].(map[string]interface{})["attributes"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected attributes type")
			}

			if attrs["m"].(map[string]interface{})["key1"].(map[string]interface{})["a"] != float64(1) {
				t.Fatal("unexpected attribute value on map")
			}
		}

	})

}

func TestMarshal_relations(t *testing.T) {
	t.Run("should correctly marshal a single resource with one-to-one relation as value", func(t *testing.T) {
		input := struct {
			ID              string `jsonapi:"primary,resource-name"`
			ExportedField   string
			DefaultRelation struct {
				ID string `jsonapi:"primary,public-name"`
			} `jsonapi:"relation"`
			ExportedRelation struct {
				ID string `jsonapi:"primary,public-name"`
			} `jsonapi:"relation,exportedRef"`
			privateRelation struct {
				ID string `jsonapi:"primary,private-name"`
			}
		}{
			ID:            "1",
			ExportedField: "test",
			DefaultRelation: struct {
				ID string `jsonapi:"primary,public-name"`
			}{
				ID: "2",
			},
			ExportedRelation: struct {
				ID string `jsonapi:"primary,public-name"`
			}{
				ID: "3",
			},
			privateRelation: struct {
				ID string `jsonapi:"primary,private-name"`
			}{
				ID: "4",
			},
		}

		plain, err := MarshalOne(input)
		if err != nil {
			t.Fatal(err)
		}

		ref, err := MarshalOne(input)
		if err != nil {
			t.Fatal(err)
		}

		for _, v := range [][]byte{plain, ref} {
			check := map[string]interface{}{}
			if err := json.Unmarshal(v, &check); err != nil {
				t.Fatal(err)
			}

			rels, ok := check["data"].(map[string]interface{})["relationships"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected relationships type")
			}

			attrs, ok := check["data"].(map[string]interface{})["attributes"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected attributes type")
			}

			if _, ok := attrs["defaultRelation"]; ok {
				t.Fatal("relation should not appear in attributes")
			}
			if _, ok := attrs["exportedRelation"]; ok {
				t.Fatal("relation should not appear in attributes")
			}
			if _, ok := attrs["exportedRef"]; ok {
				t.Fatal("relation should not appear in attributes")
			}
			if _, ok := attrs["privateRelation"]; ok {
				t.Fatal("private relation should not appear in attributes")
			}

			if _, ok := rels["defaultRelation"]; !ok {
				t.Fatal("relation should appear in relationships")
			}
			if _, ok := rels["exportedRef"]; !ok {
				t.Fatal("relation should appear in relationships")
			}
			if _, ok := rels["exportedRelation"]; ok {
				t.Fatal("overwritten relation name should not appear in relationships")
			}
			if _, ok := rels["privateRelation"]; ok {
				t.Fatal("private relation should not appear in attributes")
			}

			data, ok := rels["defaultRelation"].(map[string]interface{})["data"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected relationship type")
			}

			if data["id"] != "2" {
				t.Fatal("unexpected relationship id")
			}

			if data["type"] != "public-name" {
				t.Fatal("unexpected relationship type")
			}

			data, ok = rels["exportedRef"].(map[string]interface{})["data"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected relationship type")
			}

			if data["id"] != "3" {
				t.Fatal("unexpected relationship id")
			}

			if data["type"] != "public-name" {
				t.Fatal("unexpected relationship type")
			}
		}
	})

	t.Run("should correctly marshal a single resource with one-to-one relation as pointer", func(t *testing.T) {
		input := struct {
			ID              string `jsonapi:"primary,resource-name"`
			ExportedField   string
			DefaultRelation *struct {
				ID string `jsonapi:"primary,public-name"`
			} `jsonapi:"relation"`
			ExportedRelation *struct {
				ID string `jsonapi:"primary,public-name"`
			} `jsonapi:"relation,exportedRef"`
			privateRelation *struct {
				ID string `jsonapi:"primary,private-name"`
			}
		}{
			ID:            "1",
			ExportedField: "test",
			DefaultRelation: &struct {
				ID string `jsonapi:"primary,public-name"`
			}{
				ID: "2",
			},
			ExportedRelation: &struct {
				ID string `jsonapi:"primary,public-name"`
			}{
				ID: "3",
			},
			privateRelation: &struct {
				ID string `jsonapi:"primary,private-name"`
			}{
				ID: "4",
			},
		}

		plain, err := MarshalOne(input)
		if err != nil {
			t.Fatal(err)
		}

		ref, err := MarshalOne(input)
		if err != nil {
			t.Fatal(err)
		}

		for _, v := range [][]byte{plain, ref} {
			check := map[string]interface{}{}
			if err := json.Unmarshal(v, &check); err != nil {
				t.Fatal(err)
			}

			rels, ok := check["data"].(map[string]interface{})["relationships"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected relationships type")
			}

			attrs, ok := check["data"].(map[string]interface{})["attributes"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected attributes type")
			}

			if _, ok := attrs["defaultRelation"]; ok {
				t.Fatal("relation should not appear in attributes")
			}
			if _, ok := attrs["exportedRelation"]; ok {
				t.Fatal("relation should not appear in attributes")
			}
			if _, ok := attrs["exportedRef"]; ok {
				t.Fatal("relation should not appear in attributes")
			}
			if _, ok := attrs["privateRelation"]; ok {
				t.Fatal("private relation should not appear in attributes")
			}

			if _, ok := rels["defaultRelation"]; !ok {
				t.Fatal("relation should appear in relationships")
			}
			if _, ok := rels["exportedRef"]; !ok {
				t.Fatal("relation should appear in relationships")
			}
			if _, ok := rels["exportedRelation"]; ok {
				t.Fatal("overwritten relation should not appear in relationships")
			}
			if _, ok := rels["privateRelation"]; ok {
				t.Fatal("private relation should not appear in attributes")
			}

			data, ok := rels["defaultRelation"].(map[string]interface{})["data"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected relationship type")
			}

			if data["id"] != "2" {
				t.Fatal("unexpected relationship id")
			}

			if data["type"] != "public-name" {
				t.Fatal("unexpected relationship type")
			}

			data, ok = rels["exportedRef"].(map[string]interface{})["data"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected relationship type")
			}

			if data["id"] != "3" {
				t.Fatal("unexpected relationship id")
			}

			if data["type"] != "public-name" {
				t.Fatal("unexpected relationship type")
			}
		}
	})

	t.Run("should correctly marshal a single resource with one-to-many relation as values", func(t *testing.T) {
		input := struct {
			ID              string `jsonapi:"primary,resource-name"`
			ExportedField   string
			DefaultRelation []struct {
				ID string `jsonapi:"primary,public-name"`
			} `jsonapi:"relation"`
			ExportedRelation []struct {
				ID string `jsonapi:"primary,public-name"`
			} `jsonapi:"relation,exportedRef"`
			privateRelation struct {
				ID string `jsonapi:"primary,private-name"`
			}
		}{
			ID:            "1",
			ExportedField: "test",
			DefaultRelation: []struct {
				ID string `jsonapi:"primary,public-name"`
			}{
				{ID: "2"},
			},
			ExportedRelation: []struct {
				ID string `jsonapi:"primary,public-name"`
			}{{
				ID: "3",
			}, {
				ID: "4",
			}},
			privateRelation: struct {
				ID string `jsonapi:"primary,private-name"`
			}{
				ID: "5",
			},
		}

		plain, err := MarshalOne(input)
		if err != nil {
			t.Fatal(err)
		}

		ref, err := MarshalOne(input)
		if err != nil {
			t.Fatal(err)
		}

		for _, v := range [][]byte{plain, ref} {
			check := map[string]interface{}{}
			if err := json.Unmarshal(v, &check); err != nil {
				t.Fatal(err)
			}

			rels, ok := check["data"].(map[string]interface{})["relationships"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected relationships type")
			}

			attrs, ok := check["data"].(map[string]interface{})["attributes"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected attributes type")
			}

			if _, ok := attrs["defaultRelation"]; ok {
				t.Fatal("relation should not appear in attributes")
			}
			if _, ok := attrs["exportedRef"]; ok {
				t.Fatal("relation should not appear in attributes")
			}
			if _, ok := attrs["exportedRelation"]; ok {
				t.Fatal("relation should not appear in attributes")
			}
			if _, ok := attrs["privateRelation"]; ok {
				t.Fatal("private relation should not appear in attributes")
			}

			if _, ok := rels["defaultRelation"]; !ok {
				t.Fatal("relation should appear in relationships")
			}
			if _, ok := rels["exportedRef"]; !ok {
				t.Fatal("relation should appear in relationships")
			}
			if _, ok := rels["exportedRelation"]; ok {
				t.Fatal("overwritten relation should appear in relationships")
			}
			if _, ok := rels["privateRelation"]; ok {
				t.Fatal("private relation should not appear in attributes")
			}

			data, ok := rels["exportedRef"].(map[string]interface{})["data"].([]interface{})
			if !ok {
				t.Fatal("unexpected relationship type")
			}

			if len(data) != 2 {
				t.Fatal("unexpected relationship length")
			}

			first, ok := data[0].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected relationship doc type")
			}

			if first["id"] != "3" {
				t.Fatal("unexpected relationship id")
			}

			if first["type"] != "public-name" {
				t.Fatal("unexpected relationship type")
			}
		}
	})

	t.Run("should correctly marshal a single resource with one-to-many relation as pointers", func(t *testing.T) {
		input := struct {
			ID              string `jsonapi:"primary,resource-name"`
			ExportedField   string
			DefaultRelation []*struct {
				ID string `jsonapi:"primary,public-name"`
			} `jsonapi:"relation"`
			ExportedRelation []*struct {
				ID string `jsonapi:"primary,public-name"`
			} `jsonapi:"relation,exportedRef"`
			privateRelation struct {
				ID string `jsonapi:"primary,private-name"`
			}
		}{
			ID:            "1",
			ExportedField: "test",
			DefaultRelation: []*struct {
				ID string `jsonapi:"primary,public-name"`
			}{
				{ID: "2"},
			},
			ExportedRelation: []*struct {
				ID string `jsonapi:"primary,public-name"`
			}{{
				ID: "3",
			}, {
				ID: "4",
			}},
			privateRelation: struct {
				ID string `jsonapi:"primary,private-name"`
			}{
				ID: "5",
			},
		}

		plain, err := MarshalOne(input)
		if err != nil {
			t.Fatal(err)
		}

		ref, err := MarshalOne(input)
		if err != nil {
			t.Fatal(err)
		}

		for _, v := range [][]byte{plain, ref} {
			check := map[string]interface{}{}
			if err := json.Unmarshal(v, &check); err != nil {
				t.Fatal(err)
			}

			rels, ok := check["data"].(map[string]interface{})["relationships"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected relationships type")
			}

			attrs, ok := check["data"].(map[string]interface{})["attributes"].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected attributes type")
			}

			if _, ok := attrs["defaultRelation"]; ok {
				t.Fatal("relation should not appear in attributes")
			}
			if _, ok := attrs["exportedRef"]; ok {
				t.Fatal("relation should not appear in attributes")
			}
			if _, ok := attrs["exportedRelation"]; ok {
				t.Fatal("relation should not appear in attributes")
			}
			if _, ok := attrs["privateRelation"]; ok {
				t.Fatal("private relation should not appear in attributes")
			}

			if _, ok := rels["defaultRelation"]; !ok {
				t.Fatal("relation should appear in relationships")
			}
			if _, ok := rels["exportedRef"]; !ok {
				t.Fatal("relation should appear in relationships")
			}
			if _, ok := rels["exportedRelation"]; ok {
				t.Fatal("overwritten relation should appear in relationships")
			}
			if _, ok := rels["privateRelation"]; ok {
				t.Fatal("private relation should not appear in attributes")
			}

			data, ok := rels["exportedRef"].(map[string]interface{})["data"].([]interface{})
			if !ok {
				t.Fatal("unexpected relationship type")
			}

			if len(data) != 2 {
				t.Fatal("unexpected relationship length")
			}

			first, ok := data[0].(map[string]interface{})
			if !ok {
				t.Fatal("unexpected relationship doc type")
			}

			if first["id"] != "3" {
				t.Fatal("unexpected relationship id")
			}

			if first["type"] != "public-name" {
				t.Fatal("unexpected relationship type")

			}
		}
	})
}

func TestMarshalMany(t *testing.T) {
	type Test struct {
		ID    string `jsonapi:"primary,resource-name"`
		Value string
		Muted string `json:"-"`
	}

	t.Run("should correctly marshal a collection of resources as values", func(t *testing.T) {
		inputs := []Test{
			{ID: "1", Value: "one", Muted: "mute"},
			{ID: "2", Value: "two", Muted: "mute"},
		}

		plain, err := MarshalMany(inputs)
		if err != nil {
			t.Fatal(err)
		}

		ref, err := MarshalMany(&inputs)
		if err != nil {
			t.Fatal(err)
		}

		for _, v := range [][]byte{plain, ref} {

			check := map[string]interface{}{}
			if err := json.Unmarshal(v, &check); err != nil {
				t.Fatal(err)
			}

			if len(check["data"].([]interface{})) != 2 {
				t.Fatal("unexpected number of resources")
			}

			for _, v := range check["data"].([]interface{}) {
				doc := v.(map[string]interface{})

				if _, ok := doc["id"]; !ok {
					t.Fatal("missing id")
				}

				typeV, ok := doc["type"]
				if !ok {
					t.Fatal("missing type")
				}
				if typeV != "resource-name" {
					t.Fatal("unexpected type")
				}

				attrs, ok := doc["attributes"]
				if !ok {
					t.Fatal("missing attributes")
				}

				if attrs.(map[string]interface{})["muted"] == "mute" {
					t.Fatal("muted field should not be present")
				}
				if attrs.(map[string]interface{})["-"] == "mute" {
					t.Fatal("muted field should not be present")
				}
				if attrs.(map[string]interface{})[""] == "mute" {
					t.Fatal("muted field should not be present")
				}

			}
		}
	})

	t.Run("should correctly marshal a collection of resources as pointers", func(t *testing.T) {
		inputs := []*Test{
			{ID: "1", Value: "one", Muted: "mute"},
			{ID: "2", Value: "two", Muted: "mute"},
		}

		plain, err := MarshalMany(inputs)
		if err != nil {
			t.Fatal(err)
		}

		ref, err := MarshalMany(&inputs)
		if err != nil {
			t.Fatal(err)
		}

		for _, v := range [][]byte{plain, ref} {

			check := map[string]interface{}{}
			if err := json.Unmarshal(v, &check); err != nil {
				t.Fatal(err)
			}

			if len(check["data"].([]interface{})) != 2 {
				t.Fatal("unexpected number of resources")
			}

			for _, v := range check["data"].([]interface{}) {
				doc := v.(map[string]interface{})

				if _, ok := doc["id"]; !ok {
					t.Fatal("missing id")
				}

				typeV, ok := doc["type"]
				if !ok {
					t.Fatal("missing type")
				}
				if typeV != "resource-name" {
					t.Fatal("unexpected type")
				}

				attrs, ok := doc["attributes"]
				if !ok {
					t.Fatal("missing attributes")
				}

				if attrs.(map[string]interface{})["muted"] == "mute" {
					t.Fatal("muted field should not be present")
				}
				if attrs.(map[string]interface{})["-"] == "mute" {
					t.Fatal("muted field should not be present")
				}
				if attrs.(map[string]interface{})[""] == "mute" {
					t.Fatal("muted field should not be present")
				}

			}
		}
	})
}

func TestMarshalWithRelationships(t *testing.T) {

	t.Run("should push correct one-to-one relationships into included list", func(t *testing.T) {

		type Rel struct {
			ID  string `jsonapi:"primary,related"`
			Str string
		}

		type SUT struct {
			ID      string `jsonapi:"primary,test"`
			ByValue Rel    `jsonapi:"relation,val"`
			ByRef   *Rel   `jsonapi:"relation,ref"`
		}

		input := SUT{
			ID: "1",
			ByValue: Rel{
				ID:  "2",
				Str: "value",
			},
			ByRef: &Rel{
				ID:  "3",
				Str: "value",
			},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		check := map[string]interface{}{}
		if err := json.Unmarshal(raw, &check); err != nil {
			t.Fatal(err)
		}

		if len(check["included"].([]interface{})) != 2 {
			t.Fatal("unexpected number of included resources")
		}

		for _, v := range check["included"].([]interface{}) {
			doc, ok := v.(map[string]interface{})
			if !ok {
				t.Fatal("invalid document format")
			}

			if _, ok := doc["id"]; !ok {
				t.Fatal("missing id")
			}

			typeV, ok := doc["type"]
			if !ok {
				t.Fatal("missing type")
			}

			if typeV != "related" {
				t.Fatal("unexpected type")
			}

			if _, ok := doc["attributes"]; !ok {
				t.Fatal("missing attributes")
			}

			if _, ok := doc["relationships"]; !ok {
				t.Fatal("missing relationships")
			}
		}

		includedIds := []string{}
		for _, v := range check["included"].([]interface{}) {
			doc := v.(map[string]interface{})
			includedIds = append(includedIds, doc["id"].(string))
		}
		if len(includedIds) != 2 {
			t.Fatal("unexpected number of included resources")
		}
	})

	t.Run("should push correct one-to-many relationships into included list", func(t *testing.T) {
		type Rel struct {
			ID  string `jsonapi:"primary,related"`
			Str string
		}

		type SUT struct {
			ID      string `jsonapi:"primary,test"`
			ByValue []Rel  `jsonapi:"relation,val"`
			ByRef   []*Rel `jsonapi:"relation,ref"`
		}

		input := SUT{
			ID: "1",
			ByValue: []Rel{
				{ID: "2", Str: "value"},
				{ID: "3", Str: "value"},
			},
			ByRef: []*Rel{
				{ID: "4", Str: "value"},
				{ID: "5", Str: "value"},
			},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		check := map[string]interface{}{}
		if err := json.Unmarshal(raw, &check); err != nil {
			t.Fatal(err)
		}

		if len(check["included"].([]interface{})) != 4 {
			t.Fatal("unexpected number of included resources")
		}

		for _, v := range check["included"].([]interface{}) {
			doc, ok := v.(map[string]interface{})
			if !ok {
				t.Fatal("invalid document format")
			}

			if _, ok := doc["id"]; !ok {
				t.Fatal("missing id")
			}

			typeV, ok := doc["type"]
			if !ok {
				t.Fatal("missing type")
			}

			if typeV != "related" {
				t.Fatal("unexpected type")
			}

			if _, ok := doc["attributes"]; !ok {
				t.Fatal("missing attributes")
			}

			if _, ok := doc["relationships"]; !ok {
				t.Fatal("missing relationships")
			}
		}

		includedIds := []string{}
		for _, v := range check["included"].([]interface{}) {
			doc := v.(map[string]interface{})
			includedIds = append(includedIds, doc["id"].(string))
		}
		if len(includedIds) != 4 {
			t.Fatal("unexpected number of included resources")
		}
	})

	t.Run("should not push documents with zero ids into included", func(t *testing.T) {
		type Rel struct {
			ID  string `jsonapi:"primary,related"`
			Str string
		}

		type SUT struct {
			ID         string `jsonapi:"primary,test"`
			ByValue    Rel    `jsonapi:"relation,val"`
			SliceValue []Rel  `jsonapi:"relation,vals"`
		}

		input := SUT{
			ID: "1",
			ByValue: Rel{
				ID: "",
			},
			SliceValue: []Rel{
				{ID: ""},
			},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		check := map[string]interface{}{}
		if err := json.Unmarshal(raw, &check); err != nil {
			t.Fatal(err)
		}

		if len(check["included"].([]interface{})) != 0 {
			t.Fatal("unexpected number of included resources")
		}
	})
}

func TestMarshalRelationshipDeduplication(t *testing.T) {
	//Inclusion order matters, need to check multiple combinations of the same data in list

	t.Run("should correctly merge non-zero attributes", func(t *testing.T) {

		type Comment struct {
			ID      string `jsonapi:"primary,comments"`
			Content string
			Bool    bool
			ReplyTo *Comment `jsonapi:"relation,reply"`
		}

		type Post struct {
			ID       string     `jsonapi:"primary,posts"`
			Comments []*Comment `jsonapi:"relation,comments"`
		}

		hand := []*Comment{
			{ID: "2", Content: "first", Bool: true},
			{ID: "4", Content: "third", Bool: true, ReplyTo: &Comment{ID: "3"}},
			{ID: "3", Content: "second", Bool: true, ReplyTo: &Comment{ID: "2"}},
			{ID: "5", Content: "third", Bool: true, ReplyTo: &Comment{ID: "6"}},
		}

		for i, _ := range hand {
			for j, _ := range hand {
				hand[j], hand[i] = hand[i], hand[j]
				input := Post{
					ID:       "1",
					Comments: hand,
				}

				raw, err := Marshal(input)
				if err != nil {
					t.Fatal(err)
				}

				check := map[string]interface{}{}
				if err := json.Unmarshal(raw, &check); err != nil {
					t.Fatal(err)
				}

				if len(check["included"].([]interface{})) != 5 {
					t.Fatalf("unexpected number of included resources, got %v", len(check["included"].([]interface{})))
				}

				for _, v := range check["included"].([]interface{}) {
					doc, ok := v.(map[string]interface{})
					if !ok {
						t.Fatal("invalid document format")
					}

					if _, ok := doc["id"]; !ok {
						t.Fatal("missing id")
					}

					typeV, ok := doc["type"]
					if !ok {
						t.Fatal("missing type")
					}

					if typeV != "comments" {
						t.Fatal("unexpected type")
					}

					if _, ok := doc["attributes"]; !ok {
						t.Fatal("missing attributes")
					}

					if _, ok := doc["relationships"]; !ok {
						t.Fatal("missing relationships")
					}

					attrs := doc["attributes"].(map[string]interface{})
					if attrs["content"] == "" && doc["id"] != "6" {
						t.Fatal("unexpected string attribute value")
					}
					if attrs["bool"] == false && doc["id"] != "6" {
						t.Fatal("unexpected boolean attribute value")
					}
				}
			}
		}
	})

	t.Run("should correctly merge non-zero relationships", func(t *testing.T) {

		type Comment struct {
			ID     string   `jsonapi:"primary,comments"`
			Parent *Comment `jsonapi:"relation,parent"`
		}

		type Post struct {
			ID       string     `jsonapi:"primary,posts"`
			Comments []*Comment `jsonapi:"relation,comments"`
		}

		hand := []*Comment{
			{ID: "2"},
			{ID: "3", Parent: &Comment{ID: "2"}},
			{ID: "4", Parent: &Comment{ID: "3"}},
			{ID: "5", Parent: &Comment{ID: "4", Parent: &Comment{ID: "3"}}},
		}

		for i, _ := range hand {
			for j, _ := range hand {
				hand[i], hand[j] = hand[j], hand[i]

				input := Post{
					ID:       "1",
					Comments: hand,
				}

				raw, err := Marshal(input)
				if err != nil {
					t.Fatal(err)
				}

				check := map[string]interface{}{}
				if err := json.Unmarshal(raw, &check); err != nil {
					t.Fatal(err)
				}

				if len(check["included"].([]interface{})) != 4 {
					t.Fatalf("unexpected number of included resources, got %v", len(check["included"].([]interface{})))
				}
			}
		}
	})

	t.Run("should keep attributes in case of circular references", func(t *testing.T) {

		type SUT struct {
			ID       string     `jsonapi:"primary,mains"`
			Circular *CircularA `jsonapi:"relation,circular"`
		}

		input := SUT{
			ID: "main-id",
			Circular: &CircularA{
				ID: "1",
				B: &CircularB{
					ID:  "2",
					Val: "b",
					A: &CircularA{
						ID:  "1",
						Val: "a",
					},
				},
			},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		check := map[string]interface{}{}
		if err := json.Unmarshal(raw, &check); err != nil {
			t.Fatal(err)
		}

		included := check["included"].([]interface{})
		if len(included) != 2 {
			t.Fatalf("unexpected number of included resources, got %v", len(check["included"].([]interface{})))
		}

		//Val on A should still be set
		var a map[string]interface{}
		for _, viface := range included {
			v := viface.(map[string]interface{})
			if v["type"] == "circular-a" && v["id"] == "1" {
				a = v
			}
		}

		if a == nil {
			t.Fatal("missing included resource")
		}
		if a["attributes"].(map[string]interface{})["val"] != "a" {
			t.Fatal("erased attribute value")
		}

		if a["relationships"].(map[string]interface{})["b"].(map[string]interface{})["data"] == nil {
			t.Fatal("erased relationship value")
		}
	})
}

func TestMarshalID(t *testing.T) {

	t.Run("it should correctly marshal primitive ids ", func(t *testing.T) {
		type SUT struct {
			ID string `jsonapi:"primary,tests"`
		}

		input := SUT{
			ID: "1",
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		check := map[string]interface{}{}
		if err := json.Unmarshal(raw, &check); err != nil {
			t.Fatal(err)
		}

		if check["data"].(map[string]interface{})["id"] != "1" {
			t.Fatal("unexpected id")
		}
	})

	t.Run("it should correctly marshal binary ids ", func(t *testing.T) {
		type SUT struct {
			ID StringSerializable `jsonapi:"primary,tests"`
		}

		binary := StringSerializable{1, 2, 3, 4}
		input := SUT{
			ID: binary,
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		check := map[string]interface{}{}
		if err := json.Unmarshal(raw, &check); err != nil {
			t.Fatal(err)
		}

		if check["data"].(map[string]interface{})["id"] != "01020304" {
			t.Fatal("unexpected id")
		}
	})

	t.Run("should correctly marshal pointer ids", func(t *testing.T) {
		type SUT struct {
			ID *StringSerializable `jsonapi:"primary,tests"`
		}

		binary := StringSerializable{1, 2, 3, 4}
		input := SUT{
			ID: &binary,
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		check := map[string]interface{}{}
		if err := json.Unmarshal(raw, &check); err != nil {
			t.Fatal(err)
		}

		if check["data"].(map[string]interface{})["id"] != "01020304" {
			t.Fatal("unexpected id")
		}
	})
}

func TestMarshalTime(t *testing.T) {

	t.Run("should correctly serialise time value as RFC3339 by default", func(t *testing.T) {

		type SUT struct {
			ID    string `jsonapi:"primary,tests"`
			Time  time.Time
			PTime *time.Time
			NTime *time.Time
		}

		ptime := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)

		input := SUT{
			ID:    "1",
			Time:  time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			PTime: &ptime,
			NTime: nil,
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		check := map[string]interface{}{}
		if err := json.Unmarshal(raw, &check); err != nil {
			t.Fatal(err)
		}

		attrs := check["data"].(map[string]interface{})["attributes"].(map[string]interface{})

		if attrs["time"] != "2018-01-01T00:00:00Z" {
			t.Fatal("unexpected time value in value")
		}
		if attrs["pTime"] != "2018-01-01T00:00:00Z" {
			t.Fatal("unexpected time value in pointer")
		}
		if v, ok := attrs["nTime"]; !ok || v != nil {
			t.Fatal("expected nil time pointer to be included as nil")
		}

	})
}

func TestMarshalRecursive(t *testing.T) {

	t.Run("should correctly untangle single recursive references", func(t *testing.T) {
		type Recursive struct {
			ID  string     `jsonapi:"primary,base"`
			Ref *Recursive `jsonapi:"relation,ref"`
		}

		a := Recursive{ID: "1"}
		b := Recursive{ID: "2", Ref: &a}
		a.Ref = &b

		raw, err := Marshal(a)
		if err != nil {
			t.Fatal(err)
		}

		check := map[string]interface{}{}
		if err := json.Unmarshal(raw, &check); err != nil {
			t.Fatal(err)
		}

		if len(check["included"].([]interface{})) != 2 {
			t.Fatal("unexpected number of included resources")
		}
	})

	t.Run("should correctly untangle list recursive references", func(t *testing.T) {
		type Recursive struct {
			ID  string       `jsonapi:"primary,base"`
			Ref []*Recursive `jsonapi:"relation,ref"`
		}

		a := Recursive{ID: "1"}
		b := Recursive{ID: "2", Ref: []*Recursive{&a}}
		a.Ref = []*Recursive{&b}

		raw, err := Marshal(a)
		if err != nil {
			t.Fatal(err)
		}

		check := map[string]interface{}{}
		if err := json.Unmarshal(raw, &check); err != nil {
			t.Fatal(err)
		}

		if len(check["included"].([]interface{})) != 2 {
			t.Fatal("unexpected number of included resources")
		}
	})

}

func TestMixInMeta(t *testing.T) {

	t.Run("should correctly extend the result with metadata", func(t *testing.T) {
		type Doc struct {
			ID string `jsonapi:"primary,collection"`
		}

		docs := []Doc{
			{ID: "1"},
			{ID: "2"},
		}

		raw, err := MarshalMany(docs)

		if err != nil {
			t.Fatal(err)
		}

		ext := map[string]interface{}{
			"page":  1,
			"count": 100,
		}

		extended, err := MixInMeta(raw, ext)
		if err != nil {
			t.Fatal(err)
		}

		check := map[string]interface{}{}
		if err := json.Unmarshal(extended, &check); err != nil {
			t.Fatal(err)
		}

		if check["meta"].(map[string]interface{})["page"] != float64(1) {
			t.Fatal("unexpected metadata value")
		}
		if check["meta"].(map[string]interface{})["count"] != float64(100) {
			t.Fatal("unexpected metadata value")
		}

	})
}

func Test_shallowMerge(t *testing.T) {
	t.Run("should fill in non-zero values from both maps", func(t *testing.T) {
		a := map[string]interface{}{
			"key1": "value1",
			"key2": "",
			"key3": 0,
		}
		b := map[string]interface{}{
			"key2": "value2",
			"key3": 3,
			"key4": nil,
		}

		result := shallowMerge(a, b, isAttributeZero)

		expected := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": 3,
			"key4": nil,
		}

		if len(result) != len(expected) {
			t.Fatalf("expected %d keys, got %d", len(expected), len(result))
		}

		for k, v := range expected {
			if result[k] != v {
				t.Fatalf("expected %v for key %s, got %v", v, k, result[k])
			}
		}
	})

	t.Run("should handle empty maps", func(t *testing.T) {
		a := map[string]interface{}{}
		b := map[string]interface{}{
			"key1": "value1",
		}

		result := shallowMerge(a, b, isAttributeZero)

		expected := map[string]interface{}{
			"key1": "value1",
		}

		if len(result) != len(expected) {
			t.Fatalf("expected %d keys, got %d", len(expected), len(result))
		}

		for k, v := range expected {
			if result[k] != v {
				t.Fatalf("expected %v for key %s, got %v", v, k, result[k])
			}
		}
	})

	t.Run("should handle nil values correctly", func(t *testing.T) {
		a := map[string]interface{}{
			"key1": nil,
		}
		b := map[string]interface{}{
			"key1": "value1",
		}

		result := shallowMerge(a, b, isAttributeZero)

		expected := map[string]interface{}{
			"key1": "value1",
		}

		if len(result) != len(expected) {
			t.Fatalf("expected %d keys, got %d", len(expected), len(result))
		}

		for k, v := range expected {
			if result[k] != v {
				t.Fatalf("expected %v for key %s, got %v", v, k, result[k])
			}
		}
	})

	t.Run("should not drop off zero values on both maps if keys are defined", func(t *testing.T) {
		a := map[string]interface{}{
			"key1": "",
			"key2": 0,
		}
		b := map[string]interface{}{}

		expected := map[string]interface{}{
			"key1": "",
			"key2": 0,
		}

		leftToRight := shallowMerge(a, b, isAttributeZero)
		rightToLeft := shallowMerge(b, a, isAttributeZero)

		if len(rightToLeft) != len(expected) {
			t.Fatalf("expected %d keys, got %d", len(expected), len(rightToLeft))
		}

		if len(leftToRight) != len(expected) {
			t.Fatalf("expected %d keys, got %d", len(expected), len(leftToRight))
		}

		for k, v := range expected {
			if leftToRight[k] != v {
				t.Fatalf("expected %v for key %s, got %v", v, k, leftToRight[k])
			}
			if rightToLeft[k] != v {
				t.Fatalf("expected %v for key %s, got %v", v, k, rightToLeft[k])
			}
		}

	})
}

func Test_isRelationshipZero(t *testing.T) {
	t.Run("should return true for nil data", func(t *testing.T) {
		rel := map[string]interface{}{}
		if !isRelationshipZero(rel) {
			t.Fatal("expected true for nil data")
		}
	})

	t.Run("should return true for empty map data", func(t *testing.T) {
		rel := map[string]interface{}{
			"data": map[string]interface{}{},
		}
		if !isRelationshipZero(rel) {
			t.Fatal("expected true for empty map data")
		}
	})

	t.Run("should return true for empty slice data", func(t *testing.T) {
		rel := map[string]interface{}{
			"data": []interface{}{},
		}
		if !isRelationshipZero(rel) {
			t.Fatal("expected true for empty slice data")
		}
	})

	t.Run("should return false for non-empty map data", func(t *testing.T) {
		rel := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "1",
				"type": "example",
			},
		}
		if isRelationshipZero(rel) {
			t.Fatal("expected false for non-empty map data")
		}
	})

	t.Run("should return false for non-empty slice data", func(t *testing.T) {
		rel := map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"id":   "1",
					"type": "example",
				},
			},
		}
		if isRelationshipZero(rel) {
			t.Fatal("expected false for non-empty slice data")
		}
	})
}
