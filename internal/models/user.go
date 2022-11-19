package models

import (
	"time"
	"user-microservice/internal/pagination"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"_id" bson:"_id"`
	FirstName string    `json:"firstName" bson:"first_name"`
	LastName  string    `json:"lastName" bson:"last_name"`
	Nickname  string    `json:"nickname" bson:"nickname"`
	Password  string    `json:"password" bson:"password"` // should be json:"-" in order to hide the password
	Email     string    `json:"email" bson:"email"`
	Country   string    `json:"country" bson:"country"`
	CreatedAt time.Time `json:"createdAt" bson:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updated_at"`
}

type PaginatedUsers struct {
	pagination.Paginated
	Users []User `json:"users"`
}

// Valid - returns true if the user is valid.
// Valid means the main fields are not empty
func (u User) Valid() bool {
	return u.FirstName != "" &&
		u.LastName != "" &&
		u.Nickname != "" &&
		u.Password != "" &&
		u.Email != "" &&
		u.Country != ""
}

// Modify - sets the values from the given user to the current one
func (u *User) Modify(mod User) {
	if u.FirstName != mod.FirstName {
		u.FirstName = mod.FirstName
	}
	if u.LastName != mod.LastName {
		u.LastName = mod.LastName
	}
	if u.Nickname != mod.Nickname {
		u.Nickname = mod.Nickname
	}
	if u.Email != mod.Email {
		u.Email = mod.Email
	}
	if u.Country != mod.Country {
		u.Country = mod.Country
	}
	if u.Password != mod.Password {
		u.Password = mod.Password
	}
}
