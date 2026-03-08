package calendar

import "time"

type ICurrentTime interface {
	Now() time.Time
}

type CurrentTime struct{}

func (CurrentTime) Now() time.Time { return time.Now() }
