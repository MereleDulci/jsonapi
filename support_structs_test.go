package jsonapi

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
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

type InvalidSerializable string

const InvalidSerializableValue = "value"

func (i InvalidSerializable) MarshalText() ([]byte, error) {
	return []byte(i), nil
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
