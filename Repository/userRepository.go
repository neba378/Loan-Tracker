package repository

import (
	"Loan_Tracker/Domain"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	Save(user *Domain.User) error
	FindByID(id string) (Domain.User, error)
	FindByEmail(email string) (Domain.User, error)
	FindByUsername(username string) (Domain.User, error)
	Update(username string, UpdatedUser bson.M) error
	Delete(userID string) error
	IsDbEmpty() (bool, error)
	InsertToken(username string, accessToke string, refreshToken string) error
	ExpireToken(token string) error
	ShowUser(id string) (Domain.User, error)
	GetAllUsers() ([]Domain.User, error)
}

type userRepository struct {
	collection      *mongo.Collection
	tokenCollection *mongo.Collection
}

func NewUserRepository(collection *mongo.Collection, tokenCollection *mongo.Collection) UserRepository {
	return &userRepository{collection: collection, tokenCollection: tokenCollection}
}

func (ur *userRepository) GetAllUsers() ([]Domain.User, error) {
	var users []Domain.User
	filter := bson.M{"role": "user"}
	cursor, err := ur.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var user Domain.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (ur *userRepository) Save(user *Domain.User) error {
	_, err := ur.collection.InsertOne(context.Background(), user)
	return err
}

func (ur *userRepository) FindByID(id string) (Domain.User, error) {
	var user Domain.User
	err := ur.collection.FindOne(context.Background(), bson.M{"id": id}).Decode(&user)
	return user, err
}

func (ur *userRepository) FindByEmail(email string) (Domain.User, error) {
	var user Domain.User
	err := ur.collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	return user, err
}

func (ur *userRepository) FindByUsername(username string) (Domain.User, error) {
	var user Domain.User
	err := ur.collection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	return user, err
}

func (ur *userRepository) Update(username string, updatedUser bson.M) error {
	_, err := ur.collection.UpdateOne(context.Background(), bson.M{"username": username}, bson.M{"$set": updatedUser})
	return err
}

func (ur *userRepository) Delete(userID string) error {
	_, err := ur.collection.DeleteOne(context.Background(), bson.M{"id": userID})
	return err
}

func (ur *userRepository) IsDbEmpty() (bool, error) {
	count, err := ur.collection.CountDocuments(context.Background(), bson.M{})
	return count == 0, err
}

func (ur *userRepository) InsertToken(username string, accessToke string, refreshToken string) error {
	token := &Domain.Token{
		TokenID:      primitive.NewObjectID(),
		Username:     username,
		AccessToken:  accessToke,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(2 * time.Hour),
	}

	_, err := ur.tokenCollection.InsertOne(context.Background(), token)
	return err
}

func (ur *userRepository) ExpireToken(token string) error {
	// Define the filter to find the token
	filter := bson.M{"access_token": token}

	// Define the update to set the ExpiresAt field to the current time
	update := bson.M{
		"$set": bson.M{
			"expires_at": time.Now(), // Assuming ExpiresAt is stored as a Unix timestamp
		},
	}

	// Perform the update operation
	_, err := ur.tokenCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (ur *userRepository) ShowUser(id string) (Domain.User, error) {
	var user Domain.User
	filter := bson.M{"id": id}
	fmt.Println("i was here", id)

	// Use FindOne to get a single user
	err := ur.collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No user found with the given id
			return user, nil
		}
		// Some other error occurred
		return user, err
	}

	return user, nil
}
