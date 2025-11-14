package assertjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/bool64/shared"
	d2 "github.com/swaggest/assertjson/diff"
)

func (c Comparer) varCollected(s string, v interface{}) bool {
	if c.Vars != nil && c.Vars.IsVar(s) {
		if _, found := c.Vars.Get(s); !found {
			if n, ok := v.(json.Number); ok {
				v = shared.DecodeJSONNumber(n)
			} else if f, ok := v.(float64); ok && f == float64(int64(f)) {
				v = int64(f)
			}

			c.Vars.Set(s, v)

			return true
		}
	}

	return false
}

func (c Comparer) filterDeltas(deltas []d2.Delta, ignoreAdded bool) []d2.Delta {
	result := make([]d2.Delta, 0, len(deltas))

	for _, delta := range deltas {
		switch v := delta.(type) {
		case *d2.Modified:
			if c.IgnoreDiff == "" && c.Vars == nil {
				break
			}

			if s, ok := v.OldValue.(string); ok {
				if s == c.IgnoreDiff { // discarding ignored diff
					continue
				}

				if c.varCollected(s, v.NewValue) {
					continue
				}
			}
		case *d2.Object:
			v.Deltas = c.filterDeltas(v.Deltas, ignoreAdded)
			if len(v.Deltas) == 0 {
				continue
			}

			delta = v
		case *d2.Array:
			v.Deltas = c.filterDeltas(v.Deltas, ignoreAdded)
			if len(v.Deltas) == 0 {
				continue
			}

			delta = v

		case *d2.Added:
			if ignoreAdded {
				continue
			}
		}

		result = append(result, delta)
	}

	return result
}

type df struct {
	deltas []d2.Delta
}

func (df *df) Deltas() []d2.Delta {
	return df.deltas
}

func (df *df) Modified() bool {
	return len(df.deltas) > 0
}

func (c Comparer) filterExpected(expected []byte) ([]byte, error) {
	if c.Vars != nil {
		for k, v := range c.Vars.GetAll() {
			j, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal var %s: %w", k, err) // Not wrapping to support go1.12.
			}

			expected = bytes.ReplaceAll(expected, []byte(`"`+k+`"`), j)
		}
	}

	return expected, nil
}

func (c Comparer) compare(expDecoded, actDecoded interface{}) (d2.Diff, error) {
	switch v := expDecoded.(type) {
	case []interface{}:
		if actArray, ok := actDecoded.([]interface{}); ok {
			return d2.New().CompareArrays(v, actArray), nil
		}

		return nil, errors.New("types mismatch, array expected")

	case map[string]interface{}:
		if actObject, ok := actDecoded.(map[string]interface{}); ok {
			return d2.New().CompareObjects(v, actObject), nil
		}

		return nil, errors.New("types mismatch, object expected")

	default:
		if !reflect.DeepEqual(expDecoded, actDecoded) { // scalar value comparison
			return nil, fmt.Errorf("values %v and %v are not equal", expDecoded, actDecoded)
		}
	}

	return nil, nil
}

func unmarshal(data []byte, decoded interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	return dec.Decode(decoded)
}

func (c Comparer) fail(expected, actual []byte, ignoreAdded bool) error {
	var expDecoded, actDecoded interface{}

	expected, err := c.filterExpected(expected)
	if err != nil {
		return err
	}

	err = unmarshal(expected, &expDecoded)
	if err != nil {
		return fmt.Errorf("failed to unmarshal expected:\n%wv", err)
	}

	err = unmarshal(actual, &actDecoded)
	if err != nil {
		return fmt.Errorf("failed to unmarshal actual:\n%wv", err)
	}

	if s, ok := expDecoded.(string); ok && c.Vars != nil && c.Vars.IsVar(s) {
		if c.varCollected(s, actDecoded) {
			return nil
		}

		if v, found := c.Vars.Get(s); found {
			expDecoded = v
		}
	}

	diffValue, err := c.compare(expDecoded, actDecoded)
	if err != nil {
		return err
	}

	if diffValue == nil {
		return nil
	}

	if !diffValue.Modified() {
		return nil
	}

	diffValue = &df{deltas: c.filterDeltas(diffValue.Deltas(), ignoreAdded)}
	if !diffValue.Modified() {
		return nil
	}

	diffText, err := d2.NewAsciiFormatter(expDecoded, c.FormatterConfig).Format(diffValue)
	if err != nil {
		return fmt.Errorf("failed to format diff:\n%wv", err)
	}

	diffText = c.reduceDiff(diffText)

	return errors.New("not equal:\n" + diffText)
}

func (c Comparer) reduceDiff(diffText string) string {
	if c.KeepFullDiff {
		return diffText
	}

	if c.FullDiffMaxLines == 0 {
		c.FullDiffMaxLines = 50
	}

	if c.DiffSurroundingLines == 0 {
		c.DiffSurroundingLines = 5
	}

	diffRows := strings.Split(diffText, "\n")
	if len(diffRows) <= c.FullDiffMaxLines {
		return diffText
	}

	var result []string

	prev := 0

	for i, r := range diffRows {
		if len(r) == 0 {
			continue
		}

		if r[0] == '-' || r[0] == '+' {
			start := i - c.DiffSurroundingLines
			if start < prev {
				start = prev
			} else if start > prev {
				result = append(result, "...")
			}

			end := i + c.DiffSurroundingLines
			if end >= len(diffRows) {
				end = len(diffRows) - 1
			}

			prev = end

			for k := start; k < end; k++ {
				result = append(result, diffRows[k])
			}
		}
	}

	if prev < len(diffRows)-1 {
		result = append(result, "...")
	}

	return strings.Join(result, "\n")
}
