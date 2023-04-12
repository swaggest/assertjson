package assertjson_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swaggest/assertjson"
)

func TestFailMismatch(t *testing.T) {
	act := json.RawMessage(`{"a": 1, "b": 2, "c": {"d": 1, "e": 2}}`)
	exp := []byte(`{"a": 1, "c": {"d": 1}}`)
	expFail := []byte(`{"a": 1, "c": {"d": 2}}`)

	assert.EqualError(t, assertjson.FailNotEqualMarshal(exp, act), `not equal:
 {
   "a": 1,
   "c": {
     "d": 1
+    "e": 2
   }
+  "b": 2
 }
`)

	assert.NoError(t, assertjson.FailMismatchMarshal(exp, act))
	assert.NoError(t, assertjson.FailMismatch(exp, act))
	assert.EqualError(t, assertjson.FailMismatchMarshal(expFail, act), `not equal:
 {
   "a": 1,
   "c": {
-    "d": 2
+    "d": 1
   }
 }
`)

	assert.False(t, assertjson.MatchesMarshal(testingT(func(format string, args ...interface{}) {}), expFail, act))
	assertjson.MatchesMarshal(t, exp, act)
	assertjson.Matches(t, exp, act)
}
