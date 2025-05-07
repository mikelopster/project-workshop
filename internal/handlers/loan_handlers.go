package handlers

import (
	"time"

	"example.com/m/internal/middleware"
	"example.com/m/internal/models"
	"example.com/m/internal/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// LoanHandler contains handlers for loan application endpoints
type LoanHandler struct {
	loanRepo repository.LoanRepository
}

// NewLoanHandler creates a new LoanHandler
func NewLoanHandler(loanRepo repository.LoanRepository) *LoanHandler {
	return &LoanHandler{
		loanRepo: loanRepo,
	}
}

// ApplyForPersonalLoan handles the submission of a new personal loan application
// Endpoint: POST /loans/personal/apply
func (h *LoanHandler) ApplyForPersonalLoan(c *fiber.Ctx) error {
	// Get customer ID from context (set by auth middleware)
	customerID, err := middleware.GetCustomerIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Parse request body
	var request models.LoanApplicationRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Validate request data
	if request.AmountRequested <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Amount requested must be greater than zero",
		})
	}

	if request.Purpose == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Loan purpose is required",
		})
	}

	// Basic validation for income details
	if request.IncomeDetails.MonthlyIncome <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Monthly income must be greater than zero",
		})
	}

	if request.IncomeDetails.EmployerName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Employer name is required",
		})
	}

	// Create loan application record
	now := time.Now()
	application := &models.LoanApplication{
		ID:              uuid.New(),
		CustomerID:      customerID,
		AmountRequested: request.AmountRequested,
		Purpose:         request.Purpose,
		IncomeDetails:   request.IncomeDetails,
		Status:          models.LoanStatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Save to database
	err = h.loanRepo.CreateLoanApplication(c.Context(), application)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create loan application",
		})
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(models.LoanApplicationResponse{
		ID:              application.ID,
		Status:          application.Status,
		AmountRequested: application.AmountRequested,
		Purpose:         application.Purpose,
		CreatedAt:       application.CreatedAt,
		Message:         "Your loan application has been submitted and is pending review",
	})
}
