package controller

import (
	"Loan_Tracker/Domain"
	Usecases "Loan_Tracker/Usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LogController struct {
	LogUsecase Usecases.LogUsecase
}

// NewLogController creates a new instance of LogController
func NewLogController(logUsecase Usecases.LogUsecase) *LogController {
	return &LogController{
		LogUsecase: logUsecase,
	}
}

// GetLogs handles retrieving logs based on the filter provided
func (lc *LogController) GetLogs(c *gin.Context) {
	var filter Domain.LogFilter

	// Bind query parameters to the LogFilter struct
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter parameters"})
		return
	}

	// Retrieve logs using the usecase
	logs, err := lc.LogUsecase.GetLogs(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}
