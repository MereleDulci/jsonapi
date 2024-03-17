package jsonapi

import (
	"reflect"
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
