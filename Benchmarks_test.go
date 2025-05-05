package jsonapi

import (
	"encoding/json"
	"reflect"
	"testing"
)

type secondRel struct {
	ID string `jsonapi:"primary,second-related"`
}

type firstRel struct {
	ID        string    `jsonapi:"primary,first-related"`
	SecondRel secondRel `jsonapi:"relation,secondRel"`
}

type embed struct {
	Value   string `json:"value"`
	IntVal  int64
	Empty   *struct{} `json:"empty,omitempty"`
	ListVal []string
}

type base struct {
	ID            string `jsonapi:"primary,resource-name"`
	ExportedField string
	Embedded      embed
	ERef          *embed
	Related       firstRel   `jsonapi:"relation,related"`
	RList         []firstRel `jsonapi:"relation,rList"`
	private       string
}

func BenchmarkMarshalVsJson(b *testing.B) {
	input := base{
		ID:            "1",
		ExportedField: "test",
		Embedded: embed{
			Value:   "test",
			IntVal:  1,
			ListVal: []string{"test", "test2"},
		},
		ERef: &embed{
			Value:   "test",
			IntVal:  1,
			ListVal: []string{"test", "test2"},
		},
		Related: firstRel{
			ID: "2",
			SecondRel: secondRel{
				ID: "3",
			},
		},
		private: "private",
	}

	b.Run("encoding/json", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			enc, err := json.Marshal(input)
			if err != nil {
				b.Fatal(err)
			}
			if err := json.Unmarshal(enc, &base{}); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("jsonapi", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			enc, err := MarshalOne(input)
			if err != nil {
				b.Fatal(err)
			}
			if err := Unmarshal(enc, &base{}); err != nil {
				b.Fatal(err)
			}
		}
	})

}

func BenchmarkPatches(b *testing.B) {

	patches := []PatchOp{
		{Op: "replace", Path: "/exportedField", Value: "1"},
		{Op: "replace", Path: "/embedded/value", Value: "2"},
		{Op: "replace", Path: "/embedded/intVal", Value: 2},
		{Op: "add", Path: "/embedded/listVal", Value: "3"},
		{Op: "replace", Path: "/related", Value: "id1"},
		{Op: "add", Path: "/rList", Value: "id2"},
	}

	encoded, err := json.Marshal(patches)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("encoding/json", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var out []PatchOp
			if err := json.Unmarshal(encoded, &out); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("jsonapi", func(b *testing.B) {
		t := reflect.TypeOf(new(base))
		for i := 0; i < b.N; i++ {
			_, err := UnmarshalPatches(encoded, t)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
