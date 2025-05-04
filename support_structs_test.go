package jsonapi

import (
	"encoding/json"
)

type nestedWithMarshalInner struct {
	A int
}

func (n nestedWithMarshalInner) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]int{
		"b": n.A,
	})
}

func (n *nestedWithMarshalInner) UnmarshalJSON(data []byte) error {
	aux := struct {
		B int `json:"b"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	n.A = aux.B
	return nil
}

type CircularA struct {
	ID  string     `jsonapi:"primary,circular-a"`
	Val string     `jsonapi:"attr,val"`
	B   *CircularB `jsonapi:"relation,b"`
}

type CircularB struct {
	ID  string     `jsonapi:"primary,circular-b"`
	Val string     `jsonapi:"attr,val"`
	A   *CircularA `jsonapi:"relation,a"`
}
