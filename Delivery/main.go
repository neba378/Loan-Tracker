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

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	mongoURI := os.Getenv("MONGO_URL")
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	userDatabase := client.Database("Loan_Tracker")

	userCollection := userDatabase.Collection("User")
	tokenCollection := userDatabase.Collection("Token")
	userRepository := repository.NewUserRepository(userCollection, tokenCollection)
	emailService := infrastructure.NewEmailService()
	userUsecase := Usecases.NewUserUsecase(userRepository, emailService)
	userController := controller.NewUserController(userUsecase)
	router := router.SetupRouter(userController, tokenCollection)
	log.Fatal(router.Run(":8080"))
}
