package model

import "time"

// AvailabilityInfo represents the structure of a single day's availability from the API
type AvailabilityInfo struct {
	FreeBeds int       `json:"freeBeds"`
	Date     time.Time `json:"date"`
}
