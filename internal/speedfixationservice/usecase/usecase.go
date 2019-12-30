// Package usecase provides business logic methods
package usecase

import (
	"time"

	"github.com/IgorRybak2055/speed-control-service/internal/speedfixationservice/repo"
)

// SpeedControl represent the services usecases
type SpeedControl interface {
	CreateRecord(repo.SpeedFixation) error
	LookUpOverSpeedByDate(repo.SpeedFixation) ([]repo.SpeedFixation, error)
	LookUpMinMaxSpeedByDate(time.Time) ([]repo.SpeedFixation, error)
}
