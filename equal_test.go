package assertjson_test

import (
	"strconv"
	"testing"

	"github.com/bool64/shared"
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

		assert.Equal(t, `	Error Trace:	equal.go:77
	            				equal.go:52
	            				equal_test.go:57
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
	t.Helper()
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

func TestComparer_Equal_vars(t *testing.T) {
	v := &shared.Vars{}
	v.Set("$varB", []int{1, 2, 3})
	v.Set("$varC", "abc")

	// Properties "b" and "c" are checked against values defined in vars.
	// Properties "a" and "d" are not checked, but their values are assigned to missing vars.
	exp := []byte(`{"a": "$varA", "b": "$varB", "c": "$varC", "d": "$varD"}`)
	act := []byte(`{"a": 1.23, "b": [1, 2, 3], "c": "abc", "d": 4}`)

	c := assertjson.Comparer{Vars: v}

	c.Equal(t, exp, act)

	val, found := v.Get("$varA")
	assert.True(t, found)
	assert.Equal(t, 1.23, val)

	val, found = v.Get("$varD")
	assert.True(t, found)
	assert.Equal(t, int64(4), val)

	c.Equal(t, exp, act)

	// Change act to have difference with exp.
	act = []byte(`{"a": 1.23, "b": [1, 2, 4], "c": "abc", "d": 4}`)
	err := c.FailNotEqual(exp, act)
	assert.EqualError(t, err, `not equal:
 {
   "a": 1.23,
   "b": [
     1,
     2,
-    3
+    4
   ],
   "c": "abc",
   "d": 4
 }
`)
}

func TestComparer_Equal_vars_scalar(t *testing.T) {
	v := &shared.Vars{}
	c := assertjson.Comparer{Vars: v}

	assert.NoError(t, c.FailNotEqual([]byte(`["$varA"]`), []byte("[123]")))

	a, found := v.Get("$varA")

	assert.True(t, found)
	assert.Equal(t, int64(123), a)

	assert.NoError(t, c.FailNotEqual([]byte(`"$varB"`), []byte(`[123]`)))
	assert.EqualError(t, c.FailNotEqual([]byte(`"$varB"`), []byte(`[124]`)),
		"not equal:\n [\n-  123\n+  124\n ]\n")

	assert.NoError(t, c.FailNotEqual([]byte(`"$varB"`), []byte(`[123]`)))

	b, found := v.Get("$varB")

	assert.True(t, found)
	assert.Equal(t, []interface{}{123.0}, b)
}
