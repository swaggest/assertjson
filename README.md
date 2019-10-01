# JSON assertions

[![Build Status](https://travis-ci.org/swaggest/assertjson.svg?branch=master)](https://travis-ci.org/swaggest/assertjson)
[![Coverage Status](https://codecov.io/gh/swaggest/assertjson/branch/master/graph/badge.svg)](https://codecov.io/gh/swaggest/assertjson)
[![GoDoc](https://godoc.org/github.com/swaggest/assertjson?status.svg)](https://godoc.org/github.com/swaggest/assertjson)

This library extends awesome [`github.com/stretchr/testify/assert`](https://godoc.org/github.com/stretchr/testify/assert) 
with nice JSON equality assertions built with [`github.com/yudai/gojsondiff`](https://github.com/yudai/gojsondiff).

## Usage

Default comparer is set up to ignore difference against `"<ignore-diff>"` values. It is accessible with package function `Equal`.

```go
package my_test

import (
	"testing"

	"github.com/swaggest/assertjson"
)

func Test(t *testing.T) {
	assertjson.Equal(t,
		[]byte(`{"a": [1, {"val": "<ignore-diff>"}, 3], "b": 2, "c": 3}`),
		[]byte(`{"a": [1, {"val": 123}, 3], "c": 2, "b": 3}`),
	)

	// Output:
	// Error Trace:	....
	//	Error:      	Not equal:
	//	            	 {
	//	            	   "a": [
	//	            	     1,
	//	            	     {
	//	            	       "val": "<ignore-diff>"
	//	            	     },
	//	            	     3
	//	            	   ],
	//	            	-  "b": 2,
	//	            	+  "b": 3,
	//	            	-  "c": 3
	//	            	+  "c": 2
	//	            	 }
}

```

Custom `Comparer` can be created and used to control ignore behavior and formatter options.
