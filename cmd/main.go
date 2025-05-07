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

// StaffAuthMiddleware ตรวจสอบว่า request มาจากพนักงานที่มีสิทธิ์หรือไม่
func staffAuthMiddleware(c *fiber.Ctx) error {
	// ในแอพจริงควรจะมีการตรวจสอบ JWT token และสิทธิ์ของพนักงาน
	// ตอนนี้เป็นเพียงตัวอย่าง จึงให้ผ่านไปทุก request
	return c.Next()
}

// getCustomerDetails จัดการคำขอดูรายละเอียดลูกค้าจากพนักงานธนาคาร
func getCustomerDetails(c *fiber.Ctx) error {
	customerID := c.Params("customerId")

	// ในแอพจริงจะต้องดึงข้อมูลลูกค้าจากฐานข้อมูล
	// ตอนนี้ใช้ mock data สำหรับการทดสอบ
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
	// ในแอพจริงควรดึงข้อมูลจากฐานข้อมูล
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

	// ใส่ middleware สำหรับตรวจสอบสิทธิ์พนักงานในทุก route ในกลุ่ม
	staff.Use(staffAuthMiddleware)

	// Route สำหรับดูรายละเอียดลูกค้า
	staff.Get("/customers/:customerId", getCustomerDetails)
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

	// ตั้งค่า routes สำหรับพนักงาน
	setupStaffRoutes(app)

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
