package models

import (
	"time"

	"github.com/google/uuid"
)

// LoanApplicationStatus represents the possible status values of a loan application
type LoanApplicationStatus string

const (
	// LoanStatusPending indicates the loan application is waiting for review
	LoanStatusPending LoanApplicationStatus = "pending"
	// LoanStatusApproved indicates the loan application has been approved
	LoanStatusApproved LoanApplicationStatus = "approved"
	// LoanStatusRejected indicates the loan application has been rejected
	LoanStatusRejected LoanApplicationStatus = "rejected"
	// LoanStatusMoreInfoRequired indicates additional information is needed
	LoanStatusMoreInfoRequired LoanApplicationStatus = "more_info_required"
)

// LoanApplication represents a personal loan application
type LoanApplication struct {
	ID              uuid.UUID             `json:"id" db:"id"`
	CustomerID      uuid.UUID             `json:"customer_id" db:"customer_id"`
	AmountRequested float64               `json:"amount_requested" db:"amount_requested"`
	Purpose         string                `json:"purpose" db:"purpose"`
	IncomeDetails   IncomeDetails         `json:"income_details" db:"income_details"`
	Status          LoanApplicationStatus `json:"status" db:"status"`
	CreatedAt       time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at" db:"updated_at"`
	StatusReason    string                `json:"status_reason,omitempty" db:"status_reason"`
}

// IncomeDetails contains information about the customer's income
type IncomeDetails struct {
	MonthlyIncome    float64 `json:"monthly_income"`
	EmployerName     string  `json:"employer_name"`
	EmploymentYears  int     `json:"employment_years"`
	AdditionalIncome float64 `json:"additional_income,omitempty"`
	IncomeSource     string  `json:"income_source"`
}

// LoanApplicationRequest represents the request payload for creating a loan application
type LoanApplicationRequest struct {
	AmountRequested float64       `json:"amount_requested" validate:"required,min=1000"`
	Purpose         string        `json:"purpose" validate:"required,min=5,max=200"`
	IncomeDetails   IncomeDetails `json:"income_details" validate:"required"`
}

// LoanApplicationResponse represents the response for loan application endpoints
type LoanApplicationResponse struct {
	ID              uuid.UUID             `json:"id"`
	Status          LoanApplicationStatus `json:"status"`
	AmountRequested float64               `json:"amount_requested"`
	Purpose         string                `json:"purpose"`
	CreatedAt       time.Time             `json:"created_at"`
	Message         string                `json:"message,omitempty"`
}
