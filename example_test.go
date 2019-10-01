package assertjson_test

import (
	"fmt"

	"github.com/swaggest/assertjson"
)

var t = testingT(func(format string, args ...interface{}) {
	fmt.Printf(format, args...)
})

func Example() {
	assertjson.Equal(t,
		[]byte(`{"a": [1, {"val": "<ignore-diff>"}, 3], "b": 2, "c": 3}`),
		[]byte(`{"a": [1, {"val": 123}, 3], "c": 2, "b": 3}`),
	)

	// Output:
	// Error Trace:	equal.go:90
	//	            				equal.go:33
	//	            				example_test.go:14
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
