package models

import "time"

// Customer represents a bank customer
type Customer struct {
	ID           string    `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	IDCardNumber string    `json:"id_card_number"`
	PhoneNumber  string    `json:"phone_number"`
	Email        string    `json:"email"`
	Address      string    `json:"address"`
	Password     string    `json:"-"` // Password is not included in JSON responses
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CustomerResponse is used for API responses to avoid sending sensitive data
type CustomerResponse struct {
	ID          string    `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
	Address     string    `json:"address"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToResponse converts a Customer to CustomerResponse (removing sensitive data)
func (c *Customer) ToResponse() CustomerResponse {
	return CustomerResponse{
		ID:          c.ID,
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		PhoneNumber: c.PhoneNumber,
		Email:       c.Email,
		Address:     c.Address,
		CreatedAt:   c.CreatedAt,
	}
}
