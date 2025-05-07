package database

import (
	"database/sql"
	"errors"

	"example.com/m/internal/models"
)

// ErrCustomerNotFound is returned when a customer cannot be found in the database
var ErrCustomerNotFound = errors.New("customer not found")

// CustomerRepositoryInterface is used to define methods for customer data access
type CustomerRepositoryInterface interface {
	GetByID(id string) (*models.Customer, error)
}

// CustomerRepository handles all database operations related to customers
type CustomerRepository struct {
	DB *sql.DB
}

// NewCustomerRepository creates a new CustomerRepository instance
func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{
		DB: db,
	}
}

// GetByID retrieves a customer by their ID
func (r *CustomerRepository) GetByID(id string) (*models.Customer, error) {
	// SQL query to select customer by ID
	query := `
		SELECT id, first_name, last_name, id_card_number, phone_number, email, 
		       address, password, created_at, updated_at 
		FROM customers 
		WHERE id = $1
	`

	// Execute the query
	row := r.DB.QueryRow(query, id)

	// Parse the result into a Customer struct
	customer := &models.Customer{}
	err := row.Scan(
		&customer.ID,
		&customer.FirstName,
		&customer.LastName,
		&customer.IDCardNumber,
		&customer.PhoneNumber,
		&customer.Email,
		&customer.Address,
		&customer.Password,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	// Handle errors
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}

	return customer, nil
}
