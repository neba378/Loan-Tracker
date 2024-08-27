package router

import (
	controller "Loan_Tracker/Delivery/controller"
	"Loan_Tracker/infrastructure"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRouter(userController *controller.UserController, loanController *controller.LoanController, logController *controller.LogController, tokenCollection *mongo.Collection) *gin.Engine {
	router := gin.Default()

	// Public routes (no authentication required)
	router.POST("/users/register", userController.Register)
	router.POST("/users/login", userController.Login)
	router.POST("/users/token/refresh", userController.RefreshToken)
	router.POST("/users/password-reset", userController.ForgotPassword)
	router.GET("/users/password-reset/:token", userController.ResetPassword)
	router.GET("/users/verify-email/:token", userController.Verify)

	usersRoute := router.Group("/")
	usersRoute.Use(infrastructure.AuthMiddleware(tokenCollection))
	usersRoute.GET("/users/profile/:id", userController.FindUser)
	usersRoute.PUT("/users/password-reset", userController.ChangePassword)

	// Loan routes (authentication required)
	usersRoute.POST("/loans", loanController.CreateLoan)
	usersRoute.GET("/loans/:id", loanController.ViewLoanStatus)

	adminRoute := usersRoute.Group("/")
	adminRoute.Use(infrastructure.AdminMiddleware()) // Apply admin role middleware

	adminRoute.GET("/admin/loans", loanController.ViewAllLoans)
	adminRoute.PATCH("/admin/loans/:id/status", loanController.ApproveRejectLoan)
	adminRoute.DELETE("/admin/loans/:id", loanController.DeleteLoan)

	adminRoute.GET("/admin/users", userController.GetAllUsers)
	adminRoute.DELETE("/admin/users/:id", userController.DeleteUser)
	adminRoute.GET("/admin/logs", logController.GetLogs)
	return router
}
