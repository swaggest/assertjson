package assertjson_test

import (
	"encoding/json"
	"testing"

	"github.com/iancoleman/orderedmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swaggest/assertjson"
)

func TestMarshalIndentCompact(t *testing.T) {
	// nolint:lll // Yeah, this line is loooong, but that's ok.
	j := []byte(`{"openapi":"3.0.2","info":{"title":"","version":""},"paths":{"/test/{in-path}":{"post":{"summary":"Title","description":"","operationId":"name","parameters":[{"name":"in_query","in":"query","schema":{"type":"integer"}},{"name":"in-path","in":"path","required":true,"schema":{"type":"boolean"}},{"name":"in_cookie","in":"cookie","schema":{"type":"number"}},{"name":"X-In-Header","in":"header","schema":{"type":"string"}}],"requestBody":{"content":{"application/x-www-form-urlencoded":{"schema":{"$ref":"#/components/schemas/FormDataOpenapiTestInput"}}}},"responses":{"200":{"description":"OK","content":{"application/json":{"schema":{}}}}},"deprecated":true}}},"components":{"schemas":{"FormDataOpenapiTestInput":{"type":"object","properties":{"in_form_data":{"type":"string"}}}}}}`)
	v := orderedmap.New()
	assert.NoError(t, json.Unmarshal(j, &v))

	jjj, err := assertjson.MarshalIndentCompact(v, "XXX", "YY", 100)
	assert.NoError(t, err)
	assert.Equal(t, `{
XXXYY"openapi":"3.0.2",
XXXYY"info":{"title":"","version":""},
XXXYY"paths":{
XXXYYYY"/test/{in-path}":{
XXXYYYYYY"post":{
XXXYYYYYYYY"summary":"Title",
XXXYYYYYYYY"description":"",
XXXYYYYYYYY"operationId":"name",
XXXYYYYYYYY"parameters":[
XXXYYYYYYYYYY{"name":"in_query","in":"query","schema":{"type":"integer"}},
XXXYYYYYYYYYY{"name":"in-path","in":"path","required":true,"schema":{"type":"boolean"}},
XXXYYYYYYYYYY{"name":"in_cookie","in":"cookie","schema":{"type":"number"}},
XXXYYYYYYYYYY{"name":"X-In-Header","in":"header","schema":{"type":"string"}}
XYYYYYYYYYY],
XXXYYYYYYYY"requestBody":{
XXXYYYYYYYYYY"content":{
XXXYYYYYYYYYYYY"application/x-www-form-urlencoded":{"schema":{"$ref":"#/components/schemas/FormDataOpenapiTestInput"}}
XYYYYYYYYYYYY}
XYYYYYYYYYY},
XXXYYYYYYYY"responses":{"200":{"description":"OK","content":{"application/json":{"schema":{}}}}},
XXXYYYYYYYY"deprecated":true
XYYYYYYYY}
XYYYYYY}
XYYYY},
XXXYY"components":{
XXXYYYY"schemas":{
XXXYYYYYY"FormDataOpenapiTestInput":{"type":"object","properties":{"in_form_data":{"type":"string"}}}
XYYYYYY}
XYYYY}
XYY}`, string(jjj))

	jj, err := assertjson.MarshalIndentCompact(v, "", "  ", 100)
	require.NoError(t, err)

	assertjson.Equal(t, j, jj)
	assert.Equal(t, `{
  "openapi":"3.0.2",
  "info":{"title":"","version":""},
  "paths":{
    "/test/{in-path}":{
      "post":{
        "summary":"Title",
        "description":"",
        "operationId":"name",
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
}`, string(jj))
}
