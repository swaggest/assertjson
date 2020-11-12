// Package assertjson implements JSON equality assertion for tests.
package assertjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

// Comparer compares JSON documents.
type Comparer struct {
	// IgnoreDiff is a value in expected document to ignore difference with actual document.
	IgnoreDiff string

	// FormatterConfig controls diff formatter configuration.
	FormatterConfig formatter.AsciiFormatterConfig
}

var defaultComparer = Comparer{
	IgnoreDiff: "<ignore-diff>",
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

// Equal marshals actual value and compares two JSON documents ignoring string values "<ignore-diff>".
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

func (c Comparer) filterDeltas(deltas []gojsondiff.Delta) []gojsondiff.Delta {
	result := make([]gojsondiff.Delta, 0, len(deltas))

	for _, delta := range deltas {
		switch v := delta.(type) {
		case *gojsondiff.Modified:
			if c.IgnoreDiff == "" {
				break
			}

			if s, ok := v.OldValue.(string); ok && s == c.IgnoreDiff { // discarding ignored diff
				continue
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

// FailNotEqual returns error if JSON payloads are different, nil otherwise.
func (c Comparer) FailNotEqual(expected, actual []byte) error {
	var (
		expDecoded, actDecoded interface{}
		diffValue              gojsondiff.Diff
	)

	err := json.Unmarshal(expected, &expDecoded)
	if err != nil {
		return fmt.Errorf("failed to unmarshal expected:\n%+v", err)
	}

	err = json.Unmarshal(actual, &actDecoded)
	if err != nil {
		return fmt.Errorf("failed to unmarshal actual:\n%+v", err)
	}

	switch v := expDecoded.(type) {
	case []interface{}:
		if actArray, ok := actDecoded.([]interface{}); ok {
			diffValue = gojsondiff.New().CompareArrays(v, actArray)
		} else {
			return errors.New("types mismatch, array expected")
		}

	case map[string]interface{}:
		if actObject, ok := actDecoded.(map[string]interface{}); ok {
			diffValue = gojsondiff.New().CompareObjects(v, actObject)
		} else {
			return errors.New("types mismatch, object expected")
		}

	default:
		if !reflect.DeepEqual(expDecoded, actDecoded) { // scalar value comparison
			return fmt.Errorf("values %v and %v are not equal", expDecoded, actDecoded)
		}
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
