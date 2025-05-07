package middleware

import (
    "errors"
    "fmt"
    "strings"

    "github.com/fiber/fiber/v2"
    "github.com/golang-jwt/jwt/v4"
    "github.com/google/uuid"
)

// JWTSecret is the secret key used for signing JWT tokens.
// In production, load this from environment variables.
const JWTSecret = "your-secret-key-here"

// JWTMiddleware validates the JWT token, extracts the customer ID claim,
// and stores it in context locals.
func JWTMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Authorization header is required",
            })
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Authorization header format must be Bearer {token}",
            })
        }

        tokenString := parts[1]
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(JWTSecret), nil
        })
        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Invalid token: " + err.Error(),
            })
        }

        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            customerID, ok := claims["customer_id"].(string)
            if !ok {
                return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                    "error": "Invalid token: missing customer_id claim",
                })
            }
            c.Locals("customerID", customerID)
            return c.Next()
        }

        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Invalid token",
        })
    }
}

// StaffAuthMiddleware validates that the request is coming from bank staff.
// Sets staffID and isStaff flag in context locals.
func StaffAuthMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Missing or invalid Authorization header",
            })
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")
        if token == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Invalid token",
            })
        }

        // In production, validate the token and check staff privileges.
        // Here we mock a staff ID for demo purposes.
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

// GetCustomerIDFromContext retrieves the customer ID (string) from context locals.
func GetCustomerIDFromContext(c *fiber.Ctx) (string, error) {
    id, ok := c.Locals("customerID").(string)
    if !ok || id == "" {
        return "", errors.New("customer ID not found in context")
    }
    return id, nil
}
