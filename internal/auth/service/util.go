package service

import "time"

func timeNowAdd(d time.Duration) time.Time {
	return time.Now().Add(d)
}
