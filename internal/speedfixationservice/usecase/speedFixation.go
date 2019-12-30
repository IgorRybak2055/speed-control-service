// Package usecase provides business logic methods
package usecase

import (
	"time"

	"github.com/IgorRybak2055/speed-control-service/internal/speedfixationservice/repo"
)

type speedFixationUsecase struct {
	contactRepo repo.SpeedControlRepo
}

// NewSpeedFixationUsecase will create new an SpeedControl object representation of SpeedControlRepo interface
func NewSpeedFixationUsecase(cr repo.SpeedControlRepo) SpeedControl {
	return &speedFixationUsecase{
		contactRepo: cr,
	}
}

// CreateRecord receives information from the camera and calls the save method
func (sf speedFixationUsecase) CreateRecord(fixation repo.SpeedFixation) error {
	return sf.contactRepo.CreateRecord(fixation)
}

// LookUpOverSpeedByDate receivers the search criteria and calls the violators search function
func (sf speedFixationUsecase) LookUpOverSpeedByDate(fixation repo.SpeedFixation) ([]repo.SpeedFixation, error) {
	return sf.contactRepo.LookUpOverSpeedByDate(fixation)
}

// LookUpMinMaxSpeedByDate receivers the search criteria and calls the search method which return min & max speeds
func (sf speedFixationUsecase) LookUpMinMaxSpeedByDate(date time.Time) ([]repo.SpeedFixation, error) {
	return sf.contactRepo.LookUpMinMaxSpeedByDate(date)
}
