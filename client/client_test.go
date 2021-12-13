package client

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestOnApi(t *testing.T) {
	t.Run("test on Api1", func(t *testing.T) {
		res := DoGet("test")
		m := map[string]string{}
		json.Unmarshal([]byte(res), &m)
		//assert.Equal(t, "", m["test"])

		DoSet("test", "value")

		res = DoGet("test")
		fmt.Println(res)
		json.Unmarshal([]byte(res), &m)
		assert.Equal(t, "value", m["test"])

		DoDelete("test")

		res = DoGet("test")
		json.Unmarshal([]byte(res), &m)
		assert.Equal(t, "", m["test"])
	})

	t.Run("test on Api2", func(t *testing.T) {
		res := DoGet("2")
		m := map[string]string{}
		json.Unmarshal([]byte(res), &m)
		assert.Equal(t, "", m["2"])
	})
}
