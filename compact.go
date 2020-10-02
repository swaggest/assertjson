package assertjson

import (
	"encoding/json"

	"github.com/iancoleman/orderedmap"
)

// MarshalIndentCompact applies indentation for large chunks of JSON and uses compact format for smaller ones.
//
// Line length limits indented width of JSON structure, does not apply to long distinct scalars.
// This function is not optimized for performance, so it might be not a good fit for high load scenarios.
func MarshalIndentCompact(v interface{}, prefix, indent string, lineLen int) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	if len(b) <= lineLen {
		return b, nil
	}

	m := orderedmap.New()

	// Creating a temporary JSON object to make sure it can be unmarshaled into a map.
	tmpMap := append([]byte(`{"t":`), b...)
	tmpMap = append(tmpMap, '}')

	err = json.Unmarshal(tmpMap, m)
	if err != nil {
		return nil, err
	}

	i, ok := m.Get("t")
	if !ok {
		return nil, orderedmap.NoValueError
	}

	return marshalIndentCompact(i, indent, append([]byte(prefix), []byte(indent)...), lineLen)
}

func marshalIndentCompact(m interface{}, indent string, pad []byte, lineLen int) ([]byte, error) {
	compact, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	if len(compact)+len(pad) <= lineLen {
		return compact, nil
	}

	switch o := m.(type) {
	case orderedmap.OrderedMap:
		return marshalObject(o, len(compact), indent, pad, lineLen)
	case []interface{}:
		return marshalArray(o, len(compact), indent, pad, lineLen)
	}

	return compact, nil
}

func marshalArray(o []interface{}, compactLen int, indent string, pad []byte, lineLen int) ([]byte, error) {
	res := append(make([]byte, 0, compactLen), '[', '\n')

	for i, val := range o {
		jsonVal, err := marshalIndentCompact(val, indent, append(pad, []byte(indent)...), lineLen)
		if err != nil {
			return nil, err
		}

		res = append(res, pad...)
		res = append(res, jsonVal...)

		if i == len(o)-1 {
			res = append(res, '\n')
			res = append(res, pad[len(indent):]...)
			res = append(res, ']')
		} else {
			res = append(res, ',', '\n')
		}
	}

	return res, nil
}

func marshalObject(o orderedmap.OrderedMap, compactLen int, indent string, pad []byte, lineLen int) ([]byte, error) {
	res := append(make([]byte, 0, compactLen), '{', '\n')

	keys := o.Keys()
	for i, k := range keys {
		val, ok := o.Get(k)
		if !ok {
			return nil, orderedmap.NoValueError
		}

		jsonVal, err := marshalIndentCompact(val, indent, append(pad, []byte(indent)...), lineLen)
		if err != nil {
			return nil, err
		}

		kj, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}

		res = append(res, pad...)
		res = append(res, kj...)
		res = append(res, ':')
		res = append(res, jsonVal...)

		if i == len(keys)-1 {
			res = append(res, '\n')
			res = append(res, pad[len(indent):]...)
			res = append(res, '}')
		} else {
			res = append(res, ',', '\n')
		}
	}

	return res, nil
}
