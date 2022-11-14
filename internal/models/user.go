package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" bson:"_id"`
	FirstName string    `json:"firstName" bson:"first_name"`
	LastName  string    `json:"lastName" bson:"last_name"`
	Nickname  string    `json:"nickname" bson:"nickname"`
	Password  string    `json:"-" bson:"password"`
	Email     string    `json:"email" bson:"email"`
	Country   string    `json:"country" bson:"country"`
	CreatedAt time.Time `json:"createdAt" bson:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updated_at"`
}
