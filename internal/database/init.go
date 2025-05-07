package database

import (
	"database/sql"
	"log"
)

// InitDatabase initializes all required database tables
func InitDatabase(db *sql.DB) error {
	// Initialize loan_applications table
	err := createLoanApplicationsTable(db)
	if err != nil {
		return err
	}

	log.Println("Database tables initialized successfully")
	return nil
}

// createLoanApplicationsTable creates the loan_applications table if it doesn't exist
func createLoanApplicationsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS loan_applications (
		id UUID PRIMARY KEY,
		customer_id UUID NOT NULL,
		amount_requested DECIMAL(15, 2) NOT NULL,
		purpose TEXT NOT NULL,
		income_details JSONB NOT NULL,
		status VARCHAR(20) NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		status_reason TEXT
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	log.Println("Loan applications table initialized")
	return nil
}
