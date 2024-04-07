package api

import "time"

// Use same time zone as admin api
func getNowTime() time.Time {
	return time.Now().UTC()
}
