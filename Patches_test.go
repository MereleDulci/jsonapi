package jsonapi

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type Alias string

const Legit Alias = "Legit"

func (a *Alias) UnmarshalText(text []byte) error {
	switch string(text) {
	case "Legit":
		*a = Legit
		return nil
	default:
		return fmt.Errorf("invalid value: %s", text)
	}
}

func TestUnmarshalPatches_Replace(t *testing.T) {
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

	t.Run("should correctly unmarshal with TextUnmarshaller", func(t *testing.T) {
		raw := `[
			{"op": "replace", "path": "/alias", "value": "Legit"}
		]`

		type SUT struct {
			ID    string `jsonapi:"primary,tests"`
			Alias Alias  `jsonapi:"attr,alias"`
		}

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))

		if err != nil {
			fmt.Println(err)

			t.Fatal(err)
		}

		if len(parsed) != 1 {
			t.Fatalf("expected 1 patch, got %d", len(parsed))
		}

		if parsed[0].Value.(Alias) != Legit {
			t.Fatalf("expected \"Legit\", got %v", parsed[0].Value)
		}
	})
}

func TestUnmarshalPatches_Add(t *testing.T) {

	t.Run("should error if the target type for the update is not a slice", func(t *testing.T) {
		raw := `[
			{"op": "add", "path": "/int", "value": 1}
		]`

		type SUT struct {
			ID  string `jsonapi:"primary,tests"`
			Int int    `jsonapi:"attr,int"`
		}

		_, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))

		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "invalid patch operation - target field is not a slice" {
			t.Fatalf("expected error %v", err)
		}
	})

	t.Run("should correctly unmarshal value primitives", func(t *testing.T) {

		raw := `[
			{"op": "add", "path": "/ints", "value": 1},
			{"op": "add", "path": "/strs", "value": "test"},
			{"op": "add", "path": "/bools", "value": true}
		]`

		type SUT struct {
			ID    string   `jsonapi:"primary,tests"`
			Ints  []int    `jsonapi:"attr,ints"`
			Strs  []string `jsonapi:"attr,strs"`
			Bools []bool   `jsonapi:"attr,bools"`
		}

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 3 {
			t.Fatalf("expected 3 patch, got %d", len(parsed))
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

	t.Run("should correctly unmarshal value primitives", func(t *testing.T) {

		raw := `[
			{"op": "add", "path": "/ints", "value": 1},
			{"op": "add", "path": "/strs", "value": "test"},
			{"op": "add", "path": "/bools", "value": true}
		]`

		type SUT struct {
			ID    string    `jsonapi:"primary,tests"`
			Ints  []*int    `jsonapi:"attr,ints"`
			Strs  []*string `jsonapi:"attr,strs"`
			Bools []*bool   `jsonapi:"attr,bools"`
		}

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 3 {
			t.Fatalf("expected 3 patch, got %d", len(parsed))
		}

		if *parsed[0].Value.(*int) != 1 {
			t.Fatalf("expected 1, got %v", parsed[0].Value)
		}

		if *parsed[1].Value.(*string) != "test" {
			t.Fatalf("expected test, got %v", parsed[1].Value)
		}

		if *parsed[2].Value.(*bool) != true {
			t.Fatalf("expected true, got %v", parsed[2].Value)
		}
	})

	t.Run("should correctly unmarshal value structs", func(t *testing.T) {
		raw := `[
			{"op": "add", "path": "/structs", "value": {"a": 1, "b": "2"}}
		]`

		type Inner struct {
			A int    `json:"a"`
			B string `json:"b"`
		}
		type SUT struct {
			ID      string  `jsonapi:"primary,tests"`
			Structs []Inner `jsonapi:"attr,structs"`
		}

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))

		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 1 {
			t.Fatalf("expected 1 patch, got %d", len(parsed))
		}

		checkval := Inner{
			A: 1,
			B: "2",
		}

		if parsed[0].Value.(Inner) != checkval {
			t.Fatalf("expected Inner{A: 1, B: \"2\"}, got %v", parsed[0].Value)
		}
	})

	t.Run("should correctly unmarshal ptr structs", func(t *testing.T) {
		raw := `[
			{"op": "add", "path": "/ptrs", "value": {"a": 1, "b": "2"}}
		]`

		type Inner struct {
			A int    `json:"a"`
			B string `json:"b"`
		}
		type SUT struct {
			ID   string   `jsonapi:"primary,tests"`
			Ptrs []*Inner `jsonapi:"attr,ptrs"`
		}

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))

		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 1 {
			t.Fatalf("expected 1 patch, got %d", len(parsed))
		}

		checkval := Inner{
			A: 1,
			B: "2",
		}

		if *parsed[0].Value.(*Inner) != checkval {
			t.Fatalf("expected Inner{A: 1, B: \"2\"}, got %v", parsed[0].Value)
		}
	})

	t.Run("should correctly handle pointer to slice itself", func(t *testing.T) {
		raw := `[
			{"op": "add", "path": "/ints", "value": 1},
			{"op": "add", "path": "/strs", "value": "test"},
			{"op": "add", "path": "/bools", "value": true}
		]`

		type SUT struct {
			ID    string    `jsonapi:"primary,tests"`
			Ints  *[]int    `jsonapi:"attr,ints"`
			Strs  *[]string `jsonapi:"attr,strs"`
			Bools *[]bool   `jsonapi:"attr,bools"`
		}

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 3 {
			t.Fatalf("expected 3 patch, got %d", len(parsed))
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

	t.Run("should correctly handle pointer to slice with pointers inside", func(t *testing.T) {
		raw := `[
			{"op": "add", "path": "/ints", "value": 1},
			{"op": "add", "path": "/strs", "value": "test"},
			{"op": "add", "path": "/bools", "value": true}
		]`

		type SUT struct {
			ID    string     `jsonapi:"primary,tests"`
			Ints  *[]*int    `jsonapi:"attr,ints"`
			Strs  *[]*string `jsonapi:"attr,strs"`
			Bools *[]*bool   `jsonapi:"attr,bools"`
		}

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 3 {
			t.Fatalf("expected 3 patch, got %d", len(parsed))
		}

		if *parsed[0].Value.(*int) != 1 {
			t.Fatalf("expected 1, got %v", parsed[0].Value)
		}

		if *parsed[1].Value.(*string) != "test" {
			t.Fatalf("expected test, got %v", parsed[1].Value)
		}

		if *parsed[2].Value.(*bool) != true {
			t.Fatalf("expected true, got %v", parsed[2].Value)
		}
	})

	t.Run("should correctly handle operations on reference fields", func(t *testing.T) {
		raw := `[
			{"op": "add", "path": "/byRef", "value": "1"},
			{"op": "add", "path": "/byVal", "value": "2"}
		]`

		type Referenced struct {
			ID  string `jsonapi:"primary,referenced"`
			Any string `jsonapi:"attr,any"`
		}

		type SUT struct {
			ID    string        `jsonapi:"primary,tests"`
			ByRef []*Referenced `jsonapi:"relation,byRef"`
			ByVal []Referenced  `jsonapi:"relation,byVal"`
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
