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

		assert.Equal(t, `	Error Trace:	equal.go:48
	            				equal.go:36
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
