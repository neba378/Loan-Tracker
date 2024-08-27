package repository

import (
	"Loan_Tracker/Domain"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LoanRepository interface {
	Save(loan *Domain.Loan) error
	FindByID(id primitive.ObjectID) (Domain.Loan, error)
	GetAllLoans(status string, order string) ([]Domain.Loan, error)
	UpdateStatus(status *Domain.LoanStatus) error
	Delete(id primitive.ObjectID) error
}

type loanRepository struct {
	collection *mongo.Collection
}

func NewLoanRepository(collection *mongo.Collection) LoanRepository {
	return &loanRepository{
		collection: collection,
	}
}

func (r *loanRepository) Save(loan *Domain.Loan) error {
	_, err := r.collection.InsertOne(context.Background(), loan)
	if err != nil {
		return fmt.Errorf("failed to save loan: %v", err)
	}
	return nil
}

func (r *loanRepository) FindByID(id primitive.ObjectID) (Domain.Loan, error) {
	var loan Domain.Loan
	filter := bson.M{"id": id}
	err := r.collection.FindOne(context.Background(), filter).Decode(&loan)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Domain.Loan{}, fmt.Errorf("loan not found: %v", err)
		}
		return Domain.Loan{}, fmt.Errorf("failed to find loan: %v", err)
	}
	return loan, nil
}

func (r *loanRepository) GetAllLoans(status string, order string) ([]Domain.Loan, error) {
	var filter bson.M
	if status != "" {
		filter = bson.M{"status": status}
	}

	// Determine sort order based on status and user input
	sortOrder := 1 // Default ascending order
	if status == "approved" || status == "rejected" {
		sortOrder = -1 // Descending order for reviewed statuses
	} else if order == "desc" {
		sortOrder = -1 // User-specified descending order
	}

	// Set sorting options
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: sortOrder}})

	// Execute the query
	cursor, err := r.collection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get loans: %v", err)
	}
	defer cursor.Close(context.Background())

	// Parse results
	var loans []Domain.Loan
	if err = cursor.All(context.Background(), &loans); err != nil {
		return nil, fmt.Errorf("failed to parse loans: %v", err)
	}
	return loans, nil
}

func (r *loanRepository) UpdateStatus(status *Domain.LoanStatus) error {
	filter := bson.M{"id": status.LoanID}
	update := bson.M{"$set": bson.M{"status": status.Status, "updated_at": status.ChangedAt}}
	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update loan status: %v", err)
	}
	return nil
}

func (r *loanRepository) Delete(id primitive.ObjectID) error {
	filter := bson.M{"id": id}
	_, err := r.collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return fmt.Errorf("failed to delete loan: %v", err)
	}
	return nil
}
