package model

import (
	"fmt"
	"time"
)

type BagirataTime struct {
	time.Time
}

const (
	layout24Hour = time.RFC3339
	layout12Hour = "2006-01-02T3:04:05PMZ"
)

func (ct *BagirataTime) UnmarshalJSON(b []byte) error {
	dateString := string(b)
	dateString = dateString[1 : len(dateString)-1]

	parsedTime, err := time.Parse(layout24Hour, dateString)
	if err == nil {
		ct.Time = parsedTime
		return nil
	}

	parsedTime, err = time.Parse(layout12Hour, dateString)
	if err == nil {
		ct.Time = parsedTime
		return nil
	}

	return fmt.Errorf("could not parse date: %s, got error: %v", dateString, err)
}
