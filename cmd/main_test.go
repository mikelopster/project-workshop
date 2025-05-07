package main

import (
	"io"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/gofiber/fiber/v2"
)

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

package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestInternalTransfer(t *testing.T) {
	// Reset mock data before each test
	mockAccounts = []Account{
		{ID: "ACC001", Name: "John Doe", Balance: 10000.00},
		{ID: "ACC002", Name: "Jane Smith", Balance: 5000.00},
		{ID: "ACC003", Name: "Bob Johnson", Balance: 7500.00},
	}
	mockTransactions = []Transaction{}

	// Setup the app for testing
	app := setupApp()

	// Test cases
	tests := []struct {
		name           string
		request        InternalTransferRequest
		token          string
		expectedStatus int
		checkBalance   bool
		expectedSrcBal float64
		expectedDstBal float64
		txnCount       int
	}{
		{
			name: "Successful Transfer",
			request: InternalTransferRequest{
				FromAccountID: "ACC001",
				ToAccountID:   "ACC002",
				Amount:        1000.00,
				Note:          "Test transfer",
			},
			token:          "Bearer valid-token",
			expectedStatus: 200,
			checkBalance:   true,
			expectedSrcBal: 9000.00,
			expectedDstBal: 6000.00,
			txnCount:       2,
		},
		{
			name: "Missing Token",
			request: InternalTransferRequest{
				FromAccountID: "ACC001",
				ToAccountID:   "ACC002",
				Amount:        1000.00,
				Note:          "Test transfer",
			},
			token:          "",
			expectedStatus: 401,
			checkBalance:   false,
			txnCount:       0,
		},
		{
			name: "Invalid Token Format",
			request: InternalTransferRequest{
				FromAccountID: "ACC001",
				ToAccountID:   "ACC002",
				Amount:        1000.00,
				Note:          "Test transfer",
			},
			token:          "InvalidToken",
			expectedStatus: 401,
			checkBalance:   false,
			txnCount:       0,
		},
		{
			name: "Source Account Not Found",
			request: InternalTransferRequest{
				FromAccountID: "NONEXISTENT",
				ToAccountID:   "ACC002",
				Amount:        1000.00,
				Note:          "Test transfer",
			},
			token:          "Bearer valid-token",
			expectedStatus: 404,
			checkBalance:   false,
			txnCount:       0,
		},
		{
			name: "Destination Account Not Found",
			request: InternalTransferRequest{
				FromAccountID: "ACC001",
				ToAccountID:   "NONEXISTENT",
				Amount:        1000.00,
				Note:          "Test transfer",
			},
			token:          "Bearer valid-token",
			expectedStatus: 404,
			checkBalance:   false,
			txnCount:       0,
		},
		{
			name: "Insufficient Funds",
			request: InternalTransferRequest{
				FromAccountID: "ACC001",
				ToAccountID:   "ACC002",
				Amount:        20000.00,
				Note:          "Test transfer",
			},
			token:          "Bearer valid-token",
			expectedStatus: 400,
			checkBalance:   false,
			txnCount:       0,
		},
		{
			name: "Invalid Amount (Zero)",
			request: InternalTransferRequest{
				FromAccountID: "ACC001",
				ToAccountID:   "ACC002",
				Amount:        0,
				Note:          "Test transfer",
			},
			token:          "Bearer valid-token",
			expectedStatus: 400,
			checkBalance:   false,
			txnCount:       0,
		},
		{
			name: "Invalid Amount (Negative)",
			request: InternalTransferRequest{
				FromAccountID: "ACC001",
				ToAccountID:   "ACC002",
				Amount:        -100.00,
				Note:          "Test transfer",
			},
			token:          "Bearer valid-token",
			expectedStatus: 400,
			checkBalance:   false,
			txnCount:       0,
		},
		{
			name: "Missing Source Account",
			request: InternalTransferRequest{
				FromAccountID: "",
				ToAccountID:   "ACC002",
				Amount:        1000.00,
				Note:          "Test transfer",
			},
			token:          "Bearer valid-token",
			expectedStatus: 400,
			checkBalance:   false,
			txnCount:       0,
		},
		{
			name: "Missing Destination Account",
			request: InternalTransferRequest{
				FromAccountID: "ACC001",
				ToAccountID:   "",
				Amount:        1000.00,
				Note:          "Test transfer",
			},
			token:          "Bearer valid-token",
			expectedStatus: 400,
			checkBalance:   false,
			txnCount:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset transactions for each test
			mockTransactions = []Transaction{}
			
			// Create request body
			reqBody, _ := json.Marshal(tt.request)
			
			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/transactions/transfer/internal", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}
			
			// Execute request
			resp, _ := app.Test(req)
			
			// Check status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			
			// For successful transfers, check account balances and transaction count
			if tt.checkBalance {
				// Check source account balance
				srcAcc, found := findAccount(tt.request.FromAccountID)
				assert.True(t, found)
				assert.Equal(t, tt.expectedSrcBal, srcAcc.Balance)
				
				// Check destination account balance
				dstAcc, found := findAccount(tt.request.ToAccountID)
				assert.True(t, found)
				assert.Equal(t, tt.expectedDstBal, dstAcc.Balance)
				
				// Check transaction count
				assert.Equal(t, tt.txnCount, len(mockTransactions))
				
				// For successful transfers, also check response body
				if resp.StatusCode == 200 {
					body, _ := ioutil.ReadAll(resp.Body)
					var response map[string]interface{}
					json.Unmarshal(body, &response)
					
					assert.Equal(t, "success", response["status"])
					assert.Equal(t, "Funds transferred successfully", response["message"])
					
					data, ok := response["data"].(map[string]interface{})
					assert.True(t, ok)
					assert.Equal(t, tt.request.FromAccountID, data["fromAccountId"])
					assert.Equal(t, tt.request.ToAccountID, data["toAccountId"])
					assert.Equal(t, tt.request.Amount, data["amount"])
					assert.Equal(t, tt.request.Note, data["note"])
					
					transactions, ok := data["transactions"].([]interface{})
					assert.True(t, ok)
					assert.Equal(t, 2, len(transactions))
					
					// Check transaction types
					debitFound := false
					creditFound := false
					for _, txn := range transactions {
						txnMap, ok := txn.(map[string]interface{})
						assert.True(t, ok)
						
						txnType := txnMap["transactionType"].(string)
						if txnType == "DEBIT" {
							debitFound = true
						} else if txnType == "CREDIT" {
							creditFound = true
						}
					}
					assert.True(t, debitFound, "Debit transaction not found")
					assert.True(t, creditFound, "Credit transaction not found")
				}
			}
		})
	}
}

// Test the helper functions
func TestFindAccount(t *testing.T) {
	// Reset mock data
	mockAccounts = []Account{
		{ID: "ACC001", Name: "John Doe", Balance: 10000.00},
		{ID: "ACC002", Name: "Jane Smith", Balance: 5000.00},
	}
	
	// Test finding existing account
	acc, found := findAccount("ACC001")
	assert.True(t, found)
	assert.Equal(t, "ACC001", acc.ID)
	assert.Equal(t, "John Doe", acc.Name)
	assert.Equal(t, 10000.00, acc.Balance)
	
	// Test finding non-existent account
	_, found = findAccount("NONEXISTENT")
	assert.False(t, found)
}

func TestGenerateTransactionID(t *testing.T) {
	// Test that IDs are unique
	id1 := generateTransactionID()
	id2 := generateTransactionID()
	assert.NotEqual(t, id1, id2)
	
	// Test ID format
	assert.Contains(t, id1, "TXN")
	assert.Equal(t, 13, len(id1)) // "TXN" + 10 digits
}

func TestValidateToken(t *testing.T) {
	app := fiber.New()
	
	// Setup test route with token validation
	app.Get("/protected", validateToken, func(c *fiber.Ctx) error {
		return c.SendString("Protected content")
	})
	
	// Test with valid token
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
	
	// Test with missing token
	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp, _ = app.Test(req)
	assert.Equal(t, 401, resp.StatusCode)
	
	// Test with invalid token format
	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "InvalidToken")
	resp, _ = app.Test(req)
	assert.Equal(t, 401, resp.StatusCode)
}
