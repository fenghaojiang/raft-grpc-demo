package client

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestOnApi(t *testing.T) {
	t.Run("test on Api1", func(t *testing.T) {
		res := doGet("test1")
		m := map[string]string{}
		json.Unmarshal([]byte(res), &m)
		assert.Equal(t, "", m["test1"])

		doSet("test1", "value")

		res = doGet("test1")
		fmt.Println(res)
		json.Unmarshal([]byte(res), &m)
		assert.Equal(t, "value", m["test1"])

		doDelete("test1")

		res = doGet("test1")
		json.Unmarshal([]byte(res), &m)
		assert.Equal(t, "", m["test1"])
	})

	t.Run("test on Api2", func(t *testing.T) {
		res := doGet("test2")
		m := map[string]string{}
		json.Unmarshal([]byte(res), &m)
		assert.Equal(t, "", m["test2"])

		doSet("test2", "testvalue")
		res = doGet("test2")
		json.Unmarshal([]byte(res), &m)
		assert.Equal(t, "testvalue", m["test2"])

	})
}
