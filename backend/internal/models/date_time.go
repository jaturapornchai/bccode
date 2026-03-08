package models

import (
	"strings"
	"time"
)

type Datetime struct {
	time.Time
}

func (t *Datetime) UnmarshalJSON(input []byte) error {
	strInput := strings.Trim(string(input), `"`)
	newTime, err := time.Parse("2006-01-02T15:04:05Z", strInput)
	if err != nil {
		return err
	}

	t.Time = newTime
	return nil
}
