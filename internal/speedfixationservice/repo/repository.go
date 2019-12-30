// Package repo provides all needs methods to work with data storage
package repo

import (
	"time"
)

// SpeedControlRepo represent the speedFixationRepo repository contract
type SpeedControlRepo interface {
	CreateRecord(SpeedFixation) error
	LookUpOverSpeedByDate(SpeedFixation) ([]SpeedFixation, error)
	LookUpMinMaxSpeedByDate(time.Time) ([]SpeedFixation, error)
}
