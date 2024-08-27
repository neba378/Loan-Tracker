package Domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LogEntry struct {
	ID        primitive.ObjectID `json:"id" bson:"id"`                               // Unique identifier for the log entry
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`                 // Time when the log was created
	LogType   string             `json:"log_type" bson:"log_type"`                   // Type of log (e.g., login_attempt, loan_submission)
	Message   string             `json:"message" bson:"message"`                     // Detailed message about the log entry
	UserID    string             `json:"user_id,omitempty" bson:"user_id,omitempty"` // User associated with the log (if applicable)
}

type LogFilter struct {
	LogType   string    // Optional: filter by log type (e.g., login_attempt, loan_submission)
	StartDate time.Time // Optional: filter logs starting from this date
	EndDate   time.Time // Optional: filter logs up to this date
	UserID    string    // Optional: filter logs by user ID
}
