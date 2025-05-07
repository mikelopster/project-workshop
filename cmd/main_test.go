ppackage main

import (
    "bytes"
    "encoding/json"
    "errors"
    "io"
    "io/ioutil"
    "net/http"
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

// MockCustomerRepository is a mock for CustomerRepositoryInterface
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

func generateTestToken(customerID string) (string, error) {
    token := jwt.New(jwt.SigningMethodHS256)
    claims := token.Claims.(jwt.MapClaims)
    claims["customer_id"] = customerID
    claims["exp"] = time.Now().Add(72 * time.Hour).Unix()
    return token.SignedString([]byte("your-secret-key-here"))
}

func TestGetRoot(t *testing.T) {
    app := setupApp()
    req := httptest.NewRequest("GET", "/", nil)
    resp, err := app.Test(req)
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)

    body, err := io.ReadAll(resp.Body)
    assert.NoError(t, err)
    assert.Equal(t, "Hello World", string(body))
}

func TestGetCurrentCustomerProfile(t *testing.T) {
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
                created := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
                updated := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
                customer := &models.Customer{
                    ID:           "cust-123",
                    FirstName:    "John",
                    LastName:     "Doe",
                    IDCardNumber: "1234567890123",
                    PhoneNumber:  "0812345678",
                    Email:        "john.doe@example.com",
                    Address:      "123 Main St, Bangkok, Thailand",
                    Password:     "hashed_password",
                    CreatedAt:    created,
                    UpdatedAt:    updated,
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
            expectedStatus:   404,
            expectedResponse: fiber.Map{"error": "Customer not found"},
        },
        {
            name:       "Database Error",
            customerID: "cust-789",
            mockSetup: func(repo *MockCustomerRepository) {
                repo.On("GetByID", "cust-789").Return(nil, errors.New("database error"))
            },
            expectedStatus:   500,
            expectedResponse: fiber.Map{"error": "Failed to retrieve customer profile"},
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            mockRepo := new(MockCustomerRepository)
            tc.mockSetup(mockRepo)

            app := fiber.New()
            handler := &handlers.CustomerHandler{CustomerRepo: mockRepo}
            app.Get("/customers/me", func(c *fiber.Ctx) error {
                c.Locals("customerID", tc.customerID)
                return handler.GetCurrentCustomerProfile(c)
            })

            token, err := generateTestToken(tc.customerID)
            assert.NoError(t, err)

            req := httptest.NewRequest("GET", "/customers/me", nil)
            req.Header.Set("Authorization", "Bearer "+token)
            resp, err := app.Test(req)
            assert.NoError(t, err)
            assert.Equal(t, tc.expectedStatus, resp.StatusCode)

            body, err := io.ReadAll(resp.Body)
            assert.NoError(t, err)

            if tc.expectedStatus == 200 {
                var res models.CustomerResponse
                err = json.Unmarshal(body, &res)
                assert.NoError(t, err)
                expected := tc.expectedResponse.(models.CustomerResponse)
                assert.Equal(t, expected.ID, res.ID)
                assert.Equal(t, expected.FirstName, res.FirstName)
                assert.Equal(t, expected.LastName, res.LastName)
                assert.Equal(t, expected.PhoneNumber, res.PhoneNumber)
                assert.Equal(t, expected.Email, res.Email)
                assert.Equal(t, expected.Address, res.Address)
                assert.Equal(t, expected.CreatedAt.Format(time.RFC3339), res.CreatedAt.Format(time.RFC3339))
            } else {
                var errRes map[string]string
                err = json.Unmarshal(body, &errRes)
                assert.NoError(t, err)
                expected := tc.expectedResponse.(fiber.Map)
                assert.Equal(t, expected["error"], errRes["error"])
            }

            mockRepo.AssertExpectations(t)
        })
    }
}

func TestInternalTransfer(t *testing.T) {
    // Reset mock data
    mockAccounts = []Account{
        {ID: "ACC001", Name: "John Doe", Balance: 10000.00},
        {ID: "ACC002", Name: "Jane Smith", Balance: 5000.00},
        {ID: "ACC003", Name: "Bob Johnson", Balance: 7500.00},
    }
    mockTransactions = []Transaction{}

    app := setupApp()

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
        // ... เพิ่มกรณีทดสอบอื่นๆ ตามที่มีในไฟล์เดิม ...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockTransactions = []Transaction{}

            reqBody, _ := json.Marshal(tt.request)
            req := httptest.NewRequest(http.MethodPost, "/transactions/transfer/internal", bytes.NewReader(reqBody))
            req.Header.Set("Content-Type", "application/json")
            if tt.token != "" {
                req.Header.Set("Authorization", tt.token)
            }

            resp, _ := app.Test(req)
            assert.Equal(t, tt.expectedStatus, resp.StatusCode)

            if tt.checkBalance {
                srcAcc, found := findAccount(tt.request.FromAccountID)
                assert.True(t, found)
                assert.Equal(t, tt.expectedSrcBal, srcAcc.Balance)

                dstAcc, found := findAccount(tt.request.ToAccountID)
                assert.True(t, found)
                assert.Equal(t, tt.expectedDstBal, dstAcc.Balance)

                assert.Equal(t, tt.txnCount, len(mockTransactions))

                if resp.StatusCode == 200 {
                    body, _ := ioutil.ReadAll(resp.Body)
                    var response map[string]interface{}
                    json.Unmarshal(body, &response)

                    assert.Equal(t, "success", response["status"])
                    assert.Equal(t, "Funds transferred successfully", response["message"])

                    data := response["data"].(map[string]interface{})
                    assert.Equal(t, tt.request.FromAccountID, data["fromAccountId"])
                    assert.Equal(t, tt.request.ToAccountID, data["toAccountId"])
                    assert.Equal(t, tt.request.Amount, data["amount"])
                    assert.Equal(t, tt.request.Note, data["note"])

                    transactions := data["transactions"].([]interface{})
                    assert.Equal(t, 2, len(transactions))
                }
            }
        })
    }
}

func TestFindAccount(t *testing.T) {
    mockAccounts = []Account{
        {ID: "ACC001", Name: "John Doe", Balance: 10000.00},
        {ID: "ACC002", Name: "Jane Smith", Balance: 5000.00},
    }

    acc, found := findAccount("ACC001")
    assert.True(t, found)
    assert.Equal(t, "ACC001", acc.ID)
    assert.Equal(t, 10000.00, acc.Balance)

    _, found = findAccount("NONEXISTENT")
    assert.False(t, found)
}

func TestGenerateTransactionID(t *testing.T) {
    id1 := generateTransactionID()
    id2 := generateTransactionID()
    assert.NotEqual(t, id1, id2)
    assert.Contains(t, id1, "TXN")
    assert.Len(t, id1, 13)
}

func TestValidateToken(t *testing.T) {
    app := fiber.New()
    app.Get("/protected", validateToken, func(c *fiber.Ctx) error {
        return c.SendString("Protected content")
    })

    req := httptest.NewRequest(http.MethodGet, "/protected", nil)
    req.Header.Set("Authorization", "Bearer valid-token")
    resp, _ := app.Test(req)
    assert.Equal(t, 200, resp.StatusCode)

    req = httptest.NewRequest(http.MethodGet, "/protected", nil)
    resp, _ = app.Test(req)
    assert.Equal(t, 401, resp.StatusCode)

    req = httptest.NewRequest(http.MethodGet, "/protected", nil)
    req.Header.Set("Authorization", "InvalidToken")
    resp, _ = app.Test(req)
    assert.Equal(t, 401, resp.StatusCode)
}

func TestGetCustomerDetailsSuccess(t *testing.T) {
    app := setupApp()
    req := httptest.NewRequest("GET", "/staff/customers/12345", nil)
    resp, err := app.Test(req)
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)

    body, err := io.ReadAll(resp.Body)
    assert.NoError(t, err)

    var customer map[string]interface{}
    err = json.Unmarshal(body, &customer)
    assert.NoError(t, err)
    assert.Equal(t, "12345", customer["id"])
    assert.Equal(t, "สมชาย", customer["first_name"])
    assert.Equal(t, "ใจดี", customer["last_name"])
}

func TestGetCustomerDetailsNotFound(t *testing.T) {
    app := setupApp()
    req := httptest.NewRequest("GET", "/staff/customers/99999", nil)
    resp, err := app.Test(req)
    assert.NoError(t, err)
    assert.Equal(t, 404, resp.StatusCode)

    body, err := io.ReadAll(resp.Body)
    assert.NoError(t, err)

    var errorResponse map[string]interface{}
    err = json.Unmarshal(body, &errorResponse)
    assert.NoError(t, err)
    assert.Contains(t, errorResponse, "error")
}

