package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// JWTMiddleware handles authentication by JWT token
// Note: This is a simplified version for the purpose of this example
// In a real application, you'd use a proper JWT library to verify tokens
func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")

		// Check if authorization header exists and has the right format
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or invalid Authorization header",
			})
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Note: In a real application, you would validate the token here
		// For this example, we'll just set a mock customer ID
		// This would normally come from decoding and validating the JWT token

		// Mock customer ID for development purposes only
		customerID, err := uuid.Parse("00000000-0000-0000-0000-000000000001")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to parse customer ID",
			})
		}

		// Set customer ID in context for handlers to use
		c.Locals("customerID", customerID)

		return c.Next()
	}
}

// StaffAuthMiddleware validates that the request is coming from bank staff
// Note: This is a simplified version for demo purposes
func StaffAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")

		// Check if authorization header exists and has the right format
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or invalid Authorization header",
			})
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// In a real application, you would validate the token and check if the user is staff
		// For this example, we'll just set a mock staff ID
		staffID, err := uuid.Parse("00000000-0000-0000-0000-000000000002")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to parse staff ID",
			})
		}

		c.Locals("staffID", staffID)
		c.Locals("isStaff", true)

		return c.Next()
	}
}

// GetCustomerIDFromContext extracts the customer ID from the context
func GetCustomerIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	customerID, ok := c.Locals("customerID").(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("customer ID not found in context")
	}
	return customerID, nil
}
