// Package json5 provides JSON5 decoder.
package json5

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/vearutop/json5/encoding/json5"
)

// Valid checks if bytes are a valid JSON5 payload.
func Valid(data []byte) (isValid bool) {
	defer func() {
		if r := recover(); r != nil {
			isValid = false
		}
	}()

	var v interface{}

	return Unmarshal(data, &v) == nil
}

// Downgrade converts JSON5 to JSON.
func Downgrade(data []byte) ([]byte, error) {
	v := json.RawMessage{}

	err := Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}

	return json.Marshal(v)
}

// Unmarshal parses the JSON5-encoded data and stores the result
// in the value pointed to by v.
func Unmarshal(data []byte, v interface{}) error {
	dec := json5.NewDecoder(bytes.NewReader(data))

	err := dec.Decode(&v)
	if err != nil {
		return err
	}

	var tail interface{}

	// Second decode to make sure there is only one JSON5 value in data and no garbage in tail.
	err = dec.Decode(&tail)

	if err != io.EOF {
		return errors.New("unexpected bytes after JSON5 payload")
	}

	return nil
}
