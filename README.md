# JSON assertions

[![Build Status](https://github.com/swaggest/assertjson/workflows/test/badge.svg)](https://github.com/swaggest/assertjson/actions?query=branch%3Amaster+workflow%3Atest)
[![Coverage Status](https://codecov.io/gh/swaggest/assertjson/branch/master/graph/badge.svg)](https://codecov.io/gh/swaggest/assertjson)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/swaggest/assertjson)
[![time tracker](https://wakatime.com/badge/github/swaggest/assertjson.svg)](https://wakatime.com/badge/github/swaggest/assertjson)

This library extends awesome [`github.com/stretchr/testify/assert`](https://godoc.org/github.com/stretchr/testify/assert) 
with nice JSON equality assertions built with [`github.com/yudai/gojsondiff`](https://github.com/yudai/gojsondiff).

Also it provides JSON marshaler with [compact indentation](#compact-indentation).

## Usage

Default comparer is set up to ignore difference against `"<ignore-diff>"` values. It is accessible with package functions `Equal` and `EqualMarshal`.

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

### Compact Indentation
Often `json.MarshalIndent` produces result, that is not easy to comprehend due to high count of lines that requires 
massive scrolling effort to read.

This library provides an alternative `assertjson.MarshalIndentCompact` which keeps indentation and line breaks only 
for major part of JSON document, while compacting smaller pieces.

```go
j, err := assertjson.MarshalIndentCompact(v, "", "  ", 100) // 100 is line width limit.
```
 
```json
{
  "openapi":"3.0.2","info":{"title":"","version":""},
  "paths":{
    "/test/{in-path}":{
      "post":{
        "summary":"Title","description":"","operationId":"name",
        "x-some-array":[
          "abc","def",123456,7890123456,[],{"foo":"bar"},
          "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa!"
        ],
        "parameters":[
          {"name":"in_query","in":"query","schema":{"type":"integer"}},
          {"name":"in-path","in":"path","required":true,"schema":{"type":"boolean"}},
          {"name":"in_cookie","in":"cookie","schema":{"type":"number"}},
          {"name":"X-In-Header","in":"header","schema":{"type":"string"}}
        ],
        "requestBody":{
          "content":{
            "application/x-www-form-urlencoded":{"schema":{"$ref":"#/components/schemas/FormDataOpenapiTestInput"}}
          }
        },
        "responses":{"200":{"description":"OK","content":{"application/json":{"schema":{}}}}},
        "deprecated":true
      }
    }
  },
  "components":{
    "schemas":{"FormDataOpenapiTestInput":{"type":"object","properties":{"in_form_data":{"type":"string"}}}}
  }
}
```

Available as `jsoncompact` CLI tool.
```
go get github.com/swaggest/assertjson/cmd/jsoncompact
```