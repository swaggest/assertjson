package assertjson_test

import (
	"io/ioutil"
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

		assert.Equal(t, `	Error Trace:	equal.go:82
	            				equal.go:57
	            				equal_test.go:58
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

	tt := testingT(func(format string, args ...interface{}) {})

	for i, tc := range cases {
		tc := tc

		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tc.equals, equal(tt, []byte(tc.expected), []byte(tc.actual)))
		})
	}
}

func TestComparer_Equal_EmptyIgnoreDiff(t *testing.T) {
	t.Parallel()

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
	assertjson.EqMarshal(t, `{"a":123,"b":"abc"}`, v)
}

func TestFailNotEqualMarshal(t *testing.T) {
	v := struct {
		A int    `json:"a"`
		B string `json:"b"`
	}{
		A: 123,
		B: "abc",
	}

	err := assertjson.FailNotEqualMarshal([]byte(`{"a":123,"b":"abc"}`), v)
	assert.NoError(t, err)
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
	assert.Equal(t, []interface{}{int64(123)}, b)

	assert.NoError(t, c.FailNotEqual([]byte(`"$varC"`), []byte(`{"a":17294094973108486143}`)))
	assert.EqualError(t, c.FailNotEqual([]byte(`"$varC"`), []byte(`{"a":17294094973108486144}`)),
		"not equal:\n {\n-  \"a\": 17294094973108486143\n+  \"a\": 17294094973108486144\n }\n")
	assert.NoError(t, c.FailNotEqual([]byte(`"$varC"`), []byte(`{"a":17294094973108486143}`)))

	c1, found := v.Get("$varC")

	assert.True(t, found)
	assert.Equal(t, map[string]interface{}{"a": uint64(17294094973108486143)}, c1)
}

func TestComparer_Equal_vars_uint64(t *testing.T) {
	v := &shared.Vars{}
	c := assertjson.Comparer{Vars: v}

	assert.NoError(t, c.FailNotEqual([]byte(`"$varC"`), []byte(`{"a":17294094973108486143}`)))
	assert.EqualError(t, c.FailNotEqual([]byte(`"$varC"`), []byte(`{"a":17294094973108486144}`)),
		"not equal:\n {\n-  \"a\": 17294094973108486143\n+  \"a\": 17294094973108486144\n }\n")
	assert.NoError(t, c.FailNotEqual([]byte(`"$varC"`), []byte(`{"a":17294094973108486143}`)))

	c1, found := v.Get("$varC")

	assert.True(t, found)
	assert.Equal(t, map[string]interface{}{"a": uint64(17294094973108486143)}, c1)
}

func TestComparer_Equal_long(t *testing.T) {
	long, err := ioutil.ReadFile("_testdata/long-expected.json")
	assert.NoError(t, err)

	longOther, err := ioutil.ReadFile("_testdata/long-actual.json")
	assert.NoError(t, err)

	c := assertjson.Comparer{}

	err = c.FailNotEqual(long, longOther)

	assert.EqualError(t, err, `not equal:
...
     "This file locks the dependencies of your project to a known state",
     "Read more about it at https://getcomposer.org/doc/01-basic-usage.md#installing-dependencies",
     "This file is @generated automatically"
   ],
   "aliases": [
+    "ehm"
   ],
   "content-hash": "f0ff2afe7ca18fda8104ff02b06a8d98",
   "minimum-stability": "stable",
   "packages": [
...
       "notification-url": "https://packagist.org/downloads/",
       "require": {
         "ext-json": "*"
       },
       "require-dev": {
-        "phpunit/phpunit": "^4.8.23"
+        "phpunit/phpunit": "^4.8.24"
       },
       "source": {
         "reference": "d2184358c5ef5ecaa1f6b4c2bce175fac2d25670",
         "type": "git",
         "url": "https://github.com/swaggest/json-diff.git"
       },
       "support": {
         "issues": "https://github.com/swaggest/json-diff/issues",
-        "source": "https://github.com/swaggest/json-diff/tree/v3.8.1"
+        "source": "https://github.com/swaggest/json-diff/tree/v3.8.2"
       },
       "time": "2020-09-25T17:47:07+00:00",
-      "type": "library",
+      "type": "app",
       "version": "v3.8.1"
     }
   ],
   "packages-dev": [
...
         "shasum": "",
         "type": "zip",
         "url": "https://api.github.com/repos/phpDocumentor/ReflectionDocBlock/zipball/bf329f6c1aadea3299f08ee804682b7c45b326a2"
       },
       "license": [
+        "BSD3"
+        "Apache 2.0"
       ],
       "name": "phpdocumentor/reflection-docblock",
       "notification-url": "https://packagist.org/downloads/",
       "require": {
...
         "php": "^5.5 || ^7.0",
         "phpdocumentor/reflection-common": "^1.0"
       },
       "require-dev": {
         "mockery/mockery": "^0.9.4",
-        "phpunit/phpunit": "^5.2||^4.8.24"
+        "phpunit/phpunit": "^5.2&&^4.8.24"
       },
       "source": {
         "reference": "9c977708995954784726e25d0cd1dddf4e65b0f7",
         "type": "git",
...
       "autoload": {
         "classmap": [
           "src/"
         ]
       },
-      "description": "FilterIterator implementation that filters files based on a list of suffixes.",
+      "description": "FilterIterator implementation that filterers files based on a list of suffixes.",
       "dist": {
         "reference": "730b01bc3e867237eaac355e06a36b85dd93a8b4",
         "shasum": "",
         "type": "zip",
...
         "type": "git",
         "url": "https://github.com/webmozarts/assert.git"
       },
       "support": {
         "issues": "https://github.com/webmozarts/assert/issues",
-        "source": "https://github.com/webmozarts/assert/tree/1.9.1"
+        "source": "https://github.com/webmozarts/assert/tree/1.9.2"
       },
       "time": "2020-07-08T17:02:28+00:00",
       "type": "library",
       "version": "1.9.1"
...
     "ext-json": "*",
     "ext-mbstring": "*",
     "php": ">=5.4"
   },
   "platform-dev": [
+    "whoa"
   ],
   "platform-overrides": {
     "php": "5.6.0"
   },
...`)
}
