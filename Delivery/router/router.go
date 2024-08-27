package router

import (
	controller "Loan_Tracker/Delivery/controller"
	"Loan_Tracker/infrastructure"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRouter(userController *controller.UserController, tokenCollection *mongo.Collection) *gin.Engine {
	router := gin.Default()

	// Public routes (no authentication required)
	router.POST("/users/register", userController.Register)
	router.POST("/users/login", userController.Login)
	router.POST("/users/token/refresh", userController.RefreshToken)
	router.POST("/users/password-reset", userController.ForgotPassword)
	router.GET("/users/password-reset/:token", userController.ResetPassword)
	router.GET("/users/verify-email/:token", userController.Verify)
	router.GET("/users/profile/:id", userController.FindUser)

	usersRoute := router.Group("/")
	usersRoute.Use(infrastructure.AuthMiddleware(tokenCollection))
	usersRoute.PUT("/users/password-reset", userController.ChangePassword)

	adminRoute := usersRoute.Group("/")
	adminRoute.Use(infrastructure.AdminMiddleware()) // Apply admin role middleware

	adminRoute.GET("/admin/users", userController.GetAllUsers)
	adminRoute.DELETE("/admin/users/:id", userController.DeleteUser)
	return router
}
