package jsonapi

import (
	"encoding/json"
	"maps"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestUnmarshalOne(t *testing.T) {

	t.Run("should correctly unmarshal into a simple model", func(t *testing.T) {
		type Nest struct {
			A string
		}

		type SUT struct {
			ID         string `jsonapi:"primary,test"`
			Str        string
			Slice      []string
			FloatSlice []float64
			Nested     Nest
			Point      *Nest
		}

		input := SUT{
			ID:         "1",
			Str:        "test",
			Slice:      []string{"test"},
			FloatSlice: []float64{1, 2, 3},
			Nested:     Nest{A: "test"},
			Point:      &Nest{A: "test"},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		out := SUT{}
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Fatal(err)
		}

		if *out.Point != *input.Point {
			t.Errorf("expected %+v, got %+v", input, out)
		}

		if !reflect.DeepEqual(out, input) {
			t.Errorf("expected %+v, got %+v", input, out)
		}
	})

	t.Run("should correctly unmarshal into value references", func(t *testing.T) {

		type Rel struct {
			ID  string `jsonapi:"primary,related"`
			Str string
		}

		type SUT struct {
			ID       string `jsonapi:"primary,test"`
			Relation Rel    `jsonapi:"relation,rel"`
			Empty    Rel    `jsonapi:"relation,empty"`
		}

		input := SUT{
			ID:       "1",
			Relation: Rel{ID: "2"},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		out := SUT{}
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Fatal(err)
		}

		if out.Relation != input.Relation {
			t.Errorf("expected %+v, got %+v", input, out)
		}

		if out.Empty != input.Empty {
			t.Errorf("expected %+v, got %+v", input, out)
		}

	})

	t.Run("should correctly unmarshal into pointer references", func(t *testing.T) {
		type Rel struct {
			ID  string `jsonapi:"primary,related"`
			Str string
		}

		type SUT struct {
			ID       string `jsonapi:"primary,test"`
			Relation *Rel   `jsonapi:"relation,rel"`
			Empty    *Rel   `jsonapi:"relation,empty"`
		}

		input := SUT{
			ID:       "1",
			Relation: &Rel{ID: "2"},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		out := SUT{}
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Fatal(err)
		}

		if *out.Relation != *input.Relation {
			t.Errorf("expected %+v, got %+v", input, out)
		}

		if out.Empty != nil || input.Empty != nil {
			t.Errorf("expected %+v, got %+v", input, out)
		}

	})

	t.Run("should correctly unmarshal into value slice references", func(t *testing.T) {

		type Rel struct {
			ID  string `jsonapi:"primary,related"`
			Str string
		}

		type SUT struct {
			ID       string `jsonapi:"primary,test"`
			Relation []Rel  `jsonapi:"relation,rel"`
			Empty    []Rel  `jsonapi:"relation,empty"`
		}

		input := SUT{
			ID: "1",
			Relation: []Rel{
				{ID: "2"},
				{ID: "3"},
			},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		out := SUT{}
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Fatal(err)
		}

		if len(out.Relation) != len(input.Relation) {
			t.Errorf("expected %+v, got %+v", input, out)
		}

		for i, rel := range input.Relation {
			if out.Relation[i] != rel {
				t.Errorf("expected %+v, got %+v", input, out)
			}
		}

		if len(out.Empty) != 0 || len(input.Empty) != 0 {
			t.Errorf("expected %+v, got %+v", input, out)
		}

	})

	t.Run("should correctly unmarshal into pointer slice references", func(t *testing.T) {
		type Rel struct {
			ID  string `jsonapi:"primary,related"`
			Str string
		}

		type SUT struct {
			ID       string `jsonapi:"primary,test"`
			Relation []*Rel `jsonapi:"relation,rel"`
			Empty    []*Rel `jsonapi:"relation,empty"`
		}

		input := SUT{
			ID: "1",
			Relation: []*Rel{
				{ID: "2"},
				{ID: "3"},
			},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		out := SUT{}
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Fatal(err)
		}

		if len(out.Relation) != len(input.Relation) {
			t.Errorf("expected %+v, got %+v", input, out)
		}

		for i, rel := range input.Relation {
			if *out.Relation[i] != *rel {
				t.Errorf("expected %+v, got %+v", input, out)
			}
		}

		if len(out.Empty) != 0 || len(input.Empty) != 0 {
			t.Errorf("expected %+v, got %+v", input, out)
		}

	})

	t.Run("should apply json.Unmarshaler if available", func(t *testing.T) {
		type SUT struct {
			ID  string `jsonapi:"primary,tests"`
			Val nestedWithMarshalInner
			Ref *nestedWithMarshalInner
		}

		raw := `{"data": {"type": "tests", "id": "1", "attributes": {"val": {"b": 1}, "ref": {"b": 2}}}}`
		out := SUT{}
		err := Unmarshal([]byte(raw), &out)
		if err != nil {
			t.Fatal(err)
		}
		if out.Val.A != 1 {
			t.Fatal("unexpected attribute value")
		}

		if out.Ref.A != 2 {
			t.Fatal("unexpected attribute value")
		}
	})
}

func TestUnmarshalMany(t *testing.T) {

	t.Run("should correctly unmarshal into a slice of values", func(t *testing.T) {

		type SUT struct {
			ID  string `jsonapi:"primary,test"`
			Str string
		}

		input := []SUT{
			{ID: "1", Str: "test"},
			{ID: "2", Str: "test"},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		var out []SUT
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Fatal(err)
		}

		if len(out) != len(input) {
			t.Errorf("expected %+v, got %+v", input, out)
		}

		for i, sut := range input {
			if out[i] != sut {
				t.Errorf("expected %+v, got %+v", input, out)
			}
		}

	})

	t.Run("should correctly unmarshal into a slice of pointers", func(t *testing.T) {
		type SUT struct {
			ID  string `jsonapi:"primary,test"`
			Str string
		}

		input := []*SUT{
			{ID: "1", Str: "test"},
			{ID: "2", Str: "test"},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		var out []*SUT
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Fatal(err)
		}

		if len(out) != len(input) {
			t.Errorf("expected %+v, got %+v", input, out)
		}

		for i, sut := range input {
			if *out[i] != *sut {
				t.Errorf("expected %+v, got %+v", input, out)
			}
		}
	})
}

func TestUnmarshalID(t *testing.T) {

	t.Run("it should correctly unmarshal primitive ids ", func(t *testing.T) {
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

		check := SUT{}
		if err := Unmarshal(raw, &check); err != nil {
			t.Fatal(err)
		}

		if check.ID != input.ID {
			t.Fatal("unexpected id")
		}
	})

	t.Run("it should correctly unmarshal binary ids ", func(t *testing.T) {
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

		check := SUT{}
		if err := Unmarshal(raw, &check); err != nil {
			t.Fatal(err)
		}

		if check.ID != input.ID {
			t.Fatal("unexpected id")
		}
	})

	t.Run("should correctly unmarshal pointer ids", func(t *testing.T) {
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

		check := SUT{}
		if err := Unmarshal(raw, &check); err != nil {
			t.Fatal(err)
		}

		if *check.ID != *input.ID {
			t.Fatal("unexpected id")
		}
	})

	t.Run("should omit id if it is empty", func(t *testing.T) {
		type SUT struct {
			ID  string `jsonapi:"primary,tests"`
			Val string
		}

		raw := `{"data": {"type": "tests", "attributes": {"val": "test"}}}`

		out := SUT{}
		err := Unmarshal([]byte(raw), &out)
		if err != nil {
			t.Fatal(err)
		}

		if out.ID != "" {
			t.Fatal("unexpected id")
		}
	})
}

func TestUnmarshalAsType(t *testing.T) {

	t.Run("should correctly unmarshal into a single model defined as type", func(t *testing.T) {

		type SUT struct {
			ID  string `jsonapi:"primary,test"`
			Str string
		}

		input := &SUT{
			ID:  "1",
			Str: "test",
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		out, err := UnmarshalOneAsType(raw, reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(out, input) {
			t.Errorf("expected %+v, got %+v", input, out)
		}

		if *out.(*SUT) != *input {
			t.Errorf("expected %+v, got %+v", input, out)
		}
	})

	t.Run("should correctly unmarshal into a slice of models defined as type", func(t *testing.T) {
		type SUT struct {
			ID  string `jsonapi:"primary,test"`
			Str string
		}

		input := []*SUT{
			{ID: "1", Str: "test"},
			{ID: "2", Str: "test"},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		out, err := UnmarshalManyAsType(raw, reflect.TypeOf(new(SUT)))
		if err != nil {
			t.Fatal(err)
		}

		if _, ok := out[0].(*SUT); !ok {
			t.Fatal("expecting a slice of pointers")
		}

		if !reflect.DeepEqual(out[0], input[0]) {
			t.Errorf("expected %+v, got %+v", input[0], out[0])
		}
		if !reflect.DeepEqual(out[1], input[1]) {
			t.Errorf("expected %+v, got %+v", input[1], out[1])
		}
	})
}

func TestUnmarshalTime(t *testing.T) {

	t.Run("should correctly unmarshal into a time.Time in RFC3339 format by default", func(t *testing.T) {
		type SUT struct {
			ID    string `jsonapi:"primary,test"`
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

		out := SUT{}
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Fatal(err)
		}

		if !out.Time.Equal(input.Time) {
			t.Errorf("expected %+v, got %+v", input, out)
		}
	})
}

func TestUnmarshalMap(t *testing.T) {
	t.Run("should correct unmarshal non-nil maps", func(t *testing.T) {
		type SUT struct {
			ID string `jsonapi:"primary,test"`
			M  map[string]interface{}
		}

		s := SUT{
			ID: "1",
			M: map[string]interface{}{
				"s": "test",
				"f": 1.1,
				"b": true,
			},
		}

		raw, err := Marshal(s)
		if err != nil {
			t.Fatal(err)
		}

		out := SUT{}
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Fatal(err)
		}

		if !maps.Equal(s.M, out.M) {
			t.Errorf("expected %+v, got %+v", s.M, out.M)
		}
	})

	t.Run("casts integer values of interface{} map to float", func(t *testing.T) {
		type SUT struct {
			ID string `jsonapi:"primary,test"`
			M  map[string]interface{}
		}

		s := SUT{
			ID: "1",
			M: map[string]interface{}{
				"i": 1,
			},
		}

		raw, err := Marshal(s)
		if err != nil {
			t.Fatal(err)
		}

		out := SUT{}
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Fatal(err)
		}

		cast, ok := out.M["i"].(float64)
		if cast != 1 {
			t.Errorf("expected 1, got %v", out.M["i"])
		}
		if !ok {
			t.Errorf("expected cast to float64, got %T", out.M["i"])
		}
	})

	t.Run("should correctly unmarshal nil map", func(t *testing.T) {
		type SUT struct {
			ID string `jsonapi:"primary,test"`
			M  map[string]interface{}
		}

		s := SUT{
			ID: "1",
		}

		raw, err := Marshal(s)
		if err != nil {
			t.Fatal(err)
		}

		out := SUT{}
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Fatal(err)
		}

		if out.M != nil {
			t.Errorf("expected nil map, got %+v", out.M)
		}
	})

	t.Run("should correctly marshal typed map values", func(t *testing.T) {
		type SUT struct {
			ID string `jsonapi:"primary,test"`
			Ms map[string]string
			Mi map[string]int
			Mf map[string]float64
		}

		s := SUT{
			ID: "1",
			Ms: map[string]string{
				"s": "test",
			},
			Mi: map[string]int{
				"i": 1,
			},
			Mf: map[string]float64{
				"f": 1.1,
			},
		}

		raw, err := Marshal(s)
		if err != nil {
			t.Fatal(err)
		}

		out := SUT{}
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Fatal(err)
		}

		if !maps.Equal(s.Ms, out.Ms) {
			t.Errorf("expected %+v, got %+v", s.Ms, out.Ms)
		}

		if !maps.Equal(s.Mi, out.Mi) {
			t.Errorf("expected %+v, got %+v", s.Mi, out.Mi)
		}
		if !maps.Equal(s.Mf, out.Mf) {
			t.Errorf("expected %+v, got %+v", s.Mf, out.Mf)
		}
	})
}

func TestUnmarshalWithIncluded(t *testing.T) {

	t.Run("should correctly unmarshal included doc into a single value", func(t *testing.T) {
		type Ref struct {
			ID  string `jsonapi:"primary,references"`
			Val string `jsonapi:"attr,val"`
		}

		type Main struct {
			ID  string `jsonapi:"primary,mains"`
			Ref *Ref   `jsonapi:"relation,ref"`
			Val Ref    `jsonapi:"relation,val"`
		}

		input := Main{
			ID:  "1",
			Ref: &Ref{ID: "2", Val: "test"},
			Val: Ref{ID: "3", Val: "test"},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		out := Main{}
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Fatal(err)
		}

		if out.Val != input.Val {
			t.Errorf("expected %+v, got %+v", input.Val, out.Val)
		}
		if *out.Ref != *input.Ref {
			t.Errorf("expected %+v, got %+v", input.Ref, out.Ref)
		}
	})

	t.Run("should correctly unmarshal included doc into a slice of values", func(t *testing.T) {
		type Ref struct {
			ID  string `jsonapi:"primary,references"`
			Val string `jsonapi:"attr,val"`
		}

		type Main struct {
			ID   string `jsonapi:"primary,mains"`
			Refs []*Ref `jsonapi:"relation,refs"`
			Vals []Ref  `jsonapi:"relation,vals"`
		}

		input := Main{
			ID: "1",
			Refs: []*Ref{
				{ID: "2", Val: "test"},
				{ID: "3", Val: "test"},
			},
			Vals: []Ref{
				{ID: "4", Val: "test"},
				{ID: "5", Val: "test"},
			},
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)

		}

		out := Main{}
		err = Unmarshal(raw, &out)
		if err != nil {
			t.Fatal(err)
		}

		if len(out.Refs) != len(input.Refs) {
			t.Errorf("expected %+v, got %+v", input.Refs, out.Refs)
		}

		for i, ref := range input.Refs {
			if *out.Refs[i] != *ref {
				t.Errorf("expected %+v, got %+v", input.Refs[i], out.Refs[i])
			}
		}

		if len(out.Vals) != len(input.Vals) {
			t.Errorf("expected %+v, got %+v", input.Vals, out.Vals)
		}

		for i, val := range input.Vals {
			if out.Vals[i] != val {
				t.Errorf("expected %+v, got %+v", input.Vals[i], out.Vals[i])
			}
		}

	})
}

func TestUnmarshalStability(t *testing.T) {

	t.Run("should not panic on mismatching type and return it as error value", func(t *testing.T) {

		type Main struct {
			ID     string  `jsonapi:"primary,mains"`
			Float  float64 `jsonapi:"attr,float"`
			String string  `jsonapi:"attr,string"`
		}

		floatIn := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "1",
				"type": "mains",
				"attributes": map[string]interface{}{
					"float": "as-string",
				},
			},
		}

		floatMismatchRaw, err := json.Marshal(floatIn)
		if err != nil {
			t.Fatal(err)
		}

		out := &Main{}
		err = Unmarshal(floatMismatchRaw, out)

		if err == nil {
			t.Errorf("expected error on float type mismatch")
		}
		if !strings.Contains(err.Error(), "unmarshal attribute float") {
			t.Errorf("expected error to be related to float field unmarshal, got, %+v", err)
		}

		strIn := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "1",
				"type": "mains",
				"attributes": map[string]interface{}{
					"string": 123,
				},
			},
		}

		stringMismatchRaw, err := json.Marshal(strIn)

		err = Unmarshal(stringMismatchRaw, out)

		if err == nil {
			t.Errorf("expected error on string type mismatch")
		}

		if !strings.Contains(err.Error(), "unmarshal attribute string") {
			t.Errorf("expected error to be related to string field unmarshal, got, %+v", err)
		}

	})

	t.Run("should yield meaningful error if provided type is not deserializable", func(t *testing.T) {
		type SUT struct {
			ID  string              `jsonapi:"primary,sut"`
			Val InvalidSerializable `jsonapi:"attr,val"`
		}

		input := &SUT{
			ID:  "1",
			Val: InvalidSerializableValue,
		}

		raw, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		out, err := UnmarshalOneAsType(raw, reflect.TypeOf(new(SUT)))

		if out != nil {
			t.Errorf("expected no output, got %v", out)
		}
		if err == nil {
			t.Errorf("expected error")
		}

		if !strings.Contains(err.Error(), "string is not assignable to type jsonapi.InvalidSerializable") {
			t.Fatal(err)
		}
	})
}
