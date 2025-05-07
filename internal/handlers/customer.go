package handlers

import (
	"database/sql"
	"errors"

	"example.com/m/internal/database"
	"example.com/m/internal/models"
	"github.com/gofiber/fiber/v2"
)

// CustomerRepositoryInterface defines the interface for customer repository operations
type CustomerRepositoryInterface interface {
	GetByID(id string) (*models.Customer, error)
}

// CustomerHandler handles HTTP requests related to customers
type CustomerHandler struct {
	CustomerRepo CustomerRepositoryInterface
}

// NewCustomerHandler creates a new CustomerHandler instance
func NewCustomerHandler(db *sql.DB) *CustomerHandler {
	return &CustomerHandler{
		CustomerRepo: database.NewCustomerRepository(db),
	}
}

// GetCurrentCustomerProfile handles GET /customers/me
// Returns the profile of the currently logged in customer
func (h *CustomerHandler) GetCurrentCustomerProfile(c *fiber.Ctx) error {
	// Get customer ID from context (set by auth middleware)
	customerID := c.Locals("customerID").(string)

	// Fetch customer from database
	customer, err := h.CustomerRepo.GetByID(customerID)
	if err != nil {
		if errors.Is(err, database.ErrCustomerNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Customer not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve customer profile",
		})
	}

	// Return customer profile (without sensitive information)
	return c.Status(fiber.StatusOK).JSON(customer.ToResponse())
}
