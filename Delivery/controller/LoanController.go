package controller

import (
	"Loan_Tracker/Domain"
	Usecases "Loan_Tracker/Usecase"
	"Loan_Tracker/infrastructure"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdateStatus struct {
	Status string `json:"status" bson:"status"` // "pending", "approved", "rejected"
}

type LoanController struct {
	LoanUsecase Usecases.LoanUsecase
}

// NewLoanController creates a new instance of LoanController
func NewLoanController(loanUsecase Usecases.LoanUsecase) *LoanController {
	return &LoanController{
		LoanUsecase: loanUsecase,
	}
}

type LoanInput struct {
	Amount  float64 `json:"amount" bson:"amount"`
	Term    int     `json:"term" bson:"term"` // In months
	Purpose string  `json:"purpose" bson:"purpose"`
}

// CreateLoan handles loan creation
func (lc *LoanController) CreateLoan(c *gin.Context) {
	var inp LoanInput
	if err := c.ShouldBindJSON(&inp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Extract user ID from the JWT token in the Authorization header
	authHeader := c.GetHeader("Authorization")
	tokenStr := authHeader[len("Bearer "):]

	claims, err := infrastructure.ParseToken(tokenStr, jwtKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userIDStr := claims.ID
	fmt.Println(claims)

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Set the UserID in the input
	input := Domain.LoanInput{
		UserID:  userID,
		Amount:  inp.Amount,
		Term:    inp.Term,
		Purpose: inp.Purpose,
	}

	loan, err := lc.LoanUsecase.ApplyForLoan(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"loan": loan})
}

// ViewLoanStatus handles retrieving loan status
func (lc *LoanController) ViewLoanStatus(c *gin.Context) {
	id := c.Param("id")

	loan, err := lc.LoanUsecase.ViewLoanStatus(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"loan": loan})
}

// ViewAllLoans handles retrieving all loans
func (lc *LoanController) ViewAllLoans(c *gin.Context) {
	status := c.Query("status")
	order := c.Query("order")

	loans, err := lc.LoanUsecase.ViewAllLoans(status, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"loans": loans})
}

// ApproveRejectLoan handles loan approval or rejection
func (lc *LoanController) ApproveRejectLoan(c *gin.Context) {
	id := c.Param("id")

	var input Domain.LoanStatusUpdateInput
	var inp UpdateStatus
	if err := c.ShouldBindJSON(&inp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Extract admin user ID from the context or JWT claims if necessary
	changedByStr := c.GetString("userID")
	changedBy, err := primitive.ObjectIDFromHex(changedByStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	input.Status = inp.Status
	input.ChangedBy = changedBy

	err = lc.LoanUsecase.ApproveRejectLoan(id, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Loan status updated successfully"})
}

// DeleteLoan handles loan deletion
func (lc *LoanController) DeleteLoan(c *gin.Context) {
	id := c.Param("id")

	err := lc.LoanUsecase.DeleteLoan(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Loan deleted successfully"})
}

func (lc *LoanController) GetLogs(c *gin.Context) {
	id := c.Param("id")

	err := lc.LoanUsecase.DeleteLoan(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Loan deleted successfully"})
}
