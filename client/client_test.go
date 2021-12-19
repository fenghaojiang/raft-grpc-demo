package client

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestOnApi(t *testing.T) {
	t.Run("test on Api1", func(t *testing.T) {
		res := DoGet("test1")
		m := map[string]string{}
		json.Unmarshal([]byte(res), &m)
		assert.Equal(t, "", m["test1"])

		DoSet("test1", "value")

		res = DoGet("test1")
		fmt.Println(res)
		json.Unmarshal([]byte(res), &m)
		assert.Equal(t, "value", m["test1"])

		DoDelete("test1")

		res = DoGet("test1")
		json.Unmarshal([]byte(res), &m)
		assert.Equal(t, "", m["test1"])
	})

	t.Run("test on Api2", func(t *testing.T) {
		res := DoGet("test2")
		m := map[string]string{}
		json.Unmarshal([]byte(res), &m)
		assert.Equal(t, "", m["test2"])

		DoSet("test2", "testvalue")
		res = DoGet("test2")
		json.Unmarshal([]byte(res), &m)
		assert.Equal(t, "testvalue", m["test2"])

	})
}
