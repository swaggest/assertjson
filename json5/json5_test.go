package json5_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swaggest/assertjson/json5"
)

func TestValid(t *testing.T) {
	for _, tc := range []struct {
		data  string
		valid bool
	}{
		{data: `  123   `, valid: true},
		{data: `"abc"`, valid: true},
		{data: `["abc",123]`, valid: true},
		{data: `{
					// ABC.
					"abc":123
				}`, valid: true},
		{data: `{
			// XYZ.
			"xyz": 123,
			// ABC.
			"abc": 987,
		}`, valid: true},
		{data: `{
					# ABC.
					"abc":123
				}`, valid: false},
		{data: `["abc",123`, valid: false},
		{data: `"abc",123`, valid: false},
		{data: `"abc`, valid: false},
	} {
		tc := tc
		t.Run(tc.data, func(t *testing.T) {
			assert.Equal(t, tc.valid, json5.Valid([]byte(tc.data)))
		})
	}
}

func TestDowngrade(t *testing.T) {
	j5 := `		{
		// XYZ.
					"xyz": 123,
		// ABC.
		"abc": 987
	}`

	assert.True(t, json5.Valid([]byte(j5)))
	j, err := json5.Downgrade([]byte(j5))
	require.NoError(t, err)

	assert.Equal(t, `{"xyz":123,"abc":987}`, string(j))
}

func TestUnmarshal(t *testing.T) {
	j5 := `		{
		// XYZ.
					"xyz": 123,
		// ABC.
		"abc": 987
	}`

	v := struct {
		Xyz int `json:"xyz"`
		Abc int `json:"abc"`
	}{}

	require.NoError(t, json5.Unmarshal([]byte(j5), &v))
	assert.Equal(t, 123, v.Xyz)
	assert.Equal(t, 987, v.Abc)
}
