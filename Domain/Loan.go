package Domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Loan struct {
	ID        primitive.ObjectID `json:"id" bson:"id"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	Amount    float64            `json:"amount" bson:"amount"`
	Term      int                `json:"term" bson:"term"` // In months
	Purpose   string             `json:"purpose" bson:"purpose"`
	Status    string             `json:"status" bson:"status"` // "pending", "approved", "rejected"
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

type LoanStatus struct {
	ID        primitive.ObjectID `json:"id" bson:"id"`
	LoanID    primitive.ObjectID `json:"loan_id" bson:"loan_id"`
	Status    string             `json:"status" bson:"status"` // "pending", "approved", "rejected"
	ChangedAt time.Time          `json:"changed_at" bson:"changed_at"`
	ChangedBy primitive.ObjectID `json:"changed_by" bson:"changed_by"` // UserID of the admin who changed the status
}

type LoanUpdateInput struct {
	Amount  float64 `json:"amount" bson:"amount"`
	Term    int     `json:"term" bson:"term"` // In months
	Purpose string  `json:"purpose" bson:"purpose"`
	Status  string  `json:"status" bson:"status"` // "pending", "approved", "rejected"
}

type LoanStatusUpdateInput struct {
	Status    string             `json:"status" bson:"status"`         // "pending", "approved", "rejected"
	ChangedBy primitive.ObjectID `json:"changed_by" bson:"changed_by"` // UserID of the admin who changed the status
}

type LoanInput struct {
	UserID  primitive.ObjectID `json:"user_id" bson:"user_id"`
	Amount  float64            `json:"amount" bson:"amount"`
	Term    int                `json:"term" bson:"term"` // In months
	Purpose string             `json:"purpose" bson:"purpose"`
}
