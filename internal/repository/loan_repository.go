package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"example.com/m/internal/models"
	"github.com/google/uuid"
)

// LoanRepository defines operations for loan application persistence
type LoanRepository interface {
	CreateLoanApplication(ctx context.Context, application *models.LoanApplication) error
	GetLoanApplicationByID(ctx context.Context, id uuid.UUID) (*models.LoanApplication, error)
	GetCustomerLoanApplications(ctx context.Context, customerID uuid.UUID) ([]*models.LoanApplication, error)
	UpdateLoanApplicationStatus(ctx context.Context, id uuid.UUID, status models.LoanApplicationStatus, reason string) error
}

// PostgresLoanRepository implements LoanRepository for PostgreSQL
type PostgresLoanRepository struct {
	db *sql.DB
}

// NewPostgresLoanRepository creates a new PostgresLoanRepository
func NewPostgresLoanRepository(db *sql.DB) *PostgresLoanRepository {
	return &PostgresLoanRepository{
		db: db,
	}
}

// CreateLoanApplication inserts a new loan application into the database
func (r *PostgresLoanRepository) CreateLoanApplication(ctx context.Context, application *models.LoanApplication) error {
	incomeDetailsJSON, err := json.Marshal(application.IncomeDetails)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO loan_applications (
			id, customer_id, amount_requested, purpose, income_details, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		application.ID,
		application.CustomerID,
		application.AmountRequested,
		application.Purpose,
		incomeDetailsJSON,
		application.Status,
		application.CreatedAt,
		application.UpdatedAt,
	)

	return err
}

// GetLoanApplicationByID retrieves a loan application by ID
func (r *PostgresLoanRepository) GetLoanApplicationByID(ctx context.Context, id uuid.UUID) (*models.LoanApplication, error) {
	query := `
		SELECT id, customer_id, amount_requested, purpose, income_details, status, 
		       created_at, updated_at, status_reason
		FROM loan_applications 
		WHERE id = $1
	`

	var application models.LoanApplication
	var incomeDetailsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&application.ID,
		&application.CustomerID,
		&application.AmountRequested,
		&application.Purpose,
		&incomeDetailsJSON,
		&application.Status,
		&application.CreatedAt,
		&application.UpdatedAt,
		&application.StatusReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}

	// Unmarshal the JSON income details
	if err := json.Unmarshal(incomeDetailsJSON, &application.IncomeDetails); err != nil {
		return nil, err
	}

	return &application, nil
}

// GetCustomerLoanApplications retrieves all loan applications for a customer
func (r *PostgresLoanRepository) GetCustomerLoanApplications(ctx context.Context, customerID uuid.UUID) ([]*models.LoanApplication, error) {
	query := `
		SELECT id, customer_id, amount_requested, purpose, income_details, status, 
		       created_at, updated_at, status_reason
		FROM loan_applications 
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applications := []*models.LoanApplication{}

	for rows.Next() {
		var application models.LoanApplication
		var incomeDetailsJSON []byte

		err := rows.Scan(
			&application.ID,
			&application.CustomerID,
			&application.AmountRequested,
			&application.Purpose,
			&incomeDetailsJSON,
			&application.Status,
			&application.CreatedAt,
			&application.UpdatedAt,
			&application.StatusReason,
		)

		if err != nil {
			return nil, err
		}

		// Unmarshal the JSON income details
		if err := json.Unmarshal(incomeDetailsJSON, &application.IncomeDetails); err != nil {
			return nil, err
		}

		applications = append(applications, &application)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return applications, nil
}

// UpdateLoanApplicationStatus updates the status of a loan application
func (r *PostgresLoanRepository) UpdateLoanApplicationStatus(ctx context.Context, id uuid.UUID, status models.LoanApplicationStatus, reason string) error {
	query := `
		UPDATE loan_applications
		SET status = $1, status_reason = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query, status, reason, time.Now(), id)
	return err
}
