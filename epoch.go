package config

import (
	"encoding/json"
	"time"
)

// Epoch is a custom type for time.Time that marshals and unmarshals to/from epoch time in milliseconds.
type Epoch time.Time

func (m Epoch) MarshalJSON() ([]byte, error) {
	millis := time.Time(m).UnixMilli()
	return json.Marshal(millis)
}

func (m *Epoch) UnmarshalJSON(data []byte) error {
	var millis int64
	if err := json.Unmarshal(data, &millis); err != nil {
		return err
	}
	*m = Epoch(time.UnixMilli(millis))
	return nil
}
