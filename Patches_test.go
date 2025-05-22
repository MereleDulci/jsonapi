package jsonapi

import (
	"encoding/hex"
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
			{"op": "replace", "path": "/byVal", "value": "2"},
			{"op": "replace", "path": "/listByRef", "value": ["1"]},
			{"op": "replace", "path": "/listByVal", "value": ["2"]}
		]`
		type Referenced struct {
			ID  string `jsonapi:"primary,referenced"`
			Any string `jsonapi:"attr,any"`
		}

		type SUT struct {
			ID        string        `jsonapi:"primary,tests"`
			ByRef     *Referenced   `jsonapi:"relation,byRef"`
			ByVal     Referenced    `jsonapi:"relation,byVal"`
			ListByRef []*Referenced `jsonapi:"relation,listByRef"`
			ListByVal []Referenced  `jsonapi:"relation,listByVal"`
		}

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 4 {
			t.Fatalf("expected 4 patches, got %d", len(parsed))
		}

		if parsed[0].Value.(string) != "1" {
			t.Fatalf("expected \"1\", got %v", parsed[0].Value)
		}

		if parsed[1].Value.(string) != "2" {
			t.Fatalf("expected \"2\", got %v", parsed[1].Value)
		}

		if parsed[2].Value.([]interface{})[0].(string) != "1" {
			t.Fatalf("expected \"1\", got %v", parsed[2].Value)
		}

		if parsed[3].Value.([]interface{})[0].(string) != "2" {
			t.Fatalf("expected \"2\", got %v", parsed[3].Value)
		}
	})

	t.Run("should correctly unmarshal with TextUnmarshaller", func(t *testing.T) {
		raw := `[
			{"op": "replace", "path": "/byVal", "value": "0102aaff"},
			{"op": "replace", "path": "/byRef", "value": "0202aaff"},
			{"op": "replace", "path": "/listByVal", "value": ["0302aaff"]},
			{"op": "replace", "path": "/listByRef", "value": ["0402aaff"]}
		]`

		type SUT struct {
			ID        string                `jsonapi:"primary,tests"`
			ByVal     StringSerializable    `jsonapi:"attr,byVal"`
			ByRef     *StringSerializable   `jsonapi:"attr,byRef"`
			ListByVal []StringSerializable  `jsonapi:"attr,listByVal"`
			ListByRef []*StringSerializable `jsonapi:"attr,listByRef"`
		}

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))

		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 4 {
			t.Fatalf("expected 1 patch, got %d", len(parsed))
		}

		for i, pattern := range []string{"0102aaff", "0202aaff", "0302aaff", "0402aaff"} {
			expected, _ := hex.DecodeString(pattern)

			switch i {
			case 0:
				if parsed[i].Value.(StringSerializable) != StringSerializable(expected) {
					t.Fatalf("expected %s, got %v", pattern, parsed[i].Value)
				}
			case 1:
				if *parsed[i].Value.(*StringSerializable) != StringSerializable(expected) {
					t.Fatalf("expected %s, got %v", pattern, parsed[i].Value)
				}
			case 2:
				if parsed[i].Value.([]StringSerializable)[0] != StringSerializable(expected) {
					t.Fatalf("expected %s, got %v", pattern, parsed[i].Value)
				}
			case 3:
				if *parsed[i].Value.([]*StringSerializable)[0] != StringSerializable(expected) {
					t.Fatalf("expected %s, got %v", pattern, parsed[i].Value)
				}
			}
		}

	})

	t.Run("replace on maps with TextUnmarshaler keys", func(t *testing.T) {
		type SUT struct {
			ID  string           `jsonapi:"primary,tests"`
			Map map[Alias]string `jsonapi:"attr,map"`
		}

		raw := `[
			{"op": "replace", "path": "/map/Legit", "value": "test"}
		]`

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 1 {
			t.Fatalf("expected 1 patch, got %d", len(parsed))
		}

		if parsed[0].Value.(string) != "test" {
			t.Fatalf("expected test, got %v", parsed[0].Value)
		}

		if parsed[0].Path != "/map/Legit" {
			t.Fatalf("expected /map/Legit, got %s", parsed[0].Path)
		}
	})

	t.Run("replace of map value with TextUnmarshaler key", func(t *testing.T) {
		type SUT struct {
			ID  string           `jsonapi:"primary,tests"`
			Map map[Alias]string `jsonapi:"attr,map"`
		}

		raw := `[
			{"op": "replace", "path": "/map", "value": {"Legit": "test"}}
		]`

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 1 {
			t.Fatalf("expected 1 patch, got %d", len(parsed))
		}

		typed, ok := parsed[0].Value.(map[Alias]string)
		if !ok {
			t.Fatalf("expected map[Alias]string, got %T", parsed[0].Value)
		}

		if typed["Legit"] != "test" {
			t.Fatalf("expected test, got %v", parsed[0].Value)
		}

		if parsed[0].Path != "/map" {
			t.Fatalf("expected /map, got %s", parsed[0].Path)
		}
	})

	t.Run("replace of full map value with struct value", func(t *testing.T) {

		type Inner struct {
			A int `json:"a"`
		}

		type SUT struct {
			ID  string           `jsonapi:"primary,tests"`
			Map map[string]Inner `jsonapi:"attr,map"`
		}

		raw := `[
			{"op": "replace", "path": "/map", "value": {"test": {"a": 1}}}
		]`

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 1 {
			t.Fatalf("expected 1 patch, got %d", len(parsed))
		}

		typed, ok := parsed[0].Value.(map[string]Inner)
		if !ok {
			t.Fatalf("expected map[string]Inner, got %T", parsed[0].Value)
		}

		if typed["test"].A != 1 {
			t.Fatalf("expected 1, got %v", parsed[0].Value)
		}

		if parsed[0].Path != "/map" {
			t.Fatalf("expected /map, got %s", parsed[0].Path)
		}

		if parsed[0].Value.(map[string]Inner)["test"].A != 1 {
			t.Fatalf("expected 1, got %v", parsed[0].Value)
		}
	})

	t.Run("replace of partial map value with struct value", func(t *testing.T) {
		type Inner struct {
			A int `json:"a"`
		}

		type SUT struct {
			ID  string           `jsonapi:"primary,tests"`
			Map map[string]Inner `jsonapi:"attr,map"`
		}

		raw := `[
			{"op": "replace", "path": "/map/test", "value": {"a": 1}},
			{"op": "replace", "path": "/map/test/a", "value": 2}
		]`

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 2 {
			t.Fatalf("expected 2 patches, got %d", len(parsed))
		}

		typed, ok := parsed[0].Value.(Inner)
		if !ok {
			t.Fatalf("expected Inner, got %T", parsed[0].Value)
		}

		if typed.A != 1 {
			t.Fatalf("expected 1, got %v", parsed[0].Value)
		}

		if parsed[0].Path != "/map/test" {
			t.Fatalf("expected /map/test, got %s", parsed[0].Path)
		}

		val, ok := parsed[1].Value.(int)
		if !ok {
			t.Fatalf("expected Inner, got %T", parsed[1].Value)
		}

		if val != 2 {
			t.Fatalf("expected 2, got %v", parsed[1].Value)
		}

		if parsed[1].Path != "/map/test/a" {
			t.Fatalf("expected /map/test/a, got %s", parsed[1].Path)
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

	t.Run("should correctly handle operations on reference fields with UnmarshalText", func(t *testing.T) {

		type Referenced struct {
			ID  StringSerializable `jsonapi:"primary,referenced"`
			Any string             `jsonapi:"attr,any"`
		}

		type SUT struct {
			ID    string        `jsonapi:"primary,tests"`
			ByRef []*Referenced `jsonapi:"relation,byRef"`
			ByVal []Referenced  `jsonapi:"relation,byVal"`
		}

		raw := `[
			{"op": "add", "path": "/byRef", "value": "0102aaff"},
			{"op": "add", "path": "/byVal", "value": "0202aaff"}
		]`

		parsed, err := UnmarshalPatches([]byte(raw), reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if len(parsed) != 2 {
			t.Fatalf("expected 2 patches, got %d", len(parsed))
		}

		expectedA, _ := hex.DecodeString("0102aaff")
		expectedB, _ := hex.DecodeString("0202aaff")
		if parsed[0].Value.(StringSerializable) != StringSerializable(expectedA) {
			t.Fatalf("expected 0102aaff, got %v", parsed[0].Value)
		}

		if parsed[1].Value.(StringSerializable) != StringSerializable(expectedB) {
			t.Fatalf("expected 0202aaff, got %v", parsed[1].Value)
		}

	})
}
