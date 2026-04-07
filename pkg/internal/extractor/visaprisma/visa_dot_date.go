package visaprisma

import (
	"fmt"
	"time"
)

func VISADotDateToTime(s string) (time.Time, error) {
	asTime, err := time.Parse("02.01.06", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("error converting %s to time: %w", s, err)
	}
	return asTime, err
}
