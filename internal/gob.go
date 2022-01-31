package internal

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func Marshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(v); err != nil {
		return nil, fmt.Errorf("failed to gob encode: %w", err)
	}
	return buf.Bytes(), nil
}

func Unmarshal(b []byte, v interface{}) error {
	decoder := gob.NewDecoder(bytes.NewReader(b))
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("failed to gob decode: %w", err)
	}
	return nil
}
