// Package assertjson implements JSON equality assertion for tests.
package assertjson

import (
	"strings"

	"github.com/bool64/shared"
	"github.com/stretchr/testify/assert"
	d2 "github.com/swaggest/assertjson/diff"
)

// Comparer compares JSON documents.
type Comparer struct {
	// IgnoreDiff is a value in expected document to ignore difference with actual document.
	IgnoreDiff string

	// Vars keeps state of found variables.
	Vars *shared.Vars

	// FormatterConfig controls diff formatter configuration.
	FormatterConfig d2.AsciiFormatterConfig

	// KeepFullDiff shows full diff in error message.
	KeepFullDiff bool

	// FullDiffMaxLines is a maximum number of lines to show without reductions, default 50.
	// Ignored if KeepFullDiff is true.
	FullDiffMaxLines int

	// DiffSurroundingLines is a number of lines to add before and after diff line, default 5.
	// Ignored if KeepFullDiff is true.
	DiffSurroundingLines int
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

// FailNotEqual returns error if JSON payloads are different, nil otherwise.
func FailNotEqual(expected, actual []byte) error {
	return defaultComparer.FailNotEqual(expected, actual)
}

// FailNotEqualMarshal returns error if expected JSON payload is not equal to marshaled actual value.
func FailNotEqualMarshal(expected []byte, actualValue interface{}) error {
	return defaultComparer.FailNotEqualMarshal(expected, actualValue)
}

// FailNotEqualMarshal returns error if expected JSON payload is not equal to marshaled actual value.
func (c Comparer) FailNotEqualMarshal(expected []byte, actualValue interface{}) error {
	actual, err := MarshalIndentCompact(actualValue, "", "  ", 80)
	if err != nil {
		return err
	}

	return c.FailNotEqual(expected, actual)
}

// FailNotEqual returns error if JSON payloads are different, nil otherwise.
func (c Comparer) FailNotEqual(expected, actual []byte) error {
	return c.fail(expected, actual, false)
}

// EqMarshal marshals actual value and compares two JSON documents ignoring string values "<ignore-diff>".
func EqMarshal(t TestingT, expected string, actualValue interface{}, msgAndArgs ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	return defaultComparer.EqualMarshal(t, []byte(expected), actualValue, msgAndArgs...)
}
