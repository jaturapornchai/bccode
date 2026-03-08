package models

import (
	"encoding/json"
	"time"
)

type ISODate struct {
	Format string
	time.Time
}

func (Date *ISODate) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	Date.Format = "2006-01-02"
	t, err := time.Parse(Date.Format, s)
	if err != nil {
		return err
	}

	Date.Time = t
	return nil
}

func (Date ISODate) MarshalJSON() ([]byte, error) {
	return json.Marshal(Date.Time.Format(Date.Format))
}

// func (Date *ISODate) UnmarshalBSONValue(b []byte) error {
// 	var s string
// 	if err := bson.Unmarshal(b, &s); err != nil {
// 		return err
// 	}
// 	Date.Format = "2006-01-02"
// 	t, _ := time.Parse(Date.Format, s)
// 	Date.Time = t
// 	return nil
// }

// func (Date ISODate) MarshalBSONValue() ([]byte, error) {
// 	return bson.Marshal(Date.Time.Format(Date.Format))
// }
