// add code hello world
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
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
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	fmt.Println("Connected to PostgreSQL database!")
	return db, nil
}

// setupApp configures and returns a Fiber app instance
func setupApp() *fiber.App {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	// Route to check database connection
	app.Get("/db-status", func(c *fiber.Ctx) error {
		if db == nil {
			return c.Status(500).SendString("Database not connected")
		}

		err := db.Ping()
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Database error: %v", err))
		}

		return c.SendString("Database connected successfully")
	})

	app.Post("/transactions/transfer/internal", func(c *fiber.Ctx) error {
		type TransferRequest struct {
			FromAccountId string  `json:"fromAccountId"`
			ToAccountId   string  `json:"toAccountId"`
			Amount        float64 `json:"amount"`
			Note          string  `json:"note"`
		}

		var request TransferRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(400).SendString("Invalid request")
		}

		// Mock data
		if request.FromAccountId != "ACC001" || request.ToAccountId != "ACC002" || request.Amount != 500 || request.Note != "Test transfer" {
			return c.Status(400).SendString("Invalid transfer request")
		}

		// Simulate transfer logic
		// ...

		return c.SendString("Transfer successful")
	})

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
