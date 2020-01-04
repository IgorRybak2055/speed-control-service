// Package repo provides all needs methods to work with data storage
package repo

import (
	"time"
)

type dataSlice []SpeedFixation

// SpeedFixation for working with data and storing it
type SpeedFixation struct {
	Date          time.Time `json:"date,omitempty"`
	VehicleNumber string    `json:"vehicle_number,omitempty"`
	Speed         float64   `json:"speed,omitempty"`
}