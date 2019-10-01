// Package assertjson implements JSON equality assertion for tests.
package assertjson

import (
	"encoding/json"
	"fmt"

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

// Equal compares two JSON documents ignoring string values "<ignore-diff>".
func Equal(t TestingT, expected, actual []byte, msgAndArgs ...interface{}) bool {
	return defaultComparer.Equal(t, expected, actual, msgAndArgs...)
}

// Equal compares two JSON payloads.
func (c Comparer) Equal(t TestingT, expected, actual []byte, msgAndArgs ...interface{}) bool {
	var (
		expDecoded, actDecoded interface{}
		diffValue              gojsondiff.Diff
	)
	err := json.Unmarshal(expected, &expDecoded)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("Failed to unmarshal expected:\n%+v", err), msgAndArgs...)
		return false
	}
	err = json.Unmarshal(actual, &actDecoded)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("Failed to unmarshal actual:\n%+v", err), msgAndArgs...)
		return false
	}

	if expArray, ok := expDecoded.([]interface{}); ok {
		if actArray, ok := actDecoded.([]interface{}); ok {
			diffValue = gojsondiff.New().CompareArrays(expArray, actArray)
		} else {
			assert.Fail(t, "Types mismatch, array expected", msgAndArgs...)
			return false
		}
	} else if expObject, ok := expDecoded.(map[string]interface{}); ok {
		if actObject, ok := actDecoded.(map[string]interface{}); ok {
			diffValue = gojsondiff.New().CompareObjects(expObject, actObject)
		} else {
			assert.Fail(t, "Types mismatch, object expected", msgAndArgs...)
			return false
		}
	} else if !assert.Equal(t, expDecoded, actDecoded, msgAndArgs...) { // scalar value comparison
		return false
	}

	if diffValue == nil {
		return true
	}

	if !diffValue.Modified() {
		return true
	}

	diffValue = &diff{deltas: c.filterDeltas(diffValue.Deltas())}
	if !diffValue.Modified() {
		return true
	}

	diffText, err := formatter.NewAsciiFormatter(expDecoded, c.FormatterConfig).Format(diffValue)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("Failed to format diff:\n%+v", err), msgAndArgs...)
		return false
	}

	assert.Fail(t, "Not equal:\n"+diffText, msgAndArgs...)
	return false
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
