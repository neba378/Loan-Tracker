package repository

import (
	"Loan_Tracker/Domain"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LogRepository interface {
	Save(log *Domain.LogEntry) error
	GetLogs(filter Domain.LogFilter) ([]Domain.LogEntry, error)
}

type logRepository struct {
	collection *mongo.Collection
}

func NewLogRepository(collection *mongo.Collection) LogRepository {
	return &logRepository{
		collection: collection,
	}
}

// Save saves a new log entry to the database.
func (r *logRepository) Save(log *Domain.LogEntry) error {
	_, err := r.collection.InsertOne(context.Background(), log)
	if err != nil {
		return fmt.Errorf("failed to save log entry: %v", err)
	}
	return nil
}

// GetLogs retrieves log entries based on filtering criteria.
func (r *logRepository) GetLogs(filter Domain.LogFilter) ([]Domain.LogEntry, error) {
	query := bson.M{}
	if filter.LogType != "" {
		query["log_type"] = filter.LogType
	}
	if !filter.StartDate.IsZero() {
		query["timestamp"] = bson.M{"$gte": filter.StartDate}
	}
	if !filter.EndDate.IsZero() {
		if query["timestamp"] == nil {
			query["timestamp"] = bson.M{}
		}
		query["timestamp"].(bson.M)["$lte"] = filter.EndDate
	}
	if filter.UserID != "" {
		query["user_id"] = filter.UserID
	}

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}) // Sort by timestamp in descending order
	cursor, err := r.collection.Find(context.Background(), query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %v", err)
	}
	defer cursor.Close(context.Background())

	var logs []Domain.LogEntry
	if err = cursor.All(context.Background(), &logs); err != nil {
		return nil, fmt.Errorf("failed to parse logs: %v", err)
	}
	return logs, nil
}
