package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/m/internal/database"
	"example.com/m/internal/handlers"
	"example.com/m/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// CustomerRepositoryInterface defines the interface that both the real repository and mock will implement
type CustomerRepositoryInterface interface {
	GetByID(id string) (*models.Customer, error)
}

// Create a mock for CustomerRepository
type MockCustomerRepository struct {
	mock.Mock
}

func (m *MockCustomerRepository) GetByID(id string) (*models.Customer, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Customer), args.Error(1)
}

func TestGetRoot(t *testing.T) {
	// Setup the app
	app := setupApp()

	// Create a new request
	req := httptest.NewRequest("GET", "/", nil)

	// Perform the request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	// Check status code
	assert.Equal(t, 200, resp.StatusCode, "Status code should be 200")

	// Check response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Check if body equals "Hello World"
	assert.Equal(t, "Hello World", string(body), "Response body should be 'Hello World'")
}

// Helper function to generate JWT tokens for testing
func generateTestToken(customerID string) (string, error) {
	// Create the token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["customer_id"] = customerID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Sign the token with our secret
	tokenString, err := token.SignedString([]byte("your-secret-key-here"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func TestGetCurrentCustomerProfile(t *testing.T) {
	// Create test cases
	testCases := []struct {
		name             string
		customerID       string
		mockSetup        func(*MockCustomerRepository)
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name:       "Success",
			customerID: "cust-123",
			mockSetup: func(repo *MockCustomerRepository) {
				customerCreatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
				customerUpdatedAt := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
				customer := &models.Customer{
					ID:           "cust-123",
					FirstName:    "John",
					LastName:     "Doe",
					IDCardNumber: "1234567890123",
					PhoneNumber:  "0812345678",
					Email:        "john.doe@example.com",
					Address:      "123 Main St, Bangkok, Thailand",
					Password:     "hashed_password",
					CreatedAt:    customerCreatedAt,
					UpdatedAt:    customerUpdatedAt,
				}
				repo.On("GetByID", "cust-123").Return(customer, nil)
			},
			expectedStatus: 200,
			expectedResponse: models.CustomerResponse{
				ID:          "cust-123",
				FirstName:   "John",
				LastName:    "Doe",
				PhoneNumber: "0812345678",
				Email:       "john.doe@example.com",
				Address:     "123 Main St, Bangkok, Thailand",
				CreatedAt:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:       "Customer Not Found",
			customerID: "cust-456",
			mockSetup: func(repo *MockCustomerRepository) {
				repo.On("GetByID", "cust-456").Return(nil, database.ErrCustomerNotFound)
			},
			expectedStatus: 404,
			expectedResponse: fiber.Map{
				"error": "Customer not found",
			},
		},
		{
			name:       "Database Error",
			customerID: "cust-789",
			mockSetup: func(repo *MockCustomerRepository) {
				repo.On("GetByID", "cust-789").Return(nil, errors.New("database error"))
			},
			expectedStatus: 500,
			expectedResponse: fiber.Map{
				"error": "Failed to retrieve customer profile",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock repository
			mockRepo := new(MockCustomerRepository)

			// Setup the mock expectations
			tc.mockSetup(mockRepo)

			// Create a test database for the app setup
			// We need to override the CustomerHandler in the app to use our mock
			app := fiber.New()

			// Create a custom customer handler with our mock
			customHandler := &handlers.CustomerHandler{
				CustomerRepo: mockRepo,
			}

			// Register the endpoint with middleware
			app.Get("/customers/me", func(c *fiber.Ctx) error {
				// For testing, we'll set the customerID directly in Locals
				// In production, this would be done by the middleware
				c.Locals("customerID", tc.customerID)
				return customHandler.GetCurrentCustomerProfile(c)
			})

			// Generate a JWT token for the test
			token, err := generateTestToken(tc.customerID)
			assert.NoError(t, err, "Failed to generate test token")

			// Create a new request with the token
			req := httptest.NewRequest("GET", "/customers/me", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			// Perform the request
			resp, err := app.Test(req)
			assert.NoError(t, err, "Failed to test request")

			// Check status code
			assert.Equal(t, tc.expectedStatus, resp.StatusCode, "Status code should match expected")

			// Check response body
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err, "Failed to read response body")

			// For success case, check all fields against expected response
			if tc.expectedStatus == 200 {
				var customerResponse models.CustomerResponse
				err = json.Unmarshal(body, &customerResponse)
				assert.NoError(t, err, "Failed to unmarshal response")

				expected := tc.expectedResponse.(models.CustomerResponse)
				assert.Equal(t, expected.ID, customerResponse.ID)
				assert.Equal(t, expected.FirstName, customerResponse.FirstName)
				assert.Equal(t, expected.LastName, customerResponse.LastName)
				assert.Equal(t, expected.PhoneNumber, customerResponse.PhoneNumber)
				assert.Equal(t, expected.Email, customerResponse.Email)
				assert.Equal(t, expected.Address, customerResponse.Address)
				assert.Equal(t, expected.CreatedAt.Format(time.RFC3339), customerResponse.CreatedAt.Format(time.RFC3339))
			} else {
				// For error cases, check the error message
				var errorResponse map[string]string
				err = json.Unmarshal(body, &errorResponse)
				assert.NoError(t, err, "Failed to unmarshal error response")

				expected := tc.expectedResponse.(fiber.Map)
				assert.Equal(t, expected["error"], errorResponse["error"])
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}
