package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"

    "example.com/m/internal/database"
    "example.com/m/internal/handlers"
    "example.com/m/internal/middleware"
    "example.com/m/internal/repository"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/logger"
    _ "github.com/lib/pq" // PostgreSQL driver
)

// Database connection
var db *sql.DB

// Customer คือโมเดลข้อมูลลูกค้าธนาคาร
type Customer struct {
    ID           string    `json:"id"`
    FirstName    string    `json:"first_name"`
    LastName     string    `json:"last_name"`
    IDCardNumber string    `json:"id_card_number"`
    PhoneNumber  string    `json:"phone_number"`
    Email        string    `json:"email"`
    Address      string    `json:"address"`
    CreatedAt    time.Time `json:"created_at"`
}

// setupDatabase initializes the PostgreSQL connection
func setupDatabase() (*sql.DB, error) {
    connStr := "postgresql://postgres:postgres@localhost:5432/workshop?sslmode=disable"

    var err error
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }
    if err = db.Ping(); err != nil {
        return nil, err
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)

    if err = database.InitDatabase(db); err != nil {
        return nil, fmt.Errorf("failed to initialize database schema: %w", err)
    }

    fmt.Println("Connected to PostgreSQL database!")
    return db, nil
}

// UpdateContactRequest represents the request body for updating contact information
type UpdateContactRequest struct {
    Phone string `json:"phone"`
    Email string `json:"email"`
}

// staffAuthMiddleware ตรวจสอบว่า request มาจากพนักงานที่มีสิทธิ์หรือไม่
func staffAuthMiddleware(c *fiber.Ctx) error {
    // ในแอพจริงควรจะมีการตรวจสอบ JWT token และสิทธิ์ของพนักงาน
    // ตอนนี้เป็นเพียงตัวอย่าง จึงให้ผ่านไปทุก request
    return c.Next()
}

// getCustomerDetails จัดการคำขอดูรายละเอียดลูกค้าจากพนักงานธนาคาร
func getCustomerDetails(c *fiber.Ctx) error {
    customerID := c.Params("customerId")

    customer, err := getCustomerByID(customerID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "เกิดข้อผิดพลาดในการดึงข้อมูลลูกค้า",
        })
    }
    if customer == nil {
        return c.Status(404).JSON(fiber.Map{
            "error": "ไม่พบข้อมูลลูกค้า",
        })
    }
    return c.JSON(customer)
}

// getCustomerByID เป็นฟังก์ชันช่วยในการดึงข้อมูลลูกค้าจาก ID
func getCustomerByID(id string) (*Customer, error) {
    // นี่เป็นแค่ข้อมูลตัวอย่างสำหรับการทดสอบ
    if id == "12345" {
        return &Customer{
            ID:           "12345",
            FirstName:    "สมชาย",
            LastName:     "ใจดี",
            IDCardNumber: "1234567890123",
            PhoneNumber:  "0891234567",
            Email:        "somchai@example.com",
            Address:      "123 ถนนสุขุมวิท กรุงเทพฯ",
            CreatedAt:    time.Now().Add(-24 * time.Hour),
        }, nil
    }
    return nil, nil
}

// setupStaffRoutes กำหนด routes สำหรับส่วนของพนักงาน
func setupStaffRoutes(app *fiber.App) {
    staff := app.Group("/staff")
    staff.Use(staffAuthMiddleware)
    staff.Get("/customers/:customerId", getCustomerDetails)
}

// setupApp configures and returns a Fiber app instance
func setupApp() *fiber.App {
    app := fiber.New(fiber.Config{
        ErrorHandler: func(c *fiber.Ctx, err error) error {
            code := fiber.StatusInternalServerError
            if e, ok := err.(*fiber.Error); ok {
                code = e.Code
            }
            return c.Status(code).JSON(fiber.Map{
                "error": err.Error(),
            })
        },
    })

    // Global middleware
    app.Use(logger.New())
    app.Use(cors.New())

    // Basic health routes
    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("Hello World")
    })
    app.Get("/db-status", func(c *fiber.Ctx) error {
        if db == nil {
            return c.Status(500).SendString("Database not connected")
        }
        if err := db.Ping(); err != nil {
            return c.Status(500).SendString(fmt.Sprintf("Database error: %v", err))
        }
        return c.SendString("Database connected successfully")
    })

    // Customer routes
    customerHandler := handlers.NewCustomerHandler(db)
    app.Get("/customers/me", middleware.JWTMiddleware(), customerHandler.GetCurrentCustomerProfile)
    app.Put("/customers/me/contact", func(c *fiber.Ctx) error {
        token := c.Get("Authorization")
        if token == "" {
            return c.Status(401).SendString("Unauthorized: Token is required")
        }

        var req UpdateContactRequest
        if err := json.Unmarshal(c.Body(), &req); err != nil {
            return c.Status(400).SendString("Invalid request body")
        }
        if req.Phone == "" || req.Email == "" {
            return c.Status(400).SendString("Phone and email are required")
        }

        query := "UPDATE customers SET phone = $1, email = $2 WHERE id = $3"
        customerID := 1 // TODO: extract actual customer ID from token
        if _, err := db.Exec(query, req.Phone, req.Email, customerID); err != nil {
            return c.Status(500).SendString(fmt.Sprintf("Failed to update contact information: %v", err))
        }

        return c.Status(200).JSON(fiber.Map{
            "message": "Contact information updated successfully",
        })
    })

    // Loan feature
    loanRepo := repository.NewPostgresLoanRepository(db)
    loanHandler := handlers.NewLoanHandler(loanRepo)
    api := app.Group("/api/v1")
    loans := api.Group("/loans")
    loans.Post("/personal/apply", middleware.JWTMiddleware(), loanHandler.ApplyForPersonalLoan)

    // Staff routes
    setupStaffRoutes(app)

    return app
}

func main() {
    var err error
    db, err = setupDatabase()
    if err != nil {
        log.Printf("Failed to connect to database: %v", err)
        os.Exit(1)
    }
    defer db.Close()

    app := setupApp()
    log.Println("Starting server on port 3000...")
    app.Listen(":3000")
}
