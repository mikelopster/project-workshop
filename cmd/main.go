// add code hello world
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"example.com/m/internal/handlers"
	"example.com/m/internal/middleware"
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

	// Initialize tables
	err = initializeTables(db)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to PostgreSQL database!")
	return db, nil
}

// initializeTables creates necessary tables if they don't exist
func initializeTables(db *sql.DB) error {
	// Create customers table
	customersTable := `
	CREATE TABLE IF NOT EXISTS customers (
		id VARCHAR(36) PRIMARY KEY,
		first_name VARCHAR(100) NOT NULL,
		last_name VARCHAR(100) NOT NULL,
		id_card_number VARCHAR(20) NOT NULL UNIQUE,
		phone_number VARCHAR(20) NOT NULL,
		email VARCHAR(100) NOT NULL UNIQUE,
		address TEXT NOT NULL,
		password VARCHAR(100) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);`

	_, err := db.Exec(customersTable)
	return err
}

// setupApp configures and returns a Fiber app instance
func setupApp() *fiber.App {
	app := fiber.New()

	// Create handlers
	customerHandler := handlers.NewCustomerHandler(db)

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

	// Customer routes
	// Get current customer profile - requires authentication
	app.Get("/customers/me", middleware.JWTMiddleware(), customerHandler.GetCurrentCustomerProfile)

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
