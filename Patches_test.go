package jsonapi

import (
	"reflect"
	"testing"
	"time"
)

func TestUnmarshalPatches(t *testing.T) {
	t.Run("should unmarshal primitives", func(t *testing.T) {
		raw := `[
			{"op": "replace", "path": "/int", "value": 1},
			{"op": "replace", "path": "/str", "value": "test"},
			{"op": "replace", "path": "/bool", "value": true}
		]`

		type SUT struct {
			ID   string `jsonapi:"primary,tests"`
			Int  int    `jsonapi:"attr,int"`
			Str  string `jsonapi:"attr,str"`
			Bool bool   `jsonapi:"attr,bool"`
		}

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 3 {
			t.Fatalf("expected 3 patches, got %d", len(parsed))
		}

		if parsed[0].Value.(int) != 1 {
			t.Fatalf("expected 1, got %v", parsed[0].Value)
		}
		if parsed[1].Value.(string) != "test" {
			t.Fatalf("expected test, got %v", parsed[1].Value)
		}
		if parsed[2].Value.(bool) != true {
			t.Fatalf("expected true, got %v", parsed[2].Value)
		}
	})

	t.Run("should unmarshal embedded structs", func(t *testing.T) {
		raw := `[
			{"op": "replace", "path": "/struct", "value": {"a": 1, "b": "2"}},
			{"op": "replace", "path": "/ptr", "value": {"a": 1, "b": "2"}},
			{"op": "replace", "path": "/time", "value": "2023-10-01T00:00:00Z"}
		]`

		type Inner struct {
			A int    `json:"a"`
			B string `json:"b"`
		}
		type SUT struct {
			ID     string    `jsonapi:"primary,tests"`
			Struct Inner     `jsonapi:"attr,struct"`
			Ptr    *Inner    `jsonapi:"attr,ptr"`
			Time   time.Time `jsonapi:"attr,time"`
		}

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 3 {
			t.Fatalf("expected 3 patches, got %d", len(parsed))
		}

		checkval := Inner{
			A: 1,
			B: "2",
		}

		if parsed[0].Value.(Inner) != checkval {
			t.Fatalf("expected Inner{A: 1, B: \"2\"}, got %v", parsed[0].Value)
		}
		if *parsed[1].Value.(*Inner) != checkval {
			t.Fatalf("expected &Inner{A: 1, B: \"2\"}, got %v", parsed[1].Value)
		}

		checktime := time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)
		if parsed[2].Value.(time.Time) != checktime {
			t.Fatalf("expected 2023-10-01T00:00:00Z, got %v", parsed[2].Value)
		}
	})

	t.Run("should correctly unmarshal values targeting field of embedded structs", func(t *testing.T) {
		raw := `[
			{"op": "replace", "path": "/struct/a", "value": 1},
			{"op": "replace", "path": "/ptr/a", "value": 1}
		]`
		type Inner struct {
			A int `json:"a"`
		}
		type SUT struct {
			ID     string `jsonapi:"primary,tests"`
			Struct Inner  `jsonapi:"attr,struct"`
			Ptr    *Inner `jsonapi:"attr,ptr"`
		}

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 2 {
			t.Fatalf("expected 1 patch, got %d", len(parsed))
		}

		if parsed[0].Value.(int) != 1 {
			t.Fatalf("expected 1, got %v", parsed[0].Value)
		}
		if parsed[1].Value.(int) != 1 {
			t.Fatalf("expected 1, got %v", parsed[1].Value)
		}
	})

	t.Run("should correctly unmarshal nullable values", func(t *testing.T) {
		raw := `[
			{"op": "replace", "path": "/struct", "value": null}
		]`

		type Inner struct {
			A int `json:"a"`
		}
		type SUT struct {
			ID     string `jsonapi:"primary,tests"`
			Struct *Inner `jsonapi:"attr,struct"`
		}

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 1 {
			t.Fatalf("expected 1 patch, got %d", len(parsed))
		}

		compare := SUT{}

		if parsed[0].Value != compare.Struct {
			t.Fatalf("expected nil, got %v", parsed[0].Value)
		}
	})

	t.Run("should correctly unmarshal references", func(t *testing.T) {
		//patches are limited to update reference ids only, not the properties of the related object
		raw := `[
			{"op": "replace", "path": "/byRef", "value": "1"},
			{"op": "replace", "path": "/byVal", "value": "2"}
		]`
		type Referenced struct {
			ID  string `jsonapi:"primary,referenced"`
			Any string `jsonapi:"attr,any"`
		}

		type SUT struct {
			ID    string      `jsonapi:"primary,tests"`
			ByRef *Referenced `jsonapi:"relation,byRef"`
			ByVal Referenced  `jsonapi:"relation,byVal"`
		}

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 2 {
			t.Fatalf("expected 2 patches, got %d", len(parsed))
		}

		if parsed[0].Value.(string) != "1" {
			t.Fatalf("expected \"1\", got %v", parsed[0].Value)
		}

		if parsed[1].Value.(string) != "2" {
			t.Fatalf("expected \"2\", got %v", parsed[1].Value)
		}
	})
}
