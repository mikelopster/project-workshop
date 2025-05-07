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

// setupDatabase initializes the PostgreSQL connection
func setupDatabase() (*sql.DB, error) {
    // Connection parameters
    connStr := "postgresql://postgres:postgres@localhost:5432/workshop?sslmode=disable"

    // Initialize the database connection
    var err error
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }

    // Test the connection
    if err = db.Ping(); err != nil {
        return nil, err
    }

    // Set connection pool settings
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)

    // Initialize database schema
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

    // Route to update customer contact information
    app.Put("/customers/me/contact", func(c *fiber.Ctx) error {
        // Mock token validation (replace with actual token validation logic)
        token := c.Get("Authorization")
        if token == "" {
            return c.Status(401).SendString("Unauthorized: Token is required")
        }

        // Parse request body
        var req UpdateContactRequest
        if err := json.Unmarshal(c.Body(), &req); err != nil {
            return c.Status(400).SendString("Invalid request body")
        }

        // Validate phone and email
        if req.Phone == "" || req.Email == "" {
            return c.Status(400).SendString("Phone and email are required")
        }

        // Update contact information in the database
        query := "UPDATE customers SET phone = $1, email = $2 WHERE id = $3"
        customerID := 1 // Mock customer ID (replace with actual logic to extract from token)
        if _, err := db.Exec(query, req.Phone, req.Email, customerID); err != nil {
            return c.Status(500).SendString(fmt.Sprintf("Failed to update contact information: %v", err))
        }

        return c.Status(200).JSON(fiber.Map{
            "message": "Contact information updated successfully",
        })
    })

    // Initialize repositories and handlers for loan feature
    loanRepo := repository.NewPostgresLoanRepository(db)
    loanHandler := handlers.NewLoanHandler(loanRepo)

    // API group
    api := app.Group("/api/v1")
    loans := api.Group("/loans")
    loans.Post("/personal/apply", middleware.JWTMiddleware(), loanHandler.ApplyForPersonalLoan)

    return app
}

func main() {
    // Initialize database connection
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
