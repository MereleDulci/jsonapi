package jsonapi

import (
	"encoding/json"
	"testing"
)

func BenchmarkMarshalVsJson(b *testing.B) {

	type SecondRel struct {
		ID string `jsonapi:"primary,second-related"`
	}

	type FirstRel struct {
		ID        string `jsonapi:"primary,first-related"`
		SecondRel SecondRel
	}

	type Embed struct {
		Value   string `json:"value"`
		IntVal  int64
		Empty   *struct{} `json:"empty,omitempty"`
		ListVal []string
	}

	type Base struct {
		ID            string `jsonapi:"primary,resource-name"`
		ExportedField string
		Embedded      Embed
		ERef          *Embed
		Related       FirstRel
		private       string
	}

	input := Base{
		ID:            "1",
		ExportedField: "test",
		Embedded: Embed{
			Value:   "test",
			IntVal:  1,
			ListVal: []string{"test", "test2"},
		},
		ERef: &Embed{
			Value:   "test",
			IntVal:  1,
			ListVal: []string{"test", "test2"},
		},
		Related: FirstRel{
			ID: "2",
			SecondRel: SecondRel{
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
			if err := json.Unmarshal(enc, &Base{}); err != nil {
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
			if err := Unmarshal(enc, &Base{}); err != nil {
				b.Fatal(err)
			}
		}
	})

}
