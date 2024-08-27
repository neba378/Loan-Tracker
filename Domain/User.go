package Domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID             primitive.ObjectID `json:"id" bson:"id"`
	Name           string             `json:"name" bson:"name"`
	Username       string             `json:"username" bson:"username"`
	Password       string             `json:"password" bson:"password"`
	Email          string             `json:"email" bson:"email"`
	ProfilePicture string             `json:"profile_picture" bson:"profile_picture"`
	Role           string             `json:"role" bson:"role"`
	IsActive       bool               `json:"is_active" bson:"is_active"`
}

type RegisterInput struct {
	Name           string `json:"name" bson:"name"`
	Username       string `json:"username" bson:"username"`
	Password       string `json:"password" bson:"password"`
	Email          string `json:"email" bson:"email"`
	ProfilePicture string `json:"profile_picture" bson:"profile_picture"`
	Bio            string `json:"bio" bson:"bio"`
}

type LoginInput struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}

type ChangePasswordInput struct {
	NewPassword string `json:"password" bson:"password"`
}

type ForgetPasswordInput struct {
	Email string `json:"email" bson:"email"`
}
