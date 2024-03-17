package jsonapi

import (
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestDataTypeIntegrity(t *testing.T) {

	t.Run("should error if two field utilise the same relationship name", func(t *testing.T) {

		type Rel struct {
			ID  string `jsonapi:"primary,related"`
			Str string
		}

		type Invalid struct {
			ID       string `jsonapi:"primary,test"`
			Relation Rel    `jsonapi:"relation,rel"`
			Empty    Rel    `jsonapi:"relation,rel"`
		}

		type Valid struct {
			ID       string `jsonapi:"primary,test"`
			Relation Rel    `jsonapi:"relation,rel"`
			Empty    Rel    `jsonapi:"relation,empty"`
		}

		validInput := Valid{
			ID:       "1",
			Relation: Rel{ID: "2"},
			Empty:    Rel{ID: "3"},
		}

		invalidInput := Invalid{
			ID:       "1",
			Relation: Rel{ID: "2"},
			Empty:    Rel{ID: "3"},
		}

		_, err := Marshal(invalidInput)
		if err == nil {
			t.Fatal("expected error")
		}

		if !strings.Contains(err.Error(), "relationship name already used: rel") {
			t.Fatal("expected error to contain 'relationship name already used rel', got", err.Error())
		}

		_, err = Marshal(validInput)
		if err != nil {
			t.Fatal(err)
		}

	})

	t.Run("should correctly handle all primitive data types", func(t *testing.T) {

		type Primary struct {
			ID       string `jsonapi:"primary,test"`
			Str      string
			PStr     *string
			Int      int
			PInt     *int
			Int16    int16
			PInt16   *int16
			Float    float64
			PFloat   *float64
			Bool     bool
			PBool    *bool
			LString  []string
			LPString []*string
			LInt     []int
			LPInt    []*int
			LInt16   []int16
			LPInt16  []*int16
			LFloat   []float64
			LPFloat  []*float64
			LBool    []bool
			LPBool   []*bool
		}

		input := Primary{
			ID:       "1",
			Str:      "string",
			PStr:     new(string),
			Int:      1,
			PInt:     new(int),
			Int16:    int16(16),
			PInt16:   new(int16),
			Float:    1.1,
			PFloat:   new(float64),
			Bool:     true,
			PBool:    new(bool),
			LString:  []string{"string"},
			LPString: []*string{new(string)},
			LInt:     []int{1},
			LPInt:    []*int{new(int)},
			LInt16:   []int16{16},
			LPInt16:  []*int16{new(int16)},
			LFloat:   []float64{1.1},
			LPFloat:  []*float64{new(float64)},
			LBool:    []bool{true},
			LPBool:   []*bool{new(bool)},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Error(err)
		}

		out := Primary{}
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(input, out) {
			t.Error("input and output are not equal")
		}
	})

	t.Run("should correctly handle primitive types implementing TextMarshaller / TextUnmarshaller", func(t *testing.T) {
		binary := PrimitiveSerializable(2)

		type Primary struct {
			ID  PrimitiveSerializable `jsonapi:"primary,test"`
			Val PrimitiveSerializable
			Ref *PrimitiveSerializable
			S   []PrimitiveSerializable
			P   []*PrimitiveSerializable
		}

		input := Primary{
			ID:  binary,
			Val: binary,
			Ref: &binary,
			S:   []PrimitiveSerializable{binary},
			P:   []*PrimitiveSerializable{&binary},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Error(err)
		}

		intermediate := map[string]interface{}{}
		err = json.Unmarshal(raw, &intermediate)
		if err != nil {
			t.Error(err)
		}

		if intermediate["data"].(map[string]interface{})["id"] != "2" {
			t.Error("binary id is not correct")
		}

		out := Primary{}
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(input, out) {
			t.Fatal("input and output are not equal")
		}
	})

	t.Run("should correctly handle non primitive types implementing TextMarshaller / TextUnmarshaller", func(t *testing.T) {

		binary := StringSerializable{1, 2, 170, 255}

		type Primary struct {
			ID  StringSerializable `jsonapi:"primary,test"`
			Val StringSerializable
			Ref *StringSerializable
			S   []StringSerializable
			P   []*StringSerializable
		}

		input := Primary{
			ID:  binary,
			Val: binary,
			Ref: &binary,
			S:   []StringSerializable{binary},
			P:   []*StringSerializable{&binary},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Error(err)
		}

		intermediate := map[string]interface{}{}
		err = json.Unmarshal(raw, &intermediate)
		if err != nil {
			t.Error(err)
		}

		if intermediate["data"].(map[string]interface{})["id"] != "0102aaff" {
			t.Error("binary id is not correct")
		}

		out := Primary{}
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(input, out) {
			t.Fatal("input and output are not equal")
		}
	})

}

func TestStructTags(t *testing.T) {

	t.Run("should correctly marshal a struct without tags", func(t *testing.T) {

		type SUT struct {
			ID  string `jsonapi:"primary,tests"`
			Val string
		}

		input := SUT{
			ID:  "1",
			Val: "test",
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
		attrs := check["data"].(map[string]interface{})["attributes"].(map[string]interface{})
		if attrs["val"] != "test" {
			t.Fatal("unexpected attribute value")
		}

		out := SUT{}
		if err := Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}

		if out.Val != input.Val {
			t.Fatal("unexpected attribute value")
		}
	})

	t.Run("should correctly marshal struct using json tags", func(t *testing.T) {

		type SUT struct {
			ID     string `jsonapi:"primary,tests"`
			Val    string `json:"value"`
			Ignore string `json:"-"`
			Nested struct {
				A string `json:"aValue"`
				B string `json:"-"`
			} `json:"nested"`
			Persisted *struct{} `json:"persisted"`
			Point     *struct{} `json:"point,omitempty"`
		}

		input := SUT{
			ID:     "1",
			Val:    "test",
			Ignore: "ignore",
			Nested: struct {
				A string `json:"aValue"`
				B string `json:"-"`
			}{
				A: "test",
				B: "test",
			},
			Persisted: nil,
			Point:     nil,
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
		if attrs["value"] != "test" {
			t.Fatal("unexpected attribute value")
		}

		if _, ok := attrs["ignore"]; ok {
			t.Fatal("expected ignore to be omitted")
		}

		if _, ok := attrs["persisted"]; !ok {
			t.Fatal("expected persisted to be included")
		}

		if _, ok := attrs["point"]; ok {
			t.Fatal("expected point to be omitted")
		}

		if _, ok := attrs["nested"].(map[string]interface{})["aValue"]; !ok {
			t.Fatal("expected nested attribute to be included with correct field name")
		}

		out := SUT{}
		if err := Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}

		if out.Val != input.Val {
			t.Fatal("unexpected attribute value")
		}
		if out.Nested.A != input.Nested.A {
			t.Fatal("unexpected nested attribute value")
		}
	})

	t.Run("should correctly marshal struct using jsonapi tags", func(t *testing.T) {
		type SUT struct {
			ID     string `jsonapi:"primary,tests"`
			Val    string `jsonapi:"attr,value"`
			Ignore string `jsonapi:"attr,-"`
			Nested struct {
				A string `jsonapi:"attr,aValue"`
				B string `jsonapi:"attr,-"`
			}
			Persisted *struct{} `jsonapi:"attr,persisted"`
			Point     *struct{} `jsonapi:"attr,point,omitempty"`
		}

		input := SUT{
			ID:     "1",
			Val:    "test",
			Ignore: "ignore",
			Nested: struct {
				A string `jsonapi:"attr,aValue"`
				B string `jsonapi:"attr,-"`
			}{
				A: "test",
			},
			Persisted: nil,
			Point:     nil,
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
		if attrs["value"] != "test" {
			t.Fatal("unexpected attribute value")
		}

		if _, ok := attrs["ignore"]; ok {
			t.Fatal("expected ignore to be omitted")
		}

		if _, ok := attrs["persisted"]; !ok {
			t.Fatal("expected persisted to be included")
		}

		if _, ok := attrs["point"]; ok {
			t.Fatal("expected point to be omitted")
		}

		if _, ok := attrs["nested"].(map[string]interface{})["aValue"]; !ok {
			t.Fatal("expected nested attribute to be included with correct field name")
		}

		out := SUT{}
		if err := Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}

		if out.Val != input.Val {
			t.Fatal("unexpected attribute value")
		}

		if out.Nested.A != input.Nested.A {
			t.Fatal("unexpected nested attribute value")
		}
	})
}

type StringSerializable [4]byte

func (b StringSerializable) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(b[:])), nil
}

func (b *StringSerializable) UnmarshalText(text []byte) error {
	decoded, err := hex.DecodeString(string(text))
	if err != nil {
		return err
	}
	copy(b[:], decoded)

	return nil
}

type PrimitiveSerializable int

func (p PrimitiveSerializable) MarshalText() ([]byte, error) {
	return []byte(strconv.Itoa(int(p))), nil
}

func (p *PrimitiveSerializable) UnmarshalText(text []byte) error {
	v, err := strconv.Atoi(string(text))
	if err != nil {
		return err
	}
	*p = PrimitiveSerializable(v)
	return nil
}
