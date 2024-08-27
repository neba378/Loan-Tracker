package Usecases

import (
	"Loan_Tracker/Domain"
	repository "Loan_Tracker/Repository"
)

type LogUsecase interface {
	GetLogs(filter Domain.LogFilter) ([]Domain.LogEntry, error)
}

type logUsecase struct {
	logRepo repository.LogRepository
}

// NewLogUsecase creates a new instance of logUsecase
func NewLogUsecase(logRepo repository.LogRepository) LogUsecase {
	return &logUsecase{
		logRepo: logRepo,
	}
}

// GetLogs retrieves logs based on the filter provided
func (u *logUsecase) GetLogs(filter Domain.LogFilter) ([]Domain.LogEntry, error) {
	logs, err := u.logRepo.GetLogs(filter)
	if err != nil {
		return nil, err
	}

	return logs, nil
}
