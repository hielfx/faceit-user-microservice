package models

import (
	"encoding/json"
	"time"
	"user-microservice/internal/pagination"

	"go.mongodb.org/mongo-driver/bson"
)

// User - user model
type User struct {
	ID        string `json:"id" bson:"_id" example:"ddd50d89-0cf4-4d35-b8e8-51a2b5a06ce4"`
	FirstName string `json:"firstName" bson:"first_name" example:"Alice" validate:"required"`
	LastName  string `json:"lastName" bson:"last_name" example:"Tingo" validate:"required"`
	Nickname  string `json:"nickname" bson:"nickname" example:"atingo" validate:"required"`
	// Should be json:"-" in order to hide the password
	Password  string    `json:"password" bson:"password"`
	Email     string    `json:"email" bson:"email" example:"atingo@example.com" validate:"required"`
	Country   string    `json:"country" bson:"country" example:"DE" validate:"required"`
	CreatedAt time.Time `json:"createdAt" bson:"created_at" example:"2016-05-18T16:00:00Z"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updated_at" example:"2016-05-18T16:00:00Z"`
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

// MarshalBinary - custom encoding.BinaryMarshaler implementation
// We need this to encode and decode when publishing into redis
func (u User) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}

// PaginatedUsers - users pagination data
type PaginatedUsers struct {
	pagination.Paginated
	Users []User `json:"users"`
}

// UserFilters - used when filtering users
type UserFilters struct {
	FirstName string `query:"firstName" bson:"first_name,omitempty"`
	LastName  string `query:"lastName" bson:"last_name,omitempty"`
	Nickname  string `query:"nickname" bson:"nickname,omitempty"`
	Email     string `query:"email" bson:"email,omitempty"`
	Country   string `query:"country" bson:"country,omitempty"`
}

// ToBsonM - converts the current UserFilters into bson.M in order to use it in mongodb
func (uf UserFilters) ToBsonM() bson.M {
	res := bson.M{}

	if uf.FirstName != "" {
		res["first_name"] = uf.FirstName
	}
	if uf.LastName != "" {
		res["last_name"] = uf.LastName
	}
	if uf.Email != "" {
		res["email"] = uf.Email
	}
	if uf.Country != "" {
		res["country"] = uf.Country
	}

	return res
}
