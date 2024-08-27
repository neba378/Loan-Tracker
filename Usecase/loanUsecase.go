package Usecases

import (
	"Loan_Tracker/Domain"
	repository "Loan_Tracker/Repository"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LoanUsecase interface {
	ApplyForLoan(input Domain.LoanInput) (*Domain.Loan, error)
	ViewLoanStatus(id string) (Domain.Loan, error)
	ViewAllLoans(status string, order string) ([]Domain.Loan, error)
	ApproveRejectLoan(id string, input Domain.LoanStatusUpdateInput) error
	DeleteLoan(id string) error
}

type loanUsecase struct {
	loanRepo repository.LoanRepository
	logRepo  repository.LogRepository
}

func NewLoanUsecase(loanRepo repository.LoanRepository, logrepo repository.LogRepository) LoanUsecase {
	return &loanUsecase{
		loanRepo: loanRepo,
		logRepo:  logrepo,
	}
}

func (l *loanUsecase) ApplyForLoan(input Domain.LoanInput) (*Domain.Loan, error) {
	loan := &Domain.Loan{
		ID:        primitive.NewObjectID(),
		UserID:    input.UserID,
		Amount:    input.Amount,
		Term:      input.Term,
		Purpose:   input.Purpose,
		Status:    "pending", // Initial status is pending
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := l.loanRepo.Save(loan)
	if err != nil {
		return nil, err
	}

	// Log Loan Application Submission
	log := &Domain.LogEntry{
		ID:        primitive.NewObjectID(),
		LogType:   "Loan Application Submission",
		Timestamp: time.Now(),
		UserID:    input.UserID.Hex(),
		Message:   "Loan Application Submitted",
	}
	err = l.logRepo.Save(log)
	if err != nil {
		return nil, fmt.Errorf("failed to log Loan Application Submission: %v", err)
	}

	return loan, nil
}

func (l *loanUsecase) ViewLoanStatus(id string) (Domain.Loan, error) {
	loanID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return Domain.Loan{}, err
	}

	loan, err := l.loanRepo.FindByID(loanID)
	if err != nil {
		return Domain.Loan{}, err
	}

	return loan, nil
}

func (l *loanUsecase) ViewAllLoans(status string, order string) ([]Domain.Loan, error) {
	if status != "" && status != "pending" && status != "approved" && status != "rejected" {
		return nil, errors.New("invalid status")
	}

	if order != "" && order != "asc" && order != "desc" {
		return nil, errors.New("invalid order")
	}

	loans, err := l.loanRepo.GetAllLoans(status, order)
	if err != nil {
		return nil, err
	}

	return loans, nil
}

func (l *loanUsecase) ApproveRejectLoan(id string, input Domain.LoanStatusUpdateInput) error {
	loanID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	loan, err := l.loanRepo.FindByID(loanID)
	if err != nil {
		return err
	}

	if loan.Status == "approved" || loan.Status == "rejected" {
		return errors.New("loan has already been processed")
	}

	statusUpdate := &Domain.LoanStatus{
		ID:        primitive.NewObjectID(),
		LoanID:    loanID,
		Status:    input.Status,
		ChangedAt: time.Now(),
		ChangedBy: input.ChangedBy,
	}

	err = l.loanRepo.UpdateStatus(statusUpdate)
	if err != nil {
		return err
	}

	log := &Domain.LogEntry{
		ID:        primitive.NewObjectID(),
		LogType:   "Loan Status update",
		Timestamp: time.Now(),
		UserID:    input.ChangedBy.Hex(),
		Message:   "loan status updated",
	}
	err = l.logRepo.Save(log)
	if err != nil {
		return fmt.Errorf("failed to log Loan status update: %v", err)
	}

	return nil
}

func (l *loanUsecase) DeleteLoan(id string) error {
	loanID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	err = l.loanRepo.Delete(loanID)
	if err != nil {
		return err
	}

	return nil
}
