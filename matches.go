package assertjson

import (
	"strings"

	"github.com/stretchr/testify/assert"
)

// Matches compares two JSON payloads.
// It ignores added fields in actual JSON payload.
func (c Comparer) Matches(t TestingT, expected, actual []byte, msgAndArgs ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	err := c.FailMismatch(expected, actual)
	if err == nil {
		return true
	}

	msg := err.Error()
	msg = strings.ToUpper(msg[0:1]) + msg[1:]
	assert.Fail(t, msg, msgAndArgs...)

	return false
}

// MatchesMarshal marshals actual JSON payload and compares it with expected payload.
// It ignores added fields in actual JSON payload.
func (c Comparer) MatchesMarshal(t TestingT, expected []byte, actualValue interface{}, msgAndArgs ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	actual, err := MarshalIndentCompact(actualValue, "", "  ", 80)
	assert.NoError(t, err, "failed to marshal actual value")

	if len(msgAndArgs) == 0 {
		msgAndArgs = append(msgAndArgs, string(actual))
	}

	return c.Matches(t, expected, actual, msgAndArgs...)
}

// FailMismatch returns error if expected JSON payload does not match actual JSON payload, nil otherwise.
// It ignores added fields in actual JSON payload.
func FailMismatch(expected, actual []byte) error {
	return defaultComparer.FailMismatch(expected, actual)
}

// FailMismatchMarshal returns error if expected JSON payload does not match marshaled actual value, nil otherwise.
// It ignores added fields in actual JSON payload.
func FailMismatchMarshal(expected []byte, actualValue interface{}) error {
	return defaultComparer.FailMismatchMarshal(expected, actualValue)
}

// FailMismatchMarshal returns error if expected JSON payload does not match marshaled actual value, nil otherwise.
// It ignores added fields in actual JSON payload.
func (c Comparer) FailMismatchMarshal(expected []byte, actualValue interface{}) error {
	actual, err := MarshalIndentCompact(actualValue, "", "  ", 80)
	if err != nil {
		return err
	}

	return c.FailMismatch(expected, actual)
}

// FailMismatch returns error if expected JSON payload does not match actual JSON payload, nil otherwise.
// It ignores added fields in actual JSON payload.
func (c Comparer) FailMismatch(expected, actual []byte) error {
	return c.fail(expected, actual, true)
}

// Matches compares two JSON documents ignoring string values "<ignore-diff>".
// It ignores added fields in actual JSON payload.
func Matches(t TestingT, expected, actual []byte, msgAndArgs ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	return defaultComparer.Matches(t, expected, actual, msgAndArgs...)
}

// MatchesMarshal marshals actual value and compares two JSON documents ignoring string values "<ignore-diff>".
// It ignores added fields in actual JSON payload.
func MatchesMarshal(t TestingT, expected []byte, actualValue interface{}, msgAndArgs ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	return defaultComparer.MatchesMarshal(t, expected, actualValue, msgAndArgs...)
}
