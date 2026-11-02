package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	StatusActive   = "ACTIVE"
	StatusInactive = "INACTIVE"
)

// User is the document stored in the users collection.
type User struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	FirstName   string        `bson:"firstName" json:"firstName"`
	LastName    string        `bson:"lastName" json:"lastName"`
	Email       string        `bson:"email" json:"email"`
	Phone       string        `bson:"phone" json:"phone"`
	Department  string        `bson:"department" json:"department"`
	Designation string        `bson:"designation" json:"designation"`
	Status      string        `bson:"status" json:"status"`
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt" json:"updatedAt"`
}

// UserInput is the request body for create and update.
type UserInput struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Department  string `json:"department"`
	Designation string `json:"designation"`
}

// StatusInput is the request body for status changes.
type StatusInput struct {
	Status string `json:"status"`
}

// ListQuery holds list parameters, already validated by the service.
type ListQuery struct {
	Page     int
	PageSize int
	Search   string
	Status   string
}

// Stats are the dashboard counts.
type Stats struct {
	Total    int64 `json:"total"`
	Active   int64 `json:"active"`
	Inactive int64 `json:"inactive"`
}

// Meta is pagination metadata in list responses.
type Meta struct {
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
}
