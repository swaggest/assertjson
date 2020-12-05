package assertjson_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swaggest/assertjson"
)

type testingT func(format string, args ...interface{})

func (t testingT) Errorf(format string, args ...interface{}) {
	t(format, args...)
}

func TestEquals_message(t *testing.T) {
	expected := []byte(`{
  "name": "Bob",
  "createdAt": "<ignore-diff>",
  "id": "<ignore-diff>",
  "nested": {
    "val": "<ignore-diff>"
  },
  "items": [
    {
      "val": "<ignore-diff>"
    },
    {
      "val": 123
    },
    {
      "val": "<ignore-diff>"
    }
  ]
}`)
	actual := []byte(`{
  "createdAt": "2018-08-01T00:01:02Z",
  "id": "123",
  "nested": {
    "val": "random"
  },
  "name": "Alice",
  "items": [
    {
      "val": "bar"
    },
    {
      "val": 321
    },
    {
      "val": "foo"
    }
  ]
}`)
	assert.False(t, assertjson.Equal(testingT(func(format string, args ...interface{}) {
		assert.Equal(t, "\n%s", format)
		assert.Len(t, args, 1)

		assert.Equal(t, `	Error Trace:	equal.go:69
	            				equal.go:44
	            				equal_test.go:56
	Error:      	Not equal:
	            	 {
	            	   "createdAt": "<ignore-diff>",
	            	   "id": "<ignore-diff>",
	            	   "items": [
	            	     {
	            	       "val": "<ignore-diff>"
	            	     },
	            	     {
	            	-      "val": 123
	            	+      "val": 321
	            	     },
	            	     {
	            	       "val": "<ignore-diff>"
	            	     }
	            	   ],
	            	-  "name": "Bob",
	            	+  "name": "Alice",
	            	   "nested": {
	            	     "val": "<ignore-diff>"
	            	   }
	            	 }
`, args[0])
	}), expected, actual))
}

type testcase struct {
	expected string
	actual   string
	equals   bool
}

func run(
	t *testing.T,
	cases []testcase,
	equal func(t assertjson.TestingT, expected, actual []byte, msgAndArgs ...interface{}) bool,
) {
	t.Parallel()

	tt := testingT(func(format string, args ...interface{}) {})

	for i, tc := range cases {
		tc := tc

		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tc.equals, equal(tt, []byte(tc.expected), []byte(tc.actual)))
		})
	}
}

func TestComparer_Equal_EmptyIgnoreDiff(t *testing.T) {
	c := assertjson.Comparer{}

	run(t, []testcase{
		{`{"a": [1, {"val": "<ignore-diff>"}, 3]}`, `{"a": [1, {"val": 123}, 3]}`, false},
		{`{"a": [1, {"val": 123}, 3]}`, `{"a": [1, {"val": 123}, 3]}`, true},
	}, c.Equal)
}

func TestComparer_Equal_IgnoreDiff(t *testing.T) {
	c := assertjson.Comparer{
		IgnoreDiff: "Hello, World!",
	}

	run(t, []testcase{
		{`{"a": [1, {"val": "<ignore-diff>"}, 3]}`, `{"a": [1, {"val": 123}, 3]}`, false},
		{`{"a": [1, {"val": "Hello, World!"}, 3]}`, `{"a": [1, {"val": 123}, 3]}`, true},
		{`{"a": [1, {"val": 123}, 3]}`, `{"a": [1, {"val": "Hello, World!"}, 3]}`, false},
		{`{"a": [1, {"val": 123}, 3]}`, `{"a": [1, {"val": 123}, 3]}`, true},
	}, c.Equal)
}

func TestEqual(t *testing.T) {
	run(t, []testcase{
		{`{`, `{}`, false},
		{`{}`, `{`, false},
		{`{}`, `[]`, false},
		{`[]`, `{}`, false},
		{`[]`, `[]`, true},
		{`123`, `321`, false},
		{`123`, `123`, true},
		{`[{}, {"val": "<ignore-diff>"}, {}]`, `[{"val": 123}]`, false},
		{`{"a": [1, {"val": "<ignore-diff>"}, 3]}`, `{"a": [1, {"val": 123}, 3]}`, true},
	}, assertjson.Equal)
}

func TestEqualMarshal(t *testing.T) {
	v := struct {
		A int    `json:"a"`
		B string `json:"b"`
	}{
		A: 123,
		B: "abc",
	}

	assertjson.EqualMarshal(t, []byte(`{"a":123,"b":"abc"}`), v)
}

func TestEqual_order(t *testing.T) {
	exp := []byte(`{
      "context": {
        "errors": {
          "query:box_sku": [
            "missing value"
          ],
          "query:country": [
            "missing value"
          ],
          "query:course_index[]": [
            "missing value"
          ],
          "query:customer_id": [
            "missing value"
          ],
          "query:should_propagate_charge_event": [
            "missing value"
          ],
          "query:subscription_id": [
            "missing value"
          ],
          "query:week": [
            "#: does not match pattern \"^[0-9]{4}-W(0[1-9]|[1-4][0-9]|5[0-3])$\""
          ]
        }
      },
      "error": "validation failed",
      "status": "INVALID_ARGUMENT"
    }`)

	act := []byte(`{
      "context": {
        "errors": {
          "query:box_sku": [
            "missing value"
          ],
          "query:country": [
            "missing value"
          ],
          "query:customer_id": [
            "missing value"
          ],
          "query:should_propagate_charge_event": [
            "missing value"
          ],
          "query:subscription_id": [
            "missing value"
          ],
          "query:week": [
            "#: does not match pattern \"^[0-9]{4}-W(0[1-9]|[1-4][0-9]|5[0-3])$\""
          ],
          "query:course_index[]": [
            "missing value"
          ]
        }
      },
      "error": "validation failed",
      "status": "INVALID_ARGUMENT"
    }`)

	assertjson.Equal(t, exp, act)
}
