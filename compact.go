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

	// Return early if document is small enough.
	if len(b) <= lineLen {
		return b, nil
	}

	m := orderedmap.New()

	// Create a temporary JSON object to make sure it can be unmarshaled into a map.
	tmpMap := append([]byte(`{"t":`), b...)
	tmpMap = append(tmpMap, '}')

	// Unmarshal JSON payload into ordered map to recursively walk the document.
	err = json.Unmarshal(tmpMap, m)
	if err != nil {
		return nil, err
	}

	i, ok := m.Get("t")
	if !ok {
		return nil, orderedmap.NoValueError
	}

	// Create first level padding.
	pad := append([]byte(prefix), []byte(indent)...)

	// Call recursive function to walk the document.
	return marshalIndentCompact(i, indent, pad, lineLen)
}

func marshalIndentCompact(doc interface{}, indent string, pad []byte, lineLen int) ([]byte, error) {
	// Build compact JSON for provided sub document.
	compact, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}

	// Return compact if it fits line length limit with current padding.
	if len(compact)+len(pad) <= lineLen {
		return compact, nil
	}

	// Indent arrays and objects that are too big.
	switch o := doc.(type) {
	case orderedmap.OrderedMap:
		return marshalObject(o, len(compact), indent, pad, lineLen)
	case []interface{}:
		return marshalArray(o, len(compact), indent, pad, lineLen)
	}

	// Use compact for scalar values (numbers, strings, booleans, nulls).
	return compact, nil
}

func marshalArray(o []interface{}, compactLen int, indent string, pad []byte, lineLen int) ([]byte, error) {
	// Allocate result with a size of compact form, because it is impossible to make result shorter.
	res := append(make([]byte, 0, compactLen), '[', '\n')

	for i, val := range o {
		// Build item value with an increased padding.
		jsonVal, err := marshalIndentCompact(val, indent, append(pad, []byte(indent)...), lineLen)
		if err != nil {
			return nil, err
		}

		// Add item JSON with current padding.
		res = append(res, pad...)
		res = append(res, jsonVal...)

		if i == len(o)-1 {
			// Close array at last item.
			res = append(res, '\n')
			// Strip one indent from a closing bracket.
			res = append(res, pad[len(indent):]...)
			res = append(res, ']')
		} else {
			// Add colon and new line after an item.
			res = append(res, ',', '\n')
		}
	}

	return res, nil
}

func marshalObject(o orderedmap.OrderedMap, compactLen int, indent string, pad []byte, lineLen int) ([]byte, error) {
	// Allocate result with a size of compact form, because it is impossible to make result shorter.
	res := append(make([]byte, 0, compactLen), '{', '\n')

	// Iterate object using keys slice to preserve properties order.
	keys := o.Keys()
	for i, k := range keys {
		val, ok := o.Get(k)
		if !ok {
			return nil, orderedmap.NoValueError
		}

		// Build item value with an increased padding.
		jsonVal, err := marshalIndentCompact(val, indent, append(pad, []byte(indent)...), lineLen)
		if err != nil {
			return nil, err
		}

		// Marshal key as JSON string.
		kj, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}

		// Add key JSON with current padding.
		res = append(res, pad...)
		res = append(res, kj...)
		res = append(res, ':')
		// Add value JSON to the same line.
		res = append(res, jsonVal...)

		if i == len(keys)-1 {
			// Close object at last property.
			res = append(res, '\n')
			// Strip one indent from a closing bracket.
			res = append(res, pad[len(indent):]...)
			res = append(res, '}')
		} else {
			// Add colon and new line after a property.
			res = append(res, ',', '\n')
		}
	}

	return res, nil
}
