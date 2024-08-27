package main

import (
	"Loan_Tracker/Delivery/controller"
	"Loan_Tracker/Delivery/router"
	repository "Loan_Tracker/Repository"
	Usecases "Loan_Tracker/Usecase"
	"Loan_Tracker/infrastructure"
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Setup MongoDB connection
	mongoURI := os.Getenv("MONGO_URL")
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	// Get database and collections
	database := client.Database("Loan_Tracker")
	userCollection := database.Collection("User")
	tokenCollection := database.Collection("Token")
	loanCollection := database.Collection("Loan")
	logCollection := database.Collection("Log")

	// Setup repositories
	userRepository := repository.NewUserRepository(userCollection, tokenCollection)
	loanRepository := repository.NewLoanRepository(loanCollection) // New loan repository
	logRepository := repository.NewLogRepository(logCollection)
	// Setup services
	emailService := infrastructure.NewEmailService()

	// Setup use cases
	userUsecase := Usecases.NewUserUsecase(userRepository, logRepository, emailService)
	loanUsecase := Usecases.NewLoanUsecase(loanRepository, logRepository) // New loan use case
	logUsecase := Usecases.NewLogUsecase(logRepository)

	// Setup controllers
	userController := controller.NewUserController(userUsecase)
	loanController := controller.NewLoanController(loanUsecase) // New loan controller
	logController := controller.NewLogController(logUsecase)

	// Setup router
	router := router.SetupRouter(userController, loanController, logController, tokenCollection)

	// Start the server
	log.Fatal(router.Run(":8080"))
}
