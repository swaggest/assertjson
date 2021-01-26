// Package assertjson implements JSON equality assertion for tests.
package assertjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/bool64/shared"
	"github.com/stretchr/testify/assert"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

// Comparer compares JSON documents.
type Comparer struct {
	// IgnoreDiff is a value in expected document to ignore difference with actual document.
	IgnoreDiff string

	// Vars keeps state of found variables.
	Vars *shared.Vars

	// FormatterConfig controls diff formatter configuration.
	FormatterConfig formatter.AsciiFormatterConfig
}

// IgnoreDiff is a marker to ignore difference in JSON.
const IgnoreDiff = "<ignore-diff>"

var defaultComparer = Comparer{
	IgnoreDiff: IgnoreDiff,
}

// TestingT is an interface wrapper around *testing.T.
type TestingT interface {
	Errorf(format string, args ...interface{})
}

type tHelper interface {
	Helper()
}

// Equal compares two JSON documents ignoring string values "<ignore-diff>".
func Equal(t TestingT, expected, actual []byte, msgAndArgs ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	return defaultComparer.Equal(t, expected, actual, msgAndArgs...)
}

// EqualMarshal marshals actual value and compares two JSON documents ignoring string values "<ignore-diff>".
func EqualMarshal(t TestingT, expected []byte, actualValue interface{}, msgAndArgs ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	return defaultComparer.EqualMarshal(t, expected, actualValue, msgAndArgs...)
}

// Equal compares two JSON payloads.
func (c Comparer) Equal(t TestingT, expected, actual []byte, msgAndArgs ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	err := c.FailNotEqual(expected, actual)
	if err == nil {
		return true
	}

	msg := err.Error()
	msg = strings.ToUpper(msg[0:1]) + msg[1:]
	assert.Fail(t, msg, msgAndArgs...)

	return false
}

// EqualMarshal marshals actual JSON payload and compares it with expected payload.
func (c Comparer) EqualMarshal(t TestingT, expected []byte, actualValue interface{}, msgAndArgs ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	actual, err := MarshalIndentCompact(actualValue, "", "  ", 80)
	assert.NoError(t, err, "failed to marshal actual value")

	if len(msgAndArgs) == 0 {
		msgAndArgs = append(msgAndArgs, string(actual))
	}

	return c.Equal(t, expected, actual, msgAndArgs...)
}

func (c Comparer) varCollected(s string, v interface{}) bool {
	if c.Vars != nil && c.Vars.IsVar(s) {
		if _, found := c.Vars.Get(s); !found {
			if f, ok := v.(float64); ok && f == float64(int64(f)) {
				v = int64(f)
			}

			c.Vars.Set(s, v)

			return true
		}
	}

	return false
}

func (c Comparer) filterDeltas(deltas []gojsondiff.Delta) []gojsondiff.Delta {
	result := make([]gojsondiff.Delta, 0, len(deltas))

	for _, delta := range deltas {
		switch v := delta.(type) {
		case *gojsondiff.Modified:
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
		case *gojsondiff.Object:
			v.Deltas = c.filterDeltas(v.Deltas)
			if len(v.Deltas) == 0 {
				continue
			}

			delta = v
		case *gojsondiff.Array:
			v.Deltas = c.filterDeltas(v.Deltas)
			if len(v.Deltas) == 0 {
				continue
			}

			delta = v
		}

		result = append(result, delta)
	}

	return result
}

type diff struct {
	deltas []gojsondiff.Delta
}

func (diff *diff) Deltas() []gojsondiff.Delta {
	return diff.deltas
}

func (diff *diff) Modified() bool {
	return len(diff.deltas) > 0
}

// FailNotEqual returns error if JSON payloads are different, nil otherwise.
func FailNotEqual(expected, actual []byte) error {
	return defaultComparer.FailNotEqual(expected, actual)
}

func (c Comparer) filterExpected(expected []byte) ([]byte, error) {
	if c.Vars != nil {
		for k, v := range c.Vars.GetAll() {
			j, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal var %s: %v", k, err) // Not wrapping to support go1.12.
			}

			expected = bytes.Replace(expected, []byte(`"`+k+`"`), j, -1) // nolint:gocritic // To support go1.11.
		}
	}

	return expected, nil
}

func (c Comparer) compare(expDecoded, actDecoded interface{}) (gojsondiff.Diff, error) {
	switch v := expDecoded.(type) {
	case []interface{}:
		if actArray, ok := actDecoded.([]interface{}); ok {
			return gojsondiff.New().CompareArrays(v, actArray), nil
		}

		return nil, errors.New("types mismatch, array expected")

	case map[string]interface{}:
		if actObject, ok := actDecoded.(map[string]interface{}); ok {
			return gojsondiff.New().CompareObjects(v, actObject), nil
		}

		return nil, errors.New("types mismatch, object expected")

	default:
		if !reflect.DeepEqual(expDecoded, actDecoded) { // scalar value comparison
			return nil, fmt.Errorf("values %v and %v are not equal", expDecoded, actDecoded)
		}
	}

	return nil, nil
}

// FailNotEqual returns error if JSON payloads are different, nil otherwise.
func (c Comparer) FailNotEqual(expected, actual []byte) error {
	var expDecoded, actDecoded interface{}

	expected, err := c.filterExpected(expected)
	if err != nil {
		return err
	}

	err = json.Unmarshal(expected, &expDecoded)
	if err != nil {
		return fmt.Errorf("failed to unmarshal expected:\n%+v", err)
	}

	err = json.Unmarshal(actual, &actDecoded)
	if err != nil {
		return fmt.Errorf("failed to unmarshal actual:\n%+v", err)
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

	diffValue = &diff{deltas: c.filterDeltas(diffValue.Deltas())}
	if !diffValue.Modified() {
		return nil
	}

	diffText, err := formatter.NewAsciiFormatter(expDecoded, c.FormatterConfig).Format(diffValue)
	if err != nil {
		return fmt.Errorf("failed to format diff:\n%+v", err)
	}

	return errors.New("not equal:\n" + diffText)
}
