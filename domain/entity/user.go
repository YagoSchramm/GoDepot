package entity

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `json:"id,omitempty"`
	Email    string    `json:"email,omitempty"`
	Password string    `json:"password,omitempty"`
}
type UserCredentials struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}
