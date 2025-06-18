package egob

import (
	"bytes"
	"encoding/gob"

	"github.com/abdheshnayak/inkube/pkg/fn"
)

func Marshal(obj any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(obj); err != nil {
		return nil, fn.NewE(err)
	}
	return buf.Bytes(), nil
}

func Unmarshal(data []byte, obj any) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(obj); err != nil {
		return fn.NewE(err)
	}
	return nil
}
