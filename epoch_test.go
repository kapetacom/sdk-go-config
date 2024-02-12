package config

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type StructWithTime struct {
	Time Epoch `json:"time"`
}

func TestEpochMarshalling(t *testing.T) {
	// Mock environment variables for testing

	original := StructWithTime{
		Time: Epoch(time.UnixMilli(1609459200000)),
	}

	raw, err := json.Marshal(original)
	assert.NoError(t, err)

	assert.Equal(t, `{"time":1609459200000}`, string(raw))

	var decoded StructWithTime
	err = json.Unmarshal(raw, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, original, decoded)

}
